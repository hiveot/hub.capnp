import { createApp } from 'vue';
import App from './App.vue';
import router from './router';
import PrimeVue from 'primevue/config';
import 'primevue/resources/primevue.min.css'

// import 'primevue/resources/themes/md-light-indigo/theme.css'
import 'primevue/resources/themes/saga-blue/theme.css'
// import 'primevue/resources/themes/vela-blue/theme.css'

// common components
import Tooltip from 'primevue/tooltip';
import Button from 'primevue/button';
import Checkbox from 'primevue/checkbox';
import InputText from 'primevue/inputtext';

const app = createApp(App)
    .directive("tooltip", Tooltip)
    .use(router)
    .use(PrimeVue)
    .component('Button', Button)
    .component('Checkbox', Checkbox)
    .component( 'InputText', InputText)
    .mount('#app')

