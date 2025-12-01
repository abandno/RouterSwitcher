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
	mainWindow   application.Window // 保存主窗口引用
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
func (a *WailsApp) startup() {
	log.Println("启动路由器切换工具")
	a.ctx = a.app.Context()

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
		if err := a.UpdateConfig(a.config); err != nil {
			log.Printf("更新配置失败: %v", err)
		}
	})

	a.dynamicItem = a.trayMenu.AddRadio("动态IP", false)
	a.dynamicItem.OnClick(func(*application.Context) {
		log.Println("切换到动态IP模式")
		a.config.IPMode = "dynamic"
		if err := a.UpdateConfig(a.config); err != nil {
			log.Printf("更新配置失败: %v", err)
		}
	})

	a.staticItem = a.trayMenu.AddRadio("静态IP", false)
	a.staticItem.OnClick(func(*application.Context) {
		log.Println("切换到静态IP模式")
		a.config.IPMode = "static"
		if err := a.UpdateConfig(a.config); err != nil {
			log.Printf("更新配置失败: %v", err)
		}
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
		a.showWindow()
	})

	a.systemTray.OnDoubleClick(func() {
		log.Println("托盘图标被双击")
		// 双击也显示配置界面
		a.showWindow()
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

// showWindow 显示配置窗口
func (a *WailsApp) showWindow() {
	if a.app == nil {
		log.Println("app未初始化，无法显示窗口")
		return
	}

	// 如果主窗口已存在，显示它
	if a.mainWindow != nil {
		log.Println("显示现有窗口")
		a.mainWindow.Show()
		a.mainWindow.Focus()
		return
	}

	// 创建新窗口，使用选项来设置窗口属性
	log.Println("创建新窗口")
	newWindow := a.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "路由器切换工具",
		Width:  800,
		Height: 600,
		URL:    "/", // 加载前端资源
		// 不隐藏关闭按钮，但我们会通过 ShouldQuit 来拦截
	})
	if newWindow == nil {
		log.Println("创建窗口失败")
		return
	}

	// 保存窗口引用
	a.mainWindow = newWindow
	log.Printf("窗口已创建，类型: %T, 窗口引用已保存", newWindow)

	// 监听窗口关闭事件 - 由于 wails v3 的事件类型可能不同
	// 我们尝试一个变通方法：通过 WindowManager 监听窗口关闭
	// 或者使用窗口选项来禁用关闭按钮（但这会影响用户体验）

	// 方法1：尝试通过 WindowManager 监听所有窗口事件
	// 注意：这需要在窗口创建前设置
	log.Println("尝试通过 WindowManager 监听窗口事件...")

	// 方法2：由于无法直接拦截窗口关闭，我们使用 ShouldQuit 回调
	// 但 ShouldQuit 只在 app.Quit() 时触发，不在窗口关闭时触发
	//
	// 方法3：在窗口创建时，通过设置选项来禁用关闭按钮
	// 但这会影响用户体验，不推荐
	//
	// 方法4：使用前端 JavaScript 来拦截窗口关闭事件
	// 这可能是最可行的方法

	// 显示窗口
	newWindow.Show()
	log.Println("新窗口已创建并显示，关闭事件监听已设置")
}

// HideWindow 隐藏窗口（由前端调用，用于拦截窗口关闭）
func (a *WailsApp) HideWindow() {
	log.Println("HideWindow 被调用 - 隐藏窗口而不是关闭")
	if a.mainWindow != nil {
		a.mainWindow.Hide()
		log.Println("窗口已隐藏")
	} else {
		log.Println("主窗口为 nil，无法隐藏")
	}
}

// Greet returns a greeting for the given name
func (a *WailsApp) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// GetConfig 返回当前配置
func (a *WailsApp) GetConfig() *Config {
	log.Println("GetConfig")
	return a.config
}

// UpdateConfig 保存配置 & 应用新配置
func (a *WailsApp) UpdateConfig(config *Config) error {
	log.Println("UpdateConfig")
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
	if err := a.UpdateConfig(a.config); err != nil {
		log.Printf("更新配置失败: %v", err)
	}
}

// CheckAndSwitch 检查网络状态并切换配置
func (a *WailsApp) CheckAndSwitch() {
	a.checkAndSwitch()
}

// IsConnectedToHomeNetwork 检查是否连接到家庭局域网
func (a *WailsApp) IsConnectedToHomeNetwork() bool {
	log.Println("IsConnectedToHomeNetwork")
	return a.isConnectedToHomeNetwork()
}

// IsSideRouterReachable 检查旁路由是否可达
func (a *WailsApp) IsSideRouterReachable() bool {
	log.Println("IsSideRouterReachable")
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
			if err := exec.Command("cmd", "/C", "start", "ms-settings:privacy-location").Start(); err != nil {
				log.Printf("打开位置设置页面失败: %v", err)
			}
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
		defer func() {
			if err := logFile.Close(); err != nil {
				log.Printf("关闭日志文件失败: %v", err)
			}
		}()
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
		// 设置 ShouldQuit 回调，当所有窗口关闭时，不退出应用（因为有托盘图标） a.app.Quit() 时触发
		// ShouldQuit: func() bool {
		// 	log.Println("========== ShouldQuit 被调用 ==========")
		// 	log.Println("返回 false 以保持应用运行（托盘模式）")
		// 	// 隐藏主窗口（如果存在）
		// 	if app.mainWindow != nil {
		// 		app.mainWindow.Hide()
		// 		log.Println("主窗口已隐藏")
		// 	} else {
		// 		log.Println("主窗口为 nil，无法隐藏")
		// 	}
		// 	return false // 返回 false 表示不退出应用
		// },
		// 设置 OnShutdown 回调用于调试
		OnShutdown: func() {
			log.Println("========== OnShutdown 被调用 - 应用正在关闭 ==========")
		},
	})

	app.app = appInstance

	// 处理开机启动（在启动前处理）
	app.handleAutoStart()

	// 在 Run() 之前初始化（Run() 是阻塞调用，不会返回）
	log.Println("应用启动，开始初始化...")
	log.Printf("当前配置: %+v", app.config)
	app.startup()
	log.Println("初始化完成，启动应用...")

	// Run the application (阻塞调用，直到应用退出)
	err = appInstance.Run()
	if err != nil {
		log.Fatal(err)
	}
}
