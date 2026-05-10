<template>
  <div class="message" :class="[isOwn ? 'own' : 'other', bubbleClass]">
    <!-- Avatar for others in group -->
    <div class="msg-avatar" v-if="!isOwn && senderName">
      <AppAvatar :src="senderAvatar" :name="senderName" size="sm" />
    </div>

    <div class="msg-body">
      <!-- Sender name in group -->
      <div class="msg-sender" v-if="!isOwn && senderName">{{ senderName }}</div>

      <!-- Quoted reply -->
      <div v-if="msg.reply_to" class="reply-bar" :class="replyBarClass">
        <div class="reply-line" :style="{ background: isOwn ? selfColor : otherColor }"></div>
        <div class="reply-content">
          <span class="reply-name">{{ msg.reply_to.name || '回复' }}</span>
          <span class="reply-text">{{ msg.reply_to.content }}</span>
        </div>
      </div>

      <div class="msg-bubble" :class="bubbleStyleClass" :style="bubbleStyleObj">
        <!-- Image -->
        <img
          v-if="isImage"
          :src="imageUrl"
          class="msg-image"
          :class="imgLayoutClass"
          @click="$emit('preview', imageUrl)"
          loading="lazy"
        />

        <!-- File -->
        <div v-else-if="isFile" class="msg-file">
          <span class="file-icon">📄</span>
          <span class="file-name">{{ fileName }}</span>
          <a :href="downloadUrl" :download="fileName" class="file-dl" @click.stop>下载</a>
        </div>

        <!-- Text -->
        <div v-else class="msg-text">{{ msg.content }}</div>

        <!-- Status -->
        <span v-if="isOwn && msg.send_status === 'failed'" class="msg-failed">✗</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import AppAvatar from '../common/AppAvatar.vue'
import { buildDownloadUrl } from '../../utils/formatters.js'

const props = defineProps({
  msg: { type: Object, required: true },
  isOwn: { type: Boolean, default: false },
  senderName: { type: String, default: '' },
  senderAvatar: { type: String, default: '' },
  styleSettings: { type: Object, default: () => ({}) },
})
defineEmits(['preview'])

// ──── Derived settings with defaults ────
const s = computed(() => props.styleSettings || {})

const selfColor = computed(() => s.value.selfColor || '')
const otherColor = computed(() => s.value.otherColor || '')
const borderRadius = computed(() => s.value.borderRadius || 'large')
const bubbleStyle = computed(() => s.value.bubbleStyle || 'solid')
const imageLayout = computed(() => s.value.imageLayout || 'adaptive')

// ──── Media type detection ────
const isImage = computed(() =>
  props.msg.message_type === 3 || /\.(jpg|jpeg|png|gif|webp|svg)$/i.test(props.msg.file_name || '')
)
const isFile = computed(() => props.msg.message_type === 2)
const fileName = computed(() => props.msg.file_name || '文件')
const imageUrl = computed(() => {
  const url = props.msg.file_url || props.msg.content
  if (!url) return ''
  if (url.startsWith('http') || url.startsWith('/')) return url
  return buildDownloadUrl(url, 'image', props.msg.from_id)
})
const downloadUrl = computed(() => {
  if (isImage.value) return imageUrl.value
  return buildDownloadUrl(props.msg.file_url, props.msg.message_type === 3 ? 'image' : 'file', props.msg.from_id)
})

// ──── Bubble inline style ────
const bubbleStyleObj = computed(() => {
  const style = {}

  // Own bubble color
  if (selfColor.value && props.isOwn) {
    if (bubbleStyle.value === 'outline') {
      style.border = `1px solid ${selfColor.value}`
      style.background = 'transparent'
      style.color = 'var(--text-primary)'
    } else {
      style.background = selfColor.value
    }
  }

  // Other bubble color
  if (otherColor.value && !props.isOwn) {
    if (bubbleStyle.value === 'outline') {
      style.border = `1px solid ${otherColor.value}`
      style.background = 'transparent'
    } else {
      style.background = otherColor.value
    }
  }

  // Shadow style
  if (bubbleStyle.value === 'shadow') {
    style.boxShadow = '0 2px 8px rgba(0,0,0,0.12)'
  }

  // Border radius
  if (!props.isImage) {
    if (borderRadius.value === 'small') {
      style.borderRadius = '8px'
    } else if (borderRadius.value === 'none') {
      style.borderRadius = '2px'
    }
    // 'large' uses the default CSS var
  }

  return style
})

