package handler

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

func summarizeCatpawReport(report string) string {
	if strings.TrimSpace(report) == "" {
		return "# Catpaw Inspection Summary\n\n- No valid inspection content was received."
	}
	if !strings.Contains(report, "Catpaw Windows") {
		if len([]rune(report)) > 4000 {
			return string([]rune(report)[:4000]) + "\n\n> Raw report was truncated here. The full content is stored in raw_report."
		}
		return report
	}
	return summarizeWindowsCatpawReport(report)
}

// windowsInspectionData holds extracted metrics from a Windows catpaw report
type windowsInspectionData struct {
	cpu          map[string]any
	mem          map[string]any
	disks        []map[string]any
	diskIO       []map[string]any
	netInfo      map[string]any
	tcpInfo      map[string]any
	processInfo  map[string]any
	serviceInfo  map[string]any
	events       []map[string]any
	firewallInfo map[string]any
	updateInfo   map[string]any
	plugins      []string
}

// extractWindowsData parses plugin payloads from the raw report
func extractWindowsData(report string) windowsInspectionData {
	return windowsInspectionData{
		cpu:          asMap(pluginPayload(report, "windows.cpu")),
		mem:          asMap(pluginPayload(report, "windows.mem")),
		disks:        asList(pluginPayload(report, "windows.disk")),
		diskIO:       asList(pluginPayload(report, "windows.diskio")),
		netInfo:      asMap(pluginPayload(report, "windows.net")),
		tcpInfo:      asMap(pluginPayload(report, "windows.tcpstate")),
		processInfo:  asMap(pluginPayload(report, "windows.process")),
		serviceInfo:  asMap(pluginPayload(report, "windows.service")),
		events:       normalizeWindowsEvents(asList(pluginPayload(report, "windows.eventlog"))),
		firewallInfo: asMap(pluginPayload(report, "windows.firewall")),
		updateInfo:   asMap(pluginPayload(report, "windows.update")),
		plugins:      discoveredPluginNames(report),
	}
}

// assessWindowsRisks evaluates key metrics and returns risk descriptions
func assessWindowsRisks(d windowsInspectionData) []string {
	cpuUsed := numberField(d.cpu, "CPUUsedPercent")
	cpuQueue := numberField(d.cpu, "ProcessorQueueLength")
	memUsed := numberField(d.mem, "MemUsedPercent")
	diskMax := maxNumberField(d.disks, "UsedPercent")
	autoStopped := normalizeServiceRows(asList(d.serviceInfo["AutoServicesStopped"]))

	risks := make([]string, 0, 6)
	if cpuUsed >= 80 {
		risks = append(risks, fmt.Sprintf("CPU usage %.2f%% is high. Check Top CPU processes.", cpuUsed))
	}
	if cpuQueue >= 4 {
		risks = append(risks, fmt.Sprintf("CPU queue %.2f is high. There may be runnable queue pressure.", cpuQueue))
	}
	if memUsed >= 85 {
		risks = append(risks, fmt.Sprintf("Memory usage %.2f%% is high. Check Top memory processes and page file usage.", memUsed))
	}
	if diskMax >= 85 {
		risks = append(risks, fmt.Sprintf("Maximum disk usage %.2f%% is high. Check the affected drive free space.", diskMax))
	}
	if len(autoStopped) > 0 {
		risks = append(risks, fmt.Sprintf("Found %d auto-start services that are not running. Verify whether they affect business services.", len(autoStopped)))
	}
	if len(d.events) > 0 {
		risks = append(risks, fmt.Sprintf("Found %d critical/error events in the last 24 hours.", len(d.events)))
	}
	if len(risks) == 0 {
		risks = append(risks, "No obvious high-risk pressure was found. Continue monitoring trends and business port availability.")
	}
	return risks
}

func summarizeWindowsCatpawReport(report string) string {
	d := extractWindowsData(report)
	risks := assessWindowsRisks(d)
	return renderWindowsReport(d, risks)
}

