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

type DomainCaching struct {
	cache    dao.DomainRepository
	db       dao.DomainRepository
	mu       sync.RWMutex
	onChange func() // 数据变更回调，用于通知任务调度器
}

func NewDomainCaching(cache dao.DomainRepository, db dao.DomainRepository, onChange func()) *DomainCaching {
	return &DomainCaching{
		cache:    cache,
		db:       db,
		onChange: onChange,
	}
}

func (p *DomainCaching) ReadByConfigID(ctx context.Context, configID string) ([]model.DomainConfig, error) {
	p.mu.RLock()
	cached, err := p.cache.ReadByConfigID(ctx, configID)
	p.mu.RUnlock()
	if err != nil {
		return nil, err
	}

	if len(cached) > 0 {
		return cached, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	cached, err = p.cache.ReadByConfigID(ctx, configID)
	if err != nil {
		return nil, err
	}
	if len(cached) > 0 {
		return cached, nil
	}

	data, err := p.db.ReadByConfigID(ctx, configID)
	if err != nil {
		return nil, err
	}

	if len(data) > 0 {
		for i := range data {
			p.cache.Create(ctx, &data[i])
		}
	}

	return data, nil
}

func (p *DomainCaching) Update(ctx context.Context, domain *model.DomainConfig) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	err := p.db.Update(ctx, domain)
	if err != nil {
		return err
	}

	p.cache.Delete(ctx, domain.ID)

	if p.onChange != nil {
		p.onChange()
	}

	return nil
}

func (p *DomainCaching) Create(ctx context.Context, domain *model.DomainConfig) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	err := p.db.Create(ctx, domain)
	if err != nil {
		return err
	}

	p.cache.Create(ctx, domain)

	if p.onChange != nil {
		p.onChange()
	}

	return nil
}

func (p *DomainCaching) Delete(ctx context.Context, id string) error {
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

func (p *DomainCaching) DeleteByConfigID(ctx context.Context, configID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	err := p.db.DeleteByConfigID(ctx, configID)
	if err != nil {
		return err
	}

	p.cache.DeleteByConfigID(ctx, configID)

	if p.onChange != nil {
		p.onChange()
	}

	return nil
}
