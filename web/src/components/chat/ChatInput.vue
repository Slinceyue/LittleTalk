<template>
  <div class="chat-input">
    <div class="input-row">
      <button class="attach-btn" @click="$refs.imgInput.click()" title="图片">🖼️</button>
      <input type="file" ref="imgInput" accept="image/*" hidden @change="handleImageSelect" />

      <button class="attach-btn" @click="$refs.fileInput.click()" title="文件">📎</button>
      <input type="file" ref="fileInput" hidden @change="handleFileSelect" />

      <div class="input-wrapper">
        <textarea
          ref="textInput"
          v-model="text"
          class="text-input"
          placeholder="输入消息..."
          rows="1"
          @keydown="onKeydown"
          @input="onInput"
        ></textarea>
        <button class="emoji-btn" @click="showEmoji = !showEmoji">😊</button>
      </div>

      <button class="send-btn" @click="sendText" :disabled="!text.trim() && !uploading">
        <span v-if="uploading">⏳</span>
        <span v-else>➤</span>
      </button>
    </div>

    <!-- Emoji picker -->
    <EmojiPicker v-if="showEmoji" @pick="insertEmoji" />
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import EmojiPicker from '../emoji/EmojiPicker.vue'
import { useUiStore } from '../../stores/ui.js'
import { useAuthStore } from '../../stores/auth.js'
import { MSG_TYPE, MAX_FILE_SIZE, MAX_IMAGE_DIM } from '../../utils/constants.js'
import * as messagesApi from '../../api/messages.js'

const ui = useUiStore()
const auth = useAuthStore()
const emit = defineEmits(['send', 'typing'])

const text = ref('')
const showEmoji = ref(false)
const uploading = ref(false)
const textInput = ref(null)

let typingTimer = null

function onKeydown(e) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    sendText()
  }
}

function onInput() {
  // Typing indicator
  if (text.value.trim()) {
    emit('typing', true)
    clearTimeout(typingTimer)
    typingTimer = setTimeout(() => emit('typing', false), 3000)
  } else {
    emit('typing', false)
  }
  // Auto-resize
  const el = textInput.value
  if (el) {
    el.style.height = 'auto'
    el.style.height = Math.min(el.scrollHeight, 100) + 'px'
  }
}

function sendText() {
  const content = text.value.trim()
  if (!content) return
  text.value = ''
  showEmoji.value = false
  emit('typing', false)
  emit('send', { content, messageType: MSG_TYPE.TEXT })
  // Reset height
  if (textInput.value) textInput.value.style.height = 'auto'
}

function insertEmoji(emoji) {
  text.value += emoji
  textInput.value?.focus()
}

async function handleImageSelect(e) {
  const file = e.target.files[0]
  e.target.value = ''
  if (!file) return
  await uploadAndSend(file, MSG_TYPE.IMAGE)
}

async function handleFileSelect(e) {
  const file = e.target.files[0]
  e.target.value = ''
  if (!file) return
  await uploadAndSend(file, MSG_TYPE.FILE)
}

async function uploadAndSend(file, msgType) {
  if (file.size > MAX_FILE_SIZE) { ui.showToast('文件不能超过40MB', 'error'); return }

  // Compress image
  let uploadFile = file
  if (msgType === MSG_TYPE.IMAGE && file.type.startsWith('image/')) {
    uploadFile = await compressImage(file, MAX_IMAGE_DIM, 2 * 1024 * 1024)
  }

  const formData = new FormData()
  formData.append('file', uploadFile)
  formData.append('type', msgType === MSG_TYPE.IMAGE ? 'image' : 'file')

  uploading.value = true
  try {
    const { data } = await messagesApi.uploadFile(formData)
    if (data.code === 0) {
      const fileUrl = data.data
      emit('send', {
        content: msgType === MSG_TYPE.IMAGE ? '[图片]' : `[文件] ${file.name}`,
        messageType: msgType,
        fileUrl,
        fileName: file.name,
      })
    } else {
      ui.showToast('上传失败: ' + (data.message || '未知错误'), 'error')
    }
  } catch (e) {
    ui.showToast('上传失败', 'error')
  } finally {
    uploading.value = false
  }
}

function compressImage(file, maxDim, maxSize) {
  return new Promise(resolve => {
    const img = new Image()
    img.onload = () => {
      let w = img.width, h = img.height
      if (w > maxDim || h > maxDim) {
        if (w > h) { h = Math.round(h * maxDim / w); w = maxDim }
        else { w = Math.round(w * maxDim / h); h = maxDim }
      }
      if (file.size <= maxSize) { URL.revokeObjectURL(img.src); resolve(file); return }
      const canvas = document.createElement('canvas')
      canvas.width = w; canvas.height = h
      const ctx = canvas.getContext('2d')
      ctx.drawImage(img, 0, 0, w, h)

      let quality = 0.9
      const tryCompress = () => {
        canvas.toBlob(blob => {
          if (blob.size <= maxSize || quality <= 0.1) {
            URL.revokeObjectURL(img.src)
            resolve(new File([blob], file.name, { type: file.type }))
          } else {
            quality -= 0.1
            tryCompress()
          }
        }, file.type, quality)
      }
      tryCompress()
    }
    img.src = URL.createObjectURL(file)
  })
}
</script>

<style scoped>
.chat-input {
  background: var(--bg-white);
  border-top: 1px solid var(--border-light);
  flex-shrink: 0;
  padding-bottom: var(--safe-bottom);
}
.input-row {
  display: flex; align-items: flex-end; gap: 8px;
  padding: 8px 12px;
}
.attach-btn {
  width: 36px; height: 36px;
  border: none; background: none;
  font-size: 20px; cursor: pointer; flex-shrink: 0;
  display: flex; align-items: center; justify-content: center;
}
.input-wrapper {
  flex: 1;
  display: flex; align-items: flex-end;
  background: var(--bg-input);
  border-radius: 20px;
  padding: 4px 8px 4px 14px;
}
.text-input {
  flex: 1;
  border: none; background: none;
  outline: none; resize: none;
  font-size: var(--font-md);
  line-height: 1.4;
  color: var(--text-primary);
  max-height: 100px;
  padding: 6px 0;
}
.text-input::placeholder { color: var(--text-hint); }
.emoji-btn {
  background: none; border: none;
  font-size: 18px; cursor: pointer;
  padding: 4px; flex-shrink: 0;
}
.send-btn {
  width: 36px; height: 36px;
  border: none; border-radius: 50%;
  background: var(--primary); color: #fff;
  font-size: 16px;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0; cursor: pointer;
}
.send-btn:disabled { opacity: 0.4; }
</style>
