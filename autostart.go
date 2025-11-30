package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// EnableAutoStart 启用开机启动
func EnableAutoStart() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("当前仅支持Windows系统")
	}

	// 获取可执行文件路径
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %v", err)
	}

	// 转换为绝对路径
	exePath, err = filepath.Abs(exePath)
	if err != nil {
		return fmt.Errorf("获取绝对路径失败: %v", err)
	}

	// 使用schtasks命令创建开机启动任务
	// 任务名称
	taskName := "RouterSwitcher"

	// 删除可能存在的旧任务
	exec.Command("schtasks", "/Delete", "/TN", taskName, "/F").Run()

	// 创建新任务
	cmd := exec.Command("schtasks", "/Create",
		"/TN", taskName,
		"/TR", fmt.Sprintf(`"%s"`, exePath),
		"/SC", "ONLOGON",
		"/RL", "HIGHEST",
		"/F",
	)

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("创建开机启动任务失败: %v", err)
	}

	return nil
}

// DisableAutoStart 禁用开机启动
func DisableAutoStart() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("当前仅支持Windows系统")
	}

	// 任务名称
	taskName := "RouterSwitcher"

	// 删除任务
	cmd := exec.Command("schtasks", "/Delete", "/TN", taskName, "/F")
	err := cmd.Run()
	if err != nil {
		// 如果任务不存在，不算错误
		return nil
	}

	return nil
}

// IsAutoStartEnabled 检查是否已启用开机启动
func IsAutoStartEnabled() bool {
	if runtime.GOOS != "windows" {
		return false
	}

	// 任务名称
	taskName := "RouterSwitcher"

	// 查询任务是否存在
	cmd := exec.Command("schtasks", "/Query", "/TN", taskName)
	err := cmd.Run()
	return err == nil
}
