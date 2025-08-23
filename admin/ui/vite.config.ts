import {defineConfig} from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'
import {fileURLToPath} from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [react()],
    resolve: {
        alias: {
            '@': path.resolve(__dirname, './src'),
        },
    },
    server: {
        port: 3000,
        proxy: {
            '/api': {
                target: 'http://localhost:3001',
                changeOrigin: true,
            },
            '/ws': {
                target: 'ws://localhost:3001',
                ws: true,
            },
        },
    },
    build: {
        outDir: 'dist',
        emptyOutDir: true,
    },
})