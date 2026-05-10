import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import './styles/base.css'
import './styles/animations.css'
import './styles/utilities.css'

const app = createApp(App)
app.use(createPinia())
app.mount('#app')
