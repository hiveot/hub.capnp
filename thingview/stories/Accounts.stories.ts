import AccountsTable from '../src/components/AccountsTable.vue';

// import the complete bundle
// import ElementPlus from 'element-plus';
import hubAccountStore, { AccountRecord } from '../src/store/HubAccountStore';

// this should not be necessary as vite.config.js has the ElementPlus resolver
import 'element-plus/dist/index.css';
import { withKnobs, text, boolean } from "@storybook/addon-knobs";
import { addParameters, Story } from '@storybook/vue3'
import { action } from '@storybook/addon-actions';
import { reactive } from "vue";

// The default test for accounts
export default {
  title: 'Components/Accounts',
  component: AccountsTable,
  argTypes: {
    accounts: {
      description: "Collection of AccountRecords"
    },
  },
};

//👇 Create a “template” of how args map to rendering
const Template: Story = (args: any) => ({
  components: { AccountsTable },
  setup() {
    // addParameters({
    //   actions: { handles: "onAdd" },
    // });
    const handleOnAdd = () => {
      console.log("handleOnAdd");
      // this is the only way to get actions to work
      action('onAdd')('onAdd event');
      args.accounts.push(new AccountRecord())
    }
    return { args, handleOnAdd };
  },
  template: '<AccountsTable v-bind="args" @onAdd="handleOnAdd" />',
});

// Each story then reuses the template
export const NoAccount = Template.bind({});
NoAccount.args = reactive({
  accounts: [],
})


export const OneAccount = Template.bind({});
OneAccount.args = {
  accounts: [new AccountRecord()]
}


