<template>
  <div class="config-manager">
    <h1>路由器切换工具</h1>
    
    <form @submit.prevent="saveConfig">
      <!-- 每个表单项占一行 -->
      <div class="form-item-block auto-start">
        <label>
          <input 
            type="checkbox" 
            v-model="config.AutoStart"
          >
          开机启动
        </label>
      </div>
      
      <div class="form-item-block ip-mode">
        <label>IP模式:</label>
        <div class="radio-group">
          <label class="radio-label">
            <input 
              type="radio" 
              v-model="config.IPMode" 
              value="adaptive"
            >
            自适应
          </label>
          <label class="radio-label">
            <input 
              type="radio" 
              v-model="config.IPMode" 
              value="dynamic"
            >
            动态IP
          </label>
          <label class="radio-label">
            <input 
              type="radio" 
              v-model="config.IPMode" 
              value="static"
            >
            静态IP
          </label>
        </div>
      </div>

      <div class="form-item-block ssid">
        <label for="homeSSID">使用静态IP模式的网络 (SSID):</label>
        <input 
          id="homeSSID" 
          type="text" 
          v-model="config.HomeSSID"
        >
      </div>

      <!-- 其他表单项保持不变 -->
      <fieldset>
        <legend>静态IP配置</legend>
        
        <div class="form-group">
          <label for="staticIP">IP地址:</label>
          <input 
            id="staticIP" 
            type="text" 
            v-model="config.StaticIP"
          >
        </div>
        
        <div class="form-group">
          <label for="gateway">默认网关:</label>
          <input 
            id="gateway" 
            type="text" 
            v-model="config.Gateway"
          >
        </div>
        
        <div class="form-group">
          <label for="dns">DNS:</label>
          <input 
            id="dns" 
            type="text" 
            v-model="config.DNS"
          >
        </div>
      </fieldset>

      <div class="buttons">
        <button type="submit">保存</button>
        <!-- <button type="button" @click="switchToStatic" :disabled="switching">切换到静态IP</button>
        <button type="button" @click="switchToDHCP" :disabled="switching">切换到动态IP</button> -->
      </div>
    </form>

    <div class="status">
      <h3>当前网络状态</h3>
      <ul>
        <li>
          <span class="label">是否连接到家庭网络:</span>
          <span :class="['value', isConnectedToHome ? 'yes' : 'no']">{{ isConnectedToHome ? '是' : '否' }}</span>
        </li>
        <li>
          <span class="label">旁路由是否可达:</span>
          <span :class="['value', isSideRouterReachable ? 'yes' : 'no']">{{ isSideRouterReachable ? '是' : '否' }}</span>
        </li>
      </ul>
    </div>
  </div>
</template>

<script>
import { GetConfig, UpdateConfig, SwitchToStatic, SwitchToDHCP, IsConnectedToHomeNetwork, IsSideRouterReachable } from '../../bindings/RouterSwitcher/wailsapp'
import { Events } from '@wailsio/runtime'

