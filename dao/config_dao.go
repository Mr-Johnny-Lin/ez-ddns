// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package dao

import (
	"database/sql"

	"ez-ddns/model"
)

type ConfigDAO struct {
	db *sql.DB
}

func NewConfigDAO(db *sql.DB) *ConfigDAO {
	return &ConfigDAO{db: db}
}

func (dao *ConfigDAO) Create(config *model.Config) error {
	tx, err := dao.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		"INSERT INTO config (id, access_key_id, access_key_secret, provider, interval) VALUES (?, ?, ?, ?, ?)",
		config.ID,
		config.AccessKeyId,
		config.AccessKeySecret,
		config.Provider,
		config.Interval,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, domain := range config.Domains {
		if err := Domain.CreateInTx(tx, config.ID, &domain); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (dao *ConfigDAO) ReadByID(id string) (*model.Config, error) {
	row := dao.db.QueryRow("SELECT id, access_key_id, access_key_secret, provider, interval FROM config WHERE id = ?", id)
	return scanConfig(row)
}

func (dao *ConfigDAO) ReadAll() ([]*model.Config, error) {
	rows, err := dao.db.Query("SELECT id, access_key_id, access_key_secret, provider, interval FROM config")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*model.Config
	for rows.Next() {
		config, err := scanConfigRow(rows)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}

	return configs, nil
}

func (dao *ConfigDAO) Update(config *model.Config) error {
	tx, err := dao.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		"UPDATE config SET access_key_id = ?, access_key_secret = ?, provider = ?, interval = ? WHERE id = ?",
		config.AccessKeyId,
		config.AccessKeySecret,
		config.Provider,
		config.Interval,
		config.ID,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("DELETE FROM domains WHERE config_id = ?", config.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, domain := range config.Domains {
		if err := Domain.CreateInTx(tx, config.ID, &domain); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (dao *ConfigDAO) Delete(id string) error {
	tx, err := dao.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM domains WHERE config_id = ?", id)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("DELETE FROM config WHERE id = ?", id)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func scanConfig(row *sql.Row) (*model.Config, error) {
	var id string
	var accessKeyId, accessKeySecret, provider sql.NullString
	var interval sql.NullInt64

	err := row.Scan(&id, &accessKeyId, &accessKeySecret, &provider, &interval)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return buildConfig(id, accessKeyId, accessKeySecret, provider, interval), nil
}

func scanConfigRow(rows *sql.Rows) (*model.Config, error) {
	var id string
	var accessKeyId, accessKeySecret, provider sql.NullString
	var interval sql.NullInt64

	if err := rows.Scan(&id, &accessKeyId, &accessKeySecret, &provider, &interval); err != nil {
		return nil, err
	}

	return buildConfig(id, accessKeyId, accessKeySecret, provider, interval), nil
}

func buildConfig(id string, accessKeyId, accessKeySecret, provider sql.NullString, interval sql.NullInt64) *model.Config {
	config := &model.Config{
		ID:       id,
		Provider: model.ProviderTypeAliyun,
		Interval: 180,
	}

	if accessKeyId.Valid {
		config.AccessKeyId = accessKeyId.String
	}
	if accessKeySecret.Valid {
		config.AccessKeySecret = accessKeySecret.String
	}
	if provider.Valid {
		config.Provider = model.ParseProviderType(provider.String)
	}
	if interval.Valid {
		config.Interval = int(interval.Int64)
	}

	return config
}