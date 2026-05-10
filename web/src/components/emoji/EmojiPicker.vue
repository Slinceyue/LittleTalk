<template>
  <div class="emoji-picker">
    <div class="emoji-grid">
      <button
        v-for="emoji in emojis"
        :key="emoji"
        class="emoji-item"
        @click="$emit('pick', emoji)"
      >{{ emoji }}</button>
    </div>
  </div>
</template>

<script setup>
defineEmits(['pick'])

const emojis = [
  '😀','😂','🤣','😍','😘','😜','😎','🤩','😡','😭',
  '👍','👎','👏','🙌','💪','🤝','👋','✌️','🤞','👌',
  '❤️','🧡','💛','💚','💙','💜','🖤','🤍','💔','💖',
  '🎉','🎊','🎈','🎂','🎁','🏆','⭐','🔥','✨','💯',
  '🙂','😊','😋','🤔','😴','🤗','😇','🙃','😏','😌',
  '🍎','🍕','🍔','🌮','🍣','🍦','🍺','☕','🍩','🎂',
  '🐶','🐱','🐼','🐨','🐰','🦊','🐸','🐵','🦁','🐮',
  '🌞','🌚','🌈','❄️','💧','🎵','🎶','📚','💡','🔑',
  '🚀','✈️','🚗','⛵','🏠','🏖️','🌍','🗽','⛰️','🏕️',
  '✅','❌','❓','❗','💤','🕐','📌','💻','📱','🔔',
]
</script>

<style scoped>
.emoji-picker {
  background: var(--bg-white);
  border-top: 1px solid var(--border-light);
  padding: 8px;
  max-height: 220px;
  overflow-y: auto;
}
.emoji-grid {
  display: grid;
  grid-template-columns: repeat(10, 1fr);
  gap: 4px;
}
.emoji-item {
  width: 100%;
  aspect-ratio: 1;
  border: none;
  background: none;
  font-size: 22px;
  cursor: pointer;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  line-height: 1;
}
.emoji-item:hover {
  background: var(--bg-hover);
}
.emoji-item:active {
  background: var(--bg-active);
  transform: scale(1.15);
}
</style>
