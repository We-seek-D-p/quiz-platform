import tailwindcss from '@tailwindcss/vite'

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  css: ['~/assets/styles/index.css', 'primeicons/primeicons.css'],
  vite: {
    plugins: [tailwindcss()],
  },
  runtimeConfig: {
    public: {
      authApiBase: '/api/v1/auth',
      managementApiBase: '/api/v1',
      sessionWsHostPath: '/api/v1/ws/host',
      sessionWsPlayerPath: '/api/v1/ws/player',
    },
  },
})
