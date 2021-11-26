import { createApp } from 'vue';
import App from './App.vue';
import router from './router';

// commonly used components that you don't want to import all the time
import { QBtn, QIcon, QSpace, QTooltip, QSeparator, Quasar } from 'quasar'

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

