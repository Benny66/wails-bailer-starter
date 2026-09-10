// Package config 负责应用本地配置（config.json）的读写。
//
// 配置落在用户配置目录（Windows: %AppData%，macOS: ~/Library/Application Support，
// Linux: ~/.config）下的 appName 目录，与数据库同目录。
// 首次启动无配置文件时写入默认值。
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config 是应用配置的持久化结构。
// 字段带 json tag，新增配置项在此扩展并给默认值。
type Config struct {
	// Theme 界面主题：dark / light。空串（""）表示"未设置"，语义为"跟随系统"。
	Theme string `json:"theme"`

	// path 是配置文件路径，不序列化（json:"-"）。
	path string `json:"-"`
}

// Default 返回默认配置。
// 字段默认值为空串（""），表示"未设置"——由调用方决定 fallback 行为，
// 而非硬编码一个具体默认值（如 Theme 空串 = 跟随系统）。
func Default() *Config {
	return &Config{Theme: ""}
}

// Load 读取配置；文件不存在时写入默认值并返回默认配置。
func Load(appName string) (*Config, error) {
	path, err := configPath(appName)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			c := Default()
			c.path = path
			if err := c.Save(); err != nil {
				return nil, fmt.Errorf("写入默认配置失败: %w", err)
			}
			return c, nil
		}
		return nil, fmt.Errorf("读取配置失败: %w", err)
	}

	c := Default()
	if err := json.Unmarshal(data, c); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}
	c.path = path
	return c, nil
}

// Save 将当前配置写回文件。
func (c *Config) Save() error {
	if c.path == "" {
		return fmt.Errorf("配置未绑定路径")
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}
	if err := os.WriteFile(c.path, data, 0o644); err != nil {
		return fmt.Errorf("写入配置失败: %w", err)
	}
	return nil
}

// configPath 返回配置文件路径，并确保目录存在。
func configPath(appName string) (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("获取用户配置目录失败: %w", err)
	}
	dataDir := filepath.Join(dir, appName)
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return "", fmt.Errorf("创建配置目录失败: %w", err)
	}
	return filepath.Join(dataDir, "config.json"), nil
}
