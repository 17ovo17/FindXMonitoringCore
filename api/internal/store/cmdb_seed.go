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
		// LWOPS 对齐 12 类资源模型
		{ID: "obj-db", Name: "数据库", CategoryID: "cat-system", Icon: "database", ObjectType: 102},
		{ID: "obj-middleware", Name: "中间件", CategoryID: "cat-system", Icon: "gateway", ObjectType: 103},
		{ID: "obj-network", Name: "网络设备", CategoryID: "cat-network", Icon: "router", ObjectType: 301},
		{ID: "obj-storage", Name: "存储", CategoryID: "cat-storage", Icon: "hdd", ObjectType: 401},
		{ID: "obj-link", Name: "链路", CategoryID: "cat-network", Icon: "link", ObjectType: 301},
		{ID: "obj-probe", Name: "探测", CategoryID: "cat-system", Icon: "radar", ObjectType: 101},
		{ID: "obj-cloud", Name: "云平台", CategoryID: "cat-compute", Icon: "cloud-server", ObjectType: 101},
		{ID: "obj-container", Name: "容器", CategoryID: "cat-compute", Icon: "docker", ObjectType: 101},
		{ID: "obj-iot", Name: "物联网", CategoryID: "cat-compute", Icon: "sensor", ObjectType: 101},
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
	seedLwopsModelAttributes()
}

