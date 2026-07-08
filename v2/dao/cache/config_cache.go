// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package cache

import (
	"context"
	"sync"

	"ez-ddns/v2/model"
)

type ConfigCache struct {
	mu      sync.RWMutex
	configs map[string]*model.Config
}

func NewConfigCache() *ConfigCache {
	return &ConfigCache{
		configs: make(map[string]*model.Config),
	}
}

func (c *ConfigCache) ReadByID(ctx context.Context, id string) (*model.Config, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	config, exists := c.configs[id]
	if !exists {
		return nil, nil
	}

	result := *config
	return &result, nil
}

func (c *ConfigCache) ReadAll(ctx context.Context) ([]model.Config, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]model.Config, 0, len(c.configs))
	for _, config := range c.configs {
		copied := *config
		result = append(result, copied)
	}

	return result, nil
}

func (c *ConfigCache) Create(ctx context.Context, config *model.Config) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	copied := *config
	c.configs[config.ID] = &copied
	return nil
}

func (c *ConfigCache) Update(ctx context.Context, config *model.Config) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	copied := *config
	c.configs[config.ID] = &copied
	return nil
}

func (c *ConfigCache) Delete(ctx context.Context, id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.configs, id)
	return nil
}
