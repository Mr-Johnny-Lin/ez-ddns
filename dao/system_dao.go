// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package dao

import (
	"database/sql"
	"encoding/json"
	"time"
)

type SystemSettingsDAO struct {
	db *sql.DB
}

type SystemSettings struct {
	ID        string                 `json:"id"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

func NewSystemSettingsDAO(db *sql.DB) *SystemSettingsDAO {
	return &SystemSettingsDAO{db: db}
}

func (dao *SystemSettingsDAO) Get(id string) (*SystemSettings, error) {
	var dataStr string
	var createdAt, updatedAt string

	err := dao.db.QueryRow(
		"SELECT data, created_at, updated_at FROM system_settings WHERE id = ?",
		id,
	).Scan(&dataStr, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		return nil, err
	}

	return &SystemSettings{
		ID:        id,
		Data:      data,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func (dao *SystemSettingsDAO) Set(id string, data map[string]interface{}) error {
	dataStr, err := json.Marshal(data)
	if err != nil {
		return err
	}

	now := time.Now().Format("2006-01-02 15:04:05")

	_, err = dao.db.Exec(
		`INSERT OR REPLACE INTO system_settings (id, data, created_at, updated_at)
		VALUES (?, ?, COALESCE((SELECT created_at FROM system_settings WHERE id = ?), ?), ?)`,
		id, string(dataStr), id, now, now,
	)

	return err
}

func (dao *SystemSettingsDAO) GetAll() ([]*SystemSettings, error) {
	rows, err := dao.db.Query("SELECT id, data, created_at, updated_at FROM system_settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settingsList []*SystemSettings
	for rows.Next() {
		var id string
		var dataStr string
		var createdAt, updatedAt string

		if err := rows.Scan(&id, &dataStr, &createdAt, &updatedAt); err != nil {
			return nil, err
		}

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
			return nil, err
		}

		settingsList = append(settingsList, &SystemSettings{
			ID:        id,
			Data:      data,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		})
	}

	return settingsList, nil
}

func (dao *SystemSettingsDAO) Delete(id string) error {
	_, err := dao.db.Exec("DELETE FROM system_settings WHERE id = ?", id)
	return err
}

func (dao *SystemSettingsDAO) UpdateField(id, field string, value interface{}) error {
	settings, err := dao.Get(id)
	if err != nil {
		return err
	}

	if settings == nil {
		settings = &SystemSettings{
			ID:   id,
			Data: make(map[string]interface{}),
		}
	}

	settings.Data[field] = value
	return dao.Set(id, settings.Data)
}