func seedLwopsModelAttributes() {
	attrs := []model.CmdbAttribute{
		// 数据库
		{ID: "attr-db-01", ObjectID: "obj-db", Label: "名称", Attr: "name", ValueType: "char", Tag: "基本信息", Required: true, Sort: 1},
		{ID: "attr-db-02", ObjectID: "obj-db", Label: "IP地址", Attr: "ip_address", ValueType: "ip", Tag: "基本信息", Required: true, Discovery: true, Sort: 2},
		{ID: "attr-db-03", ObjectID: "obj-db", Label: "数据库类型", Attr: "db_type", ValueType: "enum", Tag: "基本信息", Required: true, Sort: 3, Options: `["MySQL","PostgreSQL","Redis","MongoDB","Oracle","SQLServer"]`},
		{ID: "attr-db-04", ObjectID: "obj-db", Label: "版本", Attr: "version", ValueType: "char", Tag: "基本信息", Discovery: true, Sort: 4},
		{ID: "attr-db-05", ObjectID: "obj-db", Label: "端口", Attr: "port", ValueType: "int", Tag: "基本信息", Required: true, Sort: 5},
		{ID: "attr-db-06", ObjectID: "obj-db", Label: "最大连接数", Attr: "max_connections", ValueType: "int", Tag: "性能", Discovery: true, Sort: 6},
		{ID: "attr-db-07", ObjectID: "obj-db", Label: "数据大小", Attr: "data_size", ValueType: "float", Tag: "性能", Unit: "GB", Discovery: true, Sort: 7},
		{ID: "attr-db-08", ObjectID: "obj-db", Label: "复制角色", Attr: "replication_role", ValueType: "enum", Tag: "高可用", Discovery: true, Sort: 8, Options: `["主","从","单机"]`},
		// 中间件
		{ID: "attr-mw-01", ObjectID: "obj-middleware", Label: "名称", Attr: "name", ValueType: "char", Tag: "基本信息", Required: true, Sort: 1},
		{ID: "attr-mw-02", ObjectID: "obj-middleware", Label: "IP地址", Attr: "ip_address", ValueType: "ip", Tag: "基本信息", Required: true, Discovery: true, Sort: 2},
		{ID: "attr-mw-03", ObjectID: "obj-middleware", Label: "中间件类型", Attr: "mw_type", ValueType: "enum", Tag: "基本信息", Required: true, Sort: 3, Options: `["Nginx","Tomcat","Kafka","RabbitMQ","Zookeeper","Nacos"]`},
		{ID: "attr-mw-04", ObjectID: "obj-middleware", Label: "版本", Attr: "version", ValueType: "char", Tag: "基本信息", Discovery: true, Sort: 4},
		{ID: "attr-mw-05", ObjectID: "obj-middleware", Label: "端口", Attr: "port", ValueType: "int", Tag: "基本信息", Required: true, Sort: 5},
		{ID: "attr-mw-06", ObjectID: "obj-middleware", Label: "配置路径", Attr: "config_path", ValueType: "char", Tag: "运维", Sort: 6},
		{ID: "attr-mw-07", ObjectID: "obj-middleware", Label: "进程名", Attr: "process_name", ValueType: "char", Tag: "运维", Discovery: true, Sort: 7},
		// 网络设备
		{ID: "attr-net-01", ObjectID: "obj-network", Label: "名称", Attr: "name", ValueType: "char", Tag: "基本信息", Required: true, Sort: 1},
		{ID: "attr-net-02", ObjectID: "obj-network", Label: "管理IP", Attr: "management_ip", ValueType: "ip", Tag: "基本信息", Required: true, Discovery: true, Sort: 2},
		{ID: "attr-net-03", ObjectID: "obj-network", Label: "设备类型", Attr: "device_type", ValueType: "enum", Tag: "基本信息", Required: true, Sort: 3, Options: `["交换机","路由器","防火墙","负载均衡"]`},
		{ID: "attr-net-04", ObjectID: "obj-network", Label: "固件版本", Attr: "firmware", ValueType: "char", Tag: "基本信息", Discovery: true, Sort: 4},
		{ID: "attr-net-05", ObjectID: "obj-network", Label: "端口数量", Attr: "ports_count", ValueType: "int", Tag: "基本信息", Discovery: true, Sort: 5},
		{ID: "attr-net-06", ObjectID: "obj-network", Label: "厂商", Attr: "vendor", ValueType: "char", Tag: "基本信息", Sort: 6},
		// 存储
		{ID: "attr-sto-01", ObjectID: "obj-storage", Label: "名称", Attr: "name", ValueType: "char", Tag: "基本信息", Required: true, Sort: 1},
		{ID: "attr-sto-02", ObjectID: "obj-storage", Label: "IP地址", Attr: "ip_address", ValueType: "ip", Tag: "基本信息", Required: true, Discovery: true, Sort: 2},
		{ID: "attr-sto-03", ObjectID: "obj-storage", Label: "存储类型", Attr: "storage_type", ValueType: "enum", Tag: "基本信息", Required: true, Sort: 3, Options: `["SAN","NAS","对象存储","分布式"]`},
		{ID: "attr-sto-04", ObjectID: "obj-storage", Label: "总容量", Attr: "capacity", ValueType: "float", Tag: "容量", Unit: "TB", Required: true, Sort: 4},
		{ID: "attr-sto-05", ObjectID: "obj-storage", Label: "使用率", Attr: "used_percent", ValueType: "float", Tag: "容量", Unit: "%", Discovery: true, Sort: 5},
		{ID: "attr-sto-06", ObjectID: "obj-storage", Label: "协议", Attr: "protocol", ValueType: "char", Tag: "基本信息", Sort: 6},
		// 链路
		{ID: "attr-lnk-01", ObjectID: "obj-link", Label: "名称", Attr: "name", ValueType: "char", Tag: "基本信息", Required: true, Sort: 1},
		{ID: "attr-lnk-02", ObjectID: "obj-link", Label: "链路类型", Attr: "link_type", ValueType: "enum", Tag: "基本信息", Required: true, Sort: 2, Options: `["专线","VPN","互联网","内网"]`},
		{ID: "attr-lnk-03", ObjectID: "obj-link", Label: "带宽", Attr: "bandwidth", ValueType: "char", Tag: "基本信息", Required: true, Sort: 3},
		{ID: "attr-lnk-04", ObjectID: "obj-link", Label: "运营商", Attr: "provider", ValueType: "char", Tag: "基本信息", Sort: 4},
		{ID: "attr-lnk-05", ObjectID: "obj-link", Label: "源端点", Attr: "source_endpoint", ValueType: "char", Tag: "拓扑", Sort: 5},
		{ID: "attr-lnk-06", ObjectID: "obj-link", Label: "目标端点", Attr: "target_endpoint", ValueType: "char", Tag: "拓扑", Sort: 6},
		// 探测
		{ID: "attr-prb-01", ObjectID: "obj-probe", Label: "名称", Attr: "name", ValueType: "char", Tag: "基本信息", Required: true, Sort: 1},
		{ID: "attr-prb-02", ObjectID: "obj-probe", Label: "探测类型", Attr: "probe_type", ValueType: "enum", Tag: "基本信息", Required: true, Sort: 2, Options: `["HTTP","TCP","ICMP","DNS","SSL"]`},
		{ID: "attr-prb-03", ObjectID: "obj-probe", Label: "目标地址", Attr: "target_url", ValueType: "char", Tag: "基本信息", Required: true, Sort: 3},
		{ID: "attr-prb-04", ObjectID: "obj-probe", Label: "检测间隔", Attr: "interval", ValueType: "int", Tag: "策略", Unit: "s", Sort: 4},
		{ID: "attr-prb-05", ObjectID: "obj-probe", Label: "超时时间", Attr: "timeout", ValueType: "int", Tag: "策略", Unit: "s", Sort: 5},
		// 云平台
		{ID: "attr-cld-01", ObjectID: "obj-cloud", Label: "名称", Attr: "name", ValueType: "char", Tag: "基本信息", Required: true, Sort: 1},
		{ID: "attr-cld-02", ObjectID: "obj-cloud", Label: "云厂商", Attr: "cloud_provider", ValueType: "enum", Tag: "基本信息", Required: true, Sort: 2, Options: `["阿里云","腾讯云","华为云","AWS","Azure"]`},
		{ID: "attr-cld-03", ObjectID: "obj-cloud", Label: "区域", Attr: "region", ValueType: "char", Tag: "基本信息", Required: true, Sort: 3},
		{ID: "attr-cld-04", ObjectID: "obj-cloud", Label: "实例类型", Attr: "instance_type", ValueType: "char", Tag: "基本信息", Discovery: true, Sort: 4},
		{ID: "attr-cld-05", ObjectID: "obj-cloud", Label: "账号ID", Attr: "account_id", ValueType: "char", Tag: "基本信息", Sort: 5},
		{ID: "attr-cld-06", ObjectID: "obj-cloud", Label: "公网IP", Attr: "public_ip", ValueType: "ip", Tag: "网络", Discovery: true, Sort: 6},
		// 容器
		{ID: "attr-ctn-01", ObjectID: "obj-container", Label: "名称", Attr: "name", ValueType: "char", Tag: "基本信息", Required: true, Sort: 1},
		{ID: "attr-ctn-02", ObjectID: "obj-container", Label: "集群", Attr: "cluster", ValueType: "char", Tag: "基本信息", Required: true, Sort: 2},
		{ID: "attr-ctn-03", ObjectID: "obj-container", Label: "命名空间", Attr: "namespace", ValueType: "char", Tag: "基本信息", Required: true, Sort: 3},
		{ID: "attr-ctn-04", ObjectID: "obj-container", Label: "Pod名称", Attr: "pod_name", ValueType: "char", Tag: "运行时", Discovery: true, Sort: 4},
		{ID: "attr-ctn-05", ObjectID: "obj-container", Label: "镜像", Attr: "image", ValueType: "char", Tag: "运行时", Discovery: true, Sort: 5},
		{ID: "attr-ctn-06", ObjectID: "obj-container", Label: "副本数", Attr: "replicas", ValueType: "int", Tag: "运行时", Discovery: true, Sort: 6},
		// 物联网
		{ID: "attr-iot-01", ObjectID: "obj-iot", Label: "名称", Attr: "name", ValueType: "char", Tag: "基本信息", Required: true, Sort: 1},
		{ID: "attr-iot-02", ObjectID: "obj-iot", Label: "设备型号", Attr: "device_model", ValueType: "char", Tag: "基本信息", Required: true, Sort: 2},
		{ID: "attr-iot-03", ObjectID: "obj-iot", Label: "固件版本", Attr: "firmware_version", ValueType: "char", Tag: "基本信息", Discovery: true, Sort: 3},
		{ID: "attr-iot-04", ObjectID: "obj-iot", Label: "协议", Attr: "protocol", ValueType: "enum", Tag: "通信", Sort: 4, Options: `["MQTT","CoAP","HTTP","Modbus"]`},
		{ID: "attr-iot-05", ObjectID: "obj-iot", Label: "网关IP", Attr: "gateway_ip", ValueType: "ip", Tag: "通信", Sort: 5},
		{ID: "attr-iot-06", ObjectID: "obj-iot", Label: "在线状态", Attr: "online_status", ValueType: "enum", Tag: "状态", Discovery: true, Sort: 6, Options: `["在线","离线","休眠"]`},
	}
	for i := range attrs {
		GetDB().Create(&attrs[i])
	}
}

