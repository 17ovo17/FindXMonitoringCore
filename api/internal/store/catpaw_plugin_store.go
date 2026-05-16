package store

import (
	"sync"
	"time"

	"ai-workbench-api/internal/model"
)

var (
	catpawPluginsMu      sync.RWMutex
	catpawPlugins        = map[string]*model.CatpawPlugin{}
	catpawResultsMu      sync.RWMutex
	catpawResults        = map[string][]*model.CatpawInspectionResult{} // key: IP
	catpawResultsTTL     = 24 * time.Hour
)

func init() {
	initCatpawPlugins()
}

// initCatpawPlugins 预填充 35 个 catpaw 插件模板
func initCatpawPlugins() {
	plugins := []model.CatpawPlugin{
		// --- system 类 ---
		{ID: "cpu", Name: "CPU", Category: "system", Description: "CPU 使用率和 Load Average 监控，支持阈值告警和 AI 诊断", Enabled: true},
		{ID: "mem", Name: "Memory", Category: "system", Description: "物理内存和 Swap 使用率监控，支持阈值告警", Enabled: true},
		{ID: "disk", Name: "Disk", Category: "system", Description: "磁盘空间和 inode 使用率监控，支持挂载点过滤", Enabled: true},
		{ID: "diskio", Name: "DiskIO", Category: "system", Description: "磁盘 IO 读写速率和 IOPS 监控", Enabled: true},
		{ID: "net", Name: "Net", Category: "system", Description: "网络端口连通性检测，支持 TCP/UDP 协议", Enabled: true},
		{ID: "netif", Name: "NetInterface", Category: "system", Description: "网卡流量、错误包和丢包率监控", Enabled: true},
		{ID: "mount", Name: "Mount", Category: "system", Description: "文件系统挂载状态检查", Enabled: true},
		{ID: "uptime", Name: "Uptime", Category: "system", Description: "系统运行时间监控，检测意外重启", Enabled: true},
		{ID: "procnum", Name: "ProcessNum", Category: "system", Description: "进程数量监控，检测进程是否存活", Enabled: true},
		{ID: "procfd", Name: "ProcessFD", Category: "system", Description: "进程文件描述符使用量监控", Enabled: true},
		{ID: "filefd", Name: "FileFD", Category: "system", Description: "系统级文件描述符使用量监控", Enabled: true},
		{ID: "filecheck", Name: "FileCheck", Category: "system", Description: "文件存在性、大小、修改时间检查", Enabled: true},
		{ID: "hostident", Name: "HostIdent", Category: "system", Description: "主机标识信息采集（hostname、IP、OS 版本）", Enabled: true},
		{ID: "zombie", Name: "Zombie", Category: "system", Description: "僵尸进程检测", Enabled: true},
		{ID: "conntrack", Name: "Conntrack", Category: "system", Description: "连接跟踪表使用量监控", Enabled: true},
		{ID: "sockstat", Name: "SockStat", Category: "system", Description: "Socket 统计信息监控", Enabled: true},
		{ID: "tcpstate", Name: "TCPState", Category: "system", Description: "TCP 连接状态分布监控", Enabled: true},
		{ID: "sysctl", Name: "Sysctl", Category: "system", Description: "内核参数检查", Enabled: true},
		{ID: "systemd", Name: "Systemd", Category: "system", Description: "Systemd 服务状态监控", Enabled: true},
		{ID: "ntp", Name: "NTP", Category: "system", Description: "NTP 时间同步偏差监控", Enabled: true},
		{ID: "neigh", Name: "Neighbor", Category: "system", Description: "ARP/NDP 邻居表监控", Enabled: true},
		{ID: "exec", Name: "Exec", Category: "system", Description: "自定义命令执行采集", Enabled: false},
		{ID: "journaltail", Name: "JournalTail", Category: "system", Description: "Journald 日志尾部采集", Enabled: false},
		{ID: "logfile", Name: "LogFile", Category: "system", Description: "日志文件关键字匹配监控", Enabled: false},
		{ID: "scriptfilter", Name: "ScriptFilter", Category: "system", Description: "自定义脚本过滤器", Enabled: false},
		// --- service 类 ---
		{ID: "redis", Name: "Redis", Category: "service", Description: "Redis 连通性、内存、连接数、QPS、复制状态和集群健康监控", Enabled: false},
		{ID: "redis_sentinel", Name: "RedisSentinel", Category: "service", Description: "Redis Sentinel 哨兵状态监控", Enabled: false},
		{ID: "docker", Name: "Docker", Category: "service", Description: "Docker 容器运行状态、健康检查和资源使用监控", Enabled: false},
		{ID: "etcd", Name: "Etcd", Category: "service", Description: "Etcd 集群健康、Leader 状态和延迟监控", Enabled: false},
		// --- network 类 ---
		{ID: "ping", Name: "Ping", Category: "network", Description: "ICMP Ping 连通性和延迟监控", Enabled: false},
		{ID: "dns", Name: "DNS", Category: "network", Description: "DNS 解析可用性和正确性检查", Enabled: false},
		{ID: "http", Name: "HTTP", Category: "network", Description: "HTTP/HTTPS 端点可用性、状态码和响应时间监控", Enabled: false},
		// --- security 类 ---
		{ID: "secmod", Name: "SecMod", Category: "security", Description: "SELinux/AppArmor 安全模块状态基线检查", Enabled: false},
		{ID: "cert", Name: "Certificate", Category: "security", Description: "TLS 证书过期时间监控，支持远程和本地文件", Enabled: false},
	}

	defaultConfigs := map[string]string{
		"cpu": `[[instances]]
[instances.cpu_usage]
warn_ge     = 80.0
critical_ge = 90.0

[instances.load_average]
warn_ge     = 3.0
critical_ge = 5.0

[instances.alerting]
for_duration = "120s"
repeat_interval = "5m"
repeat_number = 3

[instances.diagnose]
enabled = true
`,
		"mem": `[[instances]]
[instances.memory_usage]
warn_ge     = 85.0
critical_ge = 90.0

[instances.swap_usage]
warn_ge     = 10.0
critical_ge = 30.0

[instances.alerting]
for_duration = "60s"
repeat_interval = "5m"
repeat_number = 3

[instances.diagnose]
enabled = true
`,
		"disk": `[[instances]]
ignore_mount_points = ["/dev*", "/run*", "/sys*", "/proc*", "/snap*"]
ignore_fs_types = ["tmpfs", "devtmpfs", "squashfs", "overlay"]

[instances.space_usage]
warn_ge     = 90.0
critical_ge = 99.0

[instances.inode_usage]
warn_ge     = 90.0
critical_ge = 99.0

[instances.alerting]
for_duration = 0
repeat_interval = "5m"
repeat_number = 3

[instances.diagnose]
enabled = true
`,
		"diskio": `[[instances]]
# interval = "30s"

[instances.alerting]
for_duration = "60s"
repeat_interval = "5m"
repeat_number = 3
`,
		"net": `[[partials]]
id = "default"
# concurrency = 10
# timeout = "1s"
# protocol = "tcp"

[partials.connectivity]
severity = "Critical"

[[instances]]
targets = []
partial = "default"

[instances.alerting]
for_duration = 0
repeat_interval = "5m"
repeat_number = 3
`,
		"netif": `[[instances]]
# interval = "30s"

[instances.alerting]
for_duration = "60s"
repeat_interval = "5m"
repeat_number = 3
`,
		"ping": `[[partials]]
id = "default"
# count = 3
# ping_interval = "200ms"
# timeout = "3s"

[partials.connectivity]
severity = "Critical"

[partials.packet_loss]
warn_ge = 10.0
critical_ge = 50.0

[[instances]]
targets = []
partial = "default"

[instances.alerting]
for_duration = "60s"
repeat_interval = "5m"
repeat_number = 3
`,
		"dns": `[[instances]]
targets = ["www.baidu.com"]
# servers = ["8.8.8.8"]
# timeout = "5s"

[instances.alerting]
for_duration = 0
repeat_interval = "5m"
repeat_number = 3
`,
		"http": `[[partials]]
id = "default"
# concurrency = 10
# timeout = "10s"

[partials.connectivity]
severity = "Critical"

[[instances]]
targets = []
partial = "default"

[instances.alerting]
for_duration = "60s"
repeat_interval = "5m"
repeat_number = 3
`,
		"redis": `[[partials]]
id = "default"
# timeout = "3s"
# read_timeout = "2s"
# password = ""

[partials.connectivity]
severity = "Critical"

[[instances]]
targets = []
partial = "default"

[instances.alerting]
for_duration = 0
repeat_interval = "5m"
repeat_number = 3

[instances.diagnose]
enabled = true
`,
		"redis_sentinel": `[[instances]]
targets = []
# password = ""

[instances.alerting]
for_duration = 0
repeat_interval = "5m"
repeat_number = 3
`,
		"docker": `[[instances]]
# socket = "/var/run/docker.sock"
# targets = ["my-app*"]

[instances.alerting]
for_duration = "60s"
repeat_interval = "5m"
repeat_number = 3
`,
		"etcd": `[[instances]]
targets = []
# tls_ca = ""
# tls_cert = ""
# tls_key = ""

[instances.alerting]
for_duration = 0
repeat_interval = "5m"
repeat_number = 3
`,
		"secmod": `[[instances]]
interval = "60s"

[instances.enforce_mode]
# expect = "enforcing"
# severity = "Warning"

[instances.apparmor_enabled]
# expect = "yes"
# severity = "Warning"
`,
		"cert": `[[instances]]
remote_targets = []
# file_targets = ["/etc/ssl/certs/*.pem"]

[instances.expiry]
warn_days = 30
critical_days = 7

[instances.alerting]
for_duration = 0
repeat_interval = "24h"
repeat_number = 1
`,
		"mount": `[[instances]]
# expected_mounts = ["/data", "/backup"]

[instances.alerting]
for_duration = 0
repeat_interval = "5m"
repeat_number = 3
`,
		"uptime": `[[instances]]
# interval = "60s"

[instances.reboot_detect]
severity = "Warning"
`,
		"procnum": `[[instances]]
# targets = [{name = "nginx", search = "nginx: master"}]

[instances.alerting]
for_duration = "60s"
repeat_interval = "5m"
repeat_number = 3
`,
		"procfd": `[[instances]]
# targets = [{name = "nginx", search = "nginx"}]

[instances.fd_usage]
warn_ge = 80.0
critical_ge = 95.0

[instances.alerting]
for_duration = "60s"
repeat_interval = "5m"
repeat_number = 3
`,
		"filefd": `[[instances]]
[instances.fd_usage]
warn_ge = 80.0
critical_ge = 95.0

[instances.alerting]
for_duration = "60s"
repeat_interval = "5m"
repeat_number = 3
`,
		"filecheck": `[[instances]]
# paths = ["/tmp/healthcheck"]

[instances.alerting]
for_duration = 0
repeat_interval = "5m"
repeat_number = 3
`,
		"hostident": `[[instances]]
interval = "300s"
`,
		"zombie": `[[instances]]
[instances.zombie_count]
warn_ge = 5
critical_ge = 20

[instances.alerting]
for_duration = "120s"
repeat_interval = "5m"
repeat_number = 3
`,
		"conntrack": `[[instances]]
[instances.usage]
warn_ge = 80.0
critical_ge = 95.0

[instances.alerting]
for_duration = "60s"
repeat_interval = "5m"
repeat_number = 3
`,
		"sockstat": `[[instances]]
# interval = "30s"

[instances.alerting]
for_duration = "60s"
repeat_interval = "5m"
repeat_number = 3
`,
		"tcpstate": `[[instances]]
# interval = "30s"

[instances.time_wait]
warn_ge = 5000
critical_ge = 20000

[instances.alerting]
for_duration = "60s"
repeat_interval = "5m"
repeat_number = 3
`,
		"sysctl": `[[instances]]
# checks = [{key = "net.ipv4.ip_forward", expect = "1"}]

[instances.alerting]
for_duration = 0
repeat_interval = "24h"
repeat_number = 1
`,
		"systemd": `[[instances]]
# targets = ["nginx", "docker", "sshd"]

[instances.alerting]
for_duration = "60s"
repeat_interval = "5m"
repeat_number = 3
`,
		"ntp": `[[instances]]
[instances.offset]
warn_ge_ms = 100.0
critical_ge_ms = 500.0

[instances.alerting]
for_duration = "120s"
repeat_interval = "5m"
repeat_number = 3
`,
		"neigh": `[[instances]]
# interval = "60s"

[instances.alerting]
for_duration = "60s"
repeat_interval = "5m"
repeat_number = 3
`,
		"exec": `[[instances]]
# commands = [{name = "check_app", command = "/opt/scripts/check.sh", timeout = "10s"}]

[instances.alerting]
for_duration = 0
repeat_interval = "5m"
repeat_number = 3
`,
		"journaltail": `[[instances]]
# units = ["nginx", "docker"]
# match = "error|fatal|panic"
# since = "5m"
`,
		"logfile": `[[instances]]
# paths = ["/var/log/app/*.log"]
# match = "ERROR|FATAL"
# interval = "30s"
`,
		"scriptfilter": `[[instances]]
# script = "/opt/catpaw/scripts/custom.sh"
# interval = "60s"
# timeout = "30s"
`,
	}

	for i := range plugins {
		if cfg, ok := defaultConfigs[plugins[i].ID]; ok {
			plugins[i].DefaultConfig = cfg
		}
		catpawPlugins[plugins[i].ID] = &plugins[i]
	}
}

