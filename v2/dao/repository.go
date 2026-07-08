// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package dao

import (
	"context"

	"ez-ddns/v2/model"
)

type DomainRepository interface {
	ReadByConfigID(ctx context.Context, configID string) ([]model.DomainConfig, error)
	Update(ctx context.Context, domain *model.DomainConfig) error
	Create(ctx context.Context, domain *model.DomainConfig) error
	Delete(ctx context.Context, id string) error
	DeleteByConfigID(ctx context.Context, configID string) error
}

type ConfigRepository interface {
	ReadByID(ctx context.Context, id string) (*model.Config, error)
	ReadAll(ctx context.Context) ([]model.Config, error)
	Create(ctx context.Context, config *model.Config) error
	Update(ctx context.Context, config *model.Config) error
	Delete(ctx context.Context, id string) error
}

type SystemRepository interface {
	Read(ctx context.Context) (*model.SystemConfig, error)
	Update(ctx context.Context, config *model.SystemConfig) error
}