// SeedCmdbContractProbeTopology creates a stable, real CMDB relation graph for
// contract-probe so remote browser/API checks can exercise the ready path.
func SeedCmdbContractProbeTopology() {
	visible := true
	objects := []model.CmdbObject{
		{ID: "OperatingSystem1", Name: "操作系统", CategoryID: "cat-system", Icon: "desktop", ObjectType: 101},
		{ID: "j6p8Wb2xkV1666171515", Name: "用户", CategoryID: "cat-org", Icon: "user", ObjectType: 101},
	}
	for i := range objects {
		if err := upsertCmdbObject(objects[i]); err != nil {
			logrus.WithError(err).WithField("object_id", objects[i].ID).Warn("cmdb: seed contract probe object failed")
			return
		}
	}

	instances := []model.CmdbInstance{
		{
			ID:       "contract-probe",
			ObjectID: "OperatingSystem1",
			Data:     `{"name":"FindX contract probe OS","ip_address":"198.18.126.10","OS001":"198.18.126.10","os_version":"Debian compatible","agent_status":"contract_blocked"}`,
			Creator:  "contract-seed",
			Updater:  "contract-seed",
		},
		{
			ID:       "contract-probe-user",
			ObjectID: "j6p8Wb2xkV1666171515",
			Data:     `{"name":"FindX contract probe owner","user_code":"contract-probe-owner"}`,
			Creator:  "contract-seed",
			Updater:  "contract-seed",
		},
	}
	for i := range instances {
		if err := upsertCmdbInstance(instances[i]); err != nil {
			logrus.WithError(err).WithField("instance_id", instances[i].ID).Warn("cmdb: seed contract probe instance failed")
			return
		}
	}

	relationType := model.CmdbRelationType{
		ID:             "OperatingSystem1_default_j6p8Wb2xkV1666171515",
		Name:           "default",
		Label:          "关联",
		Mapping:        "n:1",
		Visible:        &visible,
		RuleLogic:      "and",
		RuleExpression: "A",
		RulesJSON:      `[{"left_attr":"x5qvHkM1Bz1661218322","logic":"=","right_attr":"Ndg2mVtTas1666171547","tag":"A"}]`,
		LeftMin:        0,
		RightMin:       0,
		LeftMax:        1,
		RightMax:       -1,
		Source:         2,
		LeftAsstName:   "关联",
		RightAsstName:  "关联",
	}
	if err := upsertCmdbRelationType(relationType); err != nil {
		logrus.WithError(err).WithField("relation_type_id", relationType.ID).Warn("cmdb: seed contract probe relation type failed")
		return
	}

	relation := model.CmdbInstanceRelation{
		ID:               "contract-probe-relation-default-user",
		SourceInstanceID: "contract-probe",
		TargetInstanceID: "contract-probe-user",
		RelationTypeID:   relationType.ID,
	}
	if err := upsertCmdbInstanceRelation(relation); err != nil {
		logrus.WithError(err).WithField("relation_id", relation.ID).Warn("cmdb: seed contract probe relation failed")
	}
}

