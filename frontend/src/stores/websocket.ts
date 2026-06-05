import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { WSMessage } from '@/types'

export const useWebSocketStore = defineStore('websocket', () => {
  // --- 状态 ---
  const ws = ref<WebSocket | null>(null)
  const connected = ref(false)
  const streamingText = ref('')
  const lastASRText = ref('')
  const lastAudioUrl = ref('')
  const error = ref<string | null>(null)
  const reconnectCount = ref(0)
  const maxReconnect = 3

  // --- 回调注册 ---
  type MessageHandler = (msg: WSMessage) => void
  const handlers: MessageHandler[] = []

  function onMessage(fn: MessageHandler) {
    handlers.push(fn)
  }

  // --- 操作 ---

  /** 建立 WebSocket 连接 */
  function connect(conversationId: number): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        const protocol = location.protocol === 'https:' ? 'wss' : 'ws'
        const url = `${protocol}://${location.host}/ws/chat/${conversationId}`

        const socket = new WebSocket(url)
        socket.binaryType = 'arraybuffer'

        socket.onopen = () => {
          ws.value = socket
          connected.value = true
          reconnectCount.value = 0
          error.value = null
          resolve()
        }

        socket.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data) as WSMessage

            if (data.type === 'asr_result' && data.text) {
              lastASRText.value = data.text
            } else if (data.type === 'llm_chunk' && data.text) {
              streamingText.value += data.text
            } else if (data.type === 'audio_result' && data.audio_url) {
              lastAudioUrl.value = data.audio_url
            } else if (data.type === 'error') {
              error.value = data.message || '系统繁忙，请稍后重试'
            }

            // 分发给页面侧的回调
            for (const h of handlers) {
              h(data)
            }
          } catch {
            // 二进制消息（语音场景保留）
          }
        }

        socket.onclose = () => {
          connected.value = false
          ws.value = null
          // 自动重连
          if (reconnectCount.value < maxReconnect) {
            reconnectCount.value++
            setTimeout(() => connect(conversationId), 1000 * reconnectCount.value)
          }
        }

        socket.onerror = () => {
          error.value = 'WebSocket 连接失败'
          reject(new Error('WebSocket connection failed'))
        }
      } catch (e) {
        reject(e)
      }
    })
  }

  /** 发送文字消息 */
  function sendText(text: string) {
    if (ws.value && connected.value) {
      streamingText.value = ''
      lastAudioUrl.value = ''
      error.value = null
      ws.value.send(text)
    }
  }

  /** 发送语音消息（二进制 MP3） */
  function sendVoice(mp3Blob: Blob) {
    if (ws.value && connected.value) {
      streamingText.value = ''
      lastAudioUrl.value = ''
      error.value = null
      ws.value.send(mp3Blob)
    }
  }

  /** 断开连接 */
  function disconnect() {
    reconnectCount.value = maxReconnect // 阻止自动重连
    if (ws.value) {
      ws.value.close()
      ws.value = null
    }
    connected.value = false
    streamingText.value = ''
    lastASRText.value = ''
    lastAudioUrl.value = ''
    error.value = null
  }

  return {
    ws,
    connected,
    streamingText,
    lastASRText,
    lastAudioUrl,
    error,
    reconnectCount,
    onMessage,
    connect,
    sendText,
    sendVoice,
    disconnect,
  }
})
