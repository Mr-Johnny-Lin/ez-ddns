// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package dao

import (
	"database/sql"

	"ez-ddns/model"
)

type DomainDAO struct {
	db *sql.DB
}

func NewDomainDAO(db *sql.DB) *DomainDAO {
	return &DomainDAO{db: db}
}

func (dao *DomainDAO) Create(configID string, domain *model.DomainConfig) error {
	_, err := dao.db.Exec(
		"INSERT INTO domains (id, config_id, domain_name, rr, ip_type, last_ip) VALUES (?, ?, ?, ?, ?, ?)",
		domain.ID,
		configID,
		domain.DomainName,
		domain.RR,
		domain.IPType,
		domain.LastIP,
	)
	return err
}

func (dao *DomainDAO) CreateInTx(tx *sql.Tx, configID string, domain *model.DomainConfig) error {
	if tx != nil {
		_, err := tx.Exec(
			"INSERT INTO domains (id, config_id, domain_name, rr, ip_type, last_ip) VALUES (?, ?, ?, ?, ?, ?)",
			domain.ID,
			configID,
			domain.DomainName,
			domain.RR,
			domain.IPType,
			domain.LastIP,
		)
		return err
	}
	return dao.Create(configID, domain)
}

func (dao *DomainDAO) ReadByConfigID(configID string) ([]model.DomainConfig, error) {
	rows, err := dao.db.Query("SELECT id, domain_name, rr, ip_type, last_ip FROM domains WHERE config_id = ?", configID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []model.DomainConfig
	for rows.Next() {
		domain, err := scanDomainRow(rows)
		if err != nil {
			return nil, err
		}
		domains = append(domains, domain)
	}

	return domains, nil
}

func (dao *DomainDAO) ReadByID(id string) (*model.DomainConfig, error) {
	row := dao.db.QueryRow("SELECT id, domain_name, rr, ip_type, last_ip FROM domains WHERE id = ?", id)
	return scanDomain(row)
}

func (dao *DomainDAO) ReadByUniqueKey(domainName, rr, ipType string) (*model.DomainConfig, error) {
	row := dao.db.QueryRow("SELECT id, domain_name, rr, ip_type, last_ip FROM domains WHERE domain_name = ? AND rr = ? AND ip_type = ?", domainName, rr, ipType)
	return scanDomain(row)
}

func (dao *DomainDAO) UpdateLastIP(domainName, rr, ipType, lastIP string) error {
	_, err := dao.db.Exec("UPDATE domains SET last_ip = ? WHERE domain_name = ? AND rr = ? AND ip_type = ?", lastIP, domainName, rr, ipType)
	return err
}

func (dao *DomainDAO) Update(domain *model.DomainConfig) error {
	_, err := dao.db.Exec(
		"UPDATE domains SET domain_name = ?, rr = ?, ip_type = ?, last_ip = ? WHERE id = ?",
		domain.DomainName,
		domain.RR,
		domain.IPType,
		domain.LastIP,
		domain.ID,
	)
	return err
}

func (dao *DomainDAO) Delete(id string) error {
	_, err := dao.db.Exec("DELETE FROM domains WHERE id = ?", id)
	return err
}

func (dao *DomainDAO) DeleteByConfigID(configID string) error {
	_, err := dao.db.Exec("DELETE FROM domains WHERE config_id = ?", configID)
	return err
}

func scanDomain(row *sql.Row) (*model.DomainConfig, error) {
	var id string
	var domainName, rr, ipType, lastIP sql.NullString

	err := row.Scan(&id, &domainName, &rr, &ipType, &lastIP)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return buildDomain(id, domainName, rr, ipType, lastIP), nil
}

func scanDomainRow(rows *sql.Rows) (model.DomainConfig, error) {
	var id string
	var domainName, rr, ipType, lastIP sql.NullString

	if err := rows.Scan(&id, &domainName, &rr, &ipType, &lastIP); err != nil {
		return model.DomainConfig{}, err
	}

	return *buildDomain(id, domainName, rr, ipType, lastIP), nil
}

func buildDomain(id string, domainName, rr, ipType, lastIP sql.NullString) *model.DomainConfig {
	domain := &model.DomainConfig{
		ID:         id,
		DomainName: domainName.String,
		RR:         rr.String,
		IPType:     model.ParseIPType(ipType.String),
	}
	if lastIP.Valid {
		domain.LastIP = lastIP.String
	}
	return domain
}