import {app} from '@storybook/vue3';
import { Icon } from '@iconify/vue';

// no need to import Icon everywhere
app.component('icon', Icon)

// https://oh-vue-icons.netlify.app/docs/#basic-usage

export const parameters = {
  actions: { argTypesRegex: "^on[A-Z].*" },
  controls: {
    matchers: {
      color: /(background|color)$/i,
      date: /Date$/,
    },
  },
}