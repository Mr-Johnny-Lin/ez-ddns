// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package db

import (
	"context"
	"database/sql"

	"ez-ddns/v2/model"
	"ez-ddns/v2/utils"
)

type SystemDB struct {
	db *sql.DB
}

func NewSystemDB(db *sql.DB) *SystemDB {
	return &SystemDB{db: db}
}

func (s *SystemDB) Read(ctx context.Context) (*model.SystemConfig, error) {
	executor := utils.GetExecutor(ctx, s.db)

	rows, err := executor.QueryContext(ctx, `
		SELECT setting_key, setting_value
		FROM system_settings
		WHERE enabled = 1
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	config := &model.SystemConfig{}

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		switch key {
		case "log_level":
			config.LogLevel = value
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return config, nil
}

func (s *SystemDB) Update(ctx context.Context, config *model.SystemConfig) error {
	executor := utils.GetExecutor(ctx, s.db)
	_, err := executor.ExecContext(ctx, `
		INSERT OR REPLACE INTO system_settings (setting_key, setting_value, enabled)
		VALUES ('log_level', ?, 1)
	`, config.LogLevel)
	return err
}
