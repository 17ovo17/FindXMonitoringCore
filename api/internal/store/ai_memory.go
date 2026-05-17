package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"ai-workbench-api/internal/model"
)

const (
	aiMemoryRedisPrefix = "aim:"
	aiMemoryRedisTTL    = 30 * time.Minute
)

func SaveAIMemory(m *model.AIMemory) error {
	if m == nil || m.ID == "" {
		return nil
	}
	if mysqlOK {
		tagsJSON, _ := json.Marshal(m.Tags)
		_, err := db.Exec(`INSERT INTO ai_memories (id,type,content,tags,score,used_count,session_id,created_at,expires_at) VALUES (?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE content=VALUES(content),tags=VALUES(tags),score=VALUES(score),used_count=VALUES(used_count),expires_at=VALUES(expires_at)`,
			m.ID, m.Type, m.Content, string(tagsJSON), m.Score, m.UsedCount, m.SessionID, m.CreatedAt, nullableTime(m.ExpiresAt))
		if err != nil {
			return err
		}
	}
	if redisOK {
		cacheAIMemoryToRedis(m)
	}
	return nil
}

func GetAIMemory(id string) (*model.AIMemory, bool) {
	if redisOK {
		if m, ok := getAIMemoryFromRedis(id); ok {
			return m, true
		}
	}
	if mysqlOK {
		return getAIMemoryFromDB(id)
	}
	return nil, false
}

func ListAIMemories(memType string, limit int) []model.AIMemory {
	if mysqlOK {
		return listAIMemoriesFromDB(memType, limit)
	}
	return nil
}

func SearchAIMemories(query string, limit int) []model.AIMemory {
	if mysqlOK {
		return searchAIMemoriesFromDB(query, limit)
	}
	return nil
}

func DeleteAIMemory(id string) {
	if mysqlOK {
		db.Exec(`DELETE FROM ai_memories WHERE id=?`, id)
	}
	if redisOK {
		redisClient.Del(context.Background(), aiMemoryRedisPrefix+id)
	}
}

func IncrementAIMemoryUsedCount(id string) {
	if mysqlOK {
		db.Exec(`UPDATE ai_memories SET used_count=used_count+1 WHERE id=?`, id)
	}
	if redisOK {
		redisClient.Del(context.Background(), aiMemoryRedisPrefix+id)
	}
}

func cacheAIMemoryToRedis(m *model.AIMemory) {
	data, err := json.Marshal(m)
	if err != nil {
		return
	}
	redisClient.Set(context.Background(), aiMemoryRedisPrefix+m.ID, data, aiMemoryRedisTTL)
}

func getAIMemoryFromRedis(id string) (*model.AIMemory, bool) {
	data, err := redisClient.Get(context.Background(), aiMemoryRedisPrefix+id).Bytes()
	if err != nil {
		return nil, false
	}
	var m model.AIMemory
	if json.Unmarshal(data, &m) != nil {
		return nil, false
	}
	return &m, true
}

func getAIMemoryFromDB(id string) (*model.AIMemory, bool) {
	row := db.QueryRow(`SELECT id,type,content,COALESCE(tags,'[]'),score,used_count,COALESCE(session_id,''),created_at,expires_at FROM ai_memories WHERE id=?`, id)
	return scanAIMemoryRow(row)
}

func listAIMemoriesFromDB(memType string, limit int) []model.AIMemory {
	var rows *sql.Rows
	var err error
	if memType != "" {
		rows, err = db.Query(`SELECT id,type,content,COALESCE(tags,'[]'),score,used_count,COALESCE(session_id,''),created_at,expires_at FROM ai_memories WHERE type=? ORDER BY created_at DESC LIMIT ?`, memType, limit)
	} else {
		rows, err = db.Query(`SELECT id,type,content,COALESCE(tags,'[]'),score,used_count,COALESCE(session_id,''),created_at,expires_at FROM ai_memories ORDER BY created_at DESC LIMIT ?`, limit)
	}
	if err != nil {
		return nil
	}
	defer rows.Close()
	return scanAIMemoryRows(rows)
}

func searchAIMemoriesFromDB(query string, limit int) []model.AIMemory {
	rows, err := db.Query(`SELECT id,type,content,COALESCE(tags,'[]'),score,used_count,COALESCE(session_id,''),created_at,expires_at FROM ai_memories WHERE MATCH(content) AGAINST(? IN NATURAL LANGUAGE MODE) LIMIT ?`, query, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()
	return scanAIMemoryRows(rows)
}

func scanAIMemoryRow(row *sql.Row) (*model.AIMemory, bool) {
	var m model.AIMemory
	var tagsStr string
	var expiresAt sql.NullTime
	if err := row.Scan(&m.ID, &m.Type, &m.Content, &tagsStr, &m.Score, &m.UsedCount, &m.SessionID, &m.CreatedAt, &expiresAt); err != nil {
		return nil, false
	}
	_ = json.Unmarshal([]byte(tagsStr), &m.Tags)
	if expiresAt.Valid {
		m.ExpiresAt = &expiresAt.Time
	}
	return &m, true
}

func scanAIMemoryRows(rows *sql.Rows) []model.AIMemory {
	var out []model.AIMemory
	for rows.Next() {
		var m model.AIMemory
		var tagsStr string
		var expiresAt sql.NullTime
		if err := rows.Scan(&m.ID, &m.Type, &m.Content, &tagsStr, &m.Score, &m.UsedCount, &m.SessionID, &m.CreatedAt, &expiresAt); err != nil {
			continue
		}
		_ = json.Unmarshal([]byte(tagsStr), &m.Tags)
		if expiresAt.Valid {
			m.ExpiresAt = &expiresAt.Time
		}
		out = append(out, m)
	}
	return out
}
