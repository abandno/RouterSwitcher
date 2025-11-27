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

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

// WailsApp struct
type WailsApp struct {
	ctx    context.Context
	config *Config
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
	a.ctx = ctx
	
	// 创建系统托盘菜单
	a.createTrayMenu()
	
	// 启动网络监控
	go a.monitorNetwork()
}

// createTrayMenu 创建系统托盘菜单
func (a *WailsApp) createTrayMenu() {
	appMenu := menu.NewMenu()

	// 添加IP模式选项
	ipModeMenu := appMenu.AddSubmenu("IP模式")
	
	ipModeMenu.AddCheckbox("自适应", a.config.IPMode == "adaptive", &keys.Accelerator{}, func(data *menu.CallbackData) {
		a.config.IPMode = "adaptive"
		SaveConfig(a.config)
		// 触发自适应检查
		go a.checkAndSwitch()
	})
	
	ipModeMenu.AddCheckbox("动态IP", a.config.IPMode == "dynamic", &keys.Accelerator{}, func(data *menu.CallbackData) {
		a.config.IPMode = "dynamic"
		SaveConfig(a.config)
		a.switchToDHCP()
	})
	
	ipModeMenu.AddCheckbox("静态IP", a.config.IPMode == "static", &keys.Accelerator{}, func(data *menu.CallbackData) {
		a.config.IPMode = "static"
		SaveConfig(a.config)
		a.switchToStatic()
	})
	
	// 分隔线
	appMenu.AddSeparator()
	
	// 添加退出选项
	appMenu.AddText("退出", &keys.Accelerator{}, func(data *menu.CallbackData) {
		runtime.Quit(a.ctx)
	})

	runtime.MenuSetApplicationMenu(a.ctx, appMenu)
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
		result, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.InfoDialog,
			Title:   "需要开启位置服务",
			Message: `检测到位置服务被禁用，无法获取WiFi信息。

请按以下步骤开启位置服务：
1. 打开Windows设置 (Win + I)
2. 进入「隐私和安全」->「位置」
3. 开启「位置服务」开关

将自动打开位置设置页面！
或者：Win + R -> 输入: ms-settings:privacy-location
或者：终端命令行中输入: start ms-settings:privacy-location`,
			Buttons: []string{"确定", "取消"},
		})
		// 或者 Win + R -> 输入: ms-settings:privacy-location
		// 或者终端命令行中输入: start ms-settings:privacy-location
		// 或者点击确定按钮自动打开位置设置页面。
		
		// 处理可能的错误
		if err != nil {
			log.Printf("显示提示弹窗时发生错误: %v", err)
		}
		
		
		// 无论用户点击什么按钮，都重置标记以便下次可以再次显示弹窗
		locationServicePromptShown = false
		
		log.Printf("用户选择: %s", result) // x和确定, 点击TM都是 Ok
		exec.Command("cmd", "/C", "start", "ms-settings:privacy-location").Start()
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

// LogMultiWriter 同时写入控制台和文件的日志写入器
type LogMultiWriter struct {
	Console io.Writer
	File    io.Writer
}

// Write 实现io.Writer接口
func (lmw *LogMultiWriter) Write(p []byte) (n int, err error) {
	// 写入控制台
	consoleN, consoleErr := lmw.Console.Write(p)
	if consoleErr != nil {
		return consoleN, consoleErr
	}

	// 写入文件
	fileN, fileErr := lmw.File.Write(p)
	if fileErr != nil {
		return fileN, fileErr
	}

	// 返回写入字节数和错误信息
	return len(p), nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	// 创建日志文件
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("无法创建日志文件: %v", err)
	} else {
		// 将日志同时输出到文件和控制台
		log.SetOutput(&LogMultiWriter{Console: os.Stdout, File: logFile})
		defer logFile.Close()
	}

	// Create an instance of the app structure
	app := NewWailsApp()

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "路由器切换工具",
		Width:  800,
		Height: 500,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
		Mac: &mac.Options{
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			About: &mac.AboutInfo{
				Title:   "路由器切换工具",
				Message: "© 2025 RouterSwitcher",
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}