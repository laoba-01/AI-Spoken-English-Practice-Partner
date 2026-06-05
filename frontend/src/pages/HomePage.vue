<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useConversationStore } from '@/stores/conversation'
import AppLayout from '@/layout/AppLayout.vue'
import SceneSelector from '@/components/common/SceneSelector.vue'
import type { Scene } from '@/types'

const router = useRouter()
const store = useConversationStore()

const selectedScene = ref<Scene>('daily')
const loading = ref(false)

// 模拟用户ID（后续接入真实登录）
const MOCK_USER_ID = 1

function onSceneChange(scene: Scene) {
  selectedScene.value = scene
}

async function startChat() {
  loading.value = true
  try {
    store.clearMessages()
    const id = await store.createConversation(MOCK_USER_ID, selectedScene.value)
    if (id) {
      router.push(`/chat/${id}`)
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <AppLayout>
    <div class="home-page">
      <div class="hero">
        <h1>选择练习场景</h1>
        <p class="subtitle">开启你的英语口语之旅，AI 外教随时陪你练</p>
      </div>

      <SceneSelector
        :model-value="selectedScene"
        @update:model-value="onSceneChange"
      />

      <div class="actions">
        <el-button
          type="primary"
          size="large"
          :loading="loading"
          @click="startChat"
        >
          开始对话
        </el-button>
      </div>

      <div class="tips">
        <el-alert
          title="💡 提示"
          type="info"
          :closable="false"
          description="支持语音和文字两种输入方式。语音模式下按住按钮说话即可，AI 会实时语音回复。"
        />
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.home-page {
  max-width: 680px;
  margin: 0 auto;
  padding: 60px 20px;
}

.hero {
  text-align: center;
  margin-bottom: 40px;
}
.hero h1 {
  font-size: 32px;
  font-weight: 700;
  color: #303133;
  margin-bottom: 12px;
}
.subtitle {
  font-size: 16px;
  color: #909399;
}

.actions {
  text-align: center;
  margin-top: 40px;
}

.tips {
  margin-top: 32px;
}
</style>
