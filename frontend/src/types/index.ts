// types/index.ts — 全局类型定义

/** 会话场景 */
export type Scene = 'daily' | 'business' | 'exam'

/** 场景配置 */
export interface SceneConfig {
  key: Scene
  label: string
  icon: string
  description: string
  color: string
}

/** 会话 */
export interface Conversation {
  id: number
  user_id: number
  scene: Scene
  title: string
  created_at: string
  updated_at?: string
}

/** 消息角色 */
export type MessageRole = 'user' | 'assistant'

/** 消息 */
export interface Message {
  id: number
  conversation_id: number
  role: MessageRole
  content: string
  audio_url?: string
  correction?: string
  pronunciation_score?: number
  created_at: string
}

/** 服务端 WebSocket 消息 */
export interface WSMessage {
  type: 'asr_result' | 'llm_chunk' | 'audio_result' | 'error'
  text?: string
  audio_url?: string
  message?: string
}

/** API 统一响应 */
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

/** 创建会话请求 */
export interface CreateConversationRequest {
  user_id: number
  scene: Scene
}

/** 文字消息请求 */
export interface TextMessageRequest {
  conversation_id: number
  content: string
}

/** 文字消息响应 */
export interface TextMessageResponse {
  content: string
  audio_url?: string
}
