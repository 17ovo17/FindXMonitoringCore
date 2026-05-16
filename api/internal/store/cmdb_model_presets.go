package store

import (
	"sync"

	"ai-workbench-api/internal/model"

	"github.com/sirupsen/logrus"
)

// CmdbModelPreset 预置资源模型定义
type CmdbModelPreset struct {
	ID         string                    `json:"id"`
	Name       string                    `json:"name"`
	Icon       string                    `json:"icon"`
	Category   string                    `json:"category"`
	Attributes []CmdbModelPresetAttribute `json:"attributes"`
}

// CmdbModelPresetAttribute 预置模型属性
type CmdbModelPresetAttribute struct {
	Label     string `json:"label"`
	Attr      string `json:"attr"`
	ValueType string `json:"value_type"`
	Required  bool   `json:"required"`
	Discovery bool   `json:"discovery"`
}

var (
	cmdbModelPresets     []CmdbModelPreset
	cmdbModelPresetsOnce sync.Once
)

// ListCmdbModelPresets 返回全部预置资源模型
func ListCmdbModelPresets() []CmdbModelPreset {
	cmdbModelPresetsOnce.Do(initCmdbModelPresets)
	return cmdbModelPresets
}

// GetCmdbModelPreset 按 ID 获取预置模型
func GetCmdbModelPreset(id string) (*CmdbModelPreset, bool) {
	cmdbModelPresetsOnce.Do(initCmdbModelPresets)
	for i := range cmdbModelPresets {
		if cmdbModelPresets[i].ID == id {
			return &cmdbModelPresets[i], true
		}
	}
	return nil, false
}

// ApplyCmdbModelPreset 将预置模型应用到 CMDB（创建 Object 和 Attributes）
func ApplyCmdbModelPreset(presetID string) (*model.CmdbObject, error) {
	preset, ok := GetCmdbModelPreset(presetID)
	if !ok {
		return nil, nil
	}
	obj := &model.CmdbObject{
		ID:         "obj-preset-" + preset.ID,
		Name:       preset.Name,
		CategoryID: preset.Category,
		Icon:       preset.Icon,
		ObjectType: 101,
	}
	if err := CreateCmdbObject(obj); err != nil {
		logrus.WithError(err).WithField("preset_id", presetID).Warn("cmdb: apply model preset object failed")
		return nil, err
	}
	for i, attr := range preset.Attributes {
		a := &model.CmdbAttribute{
			ObjectID:  obj.ID,
			Label:     attr.Label,
			Attr:      attr.Attr,
			ValueType: attr.ValueType,
			Required:  attr.Required,
			Discovery: attr.Discovery,
			Sort:      i + 1,
			Tag:       "基本信息",
		}
		if err := CreateCmdbAttribute(a); err != nil {
			logrus.WithError(err).WithField("attr", attr.Attr).Warn("cmdb: apply model preset attribute failed")
		}
	}
	return obj, nil
}

