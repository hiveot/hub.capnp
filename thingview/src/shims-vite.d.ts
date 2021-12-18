// WTF: https://stackoverflow.com/questions/64213461/vuejs-typescript-cannot-find-module-components-navigation-or-its-correspon
//
// declare module '*.vue' {
//   import type { DefineComponent } from 'vue'
//   const component: DefineComponent<{}, {}, any>
//   export default component
// }
declare module "*.vue" {
    import Vue from "vue";
    export default Vue;
}

// workaround for 'global is not defined' error
declare module 'mqtt/dist/mqtt.min' {
    import MQTT from 'mqtt'
    export = MQTT
}

