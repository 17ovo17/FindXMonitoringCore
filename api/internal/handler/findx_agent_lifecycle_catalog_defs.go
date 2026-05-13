package handler

import "strings"

func findXAgentPackageDefs() []agentPackageDef {
	defs := baseFindXAgentPackageDefs()
	defs = append(defs, applicationProbePackageDefs()...)
	return append(defs, edgeFindXAgentPackageDefs()...)
}

func baseFindXAgentPackageDefs() []agentPackageDef {
	return []agentPackageDef{
		{
			id: "agent-core", name: "FindX Agent 核心", domain: "基础 Agent", runtime: "Agent Core",
			osList: []string{"Linux", "Windows", "Kubernetes"}, shape: "service / daemon / sidecar",
			telemetryKinds: []string{"heartbeat", "task", "config"}, configKeys: []string{"agent_id", "heartbeat_interval", "task_channel", "global_labels", "credential_ref"},
			configTemplateIDs: []string{"agent-core"},
		},
		{
			id: "host-collector", name: "主机采集能力包", domain: "基础采集", runtime: "Host",
			osList: []string{"Linux", "Windows"}, shape: "collector plugins / service",
			telemetryKinds: []string{"metrics", "process", "host"}, configKeys: []string{"scrape_interval", "plugin_set", "plugin_id", "plugin_version", "config_snippet_ref", "provider_mode", "reload_strategy", "global_labels", "credential_ref"},
			configTemplateIDs: []string{"metrics", "host-plugin"},
			pluginConfig:      pluginConfigSpec("host-plugin", pluginIDForTemplate("host-plugin"), "local-reload"),
		},
		{
			id: "container-collector", name: "容器采集能力包", domain: "基础采集", runtime: "Container",
			osList: []string{"Linux", "Kubernetes"}, shape: "daemonset / sidecar / container plugin",
			telemetryKinds: []string{"metrics", "container", "workload"}, configKeys: []string{"cluster_ref", "namespace_selector", "workload_selector", "plugin_id", "config_snippet_ref", "provider_mode", "restart_strategy", "credential_ref"},
			configTemplateIDs: []string{"metrics", "container-plugin"},
			pluginConfig:      pluginConfigSpec("container-plugin", pluginIDForTemplate("container-plugin"), "rolling-restart"),
		},
		{
			id: "log-collector", name: "日志采集能力包", domain: "日志采集", runtime: "Logs",
			osList: []string{"Linux", "Windows", "Kubernetes"}, shape: "file tailer / stdout collector / parser",
			telemetryKinds: []string{"logs", "pipeline"}, configKeys: []string{"paths", "parser", "pipeline_ref", "labels", "credential_ref"},
			configTemplateIDs: []string{"logs"},
		},
		{
			id: "inspection-runner", name: "巡检诊断能力包", domain: "巡检诊断", runtime: "Inspection",
			osList: []string{"Linux", "Windows", "Kubernetes"}, shape: "check runner / diagnostic plugin",
			telemetryKinds: []string{"inspection", "evidence"}, configKeys: []string{"check_set", "schedule", "risk_level", "evidence_chain", "credential_ref"},
			configTemplateIDs: []string{"inspection", "host-plugin"},
		},
	}
}

func applicationProbePackageDefs() []agentPackageDef {
	return []agentPackageDef{
		applicationProbe("java-app", "Java 应用能力包", "Java"),
		applicationProbe("python-app", "Python 应用能力包", "Python"),
		applicationProbe("nodejs-app", "Node.js 应用能力包", "Node.js"),
		applicationProbe("php-app", "PHP 应用能力包", "PHP"),
		applicationProbe("go-app", "Go 应用能力包", "Go"),
		applicationProbe("rust-app", "Rust 应用能力包", "Rust"),
		applicationProbe("ruby-app", "Ruby 应用能力包", "Ruby"),
	}
}

