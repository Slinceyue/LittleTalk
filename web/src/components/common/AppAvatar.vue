<template>
  <div class="avatar" :class="[sizeClass]" :style="bgStyle">
    <img v-if="src" :src="src" :alt="alt" @error="imgError = true" v-show="!imgError" />
    <span v-if="!src || imgError" class="avatar-initial">{{ initial }}</span>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
const props = defineProps({
  src: { type: String, default: '' },
  name: { type: String, default: '' },
  size: { type: String, default: 'md' },
  bg: { type: String, default: '' },
})
const emit = defineEmits(['click'])
const imgError = ref(false)
const alt = computed(() => props.name || '?')
const initial = computed(() => props.name ? props.name.charAt(0).toUpperCase() : '?')
const sizeClass = computed(() => 'avatar-' + props.size)
const bgStyle = computed(() => props.bg ? { background: props.bg } : {})
</script>

<style scoped>
.avatar {
  flex-shrink: 0;
  border-radius: var(--radius-full);
  background: linear-gradient(135deg, #667eea, #764ba2);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  color: #fff;
  font-weight: 600;
}
.avatar img {
  width: 100%; height: 100%;
  object-fit: cover;
}
.avatar-sm { width: 36px; height: 36px; }
.avatar-sm .avatar-initial { font-size: var(--font-sm); }
.avatar-md { width: 44px; height: 44px; }
.avatar-md .avatar-initial { font-size: var(--font-md); }
.avatar-lg { width: 56px; height: 56px; }
.avatar-lg .avatar-initial { font-size: var(--font-lg); }
.avatar-xl { width: 72px; height: 72px; }
.avatar-xl .avatar-initial { font-size: 28px; }
</style>
