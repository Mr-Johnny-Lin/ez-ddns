// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package main

import (
	"database/sql"
	"fmt"

	"ez-ddns/dao"

	_ "github.com/glebarez/sqlite"
)

const dbFilePath = "ez-ddns.db"

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbFilePath)
	if err != nil {
		return nil, err
	}

	createTablesSQL := `
	CREATE TABLE IF NOT EXISTS config (
		id TEXT PRIMARY KEY,
		access_key_id TEXT,
		access_key_secret TEXT,
		provider TEXT,
		interval INTEGER
	);
	CREATE TABLE IF NOT EXISTS domains (
		id TEXT PRIMARY KEY,
		config_id TEXT,
		domain_name TEXT,
		rr TEXT,
		ip_type TEXT,
		last_ip TEXT,
		FOREIGN KEY (config_id) REFERENCES config(id)
	);
	CREATE TABLE IF NOT EXISTS system_settings (
		id TEXT PRIMARY KEY,
		data JSON NOT NULL,
		created_at TEXT DEFAULT (strftime('%Y-%m-%d %H:%M:%S', 'now')),
		updated_at TEXT DEFAULT (strftime('%Y-%m-%d %H:%M:%S', 'now'))
	);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_domains_unique ON domains (domain_name, rr, ip_type);
	`

	_, err = db.Exec(createTablesSQL)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func main() {
	db, err := initDB()
	if err != nil {
		fmt.Printf("初始化数据库失败: %v\n", err)
		return
	}
	defer db.Close()

	dao.Init(db)

	command, subcommand, args := ParseArgs()
	if command == "" {
		PrintHelp()
		return
	}

	cliHandler := NewCLIHandler(dao.Config, dao.Domain, dao.System)
	if err := cliHandler.HandleCommand(command, subcommand, args); err != nil {
		fmt.Printf("操作失败: %v\n", err)
	}
}