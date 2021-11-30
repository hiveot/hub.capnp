<script lang="ts" setup>

import {onMounted} from 'vue';
import AppHeader from "./AppHeader.vue";
import {IMenuItem} from "@/components/MenuButton.vue";

import appState from '@/data/AppState'
import accountStore, {AccountRecord} from "@/data/AccountStore";
import ConnectionManager from '@/data/ConnectionManager';
import {useQuasar} from "quasar";

// debugger
let cm = new ConnectionManager()

// accountStore.Load()
const connectToHub = (accounts: Array<AccountRecord>) => {
  accounts.forEach((account) => {
    const $q = useQuasar()
    if (account.enabled) {
      cm.Connect(account)
          .then((args)=>{
            console.log("Connection to %s successful: ", account.name)
            $q.notify({
              position: 'top',
              type: 'positive',
              message: 'Connected to '+account.name,
            })
          })
          .catch((reason:any)=>{
            console.log("Connection to %s at %s failed: ", account.name, account.address, reason)
            $q.notify({
              position: 'top',
              type: 'negative',
              message: 'Connection to '+account.name+' at '+account.address+' failed: '+reason,
            })
          })
      }
  })
}

onMounted(()=>{
  connectToHub(accountStore.GetAccounts());
  // $q.notify({type:'positive', message:'Ready to rock'});
})

// future option for dark theme setting
// const $q = useQuasar()
// $q.dark.set(true) // or false or "auto"
// $q.dark.toggle()

</script>


<template>
<div class="appView">
  <AppHeader  :appState="appState" />
  <router-view></router-view>
</div>
</template>


<style>
.appView {
  display:flex;
  flex-direction: column;
}
</style>