// renderWindowsReport builds the markdown summary from extracted data and risks
func renderWindowsReport(d windowsInspectionData, risks []string) string {
	cpuUsed := numberField(d.cpu, "CPUUsedPercent")
	cpuQueue := numberField(d.cpu, "ProcessorQueueLength")
	memUsed := numberField(d.mem, "MemUsedPercent")
	diskMax := maxNumberField(d.disks, "UsedPercent")
	autoStopped := normalizeServiceRows(asList(d.serviceInfo["AutoServicesStopped"]))
	timeWait := stateCount(d.tcpInfo, "TimeWait")
	established := stateCount(d.tcpInfo, "Established")
	listen := len(asList(d.tcpInfo["ListeningTCP"]))

	var sb strings.Builder
	sb.WriteString("# Catpaw Windows Inspection Summary\n\n")
	sb.WriteString("- Data source: Catpaw Windows native inspection plus platform-side structured denoising.\n")
	sb.WriteString("- Display policy: key conclusions, Top-N details, and risks are shown by default; full raw plugin JSON is kept in raw_report for folding or download.\n")
	sb.WriteString(fmt.Sprintf("- Collected plugins: %s\n\n", strings.Join(d.plugins, ", ")))

	sb.WriteString("## Overall Conclusion\n")
	for _, risk := range risks {
		sb.WriteString("- " + risk + "\n")
	}
	sb.WriteString("\n")

	sb.WriteString("## Base Resources\n")
	sb.WriteString(fmt.Sprintf("- CPU: usage %s, queue %s, processor groups %d.\n", percentText(cpuUsed), numberText(cpuQueue), len(asList(d.cpu["Processors"]))))
	sb.WriteString(fmt.Sprintf("- Memory: usage %s, total %s MB, free %s MB, free virtual %s MB.\n", percentText(memUsed), anyText(d.mem["TotalMemoryMB"]), anyText(d.mem["FreeMemoryMB"]), anyText(d.mem["FreeVirtualMemoryMB"])))
	sb.WriteString(fmt.Sprintf("- Disk: %d local disks, highest usage %s.\n", len(d.disks), percentText(diskMax)))
	sb.WriteString(fmt.Sprintf("- TCP: Established=%d, TIME_WAIT=%d, listening port samples=%d.\n\n", established, timeWait, listen))

	renderWindowsDetails(&sb, d, autoStopped)
	return sb.String()
}