func initCmdbModelPresets() {
	cmdbModelPresets = []CmdbModelPreset{
		{
			ID: "operating_system", Name: "操作系统", Icon: "desktop", Category: "cat-system",
			Attributes: []CmdbModelPresetAttribute{
				{Label: "主机名", Attr: "hostname", ValueType: "char", Required: true, Discovery: true},
				{Label: "IP地址", Attr: "ip", ValueType: "ip", Required: true, Discovery: true},
				{Label: "内核版本", Attr: "kernel_version", ValueType: "char", Discovery: true},
				{Label: "运行天数", Attr: "uptime", ValueType: "int", Discovery: true},
				{Label: "CPU核数", Attr: "cpu_cores", ValueType: "int", Discovery: true},
				{Label: "内存总量", Attr: "memory_total", ValueType: "float", Discovery: true},
				{Label: "磁盘总量", Attr: "disk_total", ValueType: "float", Discovery: true},
				{Label: "系统版本", Attr: "os_version", ValueType: "char", Discovery: true},
			},
		},
		{
			ID: "database", Name: "数据库", Icon: "database", Category: "cat-system",
			Attributes: []CmdbModelPresetAttribute{
				{Label: "数据库类型", Attr: "db_type", ValueType: "char", Required: true},
				{Label: "版本", Attr: "version", ValueType: "char", Discovery: true},
				{Label: "端口", Attr: "port", ValueType: "int", Required: true},
				{Label: "最大连接数", Attr: "max_connections", ValueType: "int", Discovery: true},
				{Label: "数据大小", Attr: "data_size", ValueType: "float", Discovery: true},
				{Label: "复制角色", Attr: "replication_role", ValueType: "char", Discovery: true},
			},
		},
		{
			ID: "middleware", Name: "中间件", Icon: "gateway", Category: "cat-system",
			Attributes: []CmdbModelPresetAttribute{
				{Label: "中间件类型", Attr: "mw_type", ValueType: "char", Required: true},
				{Label: "版本", Attr: "version", ValueType: "char", Discovery: true},
				{Label: "端口", Attr: "port", ValueType: "int", Required: true},
				{Label: "配置路径", Attr: "config_path", ValueType: "char"},
				{Label: "进程名", Attr: "process_name", ValueType: "char", Discovery: true},
			},
		},
		{
			ID: "network_device", Name: "网络设备", Icon: "router", Category: "cat-network",
			Attributes: []CmdbModelPresetAttribute{
				{Label: "设备类型", Attr: "device_type", ValueType: "char", Required: true},
				{Label: "固件版本", Attr: "firmware", ValueType: "char", Discovery: true},
				{Label: "管理IP", Attr: "management_ip", ValueType: "ip", Required: true},
				{Label: "端口数量", Attr: "ports_count", ValueType: "int", Discovery: true},
			},
		},
		{
			ID: "server", Name: "服务器", Icon: "server", Category: "cat-compute",
			Attributes: []CmdbModelPresetAttribute{
				{Label: "服务器类型", Attr: "server_type", ValueType: "char", Required: true},
				{Label: "序列号", Attr: "sn", ValueType: "char", Required: true},
				{Label: "制造商", Attr: "manufacturer", ValueType: "char"},
				{Label: "型号", Attr: "model", ValueType: "char"},
				{Label: "机柜位置", Attr: "rack_location", ValueType: "char"},
			},
		},
		{
			ID: "storage", Name: "存储设备", Icon: "hdd", Category: "cat-storage",
			Attributes: []CmdbModelPresetAttribute{
				{Label: "存储类型", Attr: "storage_type", ValueType: "char", Required: true},
				{Label: "容量", Attr: "capacity", ValueType: "float", Required: true},
				{Label: "使用率", Attr: "used_percent", ValueType: "float", Discovery: true},
				{Label: "协议", Attr: "protocol", ValueType: "char"},
			},
		},
		{
			ID: "link", Name: "链路", Icon: "link", Category: "cat-network",
			Attributes: []CmdbModelPresetAttribute{
				{Label: "链路类型", Attr: "link_type", ValueType: "char", Required: true},
				{Label: "带宽", Attr: "bandwidth", ValueType: "char", Required: true},
				{Label: "运营商", Attr: "provider", ValueType: "char"},
				{Label: "端点", Attr: "endpoints", ValueType: "char"},
			},
		},
		{
			ID: "virtualization", Name: "虚拟化", Icon: "cloud", Category: "cat-compute",
			Attributes: []CmdbModelPresetAttribute{
				{Label: "平台", Attr: "platform", ValueType: "char", Required: true},
				{Label: "宿主机", Attr: "host_machine", ValueType: "char", Discovery: true},
				{Label: "虚拟CPU", Attr: "vcpu", ValueType: "int", Discovery: true},
				{Label: "虚拟内存", Attr: "vmemory", ValueType: "float", Discovery: true},
			},
		},
		{
			ID: "probe", Name: "探针", Icon: "radar", Category: "cat-system",
			Attributes: []CmdbModelPresetAttribute{
				{Label: "探针类型", Attr: "probe_type", ValueType: "char", Required: true},
				{Label: "目标URL", Attr: "target_url", ValueType: "char", Required: true},
				{Label: "检测间隔", Attr: "interval", ValueType: "int"},
				{Label: "超时时间", Attr: "timeout", ValueType: "int"},
			},
		},
		{
			ID: "cloud", Name: "云资源", Icon: "cloud-server", Category: "cat-compute",
			Attributes: []CmdbModelPresetAttribute{
				{Label: "云厂商", Attr: "cloud_provider", ValueType: "char", Required: true},
				{Label: "区域", Attr: "region", ValueType: "char", Required: true},
				{Label: "实例类型", Attr: "instance_type", ValueType: "char", Discovery: true},
				{Label: "账号ID", Attr: "account_id", ValueType: "char"},
			},
		},
		{
			ID: "container", Name: "容器", Icon: "docker", Category: "cat-compute",
			Attributes: []CmdbModelPresetAttribute{
				{Label: "集群", Attr: "cluster", ValueType: "char", Required: true},
				{Label: "命名空间", Attr: "namespace", ValueType: "char", Required: true},
				{Label: "Pod名称", Attr: "pod_name", ValueType: "char", Discovery: true},
				{Label: "镜像", Attr: "image", ValueType: "char", Discovery: true},
				{Label: "副本数", Attr: "replicas", ValueType: "int", Discovery: true},
			},
		},
		{
			ID: "iot", Name: "IoT设备", Icon: "sensor", Category: "cat-compute",
			Attributes: []CmdbModelPresetAttribute{
				{Label: "设备型号", Attr: "device_model", ValueType: "char", Required: true},
				{Label: "固件版本", Attr: "firmware_version", ValueType: "char", Discovery: true},
				{Label: "协议", Attr: "protocol", ValueType: "char"},
				{Label: "网关IP", Attr: "gateway_ip", ValueType: "ip"},
			},
		},
	}
}
