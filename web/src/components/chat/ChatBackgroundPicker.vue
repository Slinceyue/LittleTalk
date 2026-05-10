<template>
  <AppModal @close="ui.closeModal()">
    <template #header>
      <h3>聊天背景</h3>
      <button class="close-btn" @click="ui.closeModal()">×</button>
    </template>

    <div class="bg-section">
      <div class="bg-label">纯色背景</div>
      <div class="color-grid">
        <div
          v-for="c in solidColors"
          :key="c"
          class="color-swatch"
          :class="{ active: currentBg === c }"
          :style="{ background: c }"
          @click="selectBg(c)"
        ></div>
      </div>
    </div>

    <div class="bg-section">
      <div class="bg-label">渐变背景</div>
      <div class="color-grid">
        <div
          v-for="g in gradients"
          :key="g"
          class="color-swatch"
          :class="{ active: currentBg === g }"
          :style="{ background: g }"
          @click="selectBg(g)"
        ></div>
      </div>
    </div>

    <div class="bg-section">
      <div class="bg-label">自定义图片</div>
      <button class="upload-btn" @click="handleUpload">
        {{ currentIsImage ? '更换图片' : '选择图片' }}
      </button>
      <div v-if="currentIsImage" class="preview-wrap">
        <img :src="currentBg" class="preview-img" />
        <button class="remove-btn" @click="selectBg('')">移除图片</button>
      </div>
      <p class="hint">支持 JPG/PNG，建议 500KB 以内，仅本地保存</p>
    </div>

    <button v-if="currentBg" class="reset-btn" @click="selectBg('')">恢复默认背景</button>
  </AppModal>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useUiStore } from '../../stores/ui.js'
import { useChatStore } from '../../stores/chat.js'
import AppModal from '../common/AppModal.vue'

const ui = useUiStore()
const chat = useChatStore()

const friendId = computed(() => ui.modalData?.friendId || 0)

const currentBg = ref('')

// Load current background
if (friendId.value) {
  chat.loadBackground(friendId.value)
  currentBg.value = chat.getBackground(friendId.value) || ''
}

const currentIsImage = computed(() => currentBg.value.startsWith('data:'))

const solidColors = [
  '#ffffff', '#f5f5f5', '#ebebeb', '#fff8e1',
  '#fce4ec', '#e8eaf6', '#e0f2f1', '#e8f5e9',
  '#fff3e0', '#f3e5f5', '#e1f5fe', '#f1f8e9',
]

const gradients = [
  'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
  'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)',
  'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)',
  'linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)',
  'linear-gradient(135deg, #fa709a 0%, #fee140 100%)',
  'linear-gradient(135deg, #a18cd1 0%, #fbc2eb 100%)',
  'linear-gradient(135deg, #fccb90 0%, #d57eeb 100%)',
  'linear-gradient(135deg, #e0c3fc 0%, #8ec5fc 100%)',
]

function selectBg(value) {
  currentBg.value = value
  if (friendId.value) {
    chat.setBackground(friendId.value, value)
  }
}

function handleUpload() {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = 'image/jpeg,image/png,image/gif,image/webp'
  input.onchange = (e) => {
    const file = e.target.files[0]
    if (!file) return
    if (file.size > 512 * 1024) {
      ui.showToast('图片不能超过 500KB', 'error')
      return
    }
    const reader = new FileReader()
    reader.onload = (ev) => {
      selectBg(ev.target.result)
    }
    reader.readAsDataURL(file)
  }
  input.click()
}
</script>

<style scoped>
.close-btn { background: none; border: none; font-size: 24px; color: var(--text-hint); padding: 0; cursor: pointer; }
.bg-section { margin-bottom: 20px; }
.bg-label { font-size: var(--font-sm); color: var(--text-secondary); margin-bottom: 8px; }
.color-grid {
  display: flex; flex-wrap: wrap; gap: 8px;
}
.color-swatch {
  width: 44px; height: 44px;
  border-radius: var(--radius-md);
  border: 2px solid var(--border);
  cursor: pointer;
  transition: transform 0.15s, border-color 0.15s;
}
.color-swatch:hover { transform: scale(1.1); }
.color-swatch.active { border-color: var(--primary); }
.upload-btn {
  padding: 10px 20px;
  background: var(--primary); color: #fff;
  border: none; border-radius: var(--radius-md);
  cursor: pointer;
  font-size: var(--font-sm);
}
.preview-wrap { margin-top: 12px; }
.preview-img {
  max-width: 100%; max-height: 120px;
  border-radius: var(--radius-md);
  object-fit: cover;
  display: block;
  margin-bottom: 8px;
}
.remove-btn {
  background: none; border: 1px solid var(--danger);
  color: var(--danger); padding: 6px 14px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: var(--font-xs);
}
.hint { font-size: var(--font-xs); color: var(--text-hint); margin-top: 6px; }
.reset-btn {
  width: 100%; padding: 12px;
  background: var(--bg-input); color: var(--text-secondary);
  border: none; border-radius: var(--radius-md);
  cursor: pointer;
  font-size: var(--font-sm);
}
</style>