// ListCatpawPlugins 返回所有插件模板
func ListCatpawPlugins() []*model.CatpawPlugin {
	catpawPluginsMu.RLock()
	defer catpawPluginsMu.RUnlock()
	out := make([]*model.CatpawPlugin, 0, len(catpawPlugins))
	for _, p := range catpawPlugins {
		cp := *p
		out = append(out, &cp)
	}
	return out
}

// ListCatpawPluginsByCategory 按分类返回插件
func ListCatpawPluginsByCategory(category string) []*model.CatpawPlugin {
	catpawPluginsMu.RLock()
	defer catpawPluginsMu.RUnlock()
	out := make([]*model.CatpawPlugin, 0)
	for _, p := range catpawPlugins {
		if p.Category == category {
			cp := *p
			out = append(out, &cp)
		}
	}
	return out
}

// GetCatpawPlugin 获取单个插件模板
func GetCatpawPlugin(id string) (*model.CatpawPlugin, bool) {
	catpawPluginsMu.RLock()
	defer catpawPluginsMu.RUnlock()
	p, ok := catpawPlugins[id]
	if !ok {
		return nil, false
	}
	cp := *p
	return &cp, true
}

// AddCatpawResult 存储巡检结果
func AddCatpawResult(result *model.CatpawInspectionResult) {
	catpawResultsMu.Lock()
	defer catpawResultsMu.Unlock()
	result.ExpiresAt = time.Now().Add(catpawResultsTTL)
	catpawResults[result.IP] = append(catpawResults[result.IP], result)
}

// ListCatpawResults 获取指定 IP 的巡检结果（自动清理过期数据）
func ListCatpawResults(ip string) []*model.CatpawInspectionResult {
	catpawResultsMu.Lock()
	defer catpawResultsMu.Unlock()
	now := time.Now()
	results := catpawResults[ip]
	valid := make([]*model.CatpawInspectionResult, 0, len(results))
	for _, r := range results {
		if r.ExpiresAt.After(now) {
			valid = append(valid, r)
		}
	}
	catpawResults[ip] = valid
	return valid
}

// ListAllCatpawResults 获取所有巡检结果
func ListAllCatpawResults() []*model.CatpawInspectionResult {
	catpawResultsMu.Lock()
	defer catpawResultsMu.Unlock()
	now := time.Now()
	var all []*model.CatpawInspectionResult
	for ip, results := range catpawResults {
		valid := make([]*model.CatpawInspectionResult, 0, len(results))
		for _, r := range results {
			if r.ExpiresAt.After(now) {
				valid = append(valid, r)
			}
		}
		catpawResults[ip] = valid
		all = append(all, valid...)
	}
	return all
}
