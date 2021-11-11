import AppAboutDialog from '../views/app/AppAboutDialog.vue';

// this should not be necessary as vite.config.js has the ElementPlus resolver
import 'element-plus/dist/index.css';

export default {
  title: 'App/AppAbout',
  component: AppAboutDialog,
}



const Template = (args: any) => ({
  components: { AppAboutDialog },
  setup() {
    //👇 The args will now be passed down to the template
    return { args };
  },
  template: '<AppAboutDialog v-bind="args" />',
});




export const AppAbout = Template.bind({});
AppAbout.args = { visible: true }
