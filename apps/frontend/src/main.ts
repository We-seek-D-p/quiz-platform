import { createApp } from 'vue'
import { createPinia } from 'pinia'
import PrimeVue from 'primevue/config'
import ToastService from 'primevue/toastservice'

import App from './App.vue'
import router from './router'
import { AuraIndigoPreset } from './theme/preset'
import { initThemeMode } from './theme'
import { useAuthStore } from './stores/auth'

import 'primeicons/primeicons.css'
import './styles/index.css'

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
