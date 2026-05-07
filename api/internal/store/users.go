package store

import (
	"database/sql"

	"ai-workbench-api/internal/model"
	"time"

	"github.com/sirupsen/logrus"
)

func CreateUser(u *model.User) error {
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now
	if mysqlOK {
		_, err := db.Exec(`INSERT INTO users (id,username,password_hash,role,must_change_pwd,created_at,updated_at) VALUES (?,?,?,?,?,?,?)`,
			u.ID, u.Username, u.PasswordHash, u.Role, u.MustChangePwd, u.CreatedAt, u.UpdatedAt)
		return err
	}
	cp := *u
	mu.Lock()
	oldUser, hadOld := users[u.Username]
	users[u.Username] = &cp
	mu.Unlock()
	if err := persistFallbackSnapshot(); err != nil {
		mu.Lock()
		if hadOld {
			users[u.Username] = oldUser
		} else {
			delete(users, u.Username)
		}
		mu.Unlock()
		return err
	}
	return nil
}

func GetUserByUsername(username string) *model.User {
	if !mysqlOK {
		mu.RLock()
		defer mu.RUnlock()
		u := users[username]
		if u == nil {
			return nil
		}
		cp := *u
		return &cp
	}
	var u model.User
	err := db.QueryRow(`SELECT id,username,password_hash,role,must_change_pwd,created_at,updated_at FROM users WHERE username=?`, username).
		Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.MustChangePwd, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil
	}
	return &u
}

func UpdateUserPassword(id, hash string) error {
	if mysqlOK {
		result, err := db.Exec(`UPDATE users SET password_hash=?,must_change_pwd=0,updated_at=? WHERE id=?`, hash, time.Now(), id)
		if err != nil {
			return err
		}
		if rows, err := result.RowsAffected(); err == nil && rows == 0 {
			return sql.ErrNoRows
		}
		return nil
	}
	now := time.Now()
	mu.Lock()
	var username string
	var oldUser model.User
	for _, u := range users {
		if u.ID == id {
			username = u.Username
			oldUser = *u
			u.PasswordHash = hash
			u.MustChangePwd = false
			u.UpdatedAt = now
			break
		}
	}
	mu.Unlock()
	if username == "" {
		return sql.ErrNoRows
	}
	if err := persistFallbackSnapshot(); err != nil {
		mu.Lock()
		users[username] = &oldUser
		mu.Unlock()
		return err
	}
	return nil
}

func UserCount() int {
	if !mysqlOK {
		mu.RLock()
		defer mu.RUnlock()
		return len(users)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		logrus.Warnf("user count: %v", err)
		return 0
	}
	return count
}
