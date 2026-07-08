// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package cache

import (
	"context"
	"sync"

	"ez-ddns/v2/model"
)

type SystemCache struct {
	mu     sync.RWMutex
	config *model.SystemConfig
}

func NewSystemCache() *SystemCache {
	return &SystemCache{
		config: nil,
	}
}

func (c *SystemCache) Read(ctx context.Context) (*model.SystemConfig, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.config == nil {
		return nil, nil
	}

	result := &model.SystemConfig{
		LogLevel: c.config.LogLevel,
	}
	return result, nil
}

func (c *SystemCache) Update(ctx context.Context, config *model.SystemConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if config == nil {
		c.config = nil
		return nil
	}

	c.config = &model.SystemConfig{
		LogLevel: config.LogLevel,
	}
	return nil
}
