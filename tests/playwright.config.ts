import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: '.',
  testMatch: 'e2e_flow.spec.ts',
  timeout: 60_000,
  use: {
    baseURL: process.env.BASE_URL || 'http://127.0.0.1:28471',
    viewport: { width: 1440, height: 900 },
  },
})
