<script setup lang="ts">
import { ref } from 'vue'

const emit = defineEmits<{
  send: [text: string]
}>()

const text = ref('')
const inputRef = ref<HTMLInputElement | null>(null)

function onSend() {
  const trimmed = text.value.trim()
  if (!trimmed) return
  emit('send', trimmed)
  text.value = ''
  inputRef.value?.focus()
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    onSend()
  }
}
</script>

<template>
  <div class="text-input-wrap">
    <el-input
      ref="inputRef"
      v-model="text"
      placeholder="输入英文消息..."
      size="large"
      @keydown="onKeydown"
    >
      <template #append>
        <el-button type="primary" @click="onSend" :disabled="!text.trim()">
          发送
        </el-button>
      </template>
    </el-input>
    <p class="hint">按 Enter 发送，Shift+Enter 换行</p>
  </div>
</template>

<style scoped>
.text-input-wrap {
  flex: 1;
}
.hint {
  font-size: 11px;
  color: #c0c4cc;
  margin-top: 4px;
}
</style>
