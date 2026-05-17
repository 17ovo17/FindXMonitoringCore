package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

const cmdbHostInstanceMappingContract = "cmdb_host_instance_mapping_contract"

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// CmdbHostTerminal WebSocket 终端端点。有 SSH 凭证时建立 WebSocket SSH 连接。
func CmdbHostTerminal(c *gin.Context) {
	hostID := c.Param("id")
	host, ok := resolveCmdbHostForOps(c, hostID)
	if !ok {
		return
	}

	connInfo := parseHostConnInfo(host.Data)
	if connInfo.IP == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请先配置 SSH 凭证（主机 IP 缺失）"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"host_id": hostID,
		"ip":      connInfo.IP,
		"port":    connInfo.Port,
		"origin":  c.Request.Header.Get("Origin"),
		"action":  "terminal_executed",
	}).Info("cmdb: terminal connection initiated")

	// 升级为 WebSocket
	ws, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logrus.WithError(err).Error("cmdb: websocket upgrade failed")
		return
	}
	defer ws.Close()

	// 建立 SSH 连接
	sshConfig := &ssh.ClientConfig{
		User:            connInfo.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(connInfo.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}
	addr := fmt.Sprintf("%s:%d", connInfo.IP, connInfo.Port)
	sshClient, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("SSH 连接失败: %v\r\n", err)))
		return
	}
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("SSH 会话创建失败: %v\r\n", err)))
		return
	}
	defer session.Close()

	// 请求 PTY
	modes := ssh.TerminalModes{ssh.ECHO: 1, ssh.TTY_OP_ISPEED: 14400, ssh.TTY_OP_OSPEED: 14400}
	if err := session.RequestPty("xterm-256color", 40, 120, modes); err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("PTY 请求失败: %v\r\n", err)))
		return
	}

	stdinPipe, _ := session.StdinPipe()
	stdoutPipe, _ := session.StdoutPipe()
	if err := session.Shell(); err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Shell 启动失败: %v\r\n", err)))
		return
	}

	// SSH stdout -> WebSocket
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stdoutPipe.Read(buf)
			if n > 0 {
				ws.WriteMessage(websocket.TextMessage, buf[:n])
			}
			if err != nil {
				break
			}
		}
	}()

	// WebSocket -> SSH stdin
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			break
		}
		stdinPipe.Write(msg)
	}
}

// CmdbHostUpload 文件上传到目标主机。有 SSH 凭证时执行 SCP 上传。
func CmdbHostUpload(c *gin.Context) {
	hostID := c.Param("id")
	host, ok := resolveCmdbHostForOps(c, hostID)
	if !ok {
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件参数缺失"})
		return
	}
	defer file.Close()

	connInfo := parseHostConnInfo(host.Data)
	if connInfo.IP == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请先配置 SSH 凭证（主机 IP 缺失）"})
		return
	}

	destPath := c.DefaultPostForm("dest_path", "/tmp/"+header.Filename)

	logrus.WithFields(logrus.Fields{
		"host_id":   hostID,
		"ip":        connInfo.IP,
		"filename":  header.Filename,
		"dest_path": destPath,
		"action":    "file_upload_executed",
	}).Info("cmdb: file upload initiated")

	// 建立 SSH 连接
	sshConfig := &ssh.ClientConfig{
		User:            connInfo.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(connInfo.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}
	addr := fmt.Sprintf("%s:%d", connInfo.IP, connInfo.Port)
	sshClient, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": fmt.Sprintf("SSH 连接失败: %v", err)})
		return
	}
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": fmt.Sprintf("SSH 会话创建失败: %v", err)})
		return
	}
	defer session.Close()

	// 通过 SCP 协议上传文件
	fileContent, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取上传文件失败"})
		return
	}

	stdinPipe, err := session.StdinPipe()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "创建 stdin pipe 失败"})
		return
	}

	go func() {
		defer stdinPipe.Close()
		fmt.Fprintf(stdinPipe, "C0644 %d %s\n", len(fileContent), header.Filename)
		stdinPipe.Write(fileContent)
		fmt.Fprint(stdinPipe, "\x00")
	}()

	output, err := session.CombinedOutput("scp -t " + destPath)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":  fmt.Sprintf("SCP 上传失败: %v", err),
			"output": sanitizeOutput(string(output)),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "uploaded",
		"host_id":   hostID,
		"ip":        connInfo.IP,
		"filename":  header.Filename,
		"dest_path": destPath,
		"size":      len(fileContent),
	})
}

