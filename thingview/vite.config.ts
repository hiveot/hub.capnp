import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'
import {quasar, transformAssetUrls } from '@quasar/vite-plugin'
// import { visualizer } from 'rollup-plugin-visualizer';

// https://vitejs.dev/config/

/**
 * @type {import('vite').UserConfig}
 */
export default defineConfig({
  // if (command === 'serve') {
  // } else {
  // }
  build: {
    sourcemap:true,
  },
  plugins: [
      vue({
        template: {transformAssetUrls},
        // template: {},
      }),
      quasar(),
      // visualizer(),
    ],
    resolve: {
    alias: {
      // so we can start the import with @/
      '@': path.resolve(__dirname, '/src'),
    },
    },
})
