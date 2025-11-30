package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/windows/icon.ico
var icon []byte

// WailsApp struct
type WailsApp struct {
	ctx          context.Context
	config       *Config
	app          *application.App
	systemTray   *application.SystemTray
	trayMenu     *application.Menu
	adaptiveItem *application.MenuItem
	dynamicItem  *application.MenuItem
	staticItem   *application.MenuItem
	exitItem     *application.MenuItem
}

// NewWailsApp creates a new WailsApp application struct
func NewWailsApp() *WailsApp {
	// 加载配置
	config, err := LoadConfig()
	if err != nil {
		log.Printf("加载配置失败: %v", err)
	}

	return &WailsApp{
		config: config,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *WailsApp) startup(ctx context.Context) {
	log.Println("启动路由器切换工具")
	a.ctx = ctx

	// 创建系统托盘菜单
	a.createTrayMenu()

	// 启动网络监控
	go a.monitorNetwork()
}

// createTrayMenu 创建系统托盘菜单
func (a *WailsApp) createTrayMenu() {
	log.Println("创建系统托盘菜单")

	if a.app == nil {
		log.Println("app未初始化，无法创建托盘菜单")
		return
	}

	// 创建系统托盘
	a.systemTray = a.app.SystemTray.New()
	a.systemTray.SetIcon(icon)
	a.systemTray.SetTooltip("路由器切换工具")

	// 创建菜单
	a.trayMenu = application.NewMenu()

	// 创建菜单项（使用AddRadio方法，实现单选效果）
	a.adaptiveItem = a.trayMenu.AddRadio("自适应IP", false)
	a.adaptiveItem.OnClick(func(*application.Context) {
		log.Println("切换到自适应IP模式")
		a.config.IPMode = "adaptive"
		if err := SaveConfig(a.config); err != nil {
			log.Printf("保存配置失败: %v", err)
		}
		a.updateTrayMenuState()
		go a.checkAndSwitch()
	})

	a.dynamicItem = a.trayMenu.AddRadio("动态IP", false)
	a.dynamicItem.OnClick(func(*application.Context) {
		log.Println("切换到动态IP模式")
		a.config.IPMode = "dynamic"
		if err := SaveConfig(a.config); err != nil {
			log.Printf("保存配置失败: %v", err)
		}
		a.updateTrayMenuState()
		go a.switchToDHCP()
	})

	a.staticItem = a.trayMenu.AddRadio("静态IP", false)
	a.staticItem.OnClick(func(*application.Context) {
		log.Println("切换到静态IP模式")
		a.config.IPMode = "static"
		if err := SaveConfig(a.config); err != nil {
			log.Printf("保存配置失败: %v", err)
		}
		a.updateTrayMenuState()
		go a.switchToStatic()
	})

	a.trayMenu.AddSeparator()

	a.exitItem = a.trayMenu.Add("退出")
	a.exitItem.OnClick(func(*application.Context) {
		log.Println("退出程序")
		a.app.Quit()
	})

	// 设置托盘菜单
	a.systemTray.SetMenu(a.trayMenu)

	// 设置托盘图标点击事件
	a.systemTray.OnClick(func() {
		log.Println("托盘图标被单击")
		// 单击显示配置界面
		if a.app != nil {
			a.app.Show()
		}
	})

	a.systemTray.OnDoubleClick(func() {
		log.Println("托盘图标被双击")
		// 双击也显示配置界面
		if a.app != nil {
			a.app.Show()
		}
	})

	// 显示系统托盘
	a.systemTray.Show()

	// 更新菜单状态
	a.updateTrayMenuState()
}

// updateTrayMenuState 更新托盘菜单状态（根据当前IP模式设置勾选状态）
func (a *WailsApp) updateTrayMenuState() {
	// 清除所有勾选状态
	if a.adaptiveItem != nil {
		a.adaptiveItem.SetChecked(false)
	}
	if a.dynamicItem != nil {
		a.dynamicItem.SetChecked(false)
	}
	if a.staticItem != nil {
		a.staticItem.SetChecked(false)
	}

	// 根据当前模式设置勾选状态
	switch a.config.IPMode {
	case "adaptive":
		if a.adaptiveItem != nil {
			a.adaptiveItem.SetChecked(true)
		}
	case "dynamic":
		if a.dynamicItem != nil {
			a.dynamicItem.SetChecked(true)
		}
	case "static":
		if a.staticItem != nil {
			a.staticItem.SetChecked(true)
		}
	}

	// 更新菜单显示
	if a.trayMenu != nil {
		a.trayMenu.Update()
	}
}

// Greet returns a greeting for the given name
func (a *WailsApp) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// GetConfig 返回当前配置
func (a *WailsApp) GetConfig() *Config {
	return a.config
}

// SaveConfig 保存配置
func (a *WailsApp) SaveConfig(config *Config) error {
	a.config = config
	err := SaveConfig(a.config)
	if err != nil {
		return err
	}
	log.Printf("保存配置成功: %+v\n", config)

	// 更新托盘菜单状态
	a.updateTrayMenuState()

	// 处理开机启动
	a.handleAutoStart()

	// 触发网络检查
	go a.checkAndSwitch()
	return nil
}

// SwitchToStatic 切换到静态IP模式
func (a *WailsApp) SwitchToStatic() {
	a.switchToStatic()
}

// SwitchToDHCP 切换到动态IP模式
func (a *WailsApp) SwitchToDHCP() {
	a.switchToDHCP()
}

// SwitchToAdaptive 切换到自适应IP模式
func (a *WailsApp) SwitchToAdaptive() {
	a.config.IPMode = "adaptive"
	if err := SaveConfig(a.config); err != nil {
		log.Printf("保存配置失败: %v", err)
	}
	a.updateTrayMenuState()
	go a.checkAndSwitch()
}

// CheckAndSwitch 检查网络状态并切换配置
func (a *WailsApp) CheckAndSwitch() {
	a.checkAndSwitch()
}

// IsConnectedToHomeNetwork 检查是否连接到家庭局域网
func (a *WailsApp) IsConnectedToHomeNetwork() bool {
	return a.isConnectedToHomeNetwork()
}

// IsSideRouterReachable 检查旁路由是否可达
func (a *WailsApp) IsSideRouterReachable() bool {
	return a.isSideRouterReachable()
}

// OpenLocationSettings 打开位置设置页面
func (a *WailsApp) OpenLocationSettings() error {
	cmd := exec.Command("cmd", "/C", "start", "ms-settings:privacy-location")
	return cmd.Start()
}

// monitorNetwork 监控网络变化
func (a *WailsApp) monitorNetwork() {
	for {
		if a.config.IPMode == "adaptive" {
			a.checkAndSwitch()
		}
		time.Sleep(30 * time.Second) // 每30秒检查一次
	}
}

// checkAndSwitch 检查网络状态并切换配置
func (a *WailsApp) checkAndSwitch() {
	// 只有在自适应模式下才进行自动切换
	switch mode := a.config.IPMode; mode {
	case "adaptive":
		// 连接到家庭局域网 且旁路由可达  设置静态IP
		if a.isConnectedToHomeNetwork() && a.isSideRouterReachable() {
			a.switchToStatic()
		} else {
			// 不是家庭局域网 或 旁路由不可达，切回动态IP
			a.switchToDHCP()
		}
	case "static":
		// 强制使用静态IP
		a.switchToStatic()
	case "dynamic":
		// 强制使用动态IP
		a.switchToDHCP()
	}
}

// isConnectedToHomeNetwork 检查是否连接到家庭局域网
func (a *WailsApp) isConnectedToHomeNetwork() bool {
	// 执行命令获取当前WiFi信息
	cmd := exec.Command("netsh", "wlan", "show", "interfaces")
	output, err := cmd.Output()
	if err != nil {
		outputStr := string(output)
		log.Printf("执行netsh命令失败: %v. %v", err, outputStr)

		// 检查是否因为位置服务禁用导致无法获取SSID
		if strings.Contains(outputStr, "命令需要位置权限才能访问") ||
			strings.Contains(outputStr, "WlanQueryInterface 返回错误 5") ||
			strings.Contains(outputStr, "拒绝访问") ||
			strings.Contains(outputStr, "Network shell commands need location permission") {
			log.Println("检测到位置服务被禁用，提示用户开启位置服务以获取WiFi信息")
			a.promptUserToEnableLocationService()
		}
		return false
	}

	// 将输出转换为字符串并按行分割
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")

	// 查找包含SSID的行
	for _, line := range lines {
		// 查找包含"SSID"但不包含"BSSID"的行
		if strings.Contains(line, "SSID") && !strings.Contains(line, "BSSID") {
			// 提取SSID值
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				// 去除空格和换行符
				currentSSID := strings.TrimSpace(parts[1])
				// 比较当前SSID与配置中的HomeSSID
				if currentSSID == a.config.HomeSSID {
					return true
				}
			}
		}
	}

	return false
}

