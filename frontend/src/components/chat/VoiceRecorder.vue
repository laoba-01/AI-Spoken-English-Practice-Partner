<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'

const emit = defineEmits<{
  send: [blob: Blob]
}>()

const recording = ref(false)
const mediaRecorder = ref<MediaRecorder | null>(null)
const chunks = ref<Blob[]>([])
const startTime = ref(0)
const elapsed = ref('00:00')
let timer: ReturnType<typeof setInterval> | null = null

async function startRecording() {
  try {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
    const recorder = new MediaRecorder(stream, {
      mimeType: 'audio/webm;codecs=opus',
    })

    chunks.value = []
    recorder.ondataavailable = (e) => {
      if (e.data.size > 0) {
        chunks.value.push(e.data)
      }
    }

    recorder.onstop = () => {
      // 停止所有轨道
      stream.getTracks().forEach((t) => t.stop())

      if (chunks.value.length > 0) {
        const blob = new Blob(chunks.value, { type: 'audio/webm' })
        emit('send', blob)
      }
    }

    recorder.start()
    mediaRecorder.value = recorder
    recording.value = true
    startTime.value = Date.now()

    // 计时器
    timer = setInterval(() => {
      const secs = Math.floor((Date.now() - startTime.value) / 1000)
      const min = String(Math.floor(secs / 60)).padStart(2, '0')
      const sec = String(secs % 60).padStart(2, '0')
      elapsed.value = `${min}:${sec}`
    }, 200)
  } catch {
    ElMessage.warning('无法访问麦克风，请检查浏览器权限设置')
  }
}

function stopRecording() {
  if (mediaRecorder.value && recording.value) {
    mediaRecorder.value.stop()
    recording.value = false
    if (timer) {
      clearInterval(timer)
      timer = null
    }
    elapsed.value = '00:00'
  }
}
</script>

<template>
  <el-button
    class="voice-btn"
    :class="{ recording }"
    :type="recording ? 'danger' : 'default'"
    circle
    size="large"
    @mousedown="startRecording"
    @mouseup="stopRecording"
    @mouseleave="recording ? stopRecording() : undefined"
  >
    <el-icon :size="20">
      <Microphone v-if="!recording" />
      <span v-else class="rec-icon">⏹</span>
    </el-icon>
  </el-button>
  <span v-if="recording" class="recording-hint">🔴 录音中 {{ elapsed }} (松开发送)</span>
</template>

<style scoped>
.voice-btn {
  width: 48px;
  height: 48px;
  transition: all 0.3s;
}
.voice-btn.recording {
  animation: pulse 0.8s infinite;
}
.recording-hint {
  font-size: 13px;
  color: #f56c6c;
  white-space: nowrap;
}
.rec-icon {
  font-size: 14px;
}

@keyframes pulse {
  0%, 100% { box-shadow: 0 0 0 0 rgba(245, 108, 108, 0.4); }
  50% { box-shadow: 0 0 0 12px rgba(245, 108, 108, 0); }
}
</style>
