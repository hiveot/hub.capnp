<script lang="ts" setup>

import {onMounted, nextTick} from 'vue';
import {useQuasar} from "quasar";
import AppHeader from "./AppHeader.vue";
import {IMenuItem} from "@/components/TMenuButton.vue";

import router from '@/router'
import appState from '@/data/AppState'
import cm, { IConnectionStatus } from '@/data/accounts/ConnectionManager';
import accountStore, {AccountRecord} from "@/data/accounts/AccountStore";
import dashStore from '@/data/dashboard/DashboardStore'
import {AccountsRouteName} from "@/router";
const $q = useQuasar()


// Callback handling connection status updates
// TODO: handle multiple connections
const handleUpdate = (account:AccountRecord, status:IConnectionStatus) => {
  if (status.authenticated) {
    console.log("AppView.handleUpdate: Connection with '" + account.name + "' established.")
    $q.notify({
      position: 'top',
      type: 'positive',
      message: 'Connected to '+account.name,
    })
  } else {
    console.log("AppView.handleUpdate: Connection with '" + account.name + "' failed: ", status.statusMessage)
    $q.notify({
      position: 'top',
      type: 'negative',
      message: 'Connection to '+account.name+' at '+account.address+' failed: '+status.statusMessage,
    })
  }
  // appState.State().connectionCount = cm.connectionCount
}

// accountStore.Load()
const connectToHub = (accounts: ReadonlyArray<AccountRecord>) => {
  // Fixme: notifications for new accounts
  accounts.forEach((account) => {
    if (account.enabled) {
      cm.Connect(account, handleUpdate)
      .then()
      .catch((reason)=>{
        // popup login page
        let newPath = "/accounts/"+account.id
        console.log("AppView.connectToHub: Navigating to account edit for account '%s': path=%s", account.name, newPath)
        router.push({name: "accounts.dialog", params: { accountID: account.id}})
      })
    }
  })
}
      // cm.AuthenticationRefresh(account)
      //   .then( ()=>{
      //     cm.Connect(account, handleUpdate)
      //     console.log("AppView.connectToHub: Connection to %s successful: ", account.name)
      //     $q.notify({
      //       position: 'top',
      //       type: 'positive',
      //       message: 'Connected to '+account.name,
      //     })
      //   })
      //   .catch((reason:any)=>{
      //     console.log("AppView.connectToHub: Connection to %s at %s failed: ", account.name, account.address, reason)
      //     $q.notify({
      //       position: 'top',
      //       type: 'negative',
      //       message: 'Connection to '+account.name+' at '+account.address+' failed: '+reason,
      //     })
      //     // popup login page
      //     let newPath = "/accounts/"+account.id
      //     console.log("AppView.connectToHub: Navigating to account edit for account '%s': path=%s", account.name, newPath)
      //     // router.push(newPath)
      //     router.push({name: "accounts.dialog", params: { accountID: account.id}})
      //   })
        // cm.Connect(account, handleUpdate)
        //   .then(()=>{
        //     console.log("AppView.connectToHub: Connection to %s successful: ", account.name)
        //     $q.notify({
        //       position: 'top',
        //       type: 'positive',
        //       message: 'Connected to '+account.name,
        //     })
        //   })
        //   .then()
        //   .catch((reason:any)=>{
        //     console.log("AppView.connectToHub: Connection to %s at %s failed: ", account.name, account.address, reason)
        //     $q.notify({
        //       position: 'top',
        //       type: 'negative',
        //       message: 'Connection to '+account.name+' at '+account.address+' failed: '+reason,
        //     })
        //     // popup login page
        //     let newPath = "/accounts/"+account.id
        //     console.log("AppView.connectToHub: Navigating to account edit for account '%s': path=%s", account.name, newPath)
        //     // router.push(newPath)
        //     router.push({name: "accounts.dialog", params: { accountID: account.id}})
        //   })
  //     }
  // })
// }

onMounted(()=>{
  appState.Load()
  accountStore.Load()
  dashStore.Load()
  nextTick(()=>{
    connectToHub(accountStore.accounts);
  })
})

// future option for dark theme setting
// $q.dark.set(true) // or false or "auto"
// $q.dark.toggle()

</script>


<template>
<div class="appView">
  <AppHeader  :appState="appState"
              :cm="cm"
              :dashStore="dashStore"
              :connectionStatus="cm.connectionStatus"/>
  <router-view></router-view>
</div>
</template>


<style>
.appView {
  display:flex;
  flex-direction: column;
}
</style>