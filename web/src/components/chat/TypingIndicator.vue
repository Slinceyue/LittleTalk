<template>
  <div class="typing-indicator">
    <div class="typing-dots">
      <span class="dot"></span>
      <span class="dot"></span>
      <span class="dot"></span>
    </div>
    <span class="typing-names">{{ displayText }}</span>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  names: { type: Array, default: () => [] },
})

const displayText = computed(() => {
  if (!props.names.length) return ''
  if (props.names.length === 1) return `${props.names[0]} 正在输入...`
  if (props.names.length === 2) return `${props.names[0]} 和 ${props.names[1]} 正在输入...`
  return `${props.names[0]} 等 ${props.names.length} 人正在输入...`
})
</script>

<style scoped>
.typing-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 0;
  margin-left: 4px;
}
.typing-dots {
  display: flex;
  gap: 3px;
  align-items: center;
}
.dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--text-hint);
  animation: typingBounce 1.4s ease-in-out infinite both;
}
.dot:nth-child(1) { animation-delay: 0s; }
.dot:nth-child(2) { animation-delay: 0.16s; }
.dot:nth-child(3) { animation-delay: 0.32s; }

@keyframes typingBounce {
  0%, 60%, 100% { transform: translateY(0); opacity: 0.4; }
  30% { transform: translateY(-4px); opacity: 1; }
}

.typing-names {
  font-size: var(--font-xs);
  color: var(--text-hint);
}
</style>
