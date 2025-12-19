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
          <label class="radio-label" title="自动识别当前网络环境，在DHCP（动态）和静态IP之间切换">
            <input 
              type="radio" 
              v-model="config.IPMode" 
              value="adaptive"
            >
            自适应
          </label>
          <label class="radio-label" title="DHCP模式，IP自动分配">
            <input 
              type="radio" 
              v-model="config.IPMode" 
              value="dynamic"
            >
            动态IP
          </label>
          <label class="radio-label" title="将固定使用您下面配置IP">
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
          placeholder="输入局域网WiFi名称"
          :class="{ 'invalid': validationErrors.includes('ssid') }"
          style="font-weight: bold; font-size: 16px;"
        >
        <div v-if="validationErrors.includes('ssid')" class="error-message">
          SSID不能为空
        </div>
      </div>

      <!-- 其他表单项保持不变 -->
      <fieldset>
        <legend>静态IP配置</legend>
        
        <div class="form-group">
          <label for="staticIP">IP地址:</label>
          <IpInput
            id="staticIP"
            v-model="config.StaticIP"
            :class="{ 'invalid': validationErrors.includes('staticIP') }"
          />
          <div v-if="validationErrors.includes('staticIP')" class="error-message">
            IP地址不能为空
          </div>
        </div>
        
        <div class="form-group">
          <label for="gateway">默认网关:</label>
          <IpInput
            id="gateway"
            v-model="config.Gateway"
            :class="{ 'invalid': validationErrors.includes('gateway') }"
          />
          <div v-if="validationErrors.includes('gateway')" class="error-message">
            网关不能为空
          </div>

        </div>
        
        <div class="form-group">
          <label for="dns">DNS:</label>
          <IpInput
            id="dns"
            v-model="config.DNS"
            :class="{ 'invalid': validationErrors.includes('dns') }"
          />
          <div v-if="validationErrors.includes('dns')" class="error-message">
            DNS不能为空
          </div>
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
          <span class="label">WiFi:</span>
          <span class="value-text">{{ networkStatus.WiFiName || '未知' }}</span>
          <span class="value-text"></span>
          <span :class="['value-status', networkStatus.WiFiConnected ? 'connected' : 'disconnected', 'align-right']">
            {{ networkStatus.WiFiConnected ? '连接' : '断开' }}
          </span>
        </li>
        <li>
          <span class="label">IP:</span>
          <span class="value-text">{{ networkStatus.IPAddress || '未知' }}</span>
          <span class="value-text">{{ networkStatus.IPAssignment || '未知' }}</span>
          <span class="align-right">&nbsp;</span>
        </li>
        <li>
          <span class="label">网关:</span>
          <span class="value-text">{{ networkStatus.Gateway || '未知' }}</span>
          <span class="value-text"></span>
          <span :class="['value-status', networkStatus.GatewayReachable ? 'connected' : 'disconnected', 'align-right']">
            {{ networkStatus.GatewayReachable ? '连接' : '断开' }}
          </span>
        </li>
        <li>
          <span class="label">DNS:</span>
          <span class="value-text">{{ networkStatus.DNS || '未知' }}</span>
          <span class="value-text">{{ networkStatus.DNSAssignment || '未知' }}</span>
          <span :class="['value-status', networkStatus.DNSReachable ? 'connected' : 'disconnected', 'align-right']">
            {{ networkStatus.DNSReachable ? '连接' : '断开' }}
          </span>
        </li>
        <!--<li>
          <span class="label">IP分配:</span>
          <span class="value-text">{{ networkStatus.IPAssignment || '未知' }}</span>
        </li>
        <li>
          <span class="label">DNS分配:</span>
          <span class="value-text">{{ networkStatus.DNSAssignment || '未知' }}</span>
        </li>
        -->
      </ul>
    </div>
  </div>
</template>

<script>
import IpInput from './IpInput.vue'
import { GetConfig, UpdateConfig, SwitchToStatic, SwitchToDHCP, IsConnectedToHomeNetwork, IsSideRouterReachable, GetNetworkStatus } from '../../bindings/RouterSwitcher/wailsapp'
import { Events } from '@wailsio/runtime'
import { isValidIp } from '../utils';

