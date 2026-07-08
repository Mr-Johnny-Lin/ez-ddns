// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package cache

import (
	"context"
	"sync"

	"ez-ddns/v2/model"
)

type DomainCache struct {
	mu      sync.RWMutex
	domains map[string][]model.DomainConfig
}

func NewDomainCache() *DomainCache {
	return &DomainCache{
		domains: make(map[string][]model.DomainConfig),
	}
}

func (c *DomainCache) ReadByConfigID(ctx context.Context, configID string) ([]model.DomainConfig, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	domains, exists := c.domains[configID]
	if !exists {
		return []model.DomainConfig{}, nil
	}

	result := make([]model.DomainConfig, len(domains))
	copy(result, domains)
	return result, nil
}

func (c *DomainCache) Update(ctx context.Context, domain *model.DomainConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	configID := domain.ID
	domains, exists := c.domains[configID]
	if !exists {
		c.domains[configID] = []model.DomainConfig{*domain}
		return nil
	}

	for i, d := range domains {
		if d.ID == domain.ID {
			domains[i] = *domain
			break
		}
	}

	return nil
}

func (c *DomainCache) Create(ctx context.Context, domain *model.DomainConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	configID := domain.ID
	c.domains[configID] = append(c.domains[configID], *domain)
	return nil
}

func (c *DomainCache) Delete(ctx context.Context, id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for configID, domains := range c.domains {
		for i, domain := range domains {
			if domain.ID == id {
				c.domains[configID] = append(domains[:i], domains[i+1:]...)
				return nil
			}
		}
	}

	return nil
}

func (c *DomainCache) DeleteByConfigID(ctx context.Context, configID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.domains, configID)
	return nil
}
