<script setup lang="ts">
import AudioPlayer from '@/components/common/AudioPlayer.vue'
import type { Message } from '@/types'

defineProps<{
  message: Message
}>()
</script>

<template>
  <div class="bubble-wrapper" :class="message.role">
    <!-- 用户消息 -->
    <template v-if="message.role === 'user'">
      <div class="bubble user-bubble">
        <p class="text">{{ message.content }}</p>
      </div>
    </template>

    <!-- AI 消息 -->
    <template v-else>
      <div class="bubble ai-bubble">
        <p class="text">{{ message.content }}</p>

        <!-- 纠错内容展示 -->
        <div v-if="message.correction" class="correction">
          <el-alert
            title="纠错建议"
            :description="message.correction"
            type="warning"
            :closable="false"
            show-icon
          />
        </div>

        <!-- 发音评分 -->
        <div v-if="message.pronunciation_score" class="score">
          <span class="score-label">发音评分</span>
          <el-progress
            :percentage="message.pronunciation_score"
            :color="message.pronunciation_score >= 80 ? '#67c23a' : message.pronunciation_score >= 60 ? '#e6a23c' : '#f56c6c'"
            :stroke-width="8"
          />
        </div>

        <!-- 语音播放 -->
        <div v-if="message.audio_url" class="audio-section">
          <AudioPlayer :url="message.audio_url" />
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.bubble-wrapper {
  display: flex;
  margin-bottom: 4px;
}
.bubble-wrapper.user {
  justify-content: flex-end;
}
.bubble-wrapper.assistant {
  justify-content: flex-start;
}

.bubble {
  max-width: 75%;
  padding: 12px 16px;
  border-radius: 12px;
  line-height: 1.6;
  word-break: break-word;
}

.user-bubble {
  background: #409eff;
  color: #fff;
  border-bottom-right-radius: 4px;
}
.user-bubble .text {
  margin: 0;
}

.ai-bubble {
  background: #fff;
  color: #333;
  border: 1px solid #e8e8e8;
  border-bottom-left-radius: 4px;
  box-shadow: 0 1px 2px rgba(0,0,0,0.04);
}
.ai-bubble .text {
  margin: 0;
  white-space: pre-wrap;
}

.correction {
  margin-top: 10px;
}

.score {
  margin-top: 10px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.score-label {
  font-size: 12px;
  color: #909399;
}

.audio-section {
  margin-top: 10px;
}
</style>
