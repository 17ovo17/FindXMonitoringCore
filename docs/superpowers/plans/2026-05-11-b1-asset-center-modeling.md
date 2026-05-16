# B1 资产中心导航合并 + CMDB 对象建模 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将"基础设施"和"Agent 管理中心"合并为统一的"资产中心"导航项，并增加 CMDB 对象建模后端 API（模型分类树、模型 CRUD、属性定义）。

**Architecture:** 前端修改 navigation.js 合并导航项，后端新增 CMDB 对象建模 API（模型、属性、分类），数据存储在 MySQL。UI 风格不变（--fx-* CSS 变量体系）。

**Tech Stack:** React 18 + Go 1.21/Gin + MySQL + 现有 FindX React Shell

---

### Task 1: 后端 — CMDB 对象模型数据结构

**Files:**
- Create: `api/internal/model/cmdb.go`

- [ ] **Step 1: 创建 CMDB 数据模型**

```go
package model

import "time"

// CmdbCategory 模型分类（计算资源、系统软件、应用软件、网络、存储、机房、组织）
type CmdbCategory struct {
	ID       string `json:"id" gorm:"primaryKey;size:32"`
	Label    string `json:"label" gorm:"size:64;not null"`
	ParentID string `json:"parent_id" gorm:"size:32;index"`
	Sort     int    `json:"sort" gorm:"default:0"`
}

// CmdbObject 模型定义（操作系统、数据库、中间件、服务器...）
type CmdbObject struct {
	ID         string `json:"id" gorm:"primaryKey;size:32"`
	Name       string `json:"name" gorm:"size:64;not null"`
	CategoryID string `json:"category_id" gorm:"size:32;index"`
	ObjectType int    `json:"object_type" gorm:"default:101"`
	Icon       string `json:"icon" gorm:"size:32"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CmdbAttribute 模型属性定义
type CmdbAttribute struct {
	ID         string `json:"id" gorm:"primaryKey;size:32"`
	ObjectID   string `json:"object_id" gorm:"size:32;index;not null"`
	Label      string `json:"label" gorm:"size:64;not null"`
	Attr       string `json:"attr" gorm:"size:64;not null"`
	ValueType  string `json:"value_type" gorm:"size:16;not null"` // char, int, float, ip, boolean, enum, array, struct
	Tag        string `json:"tag" gorm:"size:32"`                 // 属性分组标签
	Required   bool   `json:"required" gorm:"default:false"`
	Unique     bool   `json:"unique" gorm:"default:false"`
	Discovery  bool   `json:"discovery" gorm:"default:false"`     // Agent 自动发现可填充
	Sort       int    `json:"sort" gorm:"default:0"`
	Unit       string `json:"unit" gorm:"size:16"`
	Options    string `json:"options" gorm:"type:text"`           // enum 选项 JSON
	DefaultVal string `json:"default_val" gorm:"size:256"`
}