// 添加一个全局变量来跟踪是否已经显示过弹窗
var locationServicePromptShown = false

// promptUserToEnableLocationService 提示用户开启位置服务
func (a *WailsApp) promptUserToEnableLocationService() {
	// 检查是否已经显示过弹窗，避免重复弹窗引起用户恐慌
	if locationServicePromptShown {
		log.Println("位置服务提示弹窗已显示过，避免重复弹窗")
		return
	}

	log.Println("==============================================")
	log.Println("检测到位置服务被禁用，无法获取WiFi信息")
	log.Println("请按以下步骤开启位置服务：")
	log.Println("1. 打开Windows设置 (Win + I)")
	log.Println("2. 进入「隐私和安全」->「位置」")
	log.Println("3. 开启「位置服务」开关")
	log.Println("或者在运行对话框中执行以下命令打开位置设置：")
	log.Println("   Win + R -> 输入: ms-settings:privacy-location")
	log.Println("或者终端命令行中输入: start ms-settings:privacy-location")
	log.Println("==============================================")

	// 在GUI中显示提示信息
	if a.ctx != nil {
		// 设置标记，表示即将显示弹窗
		locationServicePromptShown = true

		// 显示弹窗并获取用户响应
		// 		result, err := wailsruntime.MessageDialog(a.ctx, wailsruntime.MessageDialogOptions{
		// 			Type:    wailsruntime.InfoDialog,
		// 			Title:   "需要开启位置服务",
		// 			Message: `检测到位置服务被禁用，无法获取WiFi信息。
		// 请在应用界面中点击"位置服务帮助"按钮获取详细操作指南。`,
		// 			Buttons: []string{"确定", "取消"},
		// 		})

		// 使用DialogManager显示信息对话框
		if a.app != nil {
			dialog := a.app.Dialog.Info()
			dialog.SetTitle("需要开启位置服务")
			dialog.SetMessage(`检测到位置服务被禁用，无法获取WiFi信息。

请按以下步骤开启位置服务：
1. 打开Windows设置 (Win + I)
2. 进入「隐私和安全」->「位置」
3. 开启「位置服务」开关

将自动打开位置设置页面！
或者：Win + R -> 输入: ms-settings:privacy-location
或者：终端命令行中输入: start ms-settings:privacy-location`)
			dialog.AddButton("确定")
			dialog.Show()

			// 自动打开位置设置页面
			exec.Command("cmd", "/C", "start", "ms-settings:privacy-location").Start()
		}
	}
}

