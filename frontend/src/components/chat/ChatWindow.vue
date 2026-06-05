<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import MessageBubble from './MessageBubble.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import type { Message } from '@/types'

const props = defineProps<{
  messages: Message[]
}>()

const messagesContainer = ref<HTMLElement | null>(null)

// 新消息到达时自动滚动到底部
watch(
  () => props.messages.length,
  async () => {
    await nextTick()
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  }
)

// 流式更新时也滚动到底部
watch(
  () => {
    const last = props.messages[props.messages.length - 1]
    return last?.content
  },
  async () => {
    await nextTick()
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  }
)
</script>

<template>
  <div ref="messagesContainer" class="chat-window">
    <EmptyState v-if="!messages.length" message="开始你的英语口语练习吧！" />

    <MessageBubble
      v-for="msg in messages"
      :key="msg.id"
      :message="msg"
    />

    <!-- 打字指示器 -->
    <div v-if="messages.length && messages[messages.length - 1]?.id === -1" class="typing-indicator">
      <span></span><span></span><span></span>
    </div>
  </div>
</template>

<style scoped>
.chat-window {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.typing-indicator {
  display: flex;
  gap: 4px;
  padding: 4px 12px;
}
.typing-indicator span {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #c0c4cc;
  animation: typing 1.4s infinite ease-in-out both;
}
.typing-indicator span:nth-child(1) { animation-delay: 0s; }
.typing-indicator span:nth-child(2) { animation-delay: 0.2s; }
.typing-indicator span:nth-child(3) { animation-delay: 0.4s; }

@keyframes typing {
  0%, 80%, 100% { transform: scale(0.6); opacity: 0.4; }
  40% { transform: scale(1); opacity: 1; }
}
</style>
