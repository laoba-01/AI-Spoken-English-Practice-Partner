<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useConversationStore } from '@/stores/conversation'
import AppLayout from '@/layout/AppLayout.vue'
import type { Scene } from '@/types'

const router = useRouter()
const store = useConversationStore()

const filterScene = ref<Scene | ''>('')
const loading = ref(false)

// 模拟用户ID
const MOCK_USER_ID = 1

onMounted(() => {
  loadConversations()
})

async function loadConversations() {
  loading.value = true
  try {
    const scene = filterScene.value || undefined
    await store.fetchConversations(MOCK_USER_ID, scene)
  } finally {
    loading.value = false
  }
}

function onFilterChange() {
  loadConversations()
}

function goToChat(id: number) {
  store.fetchMessages(id).then(() => {
    router.push(`/chat/${id}`)
  })
}

function sceneLabel(scene: Scene) {
  return scene === 'daily' ? '日常' : scene === 'business' ? '商务' : '考试'
}

function sceneType(scene: Scene) {
  return scene === 'daily' ? '' : scene === 'business' ? 'warning' : 'danger'
}
</script>

<template>
  <AppLayout>
    <div class="history-page">
      <div class="page-header">
        <h2>历史对话记录</h2>
        <el-select
          v-model="filterScene"
          placeholder="全部场景"
          clearable
          style="width: 140px"
          @change="onFilterChange"
        >
          <el-option label="全部场景" value="" />
          <el-option label="日常对话" value="daily" />
          <el-option label="商务英语" value="business" />
          <el-option label="考试模拟" value="exam" />
        </el-select>
      </div>

      <el-table
        :data="store.conversations"
        v-loading="loading"
        stripe
        style="width: 100%"
        @row-click="(row) => goToChat(row.id)"
      >
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column label="场景" width="100">
          <template #default="{ row }">
            <el-tag :type="sceneType(row.scene)" size="small">
              {{ sceneLabel(row.scene) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题">
          <template #default="{ row }">
            {{ row.title || '未命名对话' }}
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ row.created_at?.slice(0, 19).replace('T', ' ') }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button text type="primary" size="small" @click.stop="goToChat(row.id)">
              查看
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && !store.conversations.length" description="暂无对话记录" />
    </div>
  </AppLayout>
</template>

<style scoped>
.history-page {
  max-width: 960px;
  margin: 0 auto;
  padding: 40px 20px;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}
.page-header h2 {
  font-size: 24px;
  margin: 0;
}

.el-table {
  cursor: pointer;
}
</style>
