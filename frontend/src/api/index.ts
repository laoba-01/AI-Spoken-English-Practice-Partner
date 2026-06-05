import type { ApiResponse, Conversation, Message, Scene, CreateConversationRequest, TextMessageRequest, TextMessageResponse } from '@/types'

const BASE = '/api'

/** POST /api/conversations — 创建新会话 */
export async function createConversation(userId: number, scene: Scene) {
  const res = await fetch(`${BASE}/conversations`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ user_id: userId, scene } satisfies CreateConversationRequest),
  })
  return res.json() as Promise<ApiResponse<{ conversation_id: number; created_at: string }>>
}

/** GET /api/conversations?user_id=&scene= — 获取会话列表 */
export async function getConversations(userId: number, scene?: Scene) {
  const params = new URLSearchParams({ user_id: String(userId) })
  if (scene) params.set('scene', scene)
  const res = await fetch(`${BASE}/conversations?${params}`)
  return res.json() as Promise<ApiResponse<Conversation[]>>
}

/** GET /api/conversations/:id — 获取会话历史消息 */
export async function getConversationHistory(id: number) {
  const res = await fetch(`${BASE}/conversations/${id}`)
  return res.json() as Promise<ApiResponse<{
    conversation_id: number
    scene: Scene
    title: string
    messages: Message[]
  }>>
}

/** POST /api/message/text — 文字消息兜底 */
export async function sendTextMessage(req: TextMessageRequest) {
  const res = await fetch(`${BASE}/message/text`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
  })
  return res.json() as Promise<ApiResponse<TextMessageResponse>>
}
