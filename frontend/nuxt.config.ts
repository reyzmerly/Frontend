// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: false },
  devServer: {
    port: 5173
  },
  app: {
    head: {
      title: 'University of Kigali - Invoice & Payment Management'
    }
  },
  vite: {
    server: {
      watch: {
        usePolling: false
      }
    }
  }
})
