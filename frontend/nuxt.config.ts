export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },

  // Forward task API requests to Go.
  routeRules: {
    '/api/**': {
      proxy: 'http://127.0.0.1:8080/api/**',
    },
  },
})
