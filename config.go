package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

const (
	ConfigFileName = "config.json"
)

// LoadConfig 加载配置文件
func LoadConfig() (*Config, error) {
	config := &Config{
		HomeSSID:  "HomeWiFi",
		StaticIP:  "192.168.31.100",
		Gateway:   "192.168.31.2",
		DNS:       "192.168.31.2",
		AutoStart: false,
		IPMode:    "adaptive", // 默认为自适应模式
	}

	// 获取可执行文件所在目录
	exePath, err := os.Executable()
	if err != nil {
		return config, nil // 返回默认配置
	}

	dir := filepath.Dir(exePath)
	configPath := filepath.Join(dir, ConfigFileName)

	// 在开发模式下，如果可执行文件目录下没有配置文件，尝试从当前工作目录加载
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 尝试从当前工作目录加载
		wd, err := os.Getwd()
		if err == nil {
			wdConfigPath := filepath.Join(wd, ConfigFileName)
			if _, err := os.Stat(wdConfigPath); err == nil {
				configPath = wdConfigPath
				log.Println("配置文件路径(wd):", wdConfigPath)
			}
		}
	} else {
		log.Println("配置文件路径:", configPath)
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 配置文件不存在，保存默认配置
		if err := SaveConfig(config); err != nil {
			log.Printf("保存默认配置失败: %v", err)
		} else {
			log.Printf("配置文件不存在，保存默认配置: %+v\n", config)
		}
		return config, nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, nil // 返回默认配置
	}

	// 解析配置文件
	err = json.Unmarshal(data, config)
	if err != nil {
		return config, nil // 返回默认配置
	}

	return config, nil
}

// SaveConfig 保存配置到文件
func SaveConfig(config *Config) error {
	// 获取可执行文件所在目录
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	dir := filepath.Dir(exePath)
	configPath := filepath.Join(dir, ConfigFileName)

	// 序列化配置
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// 写入配置文件
	return os.WriteFile(configPath, data, 0644)
}
