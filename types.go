package main

// Config 配置结构
type Config struct {
	HomeSSID  string // 家庭WiFi的SSID
	StaticIP  string // 静态IP地址
	Gateway   string // 网关地址
	DNS       string // DNS服务器地址
	AutoStart bool   // 是否开机自启
	IPMode    string // IP模式: adaptive(自适应), dynamic(动态IP), static(静态IP)
}

// NetworkStatus 网络状态结构
type NetworkStatus struct {
	WiFiName         string // WiFi名称
	WiFiConnected    bool   // WiFi连接状态
	IPAddress        string // 当前IP地址
	Gateway          string // 当前网关
	GatewayReachable bool   // 网关是否可达
	DNS              string // 当前DNS
	DNSReachable     bool   // DNS是否可达
	IPAssignment     string // IP分配方式: "自动(DHCP)" 或 "手动"
	DNSAssignment    string // DNS分配方式: "自动(DHCP)" 或 "手动"
}
