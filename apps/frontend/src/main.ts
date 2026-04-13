import { createApp } from 'vue'
import { createPinia } from 'pinia'
import PrimeVue from 'primevue/config'
import ToastService from 'primevue/toastservice'
import router from './router'

import 'primeicons/primeicons.css'
import './style.css'
import App from './App.vue'
import { initThemeMode } from './theme'
import { useAuthStore } from './stores/auth'



const app = createApp(App)
const pinia = createPinia()

initThemeMode()

app.use(PrimeVue, {
  theme: {
    preset: AuraIndigoPreset,
    options: {
      darkModeSelector: '.app-dark',
    },
  },
})

app.use(ToastService)
app.use(pinia)
app.use(router)

const authStore = useAuthStore(pinia)
void authStore.initializeSession()

app.mount('#app')