export default {
  name: 'ConfigManager',
  components: {
    IpInput
  },
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
      networkStatus: {
        WiFiName: '未知',
        WiFiConnected: false,
        IPAddress: '未知',
        Gateway: '未知',
        GatewayReachable: false,
        DNS: '未知',
        DNSReachable: false,
        IPAssignment: '未知',
        DNSAssignment: '未知'
      },
      configUpdatedOff: null,
      windowShownOff: null,
      windowHiddenOff: null,
      networkStatusTimer: null,
      validationErrors: [] // 用于存储验证错误信息
    }
  },
  watch: {
    // 监听config变化，实时验证表单
    config: {
      handler() {
        this.validateForm();
      },
      deep: true
    }
  },
  async mounted() {
    setTimeout(async () => {
      console.log('==mounted', this.config)
      await this.loadConfig()
    }, 200)
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
    validateForm() {
      // 清空之前的验证错误
      this.validationErrors = [];
      
      // 根据需求文档进行验证：
      // 1. 自适应时，SSID 和 静态IP配置区域 必填
      // 2. 静态IP时，静态IP配置区域 必填
      
      let requiredStaticIpConf = false
      if (this.config.IPMode === 'adaptive') {
        // 检查SSID是否为空
        if (!this.config.HomeSSID.trim()) {
          this.validationErrors.push('ssid');
        }
        requiredStaticIpConf = true
      } else if (this.config.IPMode === 'static') {
        requiredStaticIpConf = true
      }

      if (requiredStaticIpConf) {
        // 检查静态IP配置是否为空
        if (!isValidIp(this.config.StaticIP)) {
          this.validationErrors.push('staticIP');
        }
        if (!isValidIp(this.config.Gateway)) {
          this.validationErrors.push('gateway');
        }
        if (!isValidIp(this.config.DNS)) {
          this.validationErrors.push('dns');
        }
      }
      // console.log("this.validationErrors", this.validationErrors);
      
      // 返回验证是否通过
      return this.validationErrors.length === 0;
    },
    async saveConfig() {
      console.log('saveConfig', this.config);
      
      // 执行表单验证
      if (!this.validateForm()) {
        // 如果验证失败，显示错误信息并阻止提交
        alert('表单验证失败，请检查必填项');
        return;
      }
      
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
      // console.log('updateNetworkStatus start', this.isConnectedToHome, this.isSideRouterReachable)
      try {
        this.isConnectedToHome = await IsConnectedToHomeNetwork()
        this.isSideRouterReachable = await IsSideRouterReachable()
        
        // 获取详细网络状态
        const status = await GetNetworkStatus()
        if (status) {
          this.networkStatus = status
        }
        // console.log('updateNetworkStatus success', this.isConnectedToHome, this.isSideRouterReachable)
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
  width: 100%;
  margin: 0 auto;
  padding: 20px;
  box-sizing: border-box;
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
  width: 100%;
  box-sizing: border-box;
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

.form-item-block input[type="text"]:focus {
  outline: none;
  border-color: #007bff;
  box-shadow: 0 0 0 2px rgba(0, 123, 255, 0.25);
}

/* 错误状态样式 */
.form-item-block input[type="text"].invalid,
.form-item-block input[type="password"].invalid {
  border-color: #dc3545;
  box-shadow: 0 0 0 2px rgba(220, 53, 69, 0.25);
}

.form-item-block input[type="text"].invalid:focus,
.form-item-block input[type="password"].invalid:focus {
  border-color: #dc3545;
  box-shadow: 0 0 0 2px rgba(220, 53, 69, 0.25);
}

/* fieldset 错误状态样式 */
fieldset.invalid {
  border-color: #dc3545;
  box-shadow: 0 0 0 2px rgba(220, 53, 69, 0.25);
}

fieldset legend {
  font-weight: bold;
  padding: 0 5px;
}

.form-item-block .radio-group {
  display: flex;
  gap: 15px;
  flex-wrap: nowrap;
  border: none;
  padding: 0;
  margin: 0;
}

.buttons {
  display: flex;
  gap: 10px;
  justify-content: center;
  margin-top: 20px;
}

.buttons button {
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  background-color: #007bff;
  color: white;
  cursor: pointer;
  font-size: 16px;
}

.buttons button:hover:not(:disabled) {
  background-color: #0056b3;
}

.buttons button:disabled {
  background-color: #6c757d;
  cursor: not-allowed;
}

.status {
  margin-top: 30px;
  padding: 15px;
  border: 1px solid #ddd;
  border-radius: 4px;
  background-color: #f8f9fa;
  color: gray;
}

.status h3 {
  margin-top: 0;
}

.status ul {
  list-style-type: none;
  padding: 0;
}

.status li {
  display: grid;
  grid-template-columns: 80px 1fr 120px 60px;
  gap: 10px;
  align-items: center;
  margin-bottom: 10px;
  padding-bottom: 10px;
  border-bottom: 1px solid #eee;
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

.status span {
  text-align: left;
}
.status .value-text {
  font-weight: 500;
  color: #495057;
  text-align: left;
}

.status .value-status {
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 4px;
}

.status .align-right {
  text-align: right;
  justify-self: end;
}

.status .value-status.connected {
  color: #28a745;
}

.status .value-status.disconnected {
  color: #dc3545;
}

/* 错误消息样式 */
.error-message {
  color: #dc3545;
  font-size: 14px;
  margin-top: 5px;
}
</style>
