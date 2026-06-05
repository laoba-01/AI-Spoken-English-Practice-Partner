<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useConversationStore } from '@/stores/conversation'
import { useWebSocketStore } from '@/stores/websocket'
import AppLayout from '@/layout/AppLayout.vue'
import ChatWindow from '@/components/chat/ChatWindow.vue'
import TextInput from '@/components/chat/TextInput.vue'
import VoiceRecorder from '@/components/chat/VoiceRecorder.vue'
import type { WSMessage, Message } from '@/types'

const route = useRoute()
const router = useRouter()
const convStore = useConversationStore()
const wsStore = useWebSocketStore()

const conversationId = ref(Number(route.params.id))
const connected = ref(false)
const streamingContent = ref('')

// 缓存当前流式消息的文本（组件内部使用）
let currentAssistantContent = ''

onMounted(async () => {
  // 加载历史消息
  await convStore.fetchMessages(conversationId.value)

  // 建立 WebSocket 连接
  try {
    await wsStore.connect(conversationId.value)
    connected.value = true
  } catch {
    connected.value = false
  }

  // 注册 WS 消息回调
  wsStore.onMessage(handleWSMessage)
})

onUnmounted(() => {
  wsStore.disconnect()
})

function handleWSMessage(msg: WSMessage) {
  switch (msg.type) {
    case 'asr_result':
      // 语音识别结果已由 store 保存
      break
    case 'llm_chunk':
      currentAssistantContent += msg.text || ''
      streamingContent.value = currentAssistantContent
      break
    case 'audio_result':
      // TTS 语音地址已由 store 保存
      break
    case 'error':
      ElMessage.error(msg.message || '系统繁忙')
      break
  }
}

function onSendText(text: string) {
  // 添加用户消息到列表
  const userMsg: Message = {
    id: Date.now(),
    conversation_id: conversationId.value,
    role: 'user',
    content: text,
    created_at: new Date().toISOString(),
  }
  convStore.addMessage(userMsg)

  // 重置流式缓存
  currentAssistantContent = ''
  streamingContent.value = ''

  // 发送到 WebSocket
  wsStore.sendText(text)
}

function onSendVoice(blob: Blob) {
  currentAssistantContent = ''
  streamingContent.value = ''
  wsStore.sendVoice(blob)
}

// 中间态的 AI 流式消息（插入列表底部）
const displayMessages = computed(() => {
  const msgs = [...convStore.messages]
  if (streamingContent.value) {
    msgs.push({
      id: -1,
      conversation_id: conversationId.value,
      role: 'assistant',
      content: streamingContent.value,
      created_at: new Date().toISOString(),
    })
  }
  return msgs
})
</script>

<template>
  <AppLayout>
    <div class="chat-page">
      <!-- 顶部状态栏 -->
      <div class="chat-header">
        <div class="header-left">
          <el-button text @click="router.push('/')">
            <el-icon><ArrowLeft /></el-icon>
          </el-button>
          <span class="scene-badge">
            <el-tag :type="convStore.currentScene === 'daily' ? '' : convStore.currentScene === 'business' ? 'warning' : 'danger'" size="small">
              {{ convStore.currentScene === 'daily' ? '日常对话' : convStore.currentScene === 'business' ? '商务英语' : '考试模拟' }}
            </el-tag>
          </span>
        </div>
        <div class="header-center">
          <span class="connection-status">
            <span class="dot" :class="{ online: connected }"></span>
            {{ connected ? '已连接' : '连接中...' }}
          </span>
        </div>
      </div>

      <!-- 消息列表 -->
      <ChatWindow :messages="displayMessages" />

      <!-- 输入区域 -->
      <div class="input-area">
        <VoiceRecorder @send="onSendVoice" />
        <div class="divider">
          <span>或</span>
        </div>
        <TextInput @send="onSendText" />
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.chat-page {
  display: flex;
  flex-direction: column;
  height: 100vh;
}

.chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 20px;
  background: #fff;
  border-bottom: 1px solid #f0f0f0;
  flex-shrink: 0;
}
.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}
.header-center {
  font-size: 14px;
  color: #909399;
}
.connection-status {
  display: flex;
  align-items: center;
  gap: 6px;
}
.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #e6a23c;
}
.dot.online {
  background: #67c23a;
}

.input-area {
  background: #fff;
  border-top: 1px solid #f0f0f0;
  padding: 16px 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  flex-shrink: 0;
}
.divider {
  font-size: 13px;
  color: #c0c4cc;
}
</style>
