// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package caching

import (
	"context"
	"sync"

	"ez-ddns/dao"
	"ez-ddns/model"
)

type SystemCaching struct {
	cache dao.SystemRepository
	db    dao.SystemRepository
	mu    sync.RWMutex
}

func NewSystemCaching(cache dao.SystemRepository, db dao.SystemRepository) *SystemCaching {
	return &SystemCaching{
		cache: cache,
		db:    db,
	}
}

func (p *SystemCaching) Read(ctx context.Context) (*model.SystemConfig, error) {
	p.mu.RLock()
	cached, err := p.cache.Read(ctx)
	p.mu.RUnlock()
	if err != nil {
		return nil, err
	}

	if cached != nil {
		return cached, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	cached, err = p.cache.Read(ctx)
	if err != nil {
		return nil, err
	}
	if cached != nil {
		return cached, nil
	}

	data, err := p.db.Read(ctx)
	if err != nil {
		return nil, err
	}

	p.cache.Update(ctx, data)
	return data, nil
}

func (p *SystemCaching) Update(ctx context.Context, config *model.SystemConfig) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	err := p.db.Update(ctx, config)
	if err != nil {
		return err
	}

	p.cache.Update(ctx, config)

	return nil
}
