import { createApp } from 'vue';
import App from './App.vue';
import router from './router';

// commonly used components that you don't want to import all the time
import { QBtn, QIcon, QSpace, QTooltip, QSeparator, Quasar } from 'quasar'

// Application Components don't need import
import TButton from '@/components/TButton.vue';
import Dialog from '@/components/TDialog.vue';
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
        components: {QBtn, QIcon, QSeparator, QSpace, QTooltip}
    })
    .mount('#app')

