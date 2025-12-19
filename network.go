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

// GetCurrentWiFiName 获取当前连接的WiFi名称
func GetCurrentWiFiName() (string, error) {
	cmd := exec.Command("netsh", "wlan", "show", "interfaces")
	hideCmdWindow(cmd)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")

	// 查找包含SSID的行
	for _, line := range lines {
		if strings.Contains(line, "SSID") && !strings.Contains(line, "BSSID") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				wifiName := strings.TrimSpace(parts[1])
				return wifiName, nil
			}
		}
	}

	return "", fmt.Errorf("未找到WiFi信息")
}

// GetCurrentNetworkStatus 获取当前网络详细状态
func GetCurrentNetworkStatus() (*NetworkStatus, error) {
	status := &NetworkStatus{}

	// 获取活动网络接口
	iface, err := GetActiveInterface()
	if err != nil {
		return status, fmt.Errorf("获取网络接口失败: %v", err)
	}

	// 获取WiFi名称
	wifiName, err := GetCurrentWiFiName()
	if err == nil {
		status.WiFiName = wifiName
		status.WiFiConnected = true
	} else {
		status.WiFiName = "未连接"
		status.WiFiConnected = false
	}

	// 获取网络接口配置
	cmd := exec.Command("netsh", "interface", "ip", "show", "config", iface)
	hideCmdWindow(cmd)
	output, err := cmd.Output()
	if err != nil {
		return status, fmt.Errorf("获取网络配置失败: %v", err)
	}

	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")

	// 解析IP配置信息
	isIPDHCP := false
	isDNSDHCP := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 检查IP DHCP状态
		if strings.Contains(line, "DHCP enabled") || strings.Contains(line, "DHCP 已启用") {
			if strings.Contains(line, "Yes") || strings.Contains(line, "是") {
				isIPDHCP = true
			}
		}

		// 提取IP地址
		if strings.Contains(line, "IP 地址") || strings.Contains(line, "IP Address") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				ip := strings.TrimSpace(parts[1])
				// 移除子网前缀信息（如果有）
				if idx := strings.Index(ip, "/"); idx != -1 {
					ip = ip[:idx]
				}
				if idx := strings.Index(ip, " "); idx != -1 {
					ip = ip[:idx]
				}
				if ip != "" {
					status.IPAddress = ip
				}
			}
		}

		// 提取网关
		if strings.Contains(line, "默认网关") || strings.Contains(line, "Default Gateway") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				gateway := strings.TrimSpace(parts[1])
				// 移除跃点数信息（如果有）
				if idx := strings.Index(gateway, " "); idx != -1 {
					gateway = gateway[:idx]
				}
				if gateway != "" {
					status.Gateway = gateway
				}
			}
		}

		// 检查DNS DHCP状态并提取DNS地址
		if strings.Contains(line, "通过 DHCP 配置的 DNS") || strings.Contains(line, "DNS servers configured through DHCP") {
			isDNSDHCP = true
			// 提取DHCP配置的DNS地址
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				dns := strings.TrimSpace(parts[1])
				// 移除其他信息（如果有）
				if idx := strings.Index(dns, " "); idx != -1 {
					dns = dns[:idx]
				}
				if dns != "" && dns != "无" && dns != "None" {
					status.DNS = dns
				}
			}
		} else if strings.Contains(line, "DNS 服务器") || strings.Contains(line, "DNS Servers") {
			// 提取静态DNS配置
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				dns := strings.TrimSpace(parts[1])
				// 移除其他信息（如果有）
				if idx := strings.Index(dns, " "); idx != -1 {
					dns = dns[:idx]
				}
				if dns != "" && dns != "无" && dns != "None" {
					status.DNS = dns
					// 如果找到静态DNS，则不是DHCP
					isDNSDHCP = false
				}
			}
		}
	}

	// 设置分配方式
	if isIPDHCP {
		status.IPAssignment = "自动(DHCP)"
	} else {
		status.IPAssignment = "手动"
	}

	if isDNSDHCP {
		status.DNSAssignment = "自动(DHCP)"
	} else {
		status.DNSAssignment = "手动"
	}

	// 测试网关和DNS连通性
	if status.Gateway != "" {
		status.GatewayReachable = Ping(status.Gateway)
	}
	if status.DNS != "" {
		status.DNSReachable = Ping(status.DNS)
	}

	return status, nil
}
