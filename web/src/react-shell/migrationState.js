const migrationItem = (id, title, path, status, reason, sourceEvidence) => ({
  id,
  path,
  title,
  sourceEvidence,
  migration: {
    reason,
    status,
  },
})

const blocked = (id, title, path, sourceEvidence) =>
  migrationItem(
    id,
    title,
    path,
    'PENDING',
    '该页面仅建立 React-only final 迁移边界，尚未接入 FindX 后端契约、权限、审计和真实交互。',
    sourceEvidence,
  )

const tempBridge = (id, title, path, sourceEvidence) =>
  migrationItem(
    id,
    title,
    path,
    'TEMP_BRIDGE',
    '当前入口仅允许承载 Vue workbench 临时桥接，不能作为 React-only final 完成态验收。',
    sourceEvidence,
  )

export const createBlockedMigrationState = () => [
  {
    id: 'integrations',
    title: '集成中心',
    children: [
      blocked('datasources', '数据源', '/react/datasources', [
        'D:\\平台源码\\fe-main\\src\\pages\\datasource\\index.tsx',
        'D:\\平台源码\\fe-main\\src\\pages\\datasource\\services.ts',
      ]),
      blocked('components', '模板中心', '/react/components', [
        'D:\\平台源码\\fe-main\\src\\pages\\builtInComponents\\entry.tsx',
        'D:\\平台源码\\fe-main\\src\\pages\\builtInComponents\\List.tsx',
      ]),
      blocked('embedded-products', '系统集成', '/react/embedded-products', [
        'D:\\平台源码\\fe-main\\src\\components\\SideMenu\\menu.tsx',
      ]),
    ],
  },
  {
    id: 'query',
    title: '数据查询',
    children: [
      blocked('metric-explorer', '指标查询', '/react/metric/explorer', [
        'D:\\平台源码\\fe-main\\src\\pages\\explorer\\Metric.tsx',
        'D:\\平台源码\\fe-main\\src\\pages\\explorer\\Explorer.tsx',
      ]),
      blocked('built-in-metrics', '内置指标', '/react/metrics-built-in', [
        'D:\\平台源码\\fe-main\\src\\components\\SideMenu\\menu.tsx',
      ]),
      blocked('object-explorer', '对象快捷视图', '/react/object/explorer', [
        'D:\\平台源码\\fe-main\\src\\pages\\monitor\\object',
      ]),
      blocked('recording-rules', '记录规则', '/react/recording-rules', [
        'D:\\平台源码\\fe-main\\src\\pages\\recordingRules',
      ]),
    ],
  },
  {
    id: 'dashboards',
    title: '仪表盘',
    children: [
      blocked('dashboards', '仪表盘', '/react/dashboards', [
        'D:\\平台源码\\fe-main\\src\\pages\\dashboard\\index.tsx',
        'D:\\平台源码\\fe-main\\src\\pages\\dashboard\\Detail\\index.tsx',
      ]),
    ],
  },
  {
    id: 'agents',
    title: 'Agent 管理中心',
    children: [
      tempBridge('agent-overview', '概览', '/react/agents/overview', [
        'source/trace-backend',
        'source/trace-backend-ui',
      ]),
      tempBridge('agent-hosts', '主机 Agent', '/react/agents/hosts', [
        'source/trace-backend/docs',
      ]),
      tempBridge('agent-packages', '能力包', '/react/agents/packages', [
        'source/trace-backend/apm-sniffer',
      ]),
      tempBridge('agent-install', '安装向导', '/react/agents/install', [
        'source/trace-backend/docs',
      ]),
      tempBridge('agent-templates', '配置模板', '/react/agents/templates', [
        'source/trace-backend/docs',
      ]),
      tempBridge('agent-heartbeat', '心跳状态', '/react/agents/heartbeat', [
        'source/trace-backend',
      ]),
      tempBridge('agent-data-arrival', '数据到达', '/react/agents/data-arrival', [
        'source/trace-backend',
      ]),
    ],
  },
  {
    id: 'runtime-bridges',
    title: 'TEMP_BRIDGE',
    children: [
      tempBridge('vue-workbench', 'Vue workbench 临时桥', '/react/temp-bridge/vue-workbench', [
        'D:\\ai-workbench\\web\\src\\views\\*Workbench.vue',
      ]),
    ],
  },
]
