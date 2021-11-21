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
    }),
    quasar(),
    visualizer(),
  ],
  // // https://github.com/element-plus/element-plus/issues/3219
  // // except this doesn't seem to work??? :(
  // css: {
  //   preprocessorOptions: {
  //     scss: {
  //       charset: false
  //     }
  //   }
  // },  // fix import errors and @ alias
  resolve: {
    alias: {
      '@': path.resolve(__dirname, '/src'),
    },
  },
})
