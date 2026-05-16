/**
 * API 响应类型定义（JSDoc）
 * 由于项目不使用 TypeScript，通过 JSDoc @typedef 提供类型提示。
 */

/**
 * @typedef {Object} AlertRule
 * @property {number} id
 * @property {string} name - 规则名称
 * @property {string} expr - PromQL 表达式
 * @property {number} severity - 1=critical, 2=warning, 3=info
 * @property {string} status - enabled | disabled
 * @property {string} duration - 持续时间（如 "5m"）
 * @property {Object} labels - 标签键值对
 * @property {Object} annotations - 注解键值对
 * @property {string} created_at
 * @property {string} updated_at
 */

/**
 * @typedef {Object} AlertEvent
 * @property {number} id
 * @property {number} rule_id
 * @property {string} rule_name
 * @property {number} severity - 1=critical, 2=warning, 3=info
 * @property {string} status - firing | resolved | acked
 * @property {string} labels - JSON 字符串
 * @property {string} annotations - JSON 字符串
 * @property {number} value - 触发值
 * @property {string} starts_at
 * @property {string|null} ends_at
 * @property {string} acked_by
 * @property {string|null} acked_at
 * @property {string} assigned_to
 * @property {string|null} resolved_at
 * @property {string} created_at
 * @property {string} updated_at
 */

/**
 * @typedef {Object} AlertAggregateGroup
 * @property {number} rule_id
 * @property {string} rule_name
 * @property {number} severity
 * @property {string} status
 * @property {number} count
 * @property {string} first_seen
 * @property {string} last_seen
 * @property {number} sample_id
 * @property {string} labels
 */

// PLACEHOLDER_TYPES_APPEND

/**
 * @typedef {Object} Dashboard
 * @property {number} id
 * @property {string} name
 * @property {string} description
 * @property {Array<DashboardPanel>} panels
 * @property {Object} variables - 变量定义
 * @property {string} created_at
 * @property {string} updated_at
 */

/**
 * @typedef {Object} DashboardPanel
 * @property {string} id
 * @property {string} title
 * @property {string} type - graph | stat | table | text
 * @property {Object} datasource
 * @property {Array<Object>} targets - 查询目标
 * @property {Object} options - 面板配置
 * @property {Object} gridPos - 位置 {x, y, w, h}
 */

/**
 * @typedef {Object} Resource
 * @property {number} id
 * @property {string} name
 * @property {string} type - host | database | service | network
 * @property {string} ip
 * @property {string} status - online | offline | unknown
 * @property {Object} attributes - 自定义属性
 * @property {string} business_id
 * @property {string} created_at
 * @property {string} updated_at
 */

/**
 * @typedef {Object} NotificationChannel
 * @property {number} id
 * @property {string} name
 * @property {string} type - webhook | email | dingtalk | feishu | wechat
 * @property {Object} config - 渠道配置
 * @property {boolean} enabled
 * @property {string} created_at
 * @property {string} updated_at
 */

/**
 * @typedef {Object} OncallShift
 * @property {number} id
 * @property {string} user_id
 * @property {string} user_name
 * @property {string} start_time
 * @property {string} end_time
 * @property {'primary'|'backup'} type
 * @property {string} created_at
 * @property {string} updated_at
 */

/**
 * @typedef {Object} SLAService
 * @property {number} id
 * @property {string} name
 * @property {string} description
 * @property {number} target_percent
 * @property {string} alert_rule_ids - 逗号分隔
 * @property {string} created_at
 * @property {string} updated_at
 */

/**
 * @typedef {Object} SLAReport
 * @property {number} service_id
 * @property {string} service_name
 * @property {number} target_percent
 * @property {number} actual_percent
 * @property {number} total_minutes
 * @property {number} downtime_minutes
 * @property {number} uptime_minutes
 * @property {number} incident_count
 * @property {boolean} met
 * @property {string} start_time
 * @property {string} end_time
 */

/**
 * @typedef {Object} CapacityForecast
 * @property {string} metric
 * @property {string} instance
 * @property {number} current_value
 * @property {number} trend_per_day
 * @property {string|null} predicted_exhaustion_date
 * @property {number} confidence - R^2 值 (0-1)
 * @property {number} data_points
 * @property {number} days_analyzed
 */

/**
 * @typedef {Object} User
 * @property {number} id
 * @property {string} username
 * @property {string} display_name
 * @property {string} email
 * @property {string} role - admin | user | viewer
 * @property {boolean} must_change_pwd
 * @property {string} created_at
 */

/**
 * @typedef {Object} DiagnoseRecord
 * @property {string} id
 * @property {string} target_ip
 * @property {string} trigger - alert | manual
 * @property {string} source
 * @property {string} status - pending | running | done | failed
 * @property {string} alert_title
 * @property {string} result
 * @property {string} create_time
 */

/**
 * @typedef {Object} PaginatedResponse
 * @property {Array} items
 * @property {number} total
 * @property {number} page
 * @property {number} page_size
 */

export {}
