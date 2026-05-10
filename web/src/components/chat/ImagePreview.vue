<template>
  <Teleport to="body">
    <div class="preview-overlay" @click="close" @wheel.prevent @touchmove.prevent>
      <button class="preview-close" @click="close">×</button>
      <img :src="imageUrl" class="preview-image" @click.stop />
    </div>
  </Teleport>
</template>

<script setup>
import { computed } from 'vue'
import { useUiStore } from '../../stores/ui.js'

const ui = useUiStore()
const imageUrl = computed(() => ui.modalData?.url || '')

function close() {
  ui.closeModal()
}
</script>

<style scoped>
.preview-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.92);
  z-index: 2000;
  display: flex;
  align-items: center;
  justify-content: center;
}
.preview-close {
  position: absolute;
  top: 16px;
  right: 16px;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  border: none;
  background: rgba(255,255,255,0.15);
  color: #fff;
  font-size: 24px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1;
}
.preview-image {
  max-width: 95vw;
  max-height: 95vh;
  object-fit: contain;
  border-radius: 4px;
  user-select: none;
  -webkit-user-select: none;
}
</style>
