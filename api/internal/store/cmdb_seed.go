package store

import (
	"ai-workbench-api/internal/model"

	"github.com/sirupsen/logrus"
)

// SeedCmdbDefaults inserts preset categories, models, and OS attributes if empty.
func SeedCmdbDefaults() {
	if !GormOK() {
		return
	}
	var count int64
	GetDB().Model(&model.CmdbCategory{}).Count(&count)
	if count > 0 {
		return // already seeded
	}
	logrus.Info("cmdb: seeding default categories, models, and attributes")
	seedCategories()
	seedModels()
	seedOSAttributes()
}

func seedCategories() {
	cats := []model.CmdbCategory{
		{ID: "cat-compute", Label: "计算资源", ParentID: "", Sort: 1},
		{ID: "cat-system", Label: "系统软件", ParentID: "", Sort: 2},
		{ID: "cat-app", Label: "应用软件", ParentID: "", Sort: 3},
		{ID: "cat-network", Label: "网络资源", ParentID: "", Sort: 4},
		{ID: "cat-storage", Label: "存储资源", ParentID: "", Sort: 5},
		{ID: "cat-dc", Label: "机房资源", ParentID: "", Sort: 6},
		{ID: "cat-org", Label: "组织人员", ParentID: "", Sort: 7},
	}
	for i := range cats {
		GetDB().Create(&cats[i])
	}
}

func seedModels() {
	models := []model.CmdbObject{
		{ID: "obj-server", Name: "服务器", CategoryID: "cat-compute", Icon: "server", ObjectType: 101},
		{ID: "obj-vm", Name: "虚拟机", CategoryID: "cat-compute", Icon: "cloud", ObjectType: 101},
		{ID: "obj-docker", Name: "Docker", CategoryID: "cat-compute", Icon: "docker", ObjectType: 101},
		{ID: "obj-k8s", Name: "Kubernetes", CategoryID: "cat-compute", Icon: "k8s", ObjectType: 101},
		{ID: "obj-ecs", Name: "云主机(ECS)", CategoryID: "cat-compute", Icon: "cloud-server", ObjectType: 101},
		{ID: "obj-os", Name: "操作系统", CategoryID: "cat-system", Icon: "desktop", ObjectType: 101},
		{ID: "obj-mysql", Name: "MySQL", CategoryID: "cat-system", Icon: "database", ObjectType: 102},
		{ID: "obj-pgsql", Name: "PostgreSQL", CategoryID: "cat-system", Icon: "database", ObjectType: 102},
		{ID: "obj-redis", Name: "Redis", CategoryID: "cat-system", Icon: "database", ObjectType: 102},
		{ID: "obj-mongodb", Name: "MongoDB", CategoryID: "cat-system", Icon: "database", ObjectType: 102},
		{ID: "obj-nginx", Name: "Nginx", CategoryID: "cat-system", Icon: "gateway", ObjectType: 103},
		{ID: "obj-tomcat", Name: "Tomcat", CategoryID: "cat-system", Icon: "java", ObjectType: 103},
		{ID: "obj-kafka", Name: "Kafka", CategoryID: "cat-system", Icon: "queue", ObjectType: 103},
		{ID: "obj-biz", Name: "业务系统", CategoryID: "cat-app", Icon: "app", ObjectType: 201},
		{ID: "obj-netdev", Name: "网络设备", CategoryID: "cat-network", Icon: "router", ObjectType: 301},
		{ID: "obj-stor", Name: "存储设备", CategoryID: "cat-storage", Icon: "hdd", ObjectType: 401},
		{ID: "obj-room", Name: "机房", CategoryID: "cat-dc", Icon: "building", ObjectType: 501},
		{ID: "obj-rack", Name: "机柜", CategoryID: "cat-dc", Icon: "cabinet", ObjectType: 501},
	}
	for i := range models {
		GetDB().Create(&models[i])
	}
}

