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
import { GetConfig, SaveConfig, SwitchToStatic, SwitchToDHCP, IsConnectedToHomeNetwork, IsSideRouterReachable } from '../../wailsjs/go/main/WailsApp'

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
      isSideRouterReachable: false
    }
  },
  async mounted() {
    await this.loadConfig()
    await this.updateNetworkStatus()
    // 定期更新网络状态
    setInterval(this.updateNetworkStatus, 5000)
  },
  methods: {
    async loadConfig() {
      try {
        this.config = await GetConfig()
      } catch (err) {
        console.error('加载配置失败:', err)
      }
    },
    async saveConfig() {
      try {
        await SaveConfig(this.config)
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
    async updateNetworkStatus() {
      try {
        this.isConnectedToHome = await IsConnectedToHomeNetwork()
        this.isSideRouterReachable = await IsSideRouterReachable()
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