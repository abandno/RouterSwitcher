package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

var owner walk.Form

func main() {
	// 加载配置
	config, err := LoadConfig()
	if err != nil {
		log.Printf("加载配置失败: %v", err)
	}

	// 初始化应用
	app := &App{
		config: config,
	}

	// 运行应用
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

// App 应用程序结构
type App struct {
	config *Config
	mw     *walk.MainWindow
	tray   *walk.NotifyIcon
}

// Run 运行应用程序
func (app *App) Run() error {
	// 创建主窗口
	err := app.createMainWindow()
	if err != nil {
		return fmt.Errorf("创建主窗口失败: %v", err)
	}

	// 创建系统托盘
	err = app.createTrayIcon()
	if err != nil {
		return fmt.Errorf("创建系统托盘失败: %v", err)
	}

	// 启动网络监控
	go app.monitorNetwork()

	// 启动消息循环
	app.mw.Run()

	return nil
}

// createMainWindow 创建主窗口
func (app *App) createMainWindow() error {
	mw, err := walk.NewMainWindow()
	if err != nil {
		return err
	}
	app.mw = mw
	owner = mw
	return nil
}

// createTrayIcon 创建系统托盘图标
func (app *App) createTrayIcon() error {
	ni, err := walk.NewNotifyIcon(app.mw)
	if err != nil {
		return err
	}
	app.tray = ni

	// 设置托盘图标
	icon, err := walk.NewIconFromResourceId(3) // 使用默认图标
	if err != nil {
		icon = walk.IconApplication()
	}
	ni.SetIcon(icon)
	ni.SetToolTip("路由器切换工具")

	// 添加托盘菜单项
	var adaptiveIPAction, dynamicIPAction, staticIPAction *walk.Action
	
	adaptiveIPAction = walk.NewAction()
	adaptiveIPAction.SetText("自适应IP")
	adaptiveIPAction.Triggered().Attach(func() {
		app.config.IPMode = "adaptive"
		adaptiveIPAction.SetChecked(true)
		dynamicIPAction.SetChecked(false)
		staticIPAction.SetChecked(false)
		// 保存配置文件
		err := SaveConfig(app.config)
		if err != nil {
			log.Println("保存配置文件失败: ", err)
		}
		// 触发自适应检查
		go app.checkAndSwitch()
	})

	dynamicIPAction = walk.NewAction()
	dynamicIPAction.SetText("动态IP")
	dynamicIPAction.Triggered().Attach(func() {
		app.config.IPMode = "dynamic"
		// 保存配置文件
		err := SaveConfig(app.config)
		if err != nil {
			log.Println("保存配置文件失败: ", err)
		}
		app.switchToDHCP()
		dynamicIPAction.SetChecked(true)
		adaptiveIPAction.SetChecked(false)
		staticIPAction.SetChecked(false)
	})

	staticIPAction = walk.NewAction()
	staticIPAction.SetText("静态IP")
	staticIPAction.Triggered().Attach(func() {
		app.config.IPMode = "static"
		// 保存配置文件
		err := SaveConfig(app.config)
		if err != nil {
			log.Println("保存配置文件失败: ", err)
		}
		app.switchToStatic()
		staticIPAction.SetChecked(true)
		adaptiveIPAction.SetChecked(false)
		dynamicIPAction.SetChecked(false)
	})

	// 根据当前配置设置初始选中状态
	switch app.config.IPMode {
	case "adaptive":
		adaptiveIPAction.SetChecked(true)
		dynamicIPAction.SetChecked(false)
		staticIPAction.SetChecked(false)
	case "dynamic":
		dynamicIPAction.SetChecked(true)
		adaptiveIPAction.SetChecked(false)
		staticIPAction.SetChecked(false)
	case "static":
		staticIPAction.SetChecked(true)
		adaptiveIPAction.SetChecked(false)
		dynamicIPAction.SetChecked(false)
	}

	exitAction := walk.NewAction()
	exitAction.SetText("退出")
	exitAction.Triggered().Attach(func() {
		walk.App().Exit(0)
	})

	ni.ContextMenu().Actions().Add(adaptiveIPAction)
	ni.ContextMenu().Actions().Add(dynamicIPAction)
	ni.ContextMenu().Actions().Add(staticIPAction)
	ni.ContextMenu().Actions().Add(walk.NewSeparatorAction())
	ni.ContextMenu().Actions().Add(exitAction)

	// 显示托盘图标
	err = ni.SetVisible(true)
	if err != nil {
		return err
	}

	// 处理托盘图标点击事件
	ni.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
		if button == walk.LeftButton {
			// 左键单击显示配置界面
			app.showConfigDialog()
		}
	})

	return nil
}

// monitorNetwork 监控网络变化
func (app *App) monitorNetwork() {
	// if app.config.IPMode != "adaptive" {
	// 	fmt.Println("当前模式不是自适应模式，不执行网络监控")
	// 	return
	// }

	for {
		if app.config.IPMode == "adaptive" {
			app.checkAndSwitch()
		}
		time.Sleep(30 * time.Second) // 每30秒检查一次
	}
}

