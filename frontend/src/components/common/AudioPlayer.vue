<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{
  url: string
}>()

const audio = ref<HTMLAudioElement | null>(null)
const playing = ref(false)
const error = ref(false)

function togglePlay() {
  if (!audio.value || error.value) return
  if (playing.value) {
    audio.value.pause()
  } else {
    audio.value.play().catch(() => {
      error.value = true
    })
  }
}

function onPlay() {
  playing.value = true
}
function onPause() {
  playing.value = false
}
function onEnded() {
  playing.value = false
}
function onError() {
  error.value = true
}
</script>

<template>
  <div class="audio-player">
    <audio
      ref="audio"
      :src="url"
      preload="auto"
      @play="onPlay"
      @pause="onPause"
      @ended="onEnded"
      @error="onError"
    />

    <el-button
      v-if="!error"
      circle
      :type="playing ? 'warning' : 'primary'"
      size="small"
      @click="togglePlay"
    >
      <el-icon :size="16">
        <VideoPlay v-if="!playing" />
        <VideoPause v-else />
      </el-icon>
    </el-button>
    <span v-else class="error-text">无法播放</span>
    <span class="label">AI 语音回复</span>
  </div>
</template>

<style scoped>
.audio-player {
  display: flex;
  align-items: center;
  gap: 8px;
}
.label {
  font-size: 12px;
  color: #909399;
}
.error-text {
  font-size: 12px;
  color: #f56c6c;
}
</style>