func upsertCmdbObject(obj model.CmdbObject) error {
	if existing, ok := GetCmdbObject(obj.ID); ok {
		existing.Name = obj.Name
		existing.CategoryID = obj.CategoryID
		existing.Icon = obj.Icon
		existing.ObjectType = obj.ObjectType
		return UpdateCmdbObject(existing)
	}
	return CreateCmdbObject(&obj)
}

func upsertCmdbInstance(inst model.CmdbInstance) error {
	if existing, ok := GetCmdbInstance(inst.ID); ok {
		existing.ObjectID = inst.ObjectID
		existing.Data = inst.Data
		existing.Updater = inst.Updater
		if existing.Creator == "" {
			existing.Creator = inst.Creator
		}
		return UpdateCmdbInstance(existing)
	}
	return CreateCmdbInstance(&inst)
}

func upsertCmdbRelationType(rel model.CmdbRelationType) error {
	if _, ok := GetCmdbRelationType(rel.ID); ok && GormOK() {
		return GetDB().Save(&rel).Error
	}
	return CreateCmdbRelationType(&rel)
}

func upsertCmdbInstanceRelation(rel model.CmdbInstanceRelation) error {
	for _, existing := range ListCmdbInstanceRelations(rel.SourceInstanceID) {
		if existing.ID == rel.ID {
			return nil
		}
	}
	return CreateCmdbInstanceRelation(&rel)
}
