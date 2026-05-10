<template>
  <AppModal @close="$emit('close')">
    <template #header>
      <button class="close-btn" @click="$emit('close')">×</button>
      <h3>编辑个人信息</h3>
      <button class="save-btn" @click="save">保存</button>
    </template>
    <div class="edit-form">
      <label>用户名</label>
      <input v-model="form.username" maxlength="20" />
      <label>个人简介</label>
      <textarea v-model="form.intro" maxlength="255" rows="3" placeholder="介绍一下自己吧~"></textarea>
      <label>性别</label>
      <select v-model="form.sex">
        <option :value="0">未知</option>
        <option :value="1">男</option>
        <option :value="2">女</option>
      </select>
      <label>生日</label>
      <input type="date" v-model="form.birthday" />
      <label>手机号</label>
      <input v-model="form.phone" maxlength="16" placeholder="请输入手机号" />
      <label>邮箱</label>
      <input v-model="form.email" maxlength="64" placeholder="请输入邮箱" />
    </div>
  </AppModal>
</template>

<script setup>
import { reactive, onMounted } from 'vue'
import { useAuthStore } from '../../stores/auth.js'
import { useUiStore } from '../../stores/ui.js'
import AppModal from '../common/AppModal.vue'

const auth = useAuthStore()
const ui = useUiStore()
const emit = defineEmits(['close', 'saved'])

const form = reactive({
  username: '',
  intro: '',
  sex: 0,
  birthday: '',
  phone: '',
  email: '',
})

onMounted(() => {
  if (auth.user) {
    Object.assign(form, {
      username: auth.user.username || '',
      intro: auth.user.intro || '',
      sex: auth.user.sex ?? 0,
      birthday: auth.user.birthday || '',
      phone: auth.user.phone || '',
      email: auth.user.email || '',
    })
  }
})

async function save() {
  if (!form.username || form.username.length < 2) { ui.showToast('用户名至少2位', 'error'); return }
  if (form.phone && !/^1[3-9]\d{9}$/.test(form.phone)) { ui.showToast('手机号格式不正确', 'error'); return }
  if (form.email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) { ui.showToast('邮箱格式不正确', 'error'); return }
  try {
    await auth.updateProfile(form)
    ui.showToast('保存成功', 'success')
    emit('saved')
    emit('close')
  } catch (e) {
    ui.showToast(e.message || '保存失败', 'error')
  }
}
</script>

<style scoped>
.close-btn, .save-btn { background: none; border: none; font-size: 16px; padding: 0; }
.save-btn { color: var(--primary); font-size: var(--font-sm); font-weight: 500; }
.edit-form { display: flex; flex-direction: column; gap: 12px; padding: 4px 0; }
.edit-form label { font-size: var(--font-sm); color: var(--text-secondary); }
.edit-form input, .edit-form select, .edit-form textarea {
  padding: 10px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--bg-input);
  color: var(--text-primary);
  outline: none;
  font-size: var(--font-md);
  resize: none;
}
.edit-form input:focus, .edit-form select:focus, .edit-form textarea:focus {
  border-color: var(--primary);
}
</style>