func seedOSAttributes() {
	attrs := []model.CmdbAttribute{
		{ID: "attr-os-01", ObjectID: "obj-os", Label: "IP地址", Attr: "ip_address", ValueType: "ip", Tag: "基本信息", Required: true, Unique: true, Discovery: true, Sort: 1},
		{ID: "attr-os-02", ObjectID: "obj-os", Label: "名称", Attr: "name", ValueType: "char", Tag: "基本信息", Required: true, Discovery: true, Sort: 2},
		{ID: "attr-os-03", ObjectID: "obj-os", Label: "是否监控", Attr: "is_monitored", ValueType: "boolean", Tag: "基本信息", Sort: 3},
		{ID: "attr-os-04", ObjectID: "obj-os", Label: "分组", Attr: "groups", ValueType: "array", Tag: "基本信息", Discovery: true, Sort: 4},
		{ID: "attr-os-05", ObjectID: "obj-os", Label: "系统版本", Attr: "os_version", ValueType: "char", Tag: "基本信息", Discovery: true, Sort: 5},
		{ID: "attr-os-06", ObjectID: "obj-os", Label: "系统信息(uname)", Attr: "uname", ValueType: "char", Tag: "基本信息", Discovery: true, Sort: 6},
		{ID: "attr-os-07", ObjectID: "obj-os", Label: "资产负责人", Attr: "owner", ValueType: "char", Tag: "基本信息", Sort: 7},
		{ID: "attr-os-08", ObjectID: "obj-os", Label: "联系电话", Attr: "phone", ValueType: "char", Tag: "基本信息", Sort: 8},
		{ID: "attr-os-09", ObjectID: "obj-os", Label: "系统运行天数", Attr: "uptime_days", ValueType: "int", Tag: "基本信息", Discovery: true, Sort: 9},
		{ID: "attr-os-10", ObjectID: "obj-os", Label: "NTP同步状态", Attr: "ntp_status", ValueType: "enum", Tag: "基本信息", Sort: 10, Options: `["正常","异常","未配置"]`},
		{ID: "attr-os-11", ObjectID: "obj-os", Label: "DNS配置状态", Attr: "dns_status", ValueType: "enum", Tag: "基本信息", Sort: 11, Options: `["正常","异常","未配置"]`},
		{ID: "attr-os-12", ObjectID: "obj-os", Label: "厂商", Attr: "vendor", ValueType: "char", Tag: "基本信息", Sort: 12},
		{ID: "attr-os-13", ObjectID: "obj-os", Label: "附件", Attr: "attachments", ValueType: "file", Tag: "基本信息", Sort: 13},
		{ID: "attr-os-14", ObjectID: "obj-os", Label: "内存大小", Attr: "memory_total", ValueType: "float", Tag: "系统资源", Unit: "B", Discovery: true, Sort: 14},
		{ID: "attr-os-15", ObjectID: "obj-os", Label: "虚拟内存", Attr: "swap_total", ValueType: "int", Tag: "系统资源", Unit: "B", Discovery: true, Sort: 15},
		{ID: "attr-os-16", ObjectID: "obj-os", Label: "网卡数量", Attr: "nic_count", ValueType: "int", Tag: "系统资源", Discovery: true, Sort: 16},
		{ID: "attr-os-17", ObjectID: "obj-os", Label: "磁盘空间", Attr: "disk_total", ValueType: "float", Tag: "系统资源", Unit: "B", Discovery: true, Sort: 17},
		{ID: "attr-os-18", ObjectID: "obj-os", Label: "文件系统", Attr: "filesystems", ValueType: "struct", Tag: "系统资源", Discovery: true, Sort: 18},
		{ID: "attr-os-19", ObjectID: "obj-os", Label: "最大进程数", Attr: "max_procs", ValueType: "int", Tag: "系统资源", Discovery: true, Sort: 19},
		{ID: "attr-os-20", ObjectID: "obj-os", Label: "CPU核数", Attr: "cpu_cores", ValueType: "int", Tag: "系统资源", Discovery: true, Sort: 20},
		{ID: "attr-os-21", ObjectID: "obj-os", Label: "所属业务", Attr: "business", ValueType: "array", Tag: "系统信息", Sort: 21},
		{ID: "attr-os-22", ObjectID: "obj-os", Label: "管理IP", Attr: "mgmt_ip", ValueType: "ip", Tag: "系统信息", Sort: 22},
		{ID: "attr-os-23", ObjectID: "obj-os", Label: "系统属性", Attr: "sys_type", ValueType: "enum", Tag: "系统信息", Sort: 23, Options: `["物理机","虚拟机","容器"]`},
		{ID: "attr-os-24", ObjectID: "obj-os", Label: "机房", Attr: "datacenter", ValueType: "char", Tag: "系统信息", Sort: 24},
		{ID: "attr-os-25", ObjectID: "obj-os", Label: "机柜位置", Attr: "rack_position", ValueType: "char", Tag: "系统信息", Sort: 25},
		{ID: "attr-os-26", ObjectID: "obj-os", Label: "序列号", Attr: "serial_number", ValueType: "char", Tag: "系统信息", Discovery: true, Sort: 26},
		{ID: "attr-os-27", ObjectID: "obj-os", Label: "探针版本", Attr: "agent_version", ValueType: "char", Tag: "Agent", Discovery: true, Sort: 27},
		{ID: "attr-os-28", ObjectID: "obj-os", Label: "心跳时间", Attr: "heartbeat_at", ValueType: "char", Tag: "Agent", Discovery: true, Sort: 28},
		{ID: "attr-os-29", ObjectID: "obj-os", Label: "配置版本", Attr: "config_version", ValueType: "char", Tag: "Agent", Discovery: true, Sort: 29},
		{ID: "attr-os-30", ObjectID: "obj-os", Label: "数据到达时间", Attr: "data_arrival_at", ValueType: "char", Tag: "Agent", Discovery: true, Sort: 30},
		{ID: "attr-os-31", ObjectID: "obj-os", Label: "探针状态", Attr: "agent_status", ValueType: "enum", Tag: "Agent", Discovery: true, Sort: 31, Options: `["在线","离线","未安装"]`},
		{ID: "attr-os-32", ObjectID: "obj-os", Label: "备注", Attr: "remark", ValueType: "char", Tag: "自定义", Sort: 32},
	}
	for i := range attrs {
		GetDB().Create(&attrs[i])
	}
}

