<script setup lang="ts">
import type { Scene, SceneConfig } from '@/types'

const props = withDefaults(defineProps<{
  modelValue: Scene
  compact?: boolean
}>(), {
  compact: false,
})

const emit = defineEmits<{
  'update:modelValue': [scene: Scene]
}>()

const scenes: SceneConfig[] = [
  {
    key: 'daily',
    label: '日常对话',
    icon: '💬',
    description: '轻松闲聊，像和朋友聊天一样练习日常英语',
    color: '#409eff',
  },
  {
    key: 'business',
    label: '商务英语',
    icon: '💼',
    description: '职场场景模拟，会议、邮件、客户沟通',
    color: '#e6a23c',
  },
  {
    key: 'exam',
    label: '考试模拟',
    icon: '🎯',
    description: '雅思/托福口语标准，严格纠错与评分',
    color: '#f56c6c',
  },
]

function select(scene: Scene) {
  emit('update:modelValue', scene)
}
</script>

<template>
  <div class="scene-selector" :class="{ compact }">
    <div
      v-for="s in scenes"
      :key="s.key"
      class="scene-card"
      :class="{ selected: modelValue === s.key }"
      :style="{ borderColor: modelValue === s.key ? s.color : 'transparent' }"
      @click="select(s.key)"
    >
      <span class="icon">{{ s.icon }}</span>
      <div class="info">
        <h4 class="label">{{ s.label }}</h4>
        <p v-if="!compact" class="desc">{{ s.description }}</p>
      </div>
      <el-icon v-if="modelValue === s.key" class="check" :color="s.color">
        <CircleCheck />
      </el-icon>
    </div>
  </div>
</template>

<style scoped>
.scene-selector {
  display: flex;
  gap: 16px;
}
.scene-selector.compact {
  flex-direction: column;
  gap: 6px;
}

.scene-card {
  flex: 1;
  padding: 16px;
  border: 2px solid #e8e8e8;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.25s;
  display: flex;
  align-items: center;
  gap: 12px;
  background: #fff;
  position: relative;
}
.compact .scene-card {
  padding: 8px 10px;
  border-radius: 8px;
  border-width: 1px;
}

.scene-card:hover {
  border-color: #c0c4cc;
  box-shadow: 0 2px 8px rgba(0,0,0,0.06);
}
.scene-card.selected {
  box-shadow: 0 4px 12px rgba(0,0,0,0.08);
}

.icon {
  font-size: 28px;
  flex-shrink: 0;
}
.compact .icon {
  font-size: 20px;
}

.info {
  flex: 1;
}
.label {
  margin: 0 0 4px;
  font-size: 16px;
  font-weight: 600;
}
.compact .label {
  font-size: 13px;
  margin: 0;
}
.desc {
  margin: 0;
  font-size: 13px;
  color: #909399;
  line-height: 1.5;
}

.check {
  font-size: 22px;
  flex-shrink: 0;
}
</style>
