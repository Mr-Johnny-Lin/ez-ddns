// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package db

import (
	"context"
	"database/sql"
	"fmt"

	"ez-ddns/model"
	"ez-ddns/utils"
)

type DomainDB struct {
	db *sql.DB
}

func NewDomainDB(db *sql.DB) *DomainDB {
	return &DomainDB{db: db}
}

func (d *DomainDB) ReadByConfigID(ctx context.Context, configID string) ([]model.DomainConfig, error) {
	executor := utils.GetExecutor(ctx, d.db)

	rows, err := executor.QueryContext(ctx, `
		SELECT id, domain_name, rr, ip_type, last_ip
		FROM domains
		WHERE config_id = ?
	`, configID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []model.DomainConfig
	for rows.Next() {
		var domain model.DomainConfig
		var ipType string
		err := rows.Scan(
			&domain.ID,
			&domain.DomainName,
			&domain.RR,
			&ipType,
			&domain.LastIP,
		)
		if err != nil {
			return nil, err
		}
		domain.IPType = model.ParseIPType(ipType)
		domains = append(domains, domain)
	}

	return domains, nil
}

func (d *DomainDB) Update(ctx context.Context, domain *model.DomainConfig) error {
	executor := utils.GetExecutor(ctx, d.db)
	_, err := executor.ExecContext(ctx, `
		UPDATE domains
		SET domain_name = ?, rr = ?, ip_type = ?, last_ip = ?
		WHERE id = ?
	`, domain.DomainName, domain.RR, domain.IPType.String(), domain.LastIP, domain.ID)
	return err
}

func (d *DomainDB) Create(ctx context.Context, domain *model.DomainConfig) error {
	executor := utils.GetExecutor(ctx, d.db)
	result, err := executor.ExecContext(ctx, `
		INSERT INTO domains (config_id, domain_name, rr, ip_type, last_ip)
		VALUES (?, ?, ?, ?, ?)
	`, domain.ConfigID, domain.DomainName, domain.RR, domain.IPType.String(), domain.LastIP)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	domain.ID = fmt.Sprintf("%d", id)

	return nil
}

func (d *DomainDB) Delete(ctx context.Context, id string) error {
	executor := utils.GetExecutor(ctx, d.db)
	_, err := executor.ExecContext(ctx, `DELETE FROM domains WHERE id = ?`, id)
	return err
}

func (d *DomainDB) DeleteByConfigID(ctx context.Context, configID string) error {
	executor := utils.GetExecutor(ctx, d.db)
	_, err := executor.ExecContext(ctx, `DELETE FROM domains WHERE config_id = ?`, configID)
	return err
}
