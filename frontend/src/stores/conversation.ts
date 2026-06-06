import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Conversation, Message, Scene } from '@/types'
import * as api from '@/api'

export const useConversationStore = defineStore('conversation', () => {
  // --- 状态 ---
  const conversations = ref<Conversation[]>([])
  const currentConversation = ref<Conversation | null>(null)
  const messages = ref<Message[]>([])
  const loading = ref(false)
  const currentScene = ref<Scene>('daily')

  // --- 计算属性 ---
  const messagesByConversation = computed(() => {
    const map = new Map<number, Message[]>()
    for (const msg of messages.value) {
      const list = map.get(msg.conversation_id) || []
      list.push(msg)
      map.set(msg.conversation_id, list)
    }
    return map
  })

  // --- 操作 ---

  /** 创建新会话 */
  async function createConversation(userId: number, scene: Scene): Promise<number | null> {
    loading.value = true
    try {
      const res = await api.createConversation(userId, scene)
      if (res.code === 0 && res.data.conversation_id) {
        currentScene.value = scene
        await fetchConversations(userId)
        return res.data.conversation_id
      }
      return null
    } finally {
      loading.value = false
    }
  }

  /** 获取会话列表 */
  async function fetchConversations(userId: number, scene?: Scene) {
    loading.value = true
    try {
      const res = await api.getConversations(userId, scene)
      if (res.code === 0) {
        conversations.value = res.data || []
      }
    } finally {
      loading.value = false
    }
  }

  /** 获取会话历史消息 */
  async function fetchMessages(conversationId: number) {
    loading.value = true
    try {
      const res = await api.getConversationHistory(conversationId)
      if (res.code === 0) {
        const conv = res.data
        currentConversation.value = {
          id: conv.conversation_id,
          user_id: 0,
          scene: conv.scene,
          title: conv.title || '',
          created_at: '',
        }
        currentScene.value = conv.scene
        messages.value = conv.messages || []
      }
    } finally {
      loading.value = false
    }
  }

  /** 添加一条消息到当前列表 */
  function addMessage(msg: Message) {
    messages.value.push(msg)
  }

  /** 删除会话 */
  async function deleteConversation(id: number) {
    const res = await api.deleteConversation(id)
    if (res.code === 0) {
      conversations.value = conversations.value.filter(c => c.id !== id)
      if (currentConversation.value?.id === id) {
        currentConversation.value = null
        messages.value = []
      }
    }
    return res.code === 0
  }

  /** 清空消息 */
  function clearMessages() {
    messages.value = []
    currentConversation.value = null
  }

  return {
    conversations,
    currentConversation,
    messages,
    loading,
    currentScene,
    messagesByConversation,
    createConversation,
    fetchConversations,
    fetchMessages,
    addMessage,
    clearMessages,
    deleteConversation,
  }
})
