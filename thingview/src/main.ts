// https://github.com/fengyuanchen/vue-feather/issues/8
import { createApp } from 'vue';
import Vue from 'vue';
import App from './App.vue';
import router from './router';

import { Icon } from '@iconify/vue';

const app = createApp(App)
  .use(router)
  .component('v-icon', Icon)
  .mount('#app')


