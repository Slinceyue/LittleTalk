<template>
  <AppModal @close="ui.closeModal()">
    <template #header>
      <h3>聊天气泡设置</h3>
      <button class="close-btn" @click="ui.closeModal()">×</button>
    </template>

    <!-- Self bubble color -->
    <div class="section">
      <div class="section-label">我发消息的颜色</div>
      <div class="color-row">
        <div
          v-for="c in selfPresets"
          :key="c"
          class="swatch"
          :class="{ active: settings.selfColor === c }"
          :style="{ background: c }"
          @click="update({ selfColor: c })"
        ></div>
        <label class="custom-color" :style="{ background: settings.selfColor || '#95ec69' }">
          <input type="color" :value="settings.selfColor || '#95ec69'" @input="update({ selfColor: $event.target.value })" />
          <span class="picker-icon">✎</span>
        </label>
      </div>
    </div>

    <!-- Other bubble color -->
    <div class="section">
      <div class="section-label">对方消息的颜色</div>
      <div class="color-row">
        <div
          v-for="c in otherPresets"
          :key="c"
          class="swatch"
          :class="{ active: settings.otherColor === c }"
          :style="{ background: c }"
          @click="update({ otherColor: c })"
        ></div>
        <label class="custom-color" :style="{ background: settings.otherColor || '#ffffff' }">
          <input type="color" :value="settings.otherColor || '#ffffff'" @input="update({ otherColor: $event.target.value })" />
          <span class="picker-icon">✎</span>
        </label>
      </div>
    </div>

    <!-- Border radius -->
    <div class="section">
      <div class="section-label">圆角</div>
      <div class="option-row">
        <div
          v-for="opt in radiusOptions"
          :key="opt.value"
          class="option-chip"
          :class="{ active: (settings.borderRadius || 'large') === opt.value }"
          @click="update({ borderRadius: opt.value })"
        >{{ opt.label }}</div>
      </div>
    </div>

    <!-- Bubble style -->
    <div class="section">
      <div class="section-label">气泡样式</div>
      <div class="option-row">
        <div
          v-for="opt in styleOptions"
          :key="opt.value"
          class="option-chip"
          :class="{ active: (settings.bubbleStyle || 'solid') === opt.value }"
          @click="update({ bubbleStyle: opt.value })"
        >{{ opt.label }}</div>
      </div>
    </div>

    <!-- Image layout -->
    <div class="section">
      <div class="section-label">图片布局</div>
      <div class="option-row">
        <div
          v-for="opt in imgOptions"
          :key="opt.value"
          class="option-chip"
          :class="{ active: (settings.imageLayout || 'adaptive') === opt.value }"
          @click="update({ imageLayout: opt.value })"
        >{{ opt.label }}</div>
      </div>
    </div>

    <!-- Reset -->
    <button class="reset-btn" @click="handleReset">恢复默认</button>
  </AppModal>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { useUiStore } from '../../stores/ui.js'
import { useChatStore } from '../../stores/chat.js'
import AppModal from '../common/AppModal.vue'

const ui = useUiStore()
const chat = useChatStore()

const friendId = computed(() => ui.modalData?.friendId || 0)

const settings = reactive({
  selfColor: '',
  otherColor: '',
  borderRadius: 'large',
  bubbleStyle: 'solid',
  imageLayout: 'adaptive',
})

// Load current settings
if (friendId.value) {
  chat.loadBubbleStyle(friendId.value)
  const cur = chat.getBubbleStyle(friendId.value)
  if (cur.selfColor) settings.selfColor = cur.selfColor
  if (cur.otherColor) settings.otherColor = cur.otherColor
  if (cur.borderRadius) settings.borderRadius = cur.borderRadius
  if (cur.bubbleStyle) settings.bubbleStyle = cur.bubbleStyle
  if (cur.imageLayout) settings.imageLayout = cur.imageLayout
}

const selfPresets = ['#95ec69', '#07c160', '#4fc3f7', '#ff8a65', '#ce93d8', '#fff176']
const otherPresets = ['#ffffff', '#f5f5f5', '#e8eaf6', '#fce4ec', '#e0f2f1', '#fff8e1']

const radiusOptions = [
  { value: 'large', label: '大圆角' },
  { value: 'small', label: '小圆角' },
  { value: 'none', label: '直角' },
]
const styleOptions = [
  { value: 'solid', label: '实心' },
  { value: 'outline', label: '边框' },
  { value: 'shadow', label: '阴影浮起' },
]
const imgOptions = [
  { value: 'adaptive', label: '自适应' },
  { value: 'tile', label: '平铺' },
  { value: 'center', label: '居中' },
]

function update(patch) {
  Object.assign(settings, patch)
  if (friendId.value) {
    chat.setBubbleStyle(friendId.value, { ...patch })
  }
}

function handleReset() {
  settings.selfColor = ''
  settings.otherColor = ''
  settings.borderRadius = 'large'
  settings.bubbleStyle = 'solid'
  settings.imageLayout = 'adaptive'
  if (friendId.value) {
    chat.setBubbleStyle(friendId.value, {
      selfColor: '',
      otherColor: '',
      borderRadius: 'large',
      bubbleStyle: 'solid',
      imageLayout: 'adaptive',
    })
  }
}
</script>

<style scoped>
.close-btn { background: none; border: none; font-size: 24px; color: var(--text-hint); padding: 0; cursor: pointer; }
.section { margin-bottom: 20px; }
.section-label { font-size: var(--font-sm); color: var(--text-secondary); margin-bottom: 8px; font-weight: 500; }
.color-row { display: flex; flex-wrap: wrap; gap: 8px; align-items: center; }
.swatch {
  width: 40px; height: 40px;
  border-radius: 50%;
  border: 2px solid var(--border);
  cursor: pointer;
  transition: transform 0.15s;
}
.swatch:hover { transform: scale(1.15); }
.swatch.active { border-color: var(--primary); box-shadow: 0 0 0 2px var(--bg-white), 0 0 0 4px var(--primary); }
.custom-color {
  position: relative;
  width: 40px; height: 40px;
  border-radius: 50%;
  border: 2px dashed var(--border);
  cursor: pointer;
  display: flex; align-items: center; justify-content: center;
  overflow: hidden;
}
.custom-color input[type="color"] {
  position: absolute; inset: 0;
  opacity: 0; cursor: pointer;
  width: 100%; height: 100%;
}
.picker-icon { font-size: 14px; color: rgba(0,0,0,0.3); pointer-events: none; }
.option-row { display: flex; gap: 8px; flex-wrap: wrap; }
.option-chip {
  padding: 8px 16px;
  border-radius: var(--radius-full);
  border: 1px solid var(--border);
  font-size: var(--font-sm);
  cursor: pointer;
  transition: all 0.15s;
}
.option-chip:hover { border-color: var(--primary); color: var(--primary); }
.option-chip.active {
  background: var(--primary); color: #fff; border-color: var(--primary);
}
.reset-btn {
  width: 100%; padding: 12px;
  background: var(--bg-input); color: var(--text-secondary);
  border: none; border-radius: var(--radius-md);
  cursor: pointer; font-size: var(--font-sm);
}
</style>
