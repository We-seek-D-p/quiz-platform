import PrimeVue from 'primevue/config'
import { initThemeMode } from '~/theme/mode'
import { AuraIndigoPreset } from '~/theme/preset'

export default defineNuxtPlugin((nuxtApp) => {
  initThemeMode()

  nuxtApp.vueApp.use(PrimeVue, {
    theme: {
      preset: AuraIndigoPreset,
      options: {
        darkModeSelector: '.app-dark',
      },
    },
  })
})
