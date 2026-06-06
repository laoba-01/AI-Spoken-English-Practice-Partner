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

/**
 * 将 WebM/Opus 音频 blob 解码并转换为 WAV 格式（阿里云 ASR 原生支持 WAV）
 * WAV = 44 字节头 + PCM 16bit 单声道数据
 */
async function convertToWav(webmBlob: Blob): Promise<Blob> {
  const audioCtx = new AudioContext({ sampleRate: 16000 })
  try {
    const arrayBuffer = await webmBlob.arrayBuffer()
    const audioBuffer = await audioCtx.decodeAudioData(arrayBuffer)

    // 提取单声道、16kHz 的 PCM 数据
    const numChannels = 1
    const sampleRate = audioBuffer.sampleRate
    const length = audioBuffer.length
    const pcmData = audioBuffer.getChannelData(0) // Float32Array [-1, 1]

    // 转为 16-bit PCM
    const byteCount = length * 2 // 16-bit = 2 bytes per sample
    const wavBuffer = new ArrayBuffer(44 + byteCount)
    const view = new DataView(wavBuffer)

    // WAV header
    writeWavHeader(view, byteCount, sampleRate, numChannels)

    // PCM data
    let offset = 44
    for (let i = 0; i < length; i++) {
      // Float32 [-1,1] → Int16
      const sample = Math.max(-1, Math.min(1, pcmData[i]))
      const int16 = sample < 0 ? sample * 0x8000 : sample * 0x7FFF
      view.setInt16(offset, int16, true) // little-endian
      offset += 2
    }

    return new Blob([wavBuffer], { type: 'audio/wav' })
  } finally {
    audioCtx.close()
  }
}

function writeWavHeader(view: DataView, dataLength: number, sampleRate: number, numChannels: number) {
  const byteRate = sampleRate * numChannels * 2
  const blockAlign = numChannels * 2

  // RIFF chunk
  writeString(view, 0, 'RIFF')
  view.setUint32(4, 36 + dataLength, true)
  writeString(view, 8, 'WAVE')

  // fmt sub-chunk
  writeString(view, 12, 'fmt ')
  view.setUint32(16, 16, true)       // PCM = 16 bytes
  view.setUint16(20, 1, true)        // PCM format = 1
  view.setUint16(22, numChannels, true)
  view.setUint32(24, sampleRate, true)
  view.setUint32(28, byteRate, true)
  view.setUint16(32, blockAlign, true)
  view.setUint16(34, 16, true)       // bits per sample

  // data sub-chunk
  writeString(view, 36, 'data')
  view.setUint32(40, dataLength, true)
}

function writeString(view: DataView, offset: number, str: string) {
  for (let i = 0; i < str.length; i++) {
    view.setUint8(offset + i, str.charCodeAt(i))
  }
}

function getSupportedMimeType(): string {
  const types = [
    'audio/webm;codecs=opus',
    'audio/webm',
    'audio/ogg;codecs=opus',
    'audio/mp4',
  ]
  for (const t of types) {
    if (MediaRecorder.isTypeSupported(t)) {
      return t
    }
  }
  return ''
}

async function startRecording() {
  try {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
    const mimeType = getSupportedMimeType()
    const recorder = new MediaRecorder(stream, mimeType ? { mimeType } : {})

    chunks.value = []
    recorder.ondataavailable = (e) => {
      if (e.data.size > 0) {
        chunks.value.push(e.data)
      }
    }

    recorder.onstop = async () => {
      stream.getTracks().forEach((t) => t.stop())

      if (chunks.value.length === 0) return

      const webmBlob = new Blob(chunks.value, { type: 'audio/webm' })
      try {
        // 转为 WAV 格式发送给阿里云 ASR
        const wavBlob = await convertToWav(webmBlob)
        emit('send', wavBlob)
      } catch (e) {
        // WAV 转换失败，降级为直接发送 webm
        console.warn('WAV conversion failed, sending webm', e)
        emit('send', webmBlob)
      }
    }

    recorder.start()
    mediaRecorder.value = recorder
    recording.value = true
    startTime.value = Date.now()

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
  <div class="voice-recorder">
    <button
      class="record-btn"
      :class="{ recording }"
      @mousedown.prevent="startRecording"
      @mouseup.prevent="stopRecording"
      @mouseleave="recording ? stopRecording() : undefined"
      @touchstart.prevent="startRecording"
      @touchend.prevent="stopRecording"
    >
      <span v-if="!recording" class="btn-content">
        <span class="mic-icon">🎤</span>
        <span class="btn-text">按住说话</span>
      </span>
      <span v-else class="btn-content recording-content">
        <span class="pulse-ring"></span>
        <span class="rec-icon">🔴</span>
        <span class="btn-text">{{ elapsed }}</span>
        <span class="hint">松开发送</span>
      </span>
    </button>
  </div>
</template>

<style scoped>
.voice-recorder {
  display: flex;
  justify-content: center;
  align-items: center;
}

.record-btn {
  position: relative;
  width: 120px;
  height: 120px;
  border-radius: 50%;
  border: 3px solid #409eff;
  background: #fff;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s ease;
  user-select: none;
  -webkit-user-select: none;
  outline: none;
  -webkit-tap-highlight-color: transparent;
}
.record-btn:active {
  transform: scale(0.95);
}
.record-btn:hover {
  border-color: #337ecc;
  box-shadow: 0 4px 16px rgba(64, 158, 255, 0.2);
}

.record-btn.recording {
  width: 140px;
  height: 140px;
  border-color: #f56c6c;
  background: #fef0f0;
  box-shadow: 0 0 0 8px rgba(245, 108, 108, 0.15);
}

.btn-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  position: relative;
  z-index: 1;
}

.mic-icon {
  font-size: 32px;
  line-height: 1;
}

.rec-icon {
  font-size: 24px;
  animation: blink 1s infinite;
}

.btn-text {
  font-size: 14px;
  font-weight: 600;
  color: #409eff;
  white-space: nowrap;
}
.recording .btn-text {
  color: #f56c6c;
  font-size: 18px;
  font-weight: 700;
}

.hint {
  font-size: 12px;
  color: #f56c6c;
  animation: fadeInOut 1.5s infinite;
}

.pulse-ring {
  position: absolute;
  width: 100%;
  height: 100%;
  border-radius: 50%;
  border: 3px solid rgba(245, 108, 108, 0.4);
  animation: pulse 1.2s infinite;
}

@keyframes pulse {
  0% { transform: scale(0.9); opacity: 1; }
  100% { transform: scale(1.3); opacity: 0; }
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

@keyframes fadeInOut {
  0%, 100% { opacity: 0.4; }
  50% { opacity: 1; }
}
</style>
