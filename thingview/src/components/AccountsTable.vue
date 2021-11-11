<script lang="ts">

import {defineComponent} from "vue";
// vue bug. Need to use 'import type' to avoid error cannot import
import  {AccountRecord} from "@/store/HubAccountStore";
// import {hubAccountStore} from "@/store/HubAccountStore";
import { ElTable, ElTableColumn, ElButton } from "element-plus";

export default defineComponent({
  components: {ElTable, ElTableColumn, ElButton},
  emits: [ 
    "onAdd" // 'Add' button was pressed
    ],
  props: {
    /* Collection of AccountRecord objects to display */
    accounts: {
      type: Array,
      description: "aaaaa",
      required: true,
      // default: [new AccountRecord()],
    },
  },

  setup(props, {emit}) {
    const AddAccount = () => {
      console.log("Emit: onAdd")
      emit('onAdd');
    }
    return {AddAccount}
  }
})

</script>


<template>
<div style="display:flex; flex-direction: column; align-items: center;">
  <ElTable :data="accounts">
    <ElTableColumn prop="name" label="Name"></ElTableColumn>
    <ElTableColumn prop="address" label="Address"></ElTableColumn>
    <ElTableColumn prop="mqttPort" label="MQTT Port"></ElTableColumn>
    <ElTableColumn prop="directoryPort" label="Directory Port"></ElTableColumn>
    <ElTableColumn prop="enabled" label="Enabled"></ElTableColumn>
  </ElTable>
  <ElButton @click="AddAccount">Add Account</ElButton>
</div>
</template>
