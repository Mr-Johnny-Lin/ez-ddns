// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package db

import (
	"context"
	"database/sql"

	"ez-ddns/model"
	"ez-ddns/utils"
)

type ConfigDB struct {
	db *sql.DB
}

func NewConfigDB(db *sql.DB) *ConfigDB {
	return &ConfigDB{db: db}
}

func (c *ConfigDB) ReadByID(ctx context.Context, id string) (*model.Config, error) {
	var config model.Config
	var provider string

	executor := utils.GetExecutor(ctx, c.db)

	err := executor.QueryRowContext(ctx, `
		SELECT id, access_key_id, access_key_secret, provider, interval
		FROM configs
		WHERE id = ?
	`, id).Scan(
		&config.ID,
		&config.AccessKeyId,
		&config.AccessKeySecret,
		&provider,
		&config.Interval,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	config.Provider = model.ParseProviderType(provider)

	return &config, nil
}

func (c *ConfigDB) ReadAll(ctx context.Context) ([]model.Config, error) {
	executor := utils.GetExecutor(ctx, c.db)

	rows, err := executor.QueryContext(ctx, `
		SELECT id, access_key_id, access_key_secret, provider, interval
		FROM configs
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []model.Config
	for rows.Next() {
		var config model.Config
		var provider string

		err := rows.Scan(
			&config.ID,
			&config.AccessKeyId,
			&config.AccessKeySecret,
			&provider,
			&config.Interval,
		)
		if err != nil {
			return nil, err
		}

		config.Provider = model.ParseProviderType(provider)
		configs = append(configs, config)
	}

	return configs, nil
}

func (c *ConfigDB) Create(ctx context.Context, config *model.Config) error {
	executor := utils.GetExecutor(ctx, c.db)
	_, err := executor.ExecContext(ctx, `
		INSERT INTO configs (id, access_key_id, access_key_secret, provider, interval)
		VALUES (?, ?, ?, ?, ?)
	`, config.ID, config.AccessKeyId, config.AccessKeySecret, config.Provider.String(), config.Interval)
	return err
}

func (c *ConfigDB) Update(ctx context.Context, config *model.Config) error {
	executor := utils.GetExecutor(ctx, c.db)
	_, err := executor.ExecContext(ctx, `
		UPDATE configs
		SET access_key_id = ?, access_key_secret = ?, provider = ?, interval = ?
		WHERE id = ?
	`, config.AccessKeyId, config.AccessKeySecret, config.Provider.String(), config.Interval, config.ID)
	return err
}

func (c *ConfigDB) Delete(ctx context.Context, id string) error {
	executor := utils.GetExecutor(ctx, c.db)
	_, err := executor.ExecContext(ctx, `DELETE FROM configs WHERE id = ?`, id)
	return err
}
