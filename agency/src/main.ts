import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import './assets/main.css'
import { useAgencyAuthStore } from '@/stores/auth'

const app = createApp(App)
const pinia = createPinia()
const auth = useAgencyAuthStore(pinia)

auth.hydrate()
if (auth.accessToken) {
  void auth.fetchMe()
}

app.use(pinia)
app.use(router)
app.mount('#app')
