// https://github.com/fengyuanchen/vue-feather/issues/8
import { createApp } from 'vue';
import { Quasar } from 'quasar';
import quasarIconSet from 'quasar/icon-set/svg-material-icons'

// Import icon libraries: MDI and Bootstrap SVG
import '@quasar/extras/material-icons/material-icons.css'
import '@quasar/extras/mdi-v6/mdi-v6.css'
import '@quasar/extras/bootstrap-icons/bootstrap-icons.css'

// Import Quasar css
import 'quasar/src/css/index.sass'

import App from './App.vue';
import router from './router';


const app = createApp(App)
  .use(router)
  .use(Quasar,{
      plugins: {},// import Quasar plugins and add here
      iconSet: quasarIconSet,
    })
  // .component('q-icon', Icon)
  .mount('#app')