type execRequest struct {
	Command string `json:"command" binding:"required"`
	Timeout int    `json:"timeout"`
	Sudo    bool   `json:"sudo"`
}

// CmdbHostExec 远程命令执行。有 SSH 凭证时执行命令，返回真实 stdout/stderr/exit_code。
func CmdbHostExec(c *gin.Context) {
	hostID := c.Param("id")
	host, ok := resolveCmdbHostForOps(c, hostID)
	if !ok {
		return
	}

	var req execRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	if req.Timeout <= 0 {
		req.Timeout = 30
	}
	if req.Timeout > 300 {
		req.Timeout = 300
	}

	connInfo := parseHostConnInfo(host.Data)
	if connInfo.IP == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请先配置 SSH 凭证（主机 IP 缺失）"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"host_id":        hostID,
		"ip":             connInfo.IP,
		"command_length": len(req.Command),
		"command_digest": commandDigest(req.Command),
		"sudo":           req.Sudo,
		"timeout":        req.Timeout,
		"action":         "command_exec_executed",
	}).Info("cmdb: command execution initiated")

	// 建立 SSH 连接并执行命令
	sshConfig := &ssh.ClientConfig{
		User:            connInfo.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(connInfo.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Duration(req.Timeout) * time.Second,
	}
	addr := fmt.Sprintf("%s:%d", connInfo.IP, connInfo.Port)
	sshClient, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": fmt.Sprintf("SSH 连接失败: %v", err)})
		return
	}
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": fmt.Sprintf("SSH 会话创建失败: %v", err)})
		return
	}
	defer session.Close()

	command := req.Command
	if req.Sudo {
		command = "sudo " + command
	}

	output, err := session.CombinedOutput(command)
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": fmt.Sprintf("命令执行失败: %v", err)})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"host_id":   hostID,
		"ip":        connInfo.IP,
		"command":   sanitizeCommand(req.Command),
		"stdout":    sanitizeOutput(string(output)),
		"stderr":    "",
		"exit_code": exitCode,
	})
}

func resolveCmdbHostForOps(c *gin.Context, hostID string) (*model.CmdbInstance, bool) {
	host, ok := store.GetCmdbInstance(hostID)
	if ok {
		return host, true
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "主机不存在"})
	return nil, false
}

type hostConnInfo struct {
	IP       string
	Port     int
	Username string
	Password string
	Hostname string
}

func parseHostConnInfo(dataJSON string) hostConnInfo {
	info := hostConnInfo{Port: 22, Username: "root"}
	if dataJSON == "" {
		return info
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(dataJSON), &data); err != nil {
		return info
	}
	for _, key := range []string{"ip_address", "OS001", "ssh_ip", "sshIp"} {
		if ip, ok := data[key].(string); ok && ip != "" {
			info.IP = ip
			break
		}
	}
	if port := intFromAny(data["ssh_port"], 22); port > 0 {
		info.Port = port
	}
	if port := intFromAny(data["sshPort"], 22); port > 0 {
		info.Port = port
	}
	for _, key := range []string{"ssh_user", "sshName", "ssh_name"} {
		if user, ok := data[key].(string); ok && user != "" {
			info.Username = user
			break
		}
	}
	if name, ok := data["name"].(string); ok && name != "" {
		info.Hostname = name
	}
	for _, key := range []string{"ssh_password", "sshPassword", "password"} {
		if pwd, ok := data[key].(string); ok && pwd != "" {
			info.Password = pwd
			break
		}
	}
	return info
}

func sanitizeCommand(cmd string) string {
	if len(cmd) > 500 {
		return cmd[:500] + "..."
	}
	return cmd
}

func commandDigest(cmd string) string {
	sum := sha256.Sum256([]byte(cmd))
	return hex.EncodeToString(sum[:])
}

func sanitizeOutput(output string) string {
	replacer := strings.NewReplacer(
		"password=", "password=<REDACTED>",
		"passwd=", "passwd=<REDACTED>",
		"secret=", "secret=<REDACTED>",
		"token=", "token=<REDACTED>",
	)
	return replacer.Replace(output)
}