func edgeFindXAgentPackageDefs() []agentPackageDef {
	return []agentPackageDef{
		{
			id: "gateway-probe", name: "网关链路能力包", domain: "网关链路", runtime: "Gateway",
			osList: []string{"Linux", "Kubernetes"}, shape: "gateway plugin / reverse proxy module",
			telemetryKinds: []string{"tracing", "topology", "logs"}, configKeys: []string{"gateway_id", "route_selector", "collector_endpoint", "sampling", "reload_policy"},
			configTemplateIDs: []string{"gateway-plugin", "tracing", "logs"},
			pluginConfig:      pluginConfigSpec("gateway-plugin", pluginIDForTemplate("gateway-plugin"), "reload"),
		},
		{
			id: "browser-client", name: "前端体验能力包", domain: "前端体验", runtime: "Web",
			osList: []string{"Web"}, shape: "JavaScript SDK",
			telemetryKinds: []string{"rum", "tracing", "errors"}, configKeys: []string{"app_id", "domain", "collector_endpoint", "trace_context", "sampling"},
			configTemplateIDs: []string{"browser-probe"},
		},
	}
}

func applicationProbe(id, name, runtime string) agentPackageDef {
	return agentPackageDef{
		id: id, name: name, domain: "应用链路", runtime: runtime,
		osList: []string{"Linux", "Windows", "Kubernetes"}, shape: "runtime package / SDK / sidecar / init container",
		telemetryKinds:    []string{"tracing", "profiling", "topology"},
		configKeys:        []string{"service_name", "collector_endpoint", "instance_name", "environment", "sampling"},
		configTemplateIDs: []string{"tracing", "profiling"},
	}
}

func sourceHint(parts ...string) string {
	return strings.Join(parts, "")
}

func findXAgentConfigTemplateDefs() []agentConfigTemplateDef {
	defs := coreFindXAgentConfigTemplateDefs()
	defs = append(defs, diagnosticFindXAgentConfigTemplateDefs()...)
	defs = append(defs, pluginFindXAgentConfigTemplateDefs()...)
	return append(defs, applicationProbeConfigTemplateDefs()...)
}

func coreFindXAgentConfigTemplateDefs() []agentConfigTemplateDef {
	return []agentConfigTemplateDef{
		{
			id: "agent-core", name: "Agent 基础配置", kind: "基础配置", scope: "注册 / 心跳 / 任务 / 标签",
			fields:             []string{"agent_id", "heartbeat_interval", "global_labels", "task_channel", "credential_ref"},
			targetScopes:       []string{"全部 Agent", "业务组", "主机", "能力包"},
			rolloutScopes:      []string{"注册参数", "心跳周期", "任务通道", "全局标签"},
			rollbackPolicy:     "按配置版本回滚到上一套稳定基础配置",
			capabilityPackages: []string{"FindX Agent 核心"},
		},
		{
			id: "metrics", name: "指标采集配置", kind: "采集配置", scope: "主机 / 容器 / 进程",
			fields:             []string{"scrape_interval", "labels", "resource_group", "credential_ref"},
			targetScopes:       []string{"全部 Agent", "业务组", "主机", "能力包"},
			rolloutScopes:      []string{"主机指标", "容器指标", "进程指标"},
			rollbackPolicy:     "按采集配置版本回滚到上一套稳定版本",
			capabilityPackages: []string{"主机采集能力包", "容器采集能力包"},
		},
		{
			id: "logs", name: "日志采集配置", kind: "日志配置", scope: "文件 / 标准输出 / 系统日志",
			fields:             []string{"paths", "parser", "pipeline_ref", "labels"},
			targetScopes:       []string{"全部 Agent", "业务组", "主机", "能力包"},
			rolloutScopes:      []string{"文件日志", "容器标准输出", "系统日志"},
			rollbackPolicy:     "按管道引用和采集路径回滚",
			capabilityPackages: []string{"日志采集能力包"},
		},
		{
			id: "tracing", name: "应用链路配置", kind: "链路配置", scope: "应用 / 服务 / 网关",
			fields:             []string{"service_name", "collector_endpoint", "sampling", "propagation"},
			targetScopes:       []string{"业务组", "主机", "能力包", "服务"},
			rolloutScopes:      []string{"应用链路", "服务调用", "网关入口"},
			rollbackPolicy:     "保留采样率、传播协议和接入端点版本快照",
			capabilityPackages: []string{"应用链路能力包", "网关链路能力包"},
		},
	}
}

