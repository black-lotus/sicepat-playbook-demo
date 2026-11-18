import react from '@vitejs/plugin-react';
import { defineConfig } from 'vitest/config';

export default defineConfig({
  plugins: [react()],
  server: {
    // Saat development, request /api diteruskan ke backend Go lokal.
    // Ganti target dengan API_PROXY_TARGET bila backend berjalan di port lain.
    proxy: {
      '/api': process.env.API_PROXY_TARGET ?? 'http://localhost:8080',
    },
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    coverage: {
      provider: 'v8',
      include: ['src/**/*.{ts,tsx}'],
      exclude: ['src/**/*.test.{ts,tsx}', 'src/test/**', 'src/app/main.tsx', 'src/vite-env.d.ts'],
      thresholds: {
        lines: 60,
        'src/lib/**': { lines: 80 },
        'src/api/**': { lines: 80 },
      },
    },
  },
});
