import { defineConfig } from 'vite';

// In development the page comes from Vite and the API from the Go server on
// :8080 (`make run`); in production the Go server serves both (WEB_DIR, B4).
const server = 'http://localhost:8080';

export default defineConfig({
  server: {
    proxy: {
      '/rooms': server,
      '/join': server,
      '/guide': server,
      '/history': server,
      '/ws': { target: server, ws: true },
    },
  },
  build: {
    target: 'es2022',
  },
});
