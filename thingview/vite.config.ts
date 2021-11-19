import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'
import Icons from 'unplugin-icons/vite'
import IconsResolver from 'unplugin-icons/resolver'
import Components from 'unplugin-vue-components/vite'

// https://vitejs.dev/config/

/**
 * @type {import('vite').UserConfig}
 */
export default defineConfig({
  plugins: [
    vue(),
    // unplugin
    Icons({
      compiler: 'vue3',
    }),
    // for unplugin
    Components({
      dts: true,
      resolvers: [
          IconsResolver()
      ],
    }),
  ],
  // fix import errors and @ alias
  resolve: {
    alias: {
      '@': path.resolve(__dirname, '/src'),
    },
  },
})
