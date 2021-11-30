<script lang="ts" setup>

// Wrapper around the QTable for showing a list of accounts
// QTable slots are available to the parent
import {h} from 'vue';
import  {AccountRecord} from "@/data/AccountStore";
import {QBtn, QIcon, QToolbar, QTable, QTd, QToggle, QToolbarTitle, QTableProps} from "quasar";
import {mdiAccountPlus, mdiDelete, mdiAccountEdit, mdiLinkOff, mdiLink} from "@quasar/extras/mdi-v6";


// Accounts table API
interface IAccountsTable {
  // Collection of accounts to display
  accounts: Array<AccountRecord>
  // Allow editing of accounts
  editMode?: boolean
  // optional title to display above the table
  title?: string
}
const props = defineProps<IAccountsTable>()

const emit = defineEmits([
  "onEdit",
  "onDelete",
  "onToggleEnabled"
])

// table columns
interface ICol {
  name: string
  label: string
  field: string
  required?: boolean
  align?: "left"| "right" | "center" | undefined
  style?: string
  format?: (val:any, row:any)=>any
}

// calculated property
const isConnected = (accountID: string): boolean => {
  return false
}

// reactive edit mode
const isEditMode = ():boolean => {
  return !!props.editMode
}

const columns: Array<ICol> = [
  {name: "edit", label: "", field: "edit", align:"center"},
    // Use large width to minimize the other columns
  {name: "name", label: "Name", field:"name" , align:"left", required:true, style:"width:400px",
    },
  {name: "address", label: "Address", field:"address", align:"left"},
  {name: "authPort", label: "Authentication Port", field:"authPort", align:"left"},
  {name: "mqttPort", label: "MQTT Port", field:"mqttPort", align:"left"},
  {name: "directoryPort", label: "Directory Port", field:"directoryPort", align:"left"},
  {name: "enabled", label: "Enabled", field:"enabled", align:"center"},
  {name: "connected", label: "Connected", field:"connected", align:"center"},
  // {name: "delete", label: "", field:"delete", align:"center"},
]
console.log("AccountsTable: editMode=%s", props.editMode)



</script>


<template>
  <QTable :rows="props.accounts"
          :columns="columns"
          hide-pagination
          row-key="address"
          hide-selected-banner
  >
    <!-- export the slots-->
    <template v-for="(index, name) in $slots" v-slot:[name]>
      <slot :name="name" />
    </template>

    <!-- add account button-->
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

    <!-- icon for connected-->
    <template v-slot:body-cell-connected="props" >
      <QTd>
        <QIcon flat :color="isConnected(props.row.id)?'green':'red'"
               :name="isConnected(props.row.id)?mdiLink:mdiLinkOff"
               size="2em"
              />
      </QTd>
    </template>


    <!-- button for edit-->
    <template v-if="isEditMode()"  v-slot:body-cell-edit="props" >
      <QTd>
        <QBtn dense flat round color="blue" field="edit" :icon="mdiAccountEdit"
              @click="emit('onEdit', props.row)"/>
      </QTd>
    </template>
    <!-- button for delete-->
    <template v-if="isEditMode()"  v-slot:body-cell-delete="props" >
      <QTd>
        <QBtn dense flat round color="blue" field="edit" :icon="mdiDelete"
              @click="emit('onDelete', props.row)"/>
      </QTd>
    </template>
    </QTable>
</template>