export default {
  name: 'ConfigManager',
  data() {
    return {
      config: {
        HomeSSID: '',
        StaticIP: '',
        Gateway: '',
        DNS: '',
        AutoStart: false,
        IPMode: 'adaptive'
      },
      switching: false,
      isConnectedToHome: false,
      isSideRouterReachable: false,
      configUpdatedOff: null,
      windowShownOff: null,
      windowHiddenOff: null,
      networkStatusTimer: null
    }
  },
  async mounted() {
    setTimeout(async () => {
      console.log('==mounted', this.config)
      await this.loadConfig()
    }, 1 * 1000)
    // setInterval(async () => {
    //   await this.loadConfig()
    // }, 5 * 1000)
    // 监听后端发出的 configUpdated 事件，收到后重新加载配置
    this.configUpdatedOff = Events.On('configUpdated', () => {
      console.log('收到 configUpdated 事件，重新加载配置')
      this.loadConfig()
    })

    // 等待 Wails 运行时准备就绪（如需严格等待可取消注释）
    // await this.waitForRuntime()
    // await this.loadConfig()
    await this.updateNetworkStatus()
    // 监听后端发出的 configUpdated 事件，收到后重新加载配置
    this.configUpdatedOff = Events.On('configUpdated', () => {
      console.log('收到 configUpdated 事件，重新加载配置')
      this.loadConfig()
    })

    // 监听窗口显示事件：刷新配置 + 网络状态，并启动定时器
    this.windowShownOff = Events.On('windowShown', () => {
      console.log('收到 windowShown 事件，刷新配置并启动网络状态定时器')
      this.loadConfig()
      this.updateNetworkStatus()
      this.startNetworkStatusTimer()
    })

    // 监听窗口隐藏事件：停止定时器
    this.windowHiddenOff = Events.On('windowHidden', () => {
      console.log('收到 windowHidden 事件，停止网络状态定时器')
      this.stopNetworkStatusTimer()
    })

    // 首次挂载时，立即刷新一次网络状态并启动定时器
    await this.updateNetworkStatus()
    this.startNetworkStatusTimer()
  },
  beforeUnmount() {
    // 组件卸载时取消事件监听，避免内存泄漏
    if (this.configUpdatedOff) {
      this.configUpdatedOff()
      this.configUpdatedOff = null
    }
    if (this.windowShownOff) {
      this.windowShownOff()
      this.windowShownOff = null
    }
    if (this.windowHiddenOff) {
      this.windowHiddenOff()
      this.windowHiddenOff = null
    }
    this.stopNetworkStatusTimer()
  },
  methods: {
    async waitForRuntime() {
      // 等待 Wails 运行时初始化
      return new Promise((resolve) => {
        const checkRuntime = () => {
          if (window.runtime || (window.go && window.go.main && window.go.main.WailsApp)) {
            resolve()
          } else {
            setTimeout(checkRuntime, 50)
          }
        }
        checkRuntime()
      })
    },
    async loadConfig() {
      console.log('loadConfig')
      try {
        const result = await GetConfig()
        console.log('GetConfig result:', result)
        if (!result) {
          console.warn('GetConfig 返回为空, 使用当前默认配置', this.config)
          return
        }
        // 合并后端返回的配置到本地 config, 避免响应式丢失
        this.config = {
          ...this.config,
          ...result
        }
        console.log('loadConfig success', this.config)
      } catch (err) {
        console.error('加载配置失败:', err)
      }
    },
    async saveConfig() {
      console.log('saveConfig', this.config)
      try {
        await UpdateConfig(this.config)
        alert('配置保存成功')
      } catch (err) {
        console.error('保存配置失败:', err)
        alert('保存配置失败: ' + err)
      }
    },
    async switchToStatic() {
      this.switching = true
      try {
        await SwitchToStatic()
        alert('已切换到静态IP模式')
      } catch (err) {
        console.error('切换到静态IP失败:', err)
        alert('切换到静态IP失败: ' + err)
      } finally {
        this.switching = false
      }
    },
    async switchToDHCP() {
      this.switching = true
      try {
        await SwitchToDHCP()
        alert('已切换到动态IP模式')
      } catch (err) {
        console.error('切换到动态IP失败:', err)
        alert('切换到动态IP失败: ' + err)
      } finally {
        this.switching = false
      }
    },
    startNetworkStatusTimer() {
      if (this.networkStatusTimer) {
        clearInterval(this.networkStatusTimer)
      }
      this.networkStatusTimer = setInterval(this.updateNetworkStatus, 5000)
    },
    stopNetworkStatusTimer() {
      if (this.networkStatusTimer) {
        clearInterval(this.networkStatusTimer)
        this.networkStatusTimer = null
      }
    },
    async updateNetworkStatus() {
      console.log('updateNetworkStatus start', this.isConnectedToHome, this.isSideRouterReachable)
      try {
        this.isConnectedToHome = await IsConnectedToHomeNetwork()
        this.isSideRouterReachable = await IsSideRouterReachable()
        console.log('updateNetworkStatus success', this.isConnectedToHome, this.isSideRouterReachable)
      } catch (err) {
        console.error('获取网络状态失败:', err)
      }
    }
  }
}
</script>

<style scoped>
.config-manager {
  max-width: 600px;
  margin: 0 auto;
  padding: 20px;
}

.form-group {
  margin-bottom: 15px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
  font-weight: bold;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  box-sizing: border-box;
}

fieldset {
  border: 1px solid #ddd;
  border-radius: 4px;
  margin-bottom: 20px;
  padding: 15px;
}

legend {
  font-weight: bold;
  padding: 0 10px;
}

.radio-group {
  display: flex;
  gap: 15px;
  flex-wrap: wrap;
}

.radio-label {
  display: flex;
  align-items: center;
  font-weight: normal;
  gap: 5px;
}

.radio-label input[type="radio"] {
  width: auto;
  margin: 0;
}

.buttons {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
  margin-top: 20px;
}

.buttons button {
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  background-color: #007bff;
  color: white;
}

.buttons button:disabled {
  background-color: #cccccc;
  cursor: not-allowed;
}

.buttons button[type="submit"] {
  background-color: #28a745;
}

.status {
  margin-top: 30px;
  padding: 20px;
  background-color: #f8f9fa;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  border: 1px solid #e9ecef;
}

.status h3 {
  margin-top: 0;
  margin-bottom: 15px;
  color: #343a40;
  font-size: 1.1em;
}

.status ul {
  margin: 0;
  padding-left: 0;
  list-style: none;
}

.status li {
  padding: 8px 0;
  border-bottom: 1px solid #e9ecef;
  display: flex;
  justify-content: space-between;
}

.status li:last-child {
  border-bottom: none;
}

.status .label {
  font-weight: 500;
  color: #495057;
}

.status .value {
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 4px;
}

.status .value.yes {
  background-color: #d4edda;
  color: #155724;
}

.status .value.no {
  background-color: #f8d7da;
  color: #721c24;
}

.form-item-block {
  margin-top: 15px;
  margin-bottom: 15px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.form-item-block.auto-start label {
  display: flex;
  align-items: center;
  gap: 5px;
  font-weight: normal;
}

.form-item-block.ip-mode {
  flex-direction: row;
}

.form-item-block.ip-mode > label {
  font-weight: bold;
  white-space: nowrap;
}

.form-item-block.ssid {
  flex-direction: row;
}

.form-item-block.ssid > label {
  font-weight: bold;
  white-space: nowrap;
}

.form-item-block input[type="text"] {
  flex: 1;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  box-sizing: border-box;
}

.form-item-block .radio-group {
  display: flex;
  gap: 15px;
  flex-wrap: nowrap;
  border: none;
  padding: 0;
  margin: 0;
}
</style>