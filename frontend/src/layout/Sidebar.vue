<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useConversationStore } from '@/stores/conversation'
import SceneSelector from '@/components/common/SceneSelector.vue'

const router = useRouter()
const route = useRoute()
const store = useConversationStore()

const conversationList = computed(() => store.conversations)

function goToChat(id: number) {
  store.fetchMessages(id).then(() => {
    router.push(`/chat/${id}`)
  })
}

function goHome() {
  router.push('/')
}

function goHistory() {
  router.push('/history')
}

function isActive(id: number) {
  return route.params.id === String(id)
}

async function deleteConversation(id: number, event: Event) {
  event.stopPropagation()
  try {
    await ElMessageBox.confirm('确定要删除这个对话吗？删除后无法恢复。', '确认删除', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning',
    })
    const ok = await store.deleteConversation(id)
    if (ok) {
      ElMessage.success('已删除')
      // 如果正在查看被删除的会话，跳回首页
      if (route.params.id === String(id)) {
        router.push('/')
      }
    } else {
      ElMessage.error('删除失败')
    }
  } catch {
    // 用户取消
  }
}
</script>

<template>
  <div class="sidebar-wrap">
    <!-- 头部 -->
    <div class="sidebar-header">
      <h2 class="logo" @click="goHome">🎓 AI 口语陪练</h2>
    </div>

    <!-- 新建对话 -->
    <div class="new-chat-area">
      <SceneSelector compact />
    </div>

    <!-- 菜单 -->
    <div class="sidebar-nav">
      <el-button text class="nav-btn" @click="goHome">
        <el-icon><Plus /></el-icon> 新建对话
      </el-button>
      <el-button text class="nav-btn" @click="goHistory">
        <el-icon><Clock /></el-icon> 历史记录
      </el-button>
    </div>

    <!-- 会话列表 -->
    <div class="conversation-list">
      <p class="list-title">最近对话</p>
      <div
        v-for="conv in conversationList"
        :key="conv.id"
        class="conv-item"
        :class="{ active: isActive(conv.id) }"
        @click="goToChat(conv.id)"
      >
        <div class="conv-info">
          <span class="conv-title">{{ conv.title || '未命名对话' }}</span>
          <span class="conv-date">{{ conv.created_at.slice(0, 10) }}</span>
        </div>
        <div class="conv-actions">
          <el-tag size="small" :type="conv.scene === 'daily' ? '' : conv.scene === 'business' ? 'warning' : 'danger'">
            {{ conv.scene === 'daily' ? '日常' : conv.scene === 'business' ? '商务' : '考试' }}
          </el-tag>
          <el-button
            class="delete-btn"
            text
            type="danger"
            size="small"
            :icon="Delete"
            @click="(e: Event) => deleteConversation(conv.id, e)"
          />
        </div>
      </div>
      <el-empty v-if="!conversationList.length" description="暂无对话记录" :image-size="60" />
    </div>
  </div>
</template>

<style scoped>
.sidebar-wrap {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.sidebar-header {
  padding: 20px;
  border-bottom: 1px solid #f0f0f0;
}
.logo {
  font-size: 20px;
  cursor: pointer;
  margin: 0;
}

.new-chat-area {
  padding: 12px 16px;
  border-bottom: 1px solid #f0f0f0;
}

.sidebar-nav {
  display: flex;
  gap: 8px;
  padding: 8px 16px;
  border-bottom: 1px solid #f0f0f0;
}
.nav-btn {
  flex: 1;
  justify-content: center;
}

.conversation-list {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
}
.list-title {
  font-size: 12px;
  color: #999;
  padding: 4px 8px 8px;
}

.conv-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.2s;
  margin-bottom: 4px;
}
.conv-item:hover {
  background: #f5f7fa;
}
.conv-item.active {
  background: #ecf5ff;
}
.conv-item .delete-btn {
  opacity: 0;
  transition: opacity 0.2s;
}
.conv-item:hover .delete-btn {
  opacity: 1;
}

.conv-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  overflow: hidden;
  flex: 1;
  min-width: 0;
}
.conv-title {
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.conv-date {
  font-size: 12px;
  color: #999;
}

.conv-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}
</style>
