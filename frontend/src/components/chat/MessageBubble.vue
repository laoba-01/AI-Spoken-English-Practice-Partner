<script setup lang="ts">
import { computed } from 'vue'
import AudioPlayer from '@/components/common/AudioPlayer.vue'
import type { Message, CorrectionItem } from '@/types'

const props = defineProps<{
  message: Message
}>()

// 解析纠错 JSON
const corrections = computed<CorrectionItem[]>(() => {
  if (!props.message.correction) return []
  try {
    const parsed = JSON.parse(props.message.correction)
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
})

// 类型标签配置
const typeBadge = (type: string) => {
  switch (type) {
    case 'grammar':       return { label: '语法', color: '#f56c6c', bg: '#fef0f0' }
    case 'vocabulary':    return { label: '用词', color: '#e6a23c', bg: '#fdf6ec' }
    case 'pronunciation': return { label: '发音', color: '#409eff', bg: '#ecf5ff' }
    default:              return { label: type, color: '#909399', bg: '#f4f4f5' }
  }
}

// 评分颜色
const scoreColor = computed(() => {
  const s = props.message.pronunciation_score || 0
  if (s >= 80) return '#67c23a'
  if (s >= 60) return '#e6a23c'
  return '#f56c6c'
})
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

        <!-- 纠错列表 -->
        <div v-if="corrections.length" class="correction-section">
          <div
            v-for="(item, idx) in corrections"
            :key="idx"
            class="correction-item"
          >
            <div class="correction-header">
              <span
                class="type-badge"
                :style="{ color: typeBadge(item.type).color, background: typeBadge(item.type).bg }"
              >
                {{ typeBadge(item.type).label }}
              </span>
              <span class="correction-text">
                <span class="original">{{ item.original }}</span>
                <span class="arrow">→</span>
                <span class="corrected">{{ item.correction }}</span>
              </span>
            </div>
            <div v-if="item.explanation" class="correction-explanation">
              {{ item.explanation }}
            </div>
          </div>
        </div>

        <!-- 发音评分 -->
        <div v-if="message.pronunciation_score" class="score-section">
          <div class="score-header">
            <span class="score-label">发音评分</span>
            <span class="score-value" :style="{ color: scoreColor }">
              {{ message.pronunciation_score }}分
            </span>
          </div>
          <el-progress
            :percentage="message.pronunciation_score"
            :color="scoreColor"
            :stroke-width="6"
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

/* 纠错区域 */
.correction-section {
  margin-top: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.correction-item {
  background: #fafafa;
  border-radius: 8px;
  padding: 10px 12px;
  border: 1px solid #f0f0f0;
}

.correction-header {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.type-badge {
  font-size: 11px;
  padding: 1px 8px;
  border-radius: 10px;
  font-weight: 600;
  flex-shrink: 0;
}

.correction-text {
  font-size: 14px;
}

.original {
  color: #f56c6c;
  text-decoration: line-through;
}

.arrow {
  color: #c0c4cc;
  margin: 0 4px;
}

.corrected {
  color: #67c23a;
  font-weight: 500;
}

.correction-explanation {
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
  padding-left: 2px;
}

/* 评分区域 */
.score-section {
  margin-top: 12px;
}

.score-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}

.score-label {
  font-size: 12px;
  color: #909399;
}

.score-value {
  font-size: 14px;
  font-weight: 700;
}

/* 语音播放 */
.audio-section {
  margin-top: 10px;
}
</style>
