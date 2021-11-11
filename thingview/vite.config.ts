import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    // needed for vite
    Components({
      // relative paths to the directory to search for components.
      dirs: ['./src/components', './src/views', './src/store'],

      // valid file extensions for components.
      extensions: ['vue'],

      // search for subdirectories
      deep: true,

      // resolvers for custom components
      resolvers: [
        ElementPlusResolver(),
      ],

      // generate `components.d.ts` global declrations,
      // also accepts a path for custom filename
      dts: true,  // was: globalComponentsDeclaration

      // filters for transforming targets
      include: [/\.vue$/, /\.vue\?vue/, /\.md$/],
      exclude: [/node_modules/, /\.git/, /\.nuxt/],
    }),
    // ElementPlus({
    //   // options
    // }),
  ],
  // fix import errors and @ alias
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
    },
  },
  // css: {
  //   preprocessorOptions: {
  //     scss: {
  //       additionalData: `@use "~/styles/element/index.scss" as *;`,
  //     },
  //   },
  // },
})
