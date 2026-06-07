<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
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
const isAIThinking = ref(false)       // AI 思考中指示
const lastAIAudioUrl = ref('')        // 最新的 AI 语音 URL，用于自动播放
const interimASRText = ref('')        // ASR 流式中间识别文本

onMounted(async () => {
  await convStore.fetchMessages(conversationId.value)

  try {
    await wsStore.connect(conversationId.value)
    connected.value = true
  } catch {
    connected.value = false
  }

  wsStore.onMessage(handleWSMessage)
})

onUnmounted(() => {
  wsStore.disconnect()
})

function handleWSMessage(msg: WSMessage) {
  switch (msg.type) {
    case 'asr_interim': {
      // ASR 中间识别结果（边说边转，实时展示）
      interimASRText.value = msg.text || ''
      break
    }

    case 'asr_result': {
      // 语音识别结果 → 显示为用户消息气泡
      interimASRText.value = '' // 清除中间结果
      const recognizedText = msg.text || ''
      if (recognizedText) {
        const userMsg: Message = {
          id: Date.now(),
          conversation_id: conversationId.value,
          role: 'user',
          content: recognizedText,
          created_at: new Date().toISOString(),
        }
        convStore.addMessage(userMsg)
        // AI 开始思考
        isAIThinking.value = true
        streamingContent.value = ''
      }
      break
    }

    case 'llm_chunk': {
      // 收到第一个 chunk 时停止思考动画
      isAIThinking.value = false
      streamingContent.value += msg.text || ''
      break
    }

    case 'audio_result': {
      const audioUrl = msg.audio_url || ''
      // 流式结束，固化 AI 消息
      const content = streamingContent.value
      if (content || audioUrl) {
        const aiMsg: Message = {
          id: Date.now(),
          conversation_id: conversationId.value,
          role: 'assistant',
          content: content || '(语音回复)',
          audio_url: audioUrl,
          created_at: new Date().toISOString(),
        }
        convStore.addMessage(aiMsg)
        streamingContent.value = ''
        isAIThinking.value = false

        // 自动播放 AI 语音
        if (audioUrl) {
          lastAIAudioUrl.value = audioUrl
          nextTick(() => autoPlayAudio(audioUrl))
        }
      }
      break
    }

    case 'correction_result': {
      // 纠错分析结果 → 更新最后一条用户消息
      const corrections = msg.corrections || []
      const score = msg.pronunciation_score || 0
      convStore.updateLastUserMessageCorrection(
        conversationId.value,
        msg.correction || '',
        corrections,
        score
      )
      break
    }

    case 'error':
      ElMessage.error(msg.message || '系统繁忙，请稍后重试')
      isAIThinking.value = false
      streamingContent.value = ''
      break
  }
}

function autoPlayAudio(url: string) {
  // 延迟一下确保 DOM 更新
  setTimeout(() => {
    const audioEl = document.querySelector(`audio[src="${url}"]`) as HTMLAudioElement
    if (audioEl) {
      audioEl.play().catch(() => {
        // 浏览器可能阻止自动播放，静默处理
      })
    }
  }, 200)
}

function onSendText(text: string) {
  const userMsg: Message = {
    id: Date.now(),
    conversation_id: conversationId.value,
    role: 'user',
    content: text,
    created_at: new Date().toISOString(),
  }
  convStore.addMessage(userMsg)

  streamingContent.value = ''
  isAIThinking.value = true

  wsStore.sendText(text)
}

function onSendVoice(blob: Blob) {
  streamingContent.value = ''
  isAIThinking.value = true

  wsStore.sendVoice(blob)
}

// 显示的消息列表（包含流式中的 AI 消息和思考状态）
const displayMessages = computed(() => {
  const msgs = [...convStore.messages]
  // 如果 AI 正在思考但还没有内容，添加一个占位消息用于显示加载动画
  if (isAIThinking.value && !streamingContent.value) {
    msgs.push({
      id: -1,
      conversation_id: conversationId.value,
      role: 'assistant',
      content: '...',
      created_at: new Date().toISOString(),
    })
  } else if (streamingContent.value) {
    msgs.push({
      id: -2,
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
            <el-tag
              :type="convStore.currentScene === 'daily' ? '' : convStore.currentScene === 'business' ? 'warning' : 'danger'"
              size="small"
            >
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

      <!-- ASR 实时识别中间结果 -->
      <div v-if="interimASRText" class="interim-asr">
        <span class="interim-label">识别中...</span>
        <span class="interim-text">{{ interimASRText }}</span>
      </div>

      <!-- 输入区域：大按钮按住说话 + 文字输入兜底 -->
      <div class="input-area">
        <VoiceRecorder @send="onSendVoice" />
        <div class="divider">
          <span>或打字输入</span>
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

.interim-asr {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 10px 20px;
  background: #f0f9ff;
  border-top: 1px solid #b3d8ff;
  font-size: 14px;
  flex-shrink: 0;
}
.interim-label {
  color: #909399;
  font-size: 12px;
  white-space: nowrap;
  animation: blink 1s infinite;
}
.interim-text {
  color: #409eff;
  font-style: italic;
}

.input-area {
  background: #fff;
  border-top: 1px solid #f0f0f0;
  padding: 20px 20px 24px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  flex-shrink: 0;
}
.divider {
  font-size: 12px;
  color: #c0c4cc;
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  max-width: 400px;
}
.divider::before,
.divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background: #e8e8e8;
}
</style>