// isSideRouterReachable 检查旁路由是否可达
func (a *WailsApp) isSideRouterReachable() bool {
	// 使用系统ping命令检测旁路由地址是否可达
	addr := a.config.Gateway
	return Ping(addr)
}

// switchToStatic 切换到静态IP模式
func (a *WailsApp) switchToStatic() {
	log.Printf("开始切换静态IP")

	// 获取活动网络接口
	iface, err := GetActiveInterface()
	if err != nil {
		log.Printf("获取网络接口失败: %v", err)
		return
	}

	// 检查当前是否已经是目标静态IP配置
	isStatic, err := GetCurrentStaticIPConfig(iface, a.config.StaticIP, a.config.Gateway, a.config.DNS)
	if err == nil && isStatic {
		log.Printf("当前已经是目标静态IP配置, 无需重复设置: IP=%s, Gateway=%s, DNS=%s\n", a.config.StaticIP, a.config.Gateway, a.config.DNS)
		return
	}

	// 设置静态IP (这里使用默认子网掩码 255.255.255.0)
	err = SetStaticIP(iface, a.config.StaticIP, "255.255.255.0", a.config.Gateway, a.config.DNS)
	if err != nil {
		log.Printf("设置静态IP失败: %v", err)
		return
	}

	log.Printf("成功切换到静态IP模式: IP=%s, Gateway=%s, DNS=%s\n", a.config.StaticIP, a.config.Gateway, a.config.DNS)
}

// switchToDHCP 切换到自动获取IP模式
func (a *WailsApp) switchToDHCP() {
	log.Println("开始切换动态IP")

	// 获取活动网络接口
	iface, err := GetActiveInterface()
	if err != nil {
		log.Printf("获取网络接口失败: %v", err)
		return
	}

	// 检查当前是否已经是DHCP模式
	isDHCP, err := GetCurrentIPConfig(iface)
	if err == nil && isDHCP {
		log.Println("当前已经是DHCP模式, 无需重复设置")
		return
	}

	// 设置为DHCP
	err = SetDHCP(iface)
	if err != nil {
		log.Printf("设置DHCP失败: %v", err)
		return
	}

	log.Println("成功切换到DHCP模式")
}

// handleAutoStart 处理开机启动
func (a *WailsApp) handleAutoStart() {
	if a.config.AutoStart {
		err := EnableAutoStart()
		if err != nil {
			log.Printf("启用开机启动失败: %v", err)
		} else {
			log.Println("已启用开机启动")
		}
	} else {
		err := DisableAutoStart()
		if err != nil {
			log.Printf("禁用开机启动失败: %v", err)
		} else {
			log.Println("已禁用开机启动")
		}
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	// 创建日志文件
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("无法创建日志文件: %v", err)
	} else {
		// 将日志同时输出到文件和控制台
		multiWriter := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(multiWriter)
		defer logFile.Close()
	}

	// Create an instance of the app structure
	app := NewWailsApp()

	// Create application with options
	appInstance := application.New(application.Options{
		Name:   "路由器切换工具",
		Assets: application.AssetOptions{Handler: application.BundledAssetFileServer(assets)},
		Logger: nil,
		Services: []application.Service{
			application.NewService(app),
		},
	})

	app.app = appInstance

	// 处理开机启动（在启动前处理）
	app.handleAutoStart()

	// Run the application
	err = appInstance.Run()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("应用启动")

	// 应用启动后初始化
	app.startup(appInstance.Context())
}