func diagnosticFindXAgentConfigTemplateDefs() []agentConfigTemplateDef {
	return []agentConfigTemplateDef{
		{
			id: "profiling", name: "性能分析配置", kind: "性能分析配置", scope: "应用运行时",
			fields:             []string{"target_runtime", "duration", "cpu", "memory"},
			targetScopes:       []string{"业务组", "主机", "能力包"},
			rolloutScopes:      []string{"CPU 分析", "内存分析", "运行时分析"},
			rollbackPolicy:     "到期自动关闭并恢复默认采样策略",
			capabilityPackages: []string{"应用链路能力包"},
		},
		{
			id: "inspection", name: "巡检诊断配置", kind: "巡检配置", scope: "主机 / 服务 / 端口",
			fields:             []string{"check_set", "schedule", "risk_level", "evidence_chain"},
			targetScopes:       []string{"全部 Agent", "业务组", "主机", "能力包"},
			rolloutScopes:      []string{"主机巡检", "服务巡检", "端口巡检"},
			rollbackPolicy:     "回滚到上一套巡检集和调度周期",
			capabilityPackages: []string{"巡检诊断能力包"},
		},
	}
}

func pluginFindXAgentConfigTemplateDefs() []agentConfigTemplateDef {
	return []agentConfigTemplateDef{
		pluginTemplate("host-plugin", "主机插件配置", "Linux / Windows 主机插件", "主机采集能力包", "巡检诊断能力包"),
		pluginTemplate("container-plugin", "容器插件配置", "Kubernetes / Docker 插件", "容器采集能力包"),
		pluginTemplate("gateway-plugin", "网关插件配置", "网关 / 反向代理 / 边缘节点", "网关链路能力包"),
	}
}

func applicationProbeConfigTemplateDefs() []agentConfigTemplateDef {
	return []agentConfigTemplateDef{
		{
			id: "browser-probe", name: "前端体验配置", kind: "前端体验配置", scope: "Web 应用 / 页面 / 域名",
			fields:             []string{"app_id", "domain", "collector_endpoint", "trace_context", "sampling"},
			targetScopes:       []string{"业务组", "Web 应用", "域名", "能力包"},
			rolloutScopes:      []string{"前端链路", "页面性能", "错误采集"},
			rollbackPolicy:     "回滚到上一套应用标识、采样率和上下文传播配置",
			capabilityPackages: []string{"前端体验能力包"},
		},
	}
}

func pluginTemplate(id, name, scope string, packages ...string) agentConfigTemplateDef {
	return agentConfigTemplateDef{
		id: id, name: name, kind: "插件配置", scope: scope,
		fields:             []string{"plugin_id", "plugin_version", "config_format", "config_snippet_ref", "provider_mode", "reload_strategy", "remote_mutation", "canary_percent", "rollback_ref", "audit_reason", "change_ticket", "credential_ref"},
		targetScopes:       []string{"全部 Agent", "业务组", "主机", "能力包"},
		rolloutScopes:      []string{"插件配置", "HTTP Provider 配置", "远程修改"},
		rollbackPolicy:     "按插件配置版本和 TOML 片段引用回滚，并保留远程修改审计",
		capabilityPackages: packages,
		pluginConfig:       pluginConfigSpec(id, pluginIDForTemplate(id), reloadStrategyForTemplate(id)),
	}
}

func pluginIDForTemplate(id string) string {
	switch id {
	case "host-plugin":
		return "findx-agent.host-plugin-source-map"
	case "container-plugin":
		return "findx-agent.container-plugin-source-map"
	default:
		return "findx-agent.gateway-plugin-source-map"
	}
}

func reloadStrategyForTemplate(id string) string {
	if id == "container-plugin" {
		return "rolling-restart"
	}
	if id == "gateway-plugin" {
		return "reload"
	}
	return "local-reload"
}
