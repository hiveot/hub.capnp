<script lang="ts" setup>

// Wrapper around the QTable for showing a list of accounts
// QTable slots are available to the parent
import  {AccountRecord} from "@/store/HubAccountStore";
import {QBtn, QToolbar, QTable, QTd, QToggle, QToolbarTitle} from "quasar";
import {mdiAccountPlus, mdiDelete, mdiAccountEdit} from "@quasar/extras/mdi-v6";
import {reactive} from "vue";

// Accounts table API
interface IAccountsTable {
  // Collection of accounts to display
  accounts: Array<AccountRecord>
  // Allow editing of accounts
  editMode?: boolean
}
const props = defineProps<IAccountsTable>()

const emit = defineEmits([
  "onEdit",
  "onDelete",
  "onToggleEnabled"
])

interface ICol {
  name: string
  label: string
  field: string
  align: "left"| "right" | "center" | undefined
}

const columns:Array<ICol> = [
  {name: "edit", label: "", field:"edit", align:"left"},
  {name: "name", label: "Name", field:"name" , align:"left" },
  {name: "address", label: "Address", field:"address", align:"left"},
  {name: "mqttPort", label: "MQTT Port", field:"mqttPort", align:"left"},
  {name: "directoryPort", label: "directoryPort", field:"directoryPort", align:"left"},
  {name: "enabled", label: "Enabled", field:"enabled", align:"left"},
  {name: "delete", label: "", field:"delete", align:"right"},
]
console.log("AccountsTable: editMode=%s", props.editMode)

// reactive
const isEditMode = ():boolean => {
  return !!props.editMode
}

</script>


<template>
  <QTable :rows="props.accounts"
          :columns="columns"
          hide-pagination
          row-key="address"
          hide-selected-banner
  >
<!--    export the slots-->
<!--    <template v-for="(index, name) in $slots" v-slot:[name]>-->
<!--      <slot :name="name" />-->
<!--    </template>-->

    <!--    // Add account button-->
<!--    <template v-slot:top>-->
<!--      <QToolbar>-->
<!--        <QToolbarTitle  shrink>{{props.title}}</QToolbarTitle>-->
<!--        <QBtn size="sm" round color="primary" :icon="mdiAccountPlus"/>-->
<!--      </QToolbar>-->
<!--    </template>-->

    <!-- toggle 'enabled' switch. Use computed property to be reactive inside the slot -->
    <template v-slot:body-cell-enabled="props">
      <QTd :props="props">
        <QToggle :model-value="props.row.enabled"
                  @update:model-value="emit('onToggleEnabled', props.row)"
                 :disable="!isEditMode()"
        />
      </QTd>
    </template>

    <!-- button for edit-->
    <template v-if="isEditMode()"  v-slot:body-cell-edit="props" >
      <QTd style="text-align: left">
        <QBtn dense flat round color="blue" field="edit" :icon="mdiAccountEdit"
              @click="emit('onEdit', props.row)"/>
      </QTd>
    </template>

    <!-- button for delete-->
    <template v-if="isEditMode()"  v-slot:body-cell-delete="props" >
      <QTd style="text-align: right">
        <QBtn dense flat round color="blue" field="edit" :icon="mdiDelete"
              @click="emit('onDelete', props.row)"/>
      </QTd>
    </template>
    </QTable>
</template>
