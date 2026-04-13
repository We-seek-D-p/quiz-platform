import { createApp } from 'vue'
import { createPinia } from 'pinia'
import PrimeVue from 'primevue/config'
import ToastService from 'primevue/toastservice'
import router from './router'
import Aura from '@primevue/themes/aura'
import { definePreset } from '@primevue/themes'
import 'primeicons/primeicons.css'
import './style.css'
import App from './App.vue'
import { initThemeMode } from './theme'
import { useAuthStore } from './stores/auth'

const AuraIndigoPreset = definePreset(Aura, {
  semantic: {
    primary: {
      50: '{indigo.50}',
      100: '{indigo.100}',
      200: '{indigo.200}',
      300: '{indigo.300}',
      400: '{indigo.400}',
      500: '{indigo.500}',
      600: '{indigo.600}',
      700: '{indigo.700}',
      800: '{indigo.800}',
      900: '{indigo.900}',
      950: '{indigo.950}',
    },
    colorScheme: {
      light: {
        surface: {
          0: '#ffffff',
          50: '#f1f4f6',
          100: '#e2e8ee',
          200: '#c6d1dd',
          300: '#a9bbcb',
          400: '#8da4ba',
          500: '#708da9',
          600: '#5a7187',
          700: '#435565',
          800: '#2d3844',
          900: '#161c22',
          950: '#0d1117',
        },
      },
      dark: {
        surface: {
          0: '#ffffff',
          50: '#e8e9e9',
          100: '#d2d2d4',
          200: '#bbbcbe',
          300: '#a5a5a9',
          400: '#8e8f93',
          500: '#77787d',
          600: '#616268',
          700: '#4a4b52',
          800: '#34343d',
          900: '#1d1e27',
          950: '#14151d',
        },
      },
    },
  },
})

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
