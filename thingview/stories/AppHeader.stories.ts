import AppHeader from '../src/views/app/AppHeader.vue';
import { Story } from '@storybook/vue3'
import { reactive } from 'vue';

// this should not be necessary as vite.config.js has the ElementPlus resolver
import 'element-plus/dist/index.css';
import appState from '../src/store/AppState';
import dashboardStore from '../src/store/DashboardStore';

export default {
  title: 'App/AppHeader',
  component: AppHeader,
}

const Template: Story = (args: any) => ({
  components: { AppHeader },
  setup() {
    const headerState = appState.getState();

    // let storybook controls set the state
    appState.SetEditMode(args.editMode);
    appState.SetSelectedPage(args.selectedPage);
    appState.SetPages(args.pages);

    const handEditModeChange = (ev: any) => {
      console.log("received onEditModeChange event: " + ev)
      appState.SetEditMode(ev);
    }
    const handlePageSelect = (name: string) => {
      console.log("handlePageSelect: ", name)
      appState.SetSelectedPage(name);
    }
    //👇 The args will now be passed down to the template
    return { args, headerState, dashboardStore, handEditModeChange, handlePageSelect };
  },
  template: '<AppHeader \
    :editMode="headerState.editMode" \
    :pages="headerState.pages" \
    :selectedPage="headerState.selectedPage" \
    @onEditModeChange="handEditModeChange" \
    @onPageSelect="handlePageSelect" \
  /> ',
});

export const NoEditMode = Template.bind({});
NoEditMode.args = { editMode: false, pages: ['page1', 'page2', 'page3'] };


export const WithEditMode = Template.bind({});
WithEditMode.args = { editMode: true, pages: ['page1', 'page2', 'page3'] };
