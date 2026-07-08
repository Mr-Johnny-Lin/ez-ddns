// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package caching

import (
	"context"
	"sync"

	"ez-ddns/v2/dao"
	"ez-ddns/v2/model"
)

type ConfigCaching struct {
	cache    dao.ConfigRepository
	db       dao.ConfigRepository
	mu       sync.RWMutex
	onChange func() // 数据变更回调，用于通知任务调度器
}

func NewConfigCaching(cache dao.ConfigRepository, db dao.ConfigRepository, onChange func()) *ConfigCaching {
	return &ConfigCaching{
		cache:    cache,
		db:       db,
		onChange: onChange,
	}
}

func (p *ConfigCaching) ReadByID(ctx context.Context, id string) (*model.Config, error) {
	p.mu.RLock()
	cached, err := p.cache.ReadByID(ctx, id)
	p.mu.RUnlock()
	if err != nil {
		return nil, err
	}

	if cached != nil {
		return cached, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	cached, err = p.cache.ReadByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if cached != nil {
		return cached, nil
	}

	data, err := p.db.ReadByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if data != nil {
		p.cache.Create(ctx, data)
	}

	return data, nil
}

func (p *ConfigCaching) ReadAll(ctx context.Context) ([]model.Config, error) {
	p.mu.RLock()
	cached, err := p.cache.ReadAll(ctx)
	p.mu.RUnlock()
	if err != nil {
		return nil, err
	}

	if len(cached) > 0 {
		return cached, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	cached, err = p.cache.ReadAll(ctx)
	if err != nil {
		return nil, err
	}
	if len(cached) > 0 {
		return cached, nil
	}

	data, err := p.db.ReadAll(ctx)
	if err != nil {
		return nil, err
	}

	for i := range data {
		p.cache.Create(ctx, &data[i])
	}

	return data, nil
}

func (p *ConfigCaching) Create(ctx context.Context, config *model.Config) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	err := p.db.Create(ctx, config)
	if err != nil {
		return err
	}

	p.cache.Create(ctx, config)

	if p.onChange != nil {
		p.onChange()
	}

	return nil
}

func (p *ConfigCaching) Update(ctx context.Context, config *model.Config) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	err := p.db.Update(ctx, config)
	if err != nil {
		return err
	}

	p.cache.Delete(ctx, config.ID)

	if p.onChange != nil {
		p.onChange()
	}

	return nil
}

func (p *ConfigCaching) Delete(ctx context.Context, id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	err := p.db.Delete(ctx, id)
	if err != nil {
		return err
	}

	p.cache.Delete(ctx, id)

	if p.onChange != nil {
		p.onChange()
	}

	return nil
}