// renderWindowsDetails appends detailed sections to the report
func renderWindowsDetails(sb *strings.Builder, d windowsInspectionData, autoStopped []map[string]any) {
	sb.WriteString("## CPU / Memory Top Details\n")
	sb.WriteString(markdownTable("Top CPU processes", asList(d.cpu["TopCPUProcesses"]), []string{"ProcessName", "Id", "CPU", "ThreadCount"}, 8))
	sb.WriteString(markdownTable("Top memory processes", asList(d.mem["TopMemoryProcesses"]), []string{"ProcessName", "Id", "WorkingSetMB", "PagedMemoryMB"}, 8))
	sb.WriteString(markdownTable("Top handle processes", asList(d.processInfo["TopHandles"]), []string{"ProcessName", "Id", "Handles", "ThreadCount"}, 6))

	sb.WriteString("## Disk / IO\n")
	sb.WriteString(markdownTable("Logical disks", d.disks, []string{"DeviceID", "VolumeName", "FileSystem", "SizeGB", "FreeGB", "UsedPercent"}, 8))
	sb.WriteString(markdownTable("Physical disk IO samples", d.diskIO, []string{"InstanceName", "Path", "Value"}, 10))

	sb.WriteString("## Network / Ports\n")
	sb.WriteString(markdownTable("Network adapters", asList(d.netInfo["Adapters"]), []string{"Name", "Status", "LinkSpeed", "MacAddress"}, 8))
	sb.WriteString(markdownTable("Network traffic and errors", asList(d.netInfo["Statistics"]), []string{"Name", "ReceivedBytes", "SentBytes", "ReceivedDiscardedPackets", "OutboundDiscardedPackets", "ReceivedPacketErrors", "OutboundPacketErrors"}, 8))
	sb.WriteString(markdownTable("TCP state summary", asList(d.tcpInfo["TCPStateSummary"]), []string{"Name", "Count"}, 12))
	sb.WriteString(markdownTable("Listening TCP port samples", asList(d.tcpInfo["ListeningTCP"]), []string{"LocalAddress", "LocalPort", "OwningProcess"}, 12))

	sb.WriteString("## Services / Events / Security\n")
	sb.WriteString(markdownTable("Critical service status", normalizeServiceRows(asList(d.serviceInfo["CriticalServices"])), []string{"Name", "DisplayName", "Status", "StartType"}, 10))
	sb.WriteString(markdownTable("Auto-start services not running", autoStopped, []string{"Name", "DisplayName", "State", "StartMode"}, 10))
	sb.WriteString(markdownTable("Recent critical/error events", d.events, []string{"TimeCreated", "LogName", "ProviderName", "Id", "LevelDisplayName", "Message"}, 8))
	sb.WriteString(markdownTable("Firewall profiles", asList(d.firewallInfo["FirewallProfiles"]), []string{"Name", "Enabled", "DefaultInboundAction", "DefaultOutboundAction"}, 6))
	sb.WriteString(markdownTable("Remote management rules", asList(d.firewallInfo["RemoteManagementRules"]), []string{"DisplayName", "Enabled", "Direction", "Action", "Profile"}, 10))

	sb.WriteString("## OS / Patches\n")
	sb.WriteString(markdownTable("Operating system", asList(d.updateInfo["OS"]), []string{"Caption", "Version", "BuildNumber", "LastBootUpTime"}, 2))
	sb.WriteString(markdownTable("Recent hotfixes", asList(d.updateInfo["RecentHotfix"]), []string{"HotFixID", "Description", "InstalledOn"}, 10))

	sb.WriteString("## Recommendations\n")
	sb.WriteString("- Keep Top-N and key fields in the summary; expand or download raw_report for deep investigation.\n")
	sb.WriteString("- Windows license, activation, storage optimization, or service-control events must be judged by Provider, Id, error code, and business impact.\n")
	sb.WriteString("- When Prometheus/Categraf metrics are available, prefer trend-based pressure analysis over one-shot event interpretation.\n")
}

func pluginPayload(report, plugin string) any {
	pluginIndex := strings.Index(report, plugin)
	if pluginIndex < 0 {
		return nil
	}
	jsonStart := strings.Index(report[pluginIndex:], "```json")
	if jsonStart < 0 {
		return nil
	}
	payloadStart := pluginIndex + jsonStart + len("```json")
	jsonEnd := strings.Index(report[payloadStart:], "```")
	if jsonEnd < 0 {
		return nil
	}
	jsonText := report[payloadStart : payloadStart+jsonEnd]
	var payload any
	decoder := json.NewDecoder(strings.NewReader(strings.TrimSpace(jsonText)))
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
		return nil
	}
	return payload
}

func discoveredPluginNames(report string) []string {
	re := regexp.MustCompile(`windows\.[a-z0-9]+`)
	matches := re.FindAllString(report, -1)
	seen := map[string]bool{}
	plugins := make([]string, 0, len(matches))
	for _, match := range matches {
		if !seen[match] {
			seen[match] = true
			plugins = append(plugins, match)
		}
	}
	sort.Strings(plugins)
	return plugins
}

