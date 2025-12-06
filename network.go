package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetActiveInterface 获取活动网络接口名称
func GetActiveInterface() (string, error) {
	// 使用netsh命令获取网络接口信息
	cmd := exec.Command("netsh", "interface", "show", "interface")
	hideCmdWindow(cmd)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		// 查找已连接的网络接口 (支持中英文环境)
		if (strings.Contains(line, "Connected") || strings.Contains(line, "已连接")) && 
		   (strings.Contains(line, "Dedicated") || strings.Contains(line, "专用")) {
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				return fields[len(fields)-1], nil
			}
		}
	}

	return "", fmt.Errorf("未找到活动网络接口")
}

// GetCurrentIPConfig 检查当前网络接口是否为DHCP模式
func GetCurrentIPConfig(iface string) (isDHCP bool, err error) {
	// 使用netsh命令获取接口IP配置
	cmd := exec.Command("netsh", "interface", "ip", "show", "config", iface)
	hideCmdWindow(cmd)
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	// 检查输出中是否包含DHCP相关信息
	outputStr := string(output)
	if strings.Contains(outputStr, "DHCP enabled") || strings.Contains(outputStr, "DHCP 已启用") {
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			if strings.Contains(line, "DHCP enabled") || strings.Contains(line, "DHCP 已启用") {
				if strings.Contains(line, "Yes") || strings.Contains(line, "是") {
					return true, nil
				}
				break
			}
		}
	}
	
	return false, nil
}

// GetCurrentStaticIPConfig 检查当前网络接口是否已经是目标静态IP配置
func GetCurrentStaticIPConfig(iface, staticIP, gateway, dns string) (isStatic bool, err error) {
	// 使用netsh命令获取接口IP配置
	cmd := exec.Command("netsh", "interface", "ip", "show", "config", iface)
	hideCmdWindow(cmd)
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	// 将输出转换为字符串
	outputStr := string(output)
	
	// 检查是否为静态IP配置
	isDHCP, _ := GetCurrentIPConfig(iface)
	if isDHCP {
		// 如果当前是DHCP模式，则肯定不是目标静态IP配置
		return false, nil
	}
	
	// 检查IP地址、网关和DNS是否匹配目标配置
	if strings.Contains(outputStr, staticIP) && 
	   strings.Contains(outputStr, gateway) &&
	   strings.Contains(outputStr, dns) {
		return true, nil
	}
	
	return false, nil
}

// SetDHCP 设置网络接口为DHCP模式
func SetDHCP(iface string) error {
	// 设置为DHCP自动获取IP
	cmd := exec.Command("netsh", "interface", "ip", "set", "address", iface, "dhcp")
	hideCmdWindow(cmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("设置DHCP IP失败: %v", err)
	}

	// 设置DNS为自动获取
	cmd = exec.Command("netsh", "interface", "ip", "set", "dns", iface, "dhcp")
	hideCmdWindow(cmd)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("设置DHCP DNS失败: %v", err)
	}

	return nil
}

// SetStaticIP 设置网络接口为静态IP模式
func SetStaticIP(iface, ip, subnetMask, gateway, dns string) error {
	// 设置静态IP地址、子网掩码和网关
	cmd := exec.Command("netsh", "interface", "ip", "set", "address", iface, "static", ip, subnetMask, gateway)
	hideCmdWindow(cmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("设置静态IP失败: %v", err)
	}

	// 设置静态DNS服务器
	cmd = exec.Command("netsh", "interface", "ip", "set", "dns", iface, "static", dns)
	hideCmdWindow(cmd)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("设置静态DNS失败: %v", err)
	}

	return nil
}

// Ping 测试网络连通性
func Ping(host string) bool {
	cmd := exec.Command("ping", "-n", "1", "-w", "3000", host)
	hideCmdWindow(cmd)
	err := cmd.Run()
	return err == nil
}