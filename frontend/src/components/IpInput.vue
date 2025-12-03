<template>
  <div class="ip-input">
    <div v-for="(part, index) in parts" :key="index" ref="inputGroup" class="ip-input-part-box">
      <input ref="inputs" type="text" inputmode="numeric" maxlength="3" class="ip-input-part" v-model="parts[index]"
        @input="onInput($event, index)" @keydown="onKeydown($event, index)" />
      <span v-if="index < 3" :key="'dot-' + index" class="ip-input-dot">·</span>
    </div>
  </div>
</template>

<script>

function split4Parts(value = '') {
  const ret = value.split('.').slice(0, 4)
  while (ret.length < 4) {
    ret.push('')
  }
  return ret
}

export default {
  name: 'IpInput',
  props: {
    modelValue: {
      type: String,
      default: ''
    }
  },
  emits: ['update:modelValue'],
  data() {
    const initialParts = split4Parts(this.modelValue)
    return {
      parts: initialParts
    }
  },
  watch: {
    modelValue(newVal) {
      const incoming = split4Parts(newVal)
      // 仅在外部值与当前值不一致时更新，避免循环
      if (incoming.join('.') !== this.parts.join('.')) {
        this.parts = incoming
      }
    }
  },
  methods: {
    emitValue() {
      // 允许中间空段存在，但过滤掉完全空的尾部点
      const joined = this.parts.join('.').replace(/\.*$/, '')
      this.$emit('update:modelValue', joined)
    },
    normalizePart(value) {
      // 去除非数字
      let v = value.replace(/\D/g, '')
      if (v === '') return ''
      let num = parseInt(v, 10)
      if (isNaN(num)) return ''
      if (num < 0) num = 0
      if (num > 255) num = 255
      return String(num)
    },
    focusNext(index) {
      const next = this.$refs.inputs[index + 1]
      if (next && typeof next.focus === 'function') {
        next.focus()
        next.select && next.select()
      }
    },
    onInput(event, index) {
      const raw = event.target.value
      console.log('==onInput raw', raw)
      const normalized = this.normalizePart(raw)
      // Vue 3 中数组是响应式的，可以直接通过下标赋值
      this.parts[index] = normalized
      this.emitValue()

      // 自动跳到下一个输入框：当长度达到 3 位且当前值在 0~255 内
      if (normalized.length === 3 && index < 3) {
        this.focusNext(index)
      }
    },
    onKeydown(event, index) {
      console.log("onKeydown", event)
      const key = event.key

      // 允许 Tab 的默认行为，让浏览器在子输入框之间顺序切换
      if (key === 'Tab') {
        return
      }

      // 左右方向键在 4 个框之间切换
      if (key === 'ArrowLeft' && event.target.selectionStart === 0 && index > 0) {
        event.preventDefault()
        const prev = this.$refs.inputs[index - 1]
        if (prev) {
          prev.focus()
          prev.selectionStart = prev.value.length
          prev.selectionEnd = prev.value.length
        }
        return
      }
      if ((key === 'ArrowRight' || key == '.') && event.target.selectionStart === event.target.value.length && index < 3) {
        event.preventDefault()
        const next = this.$refs.inputs[index + 1]
        if (next) {
          next.focus()
          next.selectionStart = 0
          next.selectionEnd = key == '.' ? next.value.length : 0
        }
        return
      }

      // 屏蔽非数字字符（保留控制键）
      if (!/[0-9]/.test(key) && key.length === 1) {
        event.preventDefault()
      }
    }
  }
}
</script>

<style scoped>
.ip-input {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-family: system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
}

.ip-input-part-box {
  display: flex;
  flex-direction: row;
  align-items: center;
}

.ip-input-part {
  width: 100%;
  padding: 6px 4px;
  border: 1px solid #ddd;
  border-radius: 4px;
  text-align: center;
  font-size: 13px;
  box-sizing: border-box;
}

.ip-input-part:focus {
  outline: none;
  border-color: #007bff;
  box-shadow: 0 0 0 1px rgba(0, 123, 255, 0.2);
}

.ip-input-dot {
  margin: 0 2px;
  user-select: none;
  color: #555;
  font-weight: 550;
}
</style>
