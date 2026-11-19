import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
    plugins: [sveltekit()],
    server: {
        host: true,
        port: 5173,
        proxy: {
            '/api': process.env.API_PROXY_TARGET || 'http://localhost:8080',
            '/ws': { target: process.env.API_PROXY_TARGET || 'http://localhost:8080', ws: true }
        }
    }
});
