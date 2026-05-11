package store

import (
	"database/sql"
	"sort"
	"time"

	"ai-workbench-api/internal/model"

	"github.com/sirupsen/logrus"
)

var (
	cmdbCategories []model.CmdbCategory
	cmdbObjects    []model.CmdbObject
	cmdbAttributes []model.CmdbAttribute
	cmdbInstances  []model.CmdbInstance
)

func migrateCmdb() {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS cmdb_categories (id VARCHAR(64) PRIMARY KEY,label VARCHAR(64) NOT NULL,parent_id VARCHAR(32),sort INT DEFAULT 0) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS cmdb_objects (id VARCHAR(64) PRIMARY KEY,name VARCHAR(64) NOT NULL,category_id VARCHAR(32),object_type INT DEFAULT 101,icon VARCHAR(32),created_at DATETIME,updated_at DATETIME,INDEX idx_cmdb_obj_cat(category_id)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		"CREATE TABLE IF NOT EXISTS cmdb_attributes (id VARCHAR(64) PRIMARY KEY,object_id VARCHAR(64) NOT NULL,label VARCHAR(64) NOT NULL,attr VARCHAR(64) NOT NULL,value_type VARCHAR(16) NOT NULL,tag VARCHAR(32),required TINYINT(1) DEFAULT 0,`unique` TINYINT(1) DEFAULT 0,discovery TINYINT(1) DEFAULT 0,sort INT DEFAULT 0,unit VARCHAR(16),options TEXT,default_val VARCHAR(256),INDEX idx_cmdb_attr_obj(object_id)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4",
		`CREATE TABLE IF NOT EXISTS cmdb_instances (id VARCHAR(64) PRIMARY KEY,object_id VARCHAR(64) NOT NULL,data MEDIUMTEXT,creator VARCHAR(64),updater VARCHAR(64),created_at DATETIME,updated_at DATETIME,INDEX idx_cmdb_inst_obj(object_id)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS cmdb_relation_types (id VARCHAR(64) PRIMARY KEY,name VARCHAR(32) NOT NULL,label VARCHAR(64)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS cmdb_instance_relations (id VARCHAR(64) PRIMARY KEY,source_instance_id VARCHAR(64) NOT NULL,target_instance_id VARCHAR(64) NOT NULL,relation_type_id VARCHAR(64) NOT NULL,created_at DATETIME,INDEX idx_cmdb_rel_src(source_instance_id),INDEX idx_cmdb_rel_tgt(target_instance_id)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			logrus.WithError(err).Warn("cmdb migrate failed")
		}
	}
}
func ListCmdbCategories() []model.CmdbCategory {
	if mysqlOK {
		rows, err := db.Query(`SELECT id,label,parent_id,sort FROM cmdb_categories ORDER BY sort`)
		if err == nil {
			defer rows.Close()
			var out []model.CmdbCategory
			for rows.Next() {
				var c model.CmdbCategory
				if rows.Scan(&c.ID, &c.Label, &c.ParentID, &c.Sort) == nil {
					out = append(out, c)
				}
			}
			return out
		}
		logrus.WithError(err).Warn("cmdb categories query failed")
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.CmdbCategory, len(cmdbCategories))
	copy(out, cmdbCategories)
	sort.Slice(out, func(i, j int) bool { return out[i].Sort < out[j].Sort })
	return out
}
func CreateCmdbCategory(cat *model.CmdbCategory) error {
	cat.ID = NewID()
	if mysqlOK {
		_, err := db.Exec(`INSERT INTO cmdb_categories (id,label,parent_id,sort) VALUES (?,?,?,?)`,
			cat.ID, cat.Label, cat.ParentID, cat.Sort)
		if err != nil {
			return err
		}
	}
	mu.Lock()
	cmdbCategories = append(cmdbCategories, *cat)
	mu.Unlock()
	return nil
}
func ListCmdbObjects(categoryID string) []model.CmdbObject {
	if mysqlOK {
		q := `SELECT id,name,category_id,object_type,icon,created_at,updated_at FROM cmdb_objects`
		var rows *sql.Rows
		var err error
		if categoryID != "" {
			rows, err = db.Query(q+` WHERE category_id=? ORDER BY updated_at DESC`, categoryID)
		} else {
			rows, err = db.Query(q + ` ORDER BY updated_at DESC`)
		}
		if err == nil {
			defer rows.Close()
			var out []model.CmdbObject
			for rows.Next() {
				var o model.CmdbObject
				if rows.Scan(&o.ID, &o.Name, &o.CategoryID, &o.ObjectType, &o.Icon, &o.CreatedAt, &o.UpdatedAt) == nil {
					out = append(out, o)
				}
			}
			return out
		}
		logrus.WithError(err).Warn("cmdb objects query failed")
	}
	mu.RLock()
	defer mu.RUnlock()
	var out []model.CmdbObject
	for _, o := range cmdbObjects {
		if categoryID == "" || o.CategoryID == categoryID {
			out = append(out, o)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out
}
func GetCmdbObject(id string) (*model.CmdbObject, bool) {
	if mysqlOK {
		var o model.CmdbObject
		err := db.QueryRow(`SELECT id,name,category_id,object_type,icon,created_at,updated_at FROM cmdb_objects WHERE id=?`, id).
			Scan(&o.ID, &o.Name, &o.CategoryID, &o.ObjectType, &o.Icon, &o.CreatedAt, &o.UpdatedAt)
		if err == nil {
			return &o, true
		}
		if err != sql.ErrNoRows {
			logrus.WithError(err).Warn("cmdb object get failed")
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	for i := range cmdbObjects {
		if cmdbObjects[i].ID == id {
			cp := cmdbObjects[i]
			return &cp, true
		}
	}
	return nil, false
}
func CreateCmdbObject(obj *model.CmdbObject) error {
	obj.ID = NewID()
	now := time.Now()
	obj.CreatedAt = now
	obj.UpdatedAt = now
	if mysqlOK {
		_, err := db.Exec(`INSERT INTO cmdb_objects (id,name,category_id,object_type,icon,created_at,updated_at) VALUES (?,?,?,?,?,?,?)`,
			obj.ID, obj.Name, obj.CategoryID, obj.ObjectType, obj.Icon, obj.CreatedAt, obj.UpdatedAt)
		if err != nil {
			return err
		}
	}
	mu.Lock()
	cmdbObjects = append(cmdbObjects, *obj)
	mu.Unlock()
	return nil
}
func UpdateCmdbObject(obj *model.CmdbObject) error {
	obj.UpdatedAt = time.Now()
	if mysqlOK {
		_, err := db.Exec(`UPDATE cmdb_objects SET name=?,category_id=?,object_type=?,icon=?,updated_at=? WHERE id=?`,
			obj.Name, obj.CategoryID, obj.ObjectType, obj.Icon, obj.UpdatedAt, obj.ID)
		if err != nil {
			return err
		}
	}
	mu.Lock()
	for i := range cmdbObjects {
		if cmdbObjects[i].ID == obj.ID {
			cmdbObjects[i] = *obj
			break
		}
	}
	mu.Unlock()
	return nil
}
func DeleteCmdbObject(id string) error {
	if mysqlOK {
		if _, err := db.Exec(`DELETE FROM cmdb_objects WHERE id=?`, id); err != nil {
			return err
		}
	}
	mu.Lock()
	for i := range cmdbObjects {
		if cmdbObjects[i].ID == id {
			cmdbObjects = append(cmdbObjects[:i], cmdbObjects[i+1:]...)
			break
		}
	}
	mu.Unlock()
	return nil
}
func ListCmdbAttributes(objectID string) []model.CmdbAttribute {
	if mysqlOK {
		rows, err := db.Query("SELECT id,object_id,label,attr,value_type,tag,required,`unique`,discovery,sort,unit,options,default_val FROM cmdb_attributes WHERE object_id=? ORDER BY sort", objectID)
		if err == nil {
			defer rows.Close()
			var out []model.CmdbAttribute
			for rows.Next() {
				var a model.CmdbAttribute
				if rows.Scan(&a.ID, &a.ObjectID, &a.Label, &a.Attr, &a.ValueType, &a.Tag, &a.Required, &a.Unique, &a.Discovery, &a.Sort, &a.Unit, &a.Options, &a.DefaultVal) == nil {
					out = append(out, a)
				}
			}
			return out
		}
		logrus.WithError(err).Warn("cmdb attributes query failed")
	}
	mu.RLock()
	defer mu.RUnlock()
	var out []model.CmdbAttribute
	for _, a := range cmdbAttributes {
		if a.ObjectID == objectID {
			out = append(out, a)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Sort < out[j].Sort })
	return out
}
func CreateCmdbAttribute(attr *model.CmdbAttribute) error {
	attr.ID = NewID()
	if mysqlOK {
		_, err := db.Exec("INSERT INTO cmdb_attributes (id,object_id,label,attr,value_type,tag,required,`unique`,discovery,sort,unit,options,default_val) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)",
			attr.ID, attr.ObjectID, attr.Label, attr.Attr, attr.ValueType, attr.Tag, attr.Required, attr.Unique, attr.Discovery, attr.Sort, attr.Unit, attr.Options, attr.DefaultVal)
		if err != nil {
			return err
		}
	}
	mu.Lock()
	cmdbAttributes = append(cmdbAttributes, *attr)
	mu.Unlock()
	return nil
}
func UpdateCmdbAttribute(attr *model.CmdbAttribute) error {
	if mysqlOK {
		_, err := db.Exec("UPDATE cmdb_attributes SET label=?,attr=?,value_type=?,tag=?,required=?,`unique`=?,discovery=?,sort=?,unit=?,options=?,default_val=? WHERE id=?",
			attr.Label, attr.Attr, attr.ValueType, attr.Tag, attr.Required, attr.Unique, attr.Discovery, attr.Sort, attr.Unit, attr.Options, attr.DefaultVal, attr.ID)
		if err != nil {
			return err
		}
	}
	mu.Lock()
	for i := range cmdbAttributes {
		if cmdbAttributes[i].ID == attr.ID {
			cmdbAttributes[i] = *attr
			break
		}
	}
	mu.Unlock()
	return nil
}
func DeleteCmdbAttribute(id string) error {
	if mysqlOK {
		if _, err := db.Exec(`DELETE FROM cmdb_attributes WHERE id=?`, id); err != nil {
			return err
		}
	}
	mu.Lock()
	for i := range cmdbAttributes {
		if cmdbAttributes[i].ID == id {
			cmdbAttributes = append(cmdbAttributes[:i], cmdbAttributes[i+1:]...)
			break
		}
	}
	mu.Unlock()
	return nil
}
func ListCmdbInstances(objectID string, page, limit int) ([]model.CmdbInstance, int64) {
	if mysqlOK {
		var total int64
		db.QueryRow(`SELECT COUNT(*) FROM cmdb_instances WHERE object_id=?`, objectID).Scan(&total)
		offset := (page - 1) * limit
		rows, err := db.Query(`SELECT id,object_id,data,creator,updater,created_at,updated_at FROM cmdb_instances WHERE object_id=? ORDER BY created_at DESC LIMIT ? OFFSET ?`, objectID, limit, offset)
		if err == nil {
			defer rows.Close()
			var out []model.CmdbInstance
			for rows.Next() {
				var inst model.CmdbInstance
				if rows.Scan(&inst.ID, &inst.ObjectID, &inst.Data, &inst.Creator, &inst.Updater, &inst.CreatedAt, &inst.UpdatedAt) == nil {
					out = append(out, inst)
				}
			}
			return out, total
		}
		logrus.WithError(err).Warn("cmdb instances query failed")
	}
	mu.RLock()
	defer mu.RUnlock()
	var filtered []model.CmdbInstance
	for _, inst := range cmdbInstances {
		if inst.ObjectID == objectID {
			filtered = append(filtered, inst)
		}
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].CreatedAt.After(filtered[j].CreatedAt) })
	total := int64(len(filtered))
	start := (page - 1) * limit
	if start >= len(filtered) {
		return nil, total
	}
	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[start:end], total
}
func GetCmdbInstance(id string) (*model.CmdbInstance, bool) {
	if mysqlOK {
		var inst model.CmdbInstance
		err := db.QueryRow(`SELECT id,object_id,data,creator,updater,created_at,updated_at FROM cmdb_instances WHERE id=?`, id).
			Scan(&inst.ID, &inst.ObjectID, &inst.Data, &inst.Creator, &inst.Updater, &inst.CreatedAt, &inst.UpdatedAt)
		if err == nil {
			return &inst, true
		}
		if err != sql.ErrNoRows {
			logrus.WithError(err).Warn("cmdb instance get failed")
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	for i := range cmdbInstances {
		if cmdbInstances[i].ID == id {
			cp := cmdbInstances[i]
			return &cp, true
		}
	}
	return nil, false
}
func CreateCmdbInstance(inst *model.CmdbInstance) error {
	inst.ID = NewID()
	now := time.Now()
	inst.CreatedAt = now
	inst.UpdatedAt = now
	if mysqlOK {
		_, err := db.Exec(`INSERT INTO cmdb_instances (id,object_id,data,creator,updater,created_at,updated_at) VALUES (?,?,?,?,?,?,?)`,
			inst.ID, inst.ObjectID, inst.Data, inst.Creator, inst.Updater, inst.CreatedAt, inst.UpdatedAt)
		if err != nil {
			return err
		}
	}
	mu.Lock()
	cmdbInstances = append(cmdbInstances, *inst)
	mu.Unlock()
	return nil
}
func UpdateCmdbInstance(inst *model.CmdbInstance) error {
	inst.UpdatedAt = time.Now()
	if mysqlOK {
		_, err := db.Exec(`UPDATE cmdb_instances SET data=?,updater=?,updated_at=? WHERE id=?`,
			inst.Data, inst.Updater, inst.UpdatedAt, inst.ID)
		if err != nil {
			return err
		}
	}
	mu.Lock()
	for i := range cmdbInstances {
		if cmdbInstances[i].ID == inst.ID {
			cmdbInstances[i] = *inst
			break
		}
	}
	mu.Unlock()
	return nil
}
func DeleteCmdbInstance(id string) error {
	if mysqlOK {
		if _, err := db.Exec(`DELETE FROM cmdb_instances WHERE id=?`, id); err != nil {
			return err
		}
	}
	mu.Lock()
	for i := range cmdbInstances {
		if cmdbInstances[i].ID == id {
			cmdbInstances = append(cmdbInstances[:i], cmdbInstances[i+1:]...)
			break
		}
	}
	mu.Unlock()
	return nil
}
func CountCmdbInstancesByObject() map[string]int64 {
	if mysqlOK {
		rows, err := db.Query(`SELECT object_id, COUNT(*) FROM cmdb_instances GROUP BY object_id`)
		if err == nil {
			defer rows.Close()
			out := make(map[string]int64)
			for rows.Next() {
				var oid string
				var cnt int64
				if rows.Scan(&oid, &cnt) == nil {
					out[oid] = cnt
				}
			}
			return out
		}
		logrus.WithError(err).Warn("cmdb count instances failed")
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make(map[string]int64)
	for _, inst := range cmdbInstances {
		out[inst.ObjectID]++
	}
	return out
}
