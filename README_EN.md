# RouterSwitcher - Router Switching Tool

An intelligent network IP configuration switching tool that automatically switches between static IP and dynamic IP configurations based on network environment.

## 📋 Project Introduction

RouterSwitcher is a desktop application developed with Wails v3, primarily designed to automatically manage network IP configurations on Windows systems. When you connect to your home network, the application can automatically switch to static IP configuration to use a bypass router; when connecting to other networks, it automatically switches back to dynamic IP (DHCP) mode, ensuring flexibility and convenience in network connectivity.

**Typical Use Case**: Connect to your bypass router via gateway in home local area network; use dynamic IP (DHCP) mode when on external networks.

## ✨ Key Features

- **Three IP Modes**
  - **Adaptive Mode**: Automatically switches IP configuration based on network environment
  - **Dynamic IP Mode**: Forces DHCP to automatically obtain IP
  - **Static IP Mode**: Forces static IP configuration

- **Intelligent Network Detection**
  - Automatically detects currently connected WiFi SSID
  - Detects bypass router reachability (via ping detection)
  - Automatically checks network status every 30 seconds (only in adaptive mode)

- **System Tray Support**
  - Runs minimized to system tray
  - Quick IP mode switching via tray menu
  - Click tray icon to show/hide main window

- **Auto Start**
  - Supports setting automatic startup on boot
  - Implemented using Windows Task Scheduler

- **Configuration Management**
  - Graphical configuration interface
  - Real-time network status display
  - Configuration automatically saved to local file

### Interface Demo

![Interface Demo](doc/image/界面演示.gif)

### System Tray

![Tray Icon](doc/image/托盘图标.png)

![Tray Right-Click Menu](doc/image/托盘图标-右键菜单.png)

## Usage scenario

![使用情景](doc/image/使用情景.png)

## 🚀 Installation and Usage

### Installation

#### Method 1: Download and Install

Release artifacts list:

- **RouterSwitcher.exe**: Double-click to run directly
- **RouterSwitcher-amd64-installer.exe**: Windows installer, use after installation

#### Method 2: Build from Source

1. **Clone the project**
   ```bash
   git clone <repository-url>
   cd RouterSwitcher
   
   # Or build production package
   wails3 task release
   # Package Windows installer
   wails3 task windows:package
   ```
   
   After building, the executable file will be located in the `bin/` directory.

### Usage Instructions

1. **First Run**
   - After running the program, a `config.json` configuration file will be generated in the program directory
   - Default configuration:
     - IP Mode: Adaptive
     - Home SSID: `HomeWiFi`
     - Static IP: `192.168.31.100`
     - Gateway: `192.168.31.2`
     - DNS: `192.168.31.2`

2. **Configuration Settings**
   - Click the system tray icon to open the configuration interface (*the configuration interface will not be displayed automatically*)
   - Modify configuration according to your network environment:
     - **Network using static IP mode (SSID)**: Enter your home WiFi name
     - **Static IP Configuration**: Set static IP address, gateway, and DNS server address
     - **IP Mode**: Select adaptive/dynamic IP/static IP mode
     - **Auto Start**: Check to automatically run the program on system startup
   - Click the "Save" button to save configuration, which will take effect immediately

3. **System Tray Operations**
   - **Right-click** the tray icon to quickly switch IP modes (adaptive/dynamic IP/static IP)
   - **Left-click** the tray icon to show/hide the main window
   - Select "Exit" to close the program

## ⚙️ Configuration

The configuration file `config.json` is located in the same directory as the program executable, with the following format:

```json
{
  "HomeSSID": "YourWiFiName",
  "StaticIP": "192.168.31.100",
  "Gateway": "192.168.31.2",
  "DNS": "192.168.31.2",
  "AutoStart": false,
  "IPMode": "adaptive"
}
```

### Configuration Items

- `HomeSSID`: Home WiFi SSID name. When connected to this WiFi, adaptive mode will attempt to switch to static IP configuration
- `StaticIP`: Static IP address. Ensure this IP address is not occupied in your local network
- `Gateway`: Gateway address (usually the IP address of the bypass router)
- `DNS`: DNS server address, can be a single address or multiple addresses (separated by commas)
- `AutoStart`: Whether to auto-start on boot (`true`/`false`)
- `IPMode`: IP mode selection
  - `adaptive`: Adaptive mode (recommended), automatically switches based on network environment
  - `dynamic`: Dynamic IP mode, forces DHCP to obtain IP
  - `static`: Static IP mode, forces the configured static IP

## ⚠️ Important Notes

1. **Administrator Privileges**
   - Modifying network configuration requires administrator privileges. It is recommended to run the program as administrator
   - If privileges are insufficient, IP switching operations may fail

2. **Location Service Permissions**
   - Windows system needs location services enabled to obtain WiFi SSID information
   - If location services are disabled, the program will prompt you to enable them
   - How to enable: Windows Settings → Privacy & Security → Location → Enable location services

3. **Network Interface**
   - The program automatically detects active network interfaces
   - If detection fails, please check whether the network connection is normal

4. **Configuration File**
   - Configuration file is saved in the same directory as the program executable
   - If the configuration file is corrupted, the program will use default configuration and recreate the configuration file

## 📝 Development Guide

**!!Welcome to submit PRs!!**

### 🛠️ Tech Stack

- **Backend**: Go 1.24+
- **Frontend**: Vue 3 + Vite
- **Framework**: Wails v3
- **Platform**: Windows (currently only supports Windows systems)

### 📦 System Requirements

- Windows 10/11
- Administrator privileges (required for modifying network configuration)
- Location service permissions (required for obtaining WiFi SSID information)

### Project Structure

```
RouterSwitcher/
├── main.go              # Main program entry and Wails application logic
├── config.go            # Configuration file read/write
├── autostart.go         # Auto-start management
├── network.go           # Network interface and IP configuration management
├── types.go             # Data structure definitions
├── wails.json           # Wails configuration file
├── go.mod               # Go module dependencies
├── frontend/            # Frontend code
│   ├── src/
│   │   ├── main.js      # Frontend entry
│   │   └── components/
│   │       └── ConfigManager.vue  # Configuration management component
│   ├── package.json     # Frontend dependencies
│   └── vite.config.js   # Vite configuration
└── build/               # Build-related files
```

### Development Commands

```bash
# Development mode (hot reload)
wails dev

# Build test version
wails3 build

# Build production version
wails3 task release

# Package Windows installer
wails3 task windows:package
```


