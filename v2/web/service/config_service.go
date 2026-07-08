// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package service

import (
	"context"
	"database/sql"

	"ez-ddns/v2/dao"
	"ez-ddns/v2/model"
	"ez-ddns/v2/utils"
	"ez-ddns/v2/web/dto"
)

type ConfigService struct {
	configRepo dao.ConfigRepository
	domainRepo dao.DomainRepository
	db         *sql.DB
}

func NewConfigService(configRepo dao.ConfigRepository, domainRepo dao.DomainRepository, db *sql.DB) *ConfigService {
	return &ConfigService{
		configRepo: configRepo,
		domainRepo: domainRepo,
		db:         db,
	}
}

func (s *ConfigService) CreateConfigWithDomains(ctx context.Context, configDTO *dto.ConfigDTO) (*dto.ConfigDTO, error) {
	newCtx, _, commitOrRollback, err := utils.BeginTxWithPropagation(ctx, s.db)
	if err != nil {
		return nil, err
	}
	ctx = newCtx

	config := &model.Config{
		ID:              configDTO.ID,
		AccessKeyId:     configDTO.AccessKeyId,
		AccessKeySecret: configDTO.AccessKeySecret,
		Provider:        model.ParseProviderType(configDTO.Provider),
		Interval:        configDTO.Interval,
	}

	if err := s.configRepo.Create(ctx, config); err != nil {
		commitOrRollback(err)
		return nil, err
	}

	domainModels := make([]*model.DomainConfig, len(configDTO.Domains))
	for i, d := range configDTO.Domains {
		domainModels[i] = &model.DomainConfig{
			ID:         d.ID,
			ConfigID:   config.ID,
			DomainName: d.DomainName,
			RR:         d.RR,
			IPType:     model.ParseIPType(d.IPType),
		}
		if err := s.domainRepo.Create(ctx, domainModels[i]); err != nil {
			commitOrRollback(err)
			return nil, err
		}
	}

	if err := commitOrRollback(nil); err != nil {
		return nil, err
	}

	domainDTOs := make([]*dto.DomainDTO, len(domainModels))
	for i, d := range domainModels {
		domainDTOs[i] = &dto.DomainDTO{
			ID:         d.ID,
			ConfigID:   d.ConfigID,
			DomainName: d.DomainName,
			RR:         d.RR,
			IPType:     d.IPType.String(),
		}
	}

	return &dto.ConfigDTO{
		ID:              config.ID,
		AccessKeyId:     config.AccessKeyId,
		AccessKeySecret: config.AccessKeySecret,
		Provider:        config.Provider.String(),
		Interval:        config.Interval,
		Domains:         domainDTOs,
	}, nil
}

func (s *ConfigService) UpdateConfigWithDomains(ctx context.Context, configDTO *dto.ConfigDTO) (*dto.ConfigDTO, error) {
	newCtx, _, commitOrRollback, err := utils.BeginTxWithPropagation(ctx, s.db)
	if err != nil {
		return nil, err
	}
	ctx = newCtx

	config := &model.Config{
		ID:              configDTO.ID,
		AccessKeyId:     configDTO.AccessKeyId,
		AccessKeySecret: configDTO.AccessKeySecret,
		Provider:        model.ParseProviderType(configDTO.Provider),
		Interval:        configDTO.Interval,
	}

	if err := s.configRepo.Update(ctx, config); err != nil {
		commitOrRollback(err)
		return nil, err
	}

	if err := s.domainRepo.DeleteByConfigID(ctx, config.ID); err != nil {
		commitOrRollback(err)
		return nil, err
	}

	domainModels := make([]*model.DomainConfig, len(configDTO.Domains))
	for i, d := range configDTO.Domains {
		domainModels[i] = &model.DomainConfig{
			ID:         d.ID,
			ConfigID:   config.ID,
			DomainName: d.DomainName,
			RR:         d.RR,
			IPType:     model.ParseIPType(d.IPType),
		}
		if err := s.domainRepo.Create(ctx, domainModels[i]); err != nil {
			commitOrRollback(err)
			return nil, err
		}
	}

	if err := commitOrRollback(nil); err != nil {
		return nil, err
	}

	domainDTOs := make([]*dto.DomainDTO, len(domainModels))
	for i, d := range domainModels {
		domainDTOs[i] = &dto.DomainDTO{
			ID:         d.ID,
			ConfigID:   d.ConfigID,
			DomainName: d.DomainName,
			RR:         d.RR,
			IPType:     d.IPType.String(),
		}
	}

	return &dto.ConfigDTO{
		ID:              config.ID,
		AccessKeyId:     config.AccessKeyId,
		AccessKeySecret: config.AccessKeySecret,
		Provider:        config.Provider.String(),
		Interval:        config.Interval,
		Domains:         domainDTOs,
	}, nil
}

func (s *ConfigService) DeleteConfigWithDomains(ctx context.Context, configID string) error {
	return s.configRepo.Delete(ctx, configID)
}

func (s *ConfigService) GetConfigWithDomains(ctx context.Context, configID string) (*dto.ConfigDTO, error) {
	config, err := s.configRepo.ReadByID(ctx, configID)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, nil
	}

	domains, err := s.domainRepo.ReadByConfigID(ctx, configID)
	if err != nil {
		return nil, err
	}

	domainDTOs := make([]*dto.DomainDTO, len(domains))
	for i, d := range domains {
		domainDTOs[i] = &dto.DomainDTO{
			ID:         d.ID,
			ConfigID:   d.ConfigID,
			DomainName: d.DomainName,
			RR:         d.RR,
			IPType:     d.IPType.String(),
		}
	}

	return &dto.ConfigDTO{
		ID:              config.ID,
		AccessKeyId:     config.AccessKeyId,
		AccessKeySecret: config.AccessKeySecret,
		Provider:        config.Provider.String(),
		Interval:        config.Interval,
		Domains:         domainDTOs,
	}, nil
}

func (s *ConfigService) GetAllConfigs(ctx context.Context) ([]dto.ConfigDTO, error) {
	configs, err := s.configRepo.ReadAll(ctx)
	if err != nil {
		return nil, err
	}

	configDTOs := make([]dto.ConfigDTO, len(configs))
	for i, config := range configs {
		configDTOs[i] = dto.ConfigDTO{
			ID:              config.ID,
			AccessKeyId:     config.AccessKeyId,
			AccessKeySecret: config.AccessKeySecret,
			Provider:        config.Provider.String(),
			Interval:        config.Interval,
		}
	}

	return configDTOs, nil
}
