package handler

var metricDisplayNames = map[string]string{
	"cpu_usage_active":               "CPU 使用率",
	"cpu_usage_idle":                 "CPU 空闲率",
	"cpu_usage_system":               "CPU 系统态",
	"cpu_usage_user":                 "CPU 用户态",
	"cpu_usage_iowait":               "CPU IO等待",
	"cpu_usage_softirq":              "CPU 软中断",
	"cpu_usage_steal":                "CPU 窃取",
	"mem_used_percent":               "内存使用率",
	"mem_available_percent":          "内存可用率",
	"mem_used":                       "已用内存",
	"mem_available":                  "可用内存",
	"mem_cached":                     "内存缓存",
	"swap_used_percent":              "Swap 使用率",
	"disk_used_percent":              "磁盘使用率",
	"disk_total":                     "磁盘总量",
	"diskio_io_util":                 "磁盘 IO 繁忙度",
	"diskio_io_await":                "磁盘 IO 等待时间",
	"diskio_read_bytes":              "磁盘读取速率",
	"diskio_write_bytes":             "磁盘写入速率",
	"net_bits_recv":                  "网络入流量",
	"net_bits_sent":                  "网络出流量",
	"net_drop_in":                    "网络入丢包",
	"net_drop_out":                   "网络出丢包",
	"net_err_in":                     "网络入错误",
	"net_err_out":                    "网络出错误",
	"netstat_tcp_inuse":              "TCP 活跃连接数",
	"netstat_tcp_tw":                 "TCP TIME_WAIT 数",
	"netstat_sockets_used":           "Socket 使用数",
	"system_load1":                   "系统负载(1分钟)",
	"system_load5":                   "系统负载(5分钟)",
	"system_load15":                  "系统负载(15分钟)",
	"system_n_cpus":                  "CPU 核心数",
	"kernel_context_switches":        "内核上下文切换",
	"kernel_vmstat_oom_kill":         "OOM 终止次数",
	"processes":                      "进程总数",
	"processes_zombies":              "僵尸进程数",
	"processes_blocked":              "阻塞进程数",
	"linux_sysctl_fs_file_max":       "文件句柄上限",
	"oracle_up":                      "Oracle 在线状态",
	"oracle_buffer_cache_hit_ratio":  "Oracle 缓存命中率",
	"oracle_process_count":           "Oracle 进程数",
	"oracle_sessions":                "Oracle 会话数",
	"oracle_tablespace_used_percent": "Oracle 表空间使用率",
}

func MetricDisplayName(name string) string {
	if cn, ok := metricDisplayNames[name]; ok {
		return cn
	}
	return name
}
