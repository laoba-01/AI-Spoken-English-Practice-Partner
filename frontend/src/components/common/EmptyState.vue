<script setup lang="ts">
defineProps<{
  message?: string
}>()
</script>

<template>
  <div class="empty-state">
    <span class="empty-icon">🎈</span>
    <p class="empty-message">{{ message || '暂无数据' }}</p>
  </div>
</template>

<style scoped>
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #c0c4cc;
}
.empty-icon {
  font-size: 48px;
  margin-bottom: 12px;
}
.empty-message {
  font-size: 14px;
}
</style>