func markdownTable(title string, rows []map[string]any, columns []string, limit int) string {
	if len(rows) == 0 {
		return fmt.Sprintf("### %s\n\n- Not collected or no result.\n\n", title)
	}
	if limit > 0 && len(rows) > limit {
		rows = rows[:limit]
	}
	var sb strings.Builder
	sb.WriteString("### " + title + "\n\n")
	sb.WriteString("| " + strings.Join(columns, " | ") + " |\n")
	separators := make([]string, len(columns))
	for i := range separators {
		separators[i] = "---"
	}
	sb.WriteString("| " + strings.Join(separators, " | ") + " |\n")
	for _, row := range rows {
		values := make([]string, 0, len(columns))
		for _, column := range columns {
			values = append(values, tableText(row[column]))
		}
		sb.WriteString("| " + strings.Join(values, " | ") + " |\n")
	}
	sb.WriteString("\n")
	return sb.String()
}

func tableText(value any) string {
	text := strings.ReplaceAll(anyText(value), "|", "\\|")
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\n", " ")
	text = normalizeWindowsDate(text)
	if isMojibake(text) {
		text = "Unreadable encoded text; raw value is kept in raw_report. Use Provider/Id/error-code for investigation."
	}
	if len([]rune(text)) > 160 {
		return string([]rune(text)[:160]) + "..."
	}
	if strings.TrimSpace(text) == "" {
		return "-"
	}
	return text
}

func normalizeServiceRows(rows []map[string]any) []map[string]any {
	for _, row := range rows {
		if name := anyText(row["Name"]); name == "edgeupdate" && isMojibake(anyText(row["DisplayName"])) {
			row["DisplayName"] = "Microsoft Edge Update Service (edgeupdate)"
		}
	}
	return rows
}

func normalizeWindowsEvents(rows []map[string]any) []map[string]any {
	for _, row := range rows {
		row["TimeCreated"] = normalizeWindowsDate(anyText(row["TimeCreated"]))
		if isMojibake(anyText(row["LevelDisplayName"])) {
			row["LevelDisplayName"] = levelNameFromID(row)
		}
		if message := normalizeWindowsEventMessage(row); message != "" {
			row["Message"] = message
		}
	}
	return rows
}

func normalizeWindowsDate(text string) string {
	re := regexp.MustCompile(`/Date\((\d+)\)/`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}
		ms, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return match
		}
		return time.UnixMilli(ms).Format("2006-01-02 15:04:05")
	})
}

func isMojibake(text string) bool {
	if strings.TrimSpace(text) == "" {
		return false
	}
	if strings.Contains(text, "\ufffd") {
		return true
	}
	bad := strings.Count(text, "?")
	total := len([]rune(text))
	if total == 0 {
		return false
	}
	return strings.Contains(text, "????") || (bad >= 3 && float64(bad)/float64(total) > 0.3)
}

func levelNameFromID(row map[string]any) string {
	text := strings.ToLower(anyText(row["LevelDisplayName"]))
	if strings.Contains(text, "critical") {
		return "Critical"
	}
	return "Error"
}

func normalizeWindowsEventMessage(row map[string]any) string {
	provider := anyText(row["ProviderName"])
	id := int(toFloat(row["Id"]))
	message := normalizeWindowsDate(anyText(row["Message"]))
	if !isMojibake(message) && strings.TrimSpace(message) != "" {
		return message
	}
	switch provider {
	case "Microsoft-Windows-Security-SPP":
		switch id {
		case 1014:
			return "Security-SPP license acquisition failed. Commonly related to Windows activation/licensing; check hr=0xC004C060 and business impact."
		case 8200:
			return "Security-SPP detailed license acquisition failure event. Check activation state, licensing service, and recent system changes."
		case 8198:
			return "Security-SPP activation-related action failed. Check slui.exe, licensing service, and error code."
		}
	case "Microsoft-Windows-Defrag":
		return "Storage optimization event. The volume or virtual disk may not support this operation; correlate with disk health and IO metrics."
	}
	return "Event message encoding is unreadable. The original value is kept in raw_report; investigate by Provider, Id, and error code."
}
