import { createApp } from 'vue';
import App from './App.vue';
import router from './router';

// commonly used components that you don't want to import all the time
import { Dialog, Notify, Quasar} from 'quasar'

// Import SVG icon libraries (don't include the full bundle)
import iconSet from 'quasar/icon-set/svg-material-icons.js'
// import iconSet from 'quasar/icon-set/svg-mdi-v6.js'
// import '@quasar/extras/material-icons/material-icons.css'
import 'quasar/dist/quasar.css'


const app = createApp(App)
    // .directive("tooltip", Tooltip)
    .use(router)
    .use(Quasar,{
        iconSet: iconSet,
        plugins: {Notify, Dialog},
        components: {},
        // brand: {
        //     primary: '#e46262',
        //     // ... or all other brand colors
        // },
        // notify: {...}, // default set of options for Notify Quasar plugin
        // loading: {...}, // default set of options for Loading Quasar plugin
        // loadingBar: { ... }, // settings for LoadingBar Quasar plugin
        // ..and many more (check Installation card on each Quasar component/directive/plugin)
    })
    .mount('#app')

