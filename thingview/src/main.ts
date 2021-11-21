import { createApp } from 'vue';
import App from './App.vue';
import router from './router';

import { QIcon, QSpace, QTooltip,  Quasar } from 'quasar'
// Application Components don't need import
import Button from '@/components/Button.vue';
import Dialog from '@/components/Dialog.vue';
import MenuButton from '@/components/MenuButton.vue';

// Import SVG icon libraries (don't include the full bundle)
import iconSet from 'quasar/icon-set/svg-mdi-v6'
//import '@quasar/extras/mdi-v6/mdi-v6.css'
import 'quasar/dist/quasar.css'

const app = createApp(App)
    // .directive("tooltip", Tooltip)
    .use(router)
    .use(Quasar,{
        iconSet: iconSet,
        components: {QSpace, QIcon, QTooltip}
    })
    .component('Button', Button)
    .component('Dialog', Dialog)
    .component('MenuButton', MenuButton)
    .mount('#app')