// checkAndSwitch 检查网络状态并切换配置
func (app *App) checkAndSwitch() {
	// // 重新加载配置以确保使用最新设置
	// config, err := LoadConfig()
	// if err == nil && config != nil {
	// 	app.config = config
	// } else {
	// 	log.Printf("重新加载配置失败: %v", err)
	// }

	// 只有在自适应模式下才进行自动切换
	switch mode := app.config.IPMode; mode {
	case "adaptive":
		// 连接到家庭局域网 且旁路由可达  设置静态IP
		if app.isConnectedToHomeNetwork() && app.isSideRouterReachable() {
			app.switchToStatic()
		} else {
			// 不是家庭局域网 或 旁路由不可达，切回动态IP
			app.switchToDHCP()
		}
	case "static":
		// 强制使用静态IP
		app.switchToStatic()
	case "dynamic":
		// 强制使用动态IP
		app.switchToDHCP()
	}
}

// isConnectedToHomeNetwork 检查是否连接到家庭局域网
func (app *App) isConnectedToHomeNetwork() bool {
	// 执行命令获取当前WiFi信息
	cmd := exec.Command("netsh", "wlan", "show", "interfaces")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("执行netsh命令失败: %v", err)
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
				if currentSSID == app.config.HomeSSID {
					return true
				}
			}
		}
	}

	return false
}

// isSideRouterReachable 检查旁路由是否可达
func (app *App) isSideRouterReachable() bool {
	// 使用系统ping命令检测旁路由地址是否可达
	addr := app.config.Gateway
	return Ping(addr)
}

// switchToStatic 切换到静态IP模式
func (app *App) switchToStatic() {
	log.Printf("开始切换静态IP")
	
	// 获取活动网络接口
	iface, err := GetActiveInterface()
	if err != nil {
		log.Printf("获取网络接口失败: %v", err)
		return
	}
	
	// 检查当前是否已经是目标静态IP配置
	isStatic, err := GetCurrentStaticIPConfig(iface, app.config.StaticIP, app.config.Gateway, app.config.DNS)
	if err == nil && isStatic {
		log.Printf("当前已经是目标静态IP配置, 无需重复设置: IP=%s, Gateway=%s, DNS=%s\n", app.config.StaticIP, app.config.Gateway, app.config.DNS)
		return
	}
	
	// 设置静态IP (这里使用默认子网掩码 255.255.255.0)
	err = SetStaticIP(iface, app.config.StaticIP, "255.255.255.0", app.config.Gateway, app.config.DNS)
	if err != nil {
		log.Printf("设置静态IP失败: %v", err)
		return
	}
	
	log.Printf("成功切换到静态IP模式: IP=%s, Gateway=%s, DNS=%s\n", app.config.StaticIP, app.config.Gateway, app.config.DNS)
}

// switchToDHCP 切换到自动获取IP模式
func (app *App) switchToDHCP() {
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

// showConfigDialog 显示配置对话框
func (app *App) showConfigDialog() {
	var dlg *walk.Dialog
	var db *walk.DataBinder
	var acceptPB, cancelPB *walk.PushButton
	
	const spacing = 10
	
	Dialog{
		AssignTo: &dlg,
		Title: "路由器切换工具 - 配置",
		DataBinder: DataBinder{
			AssignTo: &db,
			DataSource: app.config,
		},
		MinSize: Size{Width: 300, Height: 200},
		Layout: VBox{
			Margins: Margins{Left: spacing, Top: spacing, Right: spacing, Bottom: spacing},
			Spacing: spacing,
		},
		Children: []Widget{
			Composite{
				Layout: Grid{Columns: 2},
				Children: []Widget{
					Label{Text: "IP模式:"},
					ComboBox{
						Value: Bind("IPMode"),
						Model: []string{"adaptive", "dynamic", "static"},
					},
					
					Label{Text: "家庭WiFi SSID:"},
					LineEdit{Text: Bind("HomeSSID")},
					
					Label{Text: "静态IP地址:"},
					LineEdit{Text: Bind("StaticIP")},
					
					Label{Text: "网关地址:"},
					LineEdit{Text: Bind("Gateway")},
					
					Label{Text: "DNS服务器:"},
					LineEdit{Text: Bind("DNS")},
					
					Label{Text: "自动切换:"},
					CheckBox{Checked: Bind("AutoSwitch"), Text: "启用"},
					
					Label{Text: "开机自启:"},
					CheckBox{Checked: Bind("AutoStart"), Text: "启用"},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						AssignTo: &acceptPB,
						Text: "保存",
						OnClicked: func() {
							if err := db.Submit(); err != nil {
								log.Print(err)
								return
							}
							
							// 保存配置到文件
							SaveConfig(app.config)
							
							// 根据新的配置触发网络检查
							go app.checkAndSwitch()
							
							dlg.Accept()
						},
					},
					PushButton{
						AssignTo: &cancelPB,
						Text: "取消",
						OnClicked: func() {
							dlg.Cancel()
						},
					},
				},
			},
		},
	}.Create(owner)
}