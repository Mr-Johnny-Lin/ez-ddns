// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package db

import (
	"database/sql"
	"fmt"

	_ "github.com/glebarez/sqlite"
)

func NewSQLiteDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("无法打开数据库: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("无法连接数据库: %v", err)
	}

	err = createTables(db)
	if err != nil {
		return nil, fmt.Errorf("创建表失败: %v", err)
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS configs (
			id TEXT PRIMARY KEY,
			access_key_id TEXT NOT NULL,
			access_key_secret TEXT NOT NULL,
			provider TEXT NOT NULL,
			interval INTEGER NOT NULL DEFAULT 300,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS domains (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			config_id TEXT,
			domain_name TEXT NOT NULL,
			rr TEXT NOT NULL,
			ip_type TEXT NOT NULL,
			last_ip TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (config_id) REFERENCES configs(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS system_settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			setting_key TEXT NOT NULL UNIQUE,
			setting_value TEXT NOT NULL,
			enabled BOOLEAN NOT NULL DEFAULT 1,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	return nil
}