// CmdbInstance 资产实例
type CmdbInstance struct {
	ID         string    `json:"id" gorm:"primaryKey;size:32"`
	ObjectID   string    `json:"object_id" gorm:"size:32;index;not null"`
	Data       string    `json:"data" gorm:"type:mediumtext"`     // JSON 存储属性值
	Creator    string    `json:"creator" gorm:"size:64"`
	Updater    string    `json:"updater" gorm:"size:64"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
```

- [ ] **Step 2: 验证编译**

```bash
ssh findx-ubuntu "cd /opt/ai-workbench/api && go build ./..."
```

Expected: BUILD PASS

- [ ] **Step 3: 提交**

```bash
git add -- api/internal/model/cmdb.go
git commit -m "feat(cmdb): add object modeling data structures"
```

### Task 2: 后端 — CMDB Store 层

**Files:**
- Create: `api/internal/store/cmdb.go`

- [ ] **Step 1: 实现 CMDB 存储层**

```go
package store

import (
	"ai-workbench-api/internal/model"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func migrateCmdb() {
	db := GetDB()
	db.AutoMigrate(&model.CmdbCategory{}, &model.CmdbObject{}, &model.CmdbAttribute{}, &model.CmdbInstance{})
}

func ListCategories() ([]model.CmdbCategory, error) {
	var rows []model.CmdbCategory
	err := GetDB().Order("sort asc").Find(&rows).Error
	return rows, err
}

func ListObjects(categoryID string) ([]model.CmdbObject, error) {
	var rows []model.CmdbObject
	q := GetDB().Order("created_at desc")
	if categoryID != "" {
		q = q.Where("category_id = ?", categoryID)
	}
	err := q.Find(&rows).Error
	return rows, err
}

func GetObject(id string) (*model.CmdbObject, error) {
	var obj model.CmdbObject
	err := GetDB().Where("id = ?", id).First(&obj).Error
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func CreateObject(obj *model.CmdbObject) error {
	if obj.ID == "" {
		obj.ID = uuid.New().String()[:8]
	}
	obj.CreatedAt = time.Now()
	obj.UpdatedAt = time.Now()
	return GetDB().Create(obj).Error
}

func UpdateObject(obj *model.CmdbObject) error {
	obj.UpdatedAt = time.Now()
	return GetDB().Save(obj).Error
}

func DeleteObject(id string) error {
	return GetDB().Where("id = ?", id).Delete(&model.CmdbObject{}).Error
}

func ListAttributes(objectID string) ([]model.CmdbAttribute, error) {
	var rows []model.CmdbAttribute
	err := GetDB().Where("object_id = ?", objectID).Order("sort asc").Find(&rows).Error
	return rows, err
}

func CreateAttribute(attr *model.CmdbAttribute) error {
	if attr.ID == "" {
		attr.ID = uuid.New().String()[:8]
	}
	return GetDB().Create(attr).Error
}

func UpdateAttribute(attr *model.CmdbAttribute) error {
	return GetDB().Save(attr).Error
}

func DeleteAttribute(id string) error {
	return GetDB().Where("id = ?", id).Delete(&model.CmdbAttribute{}).Error
}

func ListInstances(objectID string, page, limit int) ([]model.CmdbInstance, int64, error) {
	var rows []model.CmdbInstance
	var total int64
	q := GetDB().Where("object_id = ?", objectID)
	q.Model(&model.CmdbInstance{}).Count(&total)
	if limit <= 0 {
		limit = 20
	}
	offset := (page - 1) * limit
	err := q.Order("created_at desc").Offset(offset).Limit(limit).Find(&rows).Error
	return rows, total, err
}

func GetInstance(id string) (*model.CmdbInstance, error) {
	var inst model.CmdbInstance
	err := GetDB().Where("id = ?", id).First(&inst).Error
	if err != nil {
		return nil, err
	}
	return &inst, nil
}

func CreateInstance(inst *model.CmdbInstance) error {
	if inst.ID == "" {
		inst.ID = uuid.New().String()[:12]
	}
	inst.CreatedAt = time.Now()
	inst.UpdatedAt = time.Now()
	return GetDB().Create(inst).Error
}

func UpdateInstance(inst *model.CmdbInstance) error {
	inst.UpdatedAt = time.Now()
	return GetDB().Save(inst).Error
}

func DeleteInstance(id string) error {
	return GetDB().Where("id = ?", id).Delete(&model.CmdbInstance{}).Error
}

func CountInstancesByObject() (map[string]int64, error) {
	type result struct {
		ObjectID string
		Count    int64
	}
	var rows []result
	err := GetDB().Model(&model.CmdbInstance{}).Select("object_id, count(*) as count").Group("object_id").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	m := make(map[string]int64, len(rows))
	for _, r := range rows {
		m[r.ObjectID] = r.Count
	}
	return m, nil
}
```

- [ ] **Step 2: 在 store init 中调用 migrateCmdb**

在 `api/internal/store/init.go` 的 `Migrate()` 函数末尾添加 `migrateCmdb()` 调用。

- [ ] **Step 3: 验证编译**

```bash
ssh findx-ubuntu "cd /opt/ai-workbench/api && go build ./..."
```

- [ ] **Step 4: 提交**

```bash
git add -- api/internal/store/cmdb.go api/internal/store/init.go
git commit -m "feat(cmdb): add object modeling store layer"
```

### Task 3: 后端 — CMDB Handler API

**Files:**
- Create: `api/internal/handler/cmdb_model.go`

- [ ] **Step 1: 实现 CMDB 模型管理 API**

路由设计：
- `GET /api/v1/cmdb/tree` — 分类树 + 模型 + 实例计数
- `GET /api/v1/cmdb/objects` — 模型列表
- `POST /api/v1/cmdb/objects` — 创建模型
- `PUT /api/v1/cmdb/objects/:id` — 更新模型
- `DELETE /api/v1/cmdb/objects/:id` — 删除模型
- `GET /api/v1/cmdb/objects/:id/attributes` — 属性列表
- `POST /api/v1/cmdb/objects/:id/attributes` — 创建属性
- `PUT /api/v1/cmdb/attributes/:id` — 更新属性
- `DELETE /api/v1/cmdb/attributes/:id` — 删除属性
- `GET /api/v1/cmdb/objects/:id/instances` — 实例列表
- `POST /api/v1/cmdb/objects/:id/instances` — 创建实例
- `GET /api/v1/cmdb/instances/:id` — 实例详情
- `PUT /api/v1/cmdb/instances/:id` — 更新实例
- `DELETE /api/v1/cmdb/instances/:id` — 删除实例

- [ ] **Step 2: 注册路由到 main.go**

- [ ] **Step 3: 远端验证**

```bash
ssh findx-ubuntu "cd /opt/ai-workbench/api && go test -count=1 ./... && go build -o api-linux ."
```

- [ ] **Step 4: 提交**

```bash
git add -- api/internal/handler/cmdb_model.go api/main.go
git commit -m "feat(cmdb): add object modeling API handlers"
```

### Task 4: 后端 — 预置模型种子数据

**Files:**
- Create: `api/internal/store/cmdb_seed.go`

- [ ] **Step 1: 创建种子数据**

预置分类和模型（参考 LWOPS）：
- 计算资源：服务器、虚拟机、容器(Docker/K8s)、云资源
- 系统软件：操作系统、数据库、中间件
- 应用软件：业务系统
- 网络资源：网络设备
- 存储资源：存储
- 机房资源：机房、机柜

操作系统模型预置属性（30+ 字段，参考 LWOPS）。

- [ ] **Step 2: 验证 + 提交**

```bash
ssh findx-ubuntu "cd /opt/ai-workbench/api && go test -count=1 ./... && go build -o api-linux ."
git add -- api/internal/store/cmdb_seed.go
git commit -m "feat(cmdb): add preset model categories and OS attributes seed"
```

### Task 5: 前端 — 导航合并

**Files:**
- Modify: `web/src/react-shell/navigation.js`

- [ ] **Step 1: 合并"基础设施"和"Agent 管理中心"为"资产中心"**

将 navigation.js 中的 `infrastructure` 和 `agents` 两个 navGroup 合并为一个 `asset-center`：

```javascript
{
  key: 'asset-center',
  label: '资产中心',
  path: '/assets',
  defaultSection: 'overview',
  children: [
    { section: 'overview', label: '资产概览' },
    { section: 'models', label: '对象建模' },
    { section: 'instances', label: '实例管理' },
    { section: 'agents', label: 'Agent 状态' },
    { section: 'topology', label: '拓扑视图' },
  ],
  hiddenChildren: [
    { section: 'model-detail', label: '模型详情' },
    { section: 'instance-detail', label: '实例详情' },
  ],
}
```

删除原来的 `infrastructure` 和 `agents` 两个 group。

- [ ] **Step 2: 本地 build 验证**

```bash
cd web && npm run build
```

- [ ] **Step 3: 提交**

```bash
git add -- web/src/react-shell/navigation.js
git commit -m "feat(nav): merge infrastructure + agents into unified asset center"
```

### Task 6: 前端 — 资产中心对象建模页面

**Files:**
- Create: `web/src/react-shell/cmdb/ModelTreeSection.jsx`
- Create: `web/src/react-shell/cmdb/ModelDetailSection.jsx`
- Modify: `web/src/react-shell/cmdb/AssetsPage.jsx`

- [ ] **Step 1: 创建模型分类树组件**

ModelTreeSection.jsx — 左侧分类树 + 右侧模型卡片列表，调用 `/api/v1/cmdb/tree`。

- [ ] **Step 2: 创建模型详情组件**

ModelDetailSection.jsx — 模型属性列表、属性 CRUD、关联关系。

- [ ] **Step 3: 更新 AssetsPage 路由分发**

在 AssetsPage.jsx 中增加 `models` 和 `model-detail` section 的渲染。

- [ ] **Step 4: build 验证 + 提交**

```bash
cd web && npm run build
git add -- web/src/react-shell/cmdb/ModelTreeSection.jsx web/src/react-shell/cmdb/ModelDetailSection.jsx web/src/react-shell/cmdb/AssetsPage.jsx
git commit -m "feat(cmdb): add object modeling UI - category tree and model detail"
```

### Task 7: 里程碑验证 — 远端部署 + 浏览器

- [ ] **Step 1: 远端部署**

```bash
ssh findx-ubuntu "cd /opt/ai-workbench/api && go build -o api-linux . && sudo install -m 0755 api-linux /opt/ai-workbench-runtime/api/ai-workbench-api && sudo systemctl restart ai-workbench-api.service"
ssh findx-ubuntu "sleep 4 && curl -fsS http://127.0.0.1:8080/api/v1/health/storage"
```

- [ ] **Step 2: 前端部署**

同步 web build 产物到远端，验证 http://10.10.160.202:3000 可访问。

- [ ] **Step 3: Playwright 浏览器验证**

- 登录
- 导航到"资产中心"
- 确认导航项已合并（无"基础设施"和"Agent 管理中心"分离项）
- 点击"对象建模"，确认分类树加载
- 确认 UI 风格一致（--fx-* 变量、毛玻璃阴影、深浅主题）

- [ ] **Step 4: 确认提交历史**

```bash
git log --oneline -7
```
