<template>
  <Teleport to="body">
    <Transition name="modal-fade">
      <div v-if="show" class="modal-root" @click.self="$emit('close')">
        <Transition name="modal-slide">
          <div class="modal-panel" :class="[sizeClass]">
            <div class="modal-header" v-if="$slots.header">
              <slot name="header" />
            </div>
            <div class="modal-body">
              <slot />
            </div>
            <div class="modal-footer" v-if="$slots.footer">
              <slot name="footer" />
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  show: { type: Boolean, default: true },
  size: { type: String, default: 'md' },
})
defineEmits(['close'])
const sizeClass = computed(() => `modal-${props.size}`)
</script>

<style scoped>
.modal-root {
  position: fixed;
  inset: 0;
  background: var(--bg-modal);
  z-index: 1000;
  display: flex;
  align-items: flex-end;
  justify-content: center;
}
@media (min-width: 768px) {
  .modal-root {
    align-items: center;
  }
}
.modal-panel {
  background: var(--bg-white);
  border-radius: var(--radius-xl) var(--radius-xl) 0 0;
  width: 100%;
  max-height: 85vh;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}
@media (min-width: 768px) {
  .modal-panel {
    border-radius: var(--radius-lg);
    max-width: 400px;
    max-height: 70vh;
  }
}
.modal-md { max-width: 400px; }
.modal-header {
  padding: 16px;
  border-bottom: 1px solid var(--border-light);
  font-size: var(--font-lg);
  font-weight: 500;
  display: flex;
  align-items: center;
  justify-content: space-between;
  position: sticky;
  top: 0;
  background: var(--bg-white);
}
.modal-body {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
}
.modal-footer {
  padding: 12px 16px;
  border-top: 1px solid var(--border-light);
  display: flex;
  gap: 12px;
}
</style>
