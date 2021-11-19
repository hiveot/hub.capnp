<script lang="ts" setup>

// vue bug. Need to use 'import type' to avoid error cannot import
import  {AccountRecord} from "@/store/HubAccountStore";
// import {hubAccountStore} from "@/store/HubAccountStore";
import DataTable from 'primevue/datatable';
import Column from 'primevue/column';
import Button from 'primevue/button';

const emit = defineEmits([
    "onAdd" // 'Add' button was pressed
    ])

// Accounts table APIO
interface IAccountsTable {
  accounts: Array<AccountRecord>
}
const props = defineProps<IAccountsTable>()

const AddAccount = () => {
   console.log("Emit: onAdd")
   emit('onAdd');
 }

 const columns = [
   {name: "name", label: "Name", field:"name" },
   {name: "address", label: "Address", field:"address"},
   {name: "mqttPort", label: "MQTT Port", field:"mqttPort"},
   {name: "directoryPort", label: "directoryPort", field:"directoryPort"},
   {name: "enabled", label: "Enabled", field:"enabled"},
 ]

</script>


<template>
<div style="display:flex; flex-direction: column; align-items: center;">
  <DataTable :value="props.accounts" >
    <Column v-for="col in columns" :field="col.field" :header="col.label"/>
  </DataTable>

  <Button @click="AddAccount">Add Account</Button>
</div>
</template>