// ──── CSS classes ────
const bubbleClass = computed(() =>
  (isImage.value && imageLayout.value !== 'adaptive') ? `layout-${imageLayout.value}` : ''
)

const bubbleStyleClass = computed(() => {
  const cls = []
  if (isImage.value) cls.push('bubble-image')
  if (bubbleStyle.value === 'shadow') cls.push('bubble-shadow')
  return cls.join(' ')
})

const imgLayoutClass = computed(() => {
  if (!isImage.value) return ''
  return `img-${imageLayout.value}`
})

const replyBarClass = computed(() => {
  return props.isOwn ? 'reply-own' : 'reply-other'
})
</script>

<style scoped>
.message {
  display: flex; gap: 8px;
  padding: 2px 0;
  align-items: flex-start;
}
.message.own { flex-direction: row-reverse; }
.msg-avatar { flex-shrink: 0; }
.msg-body { max-width: 75%; display: flex; flex-direction: column; }
.message.own .msg-body { align-items: flex-end; }
.msg-sender {
  font-size: var(--font-xs);
  color: var(--text-hint);
  margin-bottom: 2px;
  margin-left: 4px;
}

/* ──── Bubble base ──── */
.msg-bubble {
  position: relative;
  padding: 10px 14px;
  border-radius: var(--radius-lg);
  font-size: var(--font-md);
  line-height: 1.5;
  word-break: break-word;
  white-space: pre-wrap;
}
.message.other .msg-bubble {
  background: var(--bubble-other);
  border-top-left-radius: 2px;
}
.message.own .msg-bubble {
  background: var(--bubble-self);
  color: var(--bubble-self-text);
  border-top-right-radius: 2px;
}
.bubble-image {
  padding: 4px;
  background: transparent !important;
}
.bubble-shadow {
  /* boxShadow set via inline style */
  border-radius: 8px;
}

/* ──── Image layouts ──── */
.msg-image {
  max-width: 240px;
  max-height: 320px;
  border-radius: var(--radius-md);
  cursor: pointer;
  display: block;
}
.img-adaptive {
  object-fit: cover;
}
.img-tile {
  width: 100%;
  max-height: 200px;
  object-fit: cover;
  border-radius: var(--radius-md);
}
.img-center {
  display: block;
  margin: 0 auto;
  max-width: 160px;
  max-height: 160px;
  object-fit: contain;
}
.layout-tile .msg-image { width: 100%; }
.layout-center { text-align: center; }

/* ──── Reply quote bar ──── */
.reply-bar {
  display: flex; gap: 6px;
  margin-bottom: 4px;
  padding: 6px 8px;
  background: var(--bg-page);
  border-radius: var(--radius-sm);
  max-width: 100%;
}
.reply-line { width: 3px; border-radius: 2px; flex-shrink: 0; }
.reply-content { display: flex; flex-direction: column; gap: 2px; min-width: 0; }
.reply-name { font-size: var(--font-xs); font-weight: 500; }
.reply-text {
  font-size: var(--font-xs);
  color: var(--text-hint);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.reply-own { align-self: flex-end; }
.reply-other { align-self: flex-start; }

/* ──── File ──── */
.msg-file {
  display: flex; align-items: center; gap: 8px;
  font-size: var(--font-sm);
}
.file-icon { font-size: 28px; }
.file-name { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.file-dl {
  color: var(--primary); text-decoration: none;
  padding: 4px 8px; border-radius: var(--radius-sm);
  background: rgba(7, 193, 96, 0.1);
  font-size: var(--font-xs);
}
.msg-failed {
  color: var(--danger);
  font-size: var(--font-xs);
  margin-left: 4px;
}
</style>
