<template>
  <div class="modal-overlay" @click="closeModal">
    <div class="modal-content" @click.stop>
      <div class="modal-header">
        <h3>需要开启位置服务</h3>
        <button class="close-button" @click="closeModal">&times;</button>
      </div>
      <div class="modal-body">
        <p>检测到位置服务被禁用，无法获取WiFi信息。</p>

        <p>请按以下步骤开启位置服务：</p>
        <ol>
          <li>打开Windows设置 (Win + I)</li>
          <li>进入「隐私和安全」->「位置」</li>
          <li>开启「位置服务」开关</li>
        </ol>

        <p>或者使用以下方法快速打开位置设置页面：</p>
        <div class="command-section">
          <div class="command-text" ref="commandText1">start ms-settings:privacy-location</div>
          <button class="copy-button" @click="copyCommand(1)">复制命令</button>
        </div>
        <div class="command-section">
          <span>WIN+R > 输入: </span>
          <div class="command-text" ref="commandText2">ms-settings:privacy-location</div>
          <button class="copy-button" @click="copyCommand(2)">复制命令</button>
        </div>

        <div class="notification" v-if="showNotification">
          命令已复制到剪贴板！
        </div>
      </div>
      <div class="modal-footer">
        <button class="btn btn-primary" @click="openSettings">自动打开设置</button>
        <button class="btn btn-secondary" @click="closeModal">关闭</button>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'LocationServiceDialog',
  data() {
    return {
      showNotification: false
    }
  },
  methods: {
    closeModal() {
      this.$emit('close', 'cancel');
    },
    
    openSettings() {
      // 调用Go后端方法打开设置页面
      window.backend.WailsApp.OpenLocationSettings().then(() => {
        this.$emit('close', 'ok');
      }).catch((error) => {
        console.error('Failed to open location settings:', error);
        // 即使出错也关闭对话框
        this.$emit('close', 'ok');
      });
    },
    
    copyCommand(t) {
      const commandText = t==1 ? this.$refs.commandText1.innerText : this.$refs.commandText2.innerText;
      
      // 使用现代Clipboard API
      if (navigator.clipboard && window.isSecureContext) {
        navigator.clipboard.writeText(commandText).then(() => {
          this.showNotification = true;
          setTimeout(() => {
            this.showNotification = false;
          }, 2000);
        }).catch(err => {
          console.error('Failed to copy text: ', err);
          // 降级到传统方法
          this.fallbackCopyTextToClipboard(commandText);
        });
      } else {
        // 降级到传统方法
        this.fallbackCopyTextToClipboard(commandText);
      }
    },
    
    fallbackCopyTextToClipboard(text) {
      const textArea = document.createElement("textarea");
      textArea.value = text;
      textArea.style.position = "fixed";
      document.body.appendChild(textArea);
      textArea.focus();
      textArea.select();
      
      try {
        const successful = document.execCommand('copy');
        if (successful) {
          this.showNotification = true;
          setTimeout(() => {
            this.showNotification = false;
          }, 2000);
        }
      } catch (err) {
        console.error('Fallback: Oops, unable to copy', err);
      }
      
      document.body.removeChild(textArea);
    }
  }
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  border-bottom: 1px solid #eee;
}

.modal-header h3 {
  margin: 0;
  color: #333;
}

.close-button {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #999;
  padding: 0;
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.close-button:hover {
  color: #333;
}

.modal-body {
  padding: 1rem;
}

.modal-body p {
  margin: 0 0 1rem 0;
  line-height: 1.5;
}

.modal-body ol {
  margin: 0 0 1rem 1.5rem;
  line-height: 1.5;
}

.command-section {
  background-color: #f8f9fa;
  border: 1px solid #e9ecef;
  border-radius: 4px;
  padding: 1rem;
  margin: 1rem 0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.command-text {
  font-family: 'Courier New', monospace;
  user-select: text;
  cursor: text;
  flex-grow: 1;
  margin-right: 1rem;
  word-break: break-all;
}

.copy-button {
  background-color: #007bff;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 0.5rem 1rem;
  cursor: pointer;
  white-space: nowrap;
}

.copy-button:hover {
  background-color: #0056b3;
}

.notification {
  position: fixed;
  top: 20px;
  right: 20px;
  background-color: #28a745;
  color: white;
  padding: 1rem;
  border-radius: 4px;
  z-index: 1001;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  padding: 1rem;
  border-top: 1px solid #eee;
}

.btn {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.btn-primary {
  background-color: #007bff;
  color: white;
}

.btn-primary:hover {
  background-color: #0056b3;
}

.btn-secondary {
  background-color: #6c757d;
  color: white;
}

.btn-secondary:hover {
  background-color: #545b62;
}

@media (max-width: 600px) {
  .command-section {
    flex-direction: column;
    align-items: stretch;
  }
  
  .command-text {
    margin-right: 0;
    margin-bottom: 0.5rem;
  }
  
  .modal-footer {
    flex-direction: column;
  }
  
  .btn {
    width: 100%;
  }
}
</style>