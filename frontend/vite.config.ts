import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import path from 'path';

export default defineConfig({
	plugins: [sveltekit()],
	resolve: {
		alias: {
			$lib: path.resolve('./src/lib'),
			$components: path.resolve('./src/lib/components'),
			$stores: path.resolve('./src/lib/stores'),
			$hooks: path.resolve('./src/lib/hooks'),
			$utils: path.resolve('./src/lib/utils'),
			$types: path.resolve('./src/lib/types'),
			$styles: path.resolve('./src/lib/styles'),
		}
	},
	build: {
		target: 'es2022',
		minify: 'esbuild',
		cssCodeSplit: true,
		rollupOptions: {
			output: {
				manualChunks: {
					monaco: ['monaco-editor'],
					vendor: ['chart.js', 'uplot', 'date-fns', 'zod', 'yaml'],
					ui: ['lucide-svelte', 'svelte-motion', 'svelte-sonner'],
				}
			}
		}
	},
	server: {
		port: 5173,
		host: true,
		fs: {
			allow: ['..']
		}
	},
	optimizeDeps: {
		include: ['monaco-editor', 'chart.js', 'uplot'],
		exclude: ['@monaco-editor/loader']
	},
	worker: {
		format: 'es',
		plugins: () => []
	}
});