import AppView from '@/views/app/AppView.vue';
import { Story } from '@storybook/vue3'
import { reactive } from 'vue';
import vueRouter from 'storybook-vue3-router'

// this should not be necessary as vite.config.js has the ElementPlus resolver
import 'element-plus/dist/index.css';

export default {
  title: 'App/AppView',
  component: AppView,
}


const Template: Story = (args: any) => ({
  components: { AppView },
  setup() {
    const headerState = reactive({
      editMode: args.editMode,
      pages: args.pages,
      selectedPage: args.selectedPage,
    })

    //👇 The args will now be passed down to the template
    return { args, headerState };
  },
  template: '<AppView \
      :editMode="headerState.editMode" \
      :pages="headerState.pages" \
      :selectedPage="headerState.selectedPage" \
      @onEditModeChange="headerState.editMode=$event" \
      @onPageSelect="headerState.selectedPage=$event" \
    />',
});


export const AppNoEditMode = Template.bind({});
AppNoEditMode.decorators = [vueRouter()]
AppNoEditMode.args = { editMode: false, selectedPage: "page1", pages: ['page1', 'page2'] }

export const AppInEditMode = Template.bind({});
AppInEditMode.decorators = [vueRouter()]
AppInEditMode.args = { editMode: true, pages: ['page3'] }