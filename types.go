package main

// Config 配置结构
type Config struct {
	HomeSSID   string // 家庭WiFi的SSID
	StaticIP   string // 静态IP地址
	Gateway    string // 网关地址
	DNS        string // DNS服务器地址
	AutoStart  bool   // 是否开机自启
	IPMode     string // IP模式: adaptive(自适应), dynamic(动态IP), static(静态IP)
}