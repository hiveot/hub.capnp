import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'
import {quasar, transformAssetUrls } from '@quasar/vite-plugin'
import { visualizer } from 'rollup-plugin-visualizer';

// https://vitejs.dev/config/

/**
 * @type {import('vite').UserConfig}
 */
export default defineConfig({
  plugins: [
    vue({
      template: {transformAssetUrls},
      // template: {},
    }),
    quasar(),
    visualizer(),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, '/src'),
    },
  },
})
