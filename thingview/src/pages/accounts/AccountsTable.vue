<script lang="ts" setup>

// Wrapper around the QTable for showing a list of accounts
// QTable slots are available to the parent
import {ref} from 'vue'
import {AccountRecord} from "@/data/accounts/AccountStore";
import {QBtn, QTooltip, QTable, QTd, QToggle} from "quasar";
import {matDelete, matEdit, matLinkOff, matLink} from "@quasar/extras/material-icons";
import {ConnectionManager, IConnectionStatus} from "@/data/ConnectionManager";
import TConnectionStatus from "@/components/TConnectionStatus.vue";
import TTable from '@/components/TTable.vue'


// Accounts table API
interface IAccountsTable {
  // Collection of accounts to display
  accounts: ReadonlyArray<AccountRecord>
  // connection manager for presenting the connection state of an account
  cm: ConnectionManager
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

// Get the account's reactive connection status
const connectState = (account: AccountRecord): IConnectionStatus => {
  let state = props.cm.GetConnectionStatus(account)
  return state
}

// get reactive edit mode
const isEditMode = ():boolean => {
  return !!props.editMode
}

// Toggle the 'enabled' status of the account
const toggleEnabled = (account:AccountRecord) => {
  emit("onToggleEnabled", account)
}

const columns: Array<ICol> = [
  {name: "edit", label: "", field: "edit", align:"center"},
    // Use large width to minimize the other columns
  {name: "name", label: "Name", field:"name" , align:"left", required:true, style:"width:400px",
    },
  {name: "address", label: "Address", field:"address", align:"left", required:true, },
  {name: "authPort", label: "Authentication Port", field:"authPort", align:"left"},
  {name: "mqttPort", label: "MQTT Port", field:"mqttPort", align:"left"},
  {name: "directoryPort", label: "Directory Port", field:"directoryPort", align:"left"},
  {name: "enabled", label: "Enabled", field:"enabled", align:"center",  },
  {name: "status", label: "Status", field:"status", align:"center", required:true, },

  // connection status message
  // {name: "message", label: "Message", field:"message", align:"left"},
  {name: "delete", label: "", field:"delete", align:"center"},
]
console.log("AccountsTable: editMode=%s", props.editMode)

const visibleColumns = ref([ 'edit', 'name', 'address', 'enabled', 'connected', 'message','delete' ])
</script>


<template>
  <TTable :rows="props.accounts"
          :columns="columns"
          :visible-columns="visibleColumns"
          hide-pagination
          row-key="address"
          hide-selected-banner
  >
    <!-- export the slots-->
    <template v-for="(index, name) in $slots" v-slot:[name]>
      <slot :name="name" />
    </template>

    <!-- Header style: large and primary -->
    <template #header-cell="propz">
      <q-th class="text-primary" style="font-size:1.1rem;" :props="propz">{{propz.col.label}}</q-th>
    </template>

    <!-- Edit button when edit is enabled -->
    <template v-if="isEditMode()"  v-slot:body-cell-edit="propz" >
      <QTd>
        <QBtn dense flat round color="blue" field="edit" :icon="matEdit"
              @click="emit('onEdit', propz.row)">
              <QTooltip>Edit the account</QTooltip>
        </QBtn>
      </QTd>
    </template>


    <!-- toggle 'enabled' switch. Use computed property to be reactive inside the slot -->
    <template v-slot:body-cell-enabled="propz">
      <QTd>
        <QToggle :model-value="propz.row.enabled"
                  @update:model-value="toggleEnabled(propz.row as AccountRecord)"
                 :disable="!isEditMode()"
        />
      </QTd>
    </template>

    <!-- Connection status -->
    <template v-slot:body-cell-status="propz" >
      <QTd>
        <TConnectionStatus :value="connectState(propz.row as AccountRecord)" withText/>
      </QTd>
    </template>

    <!-- button for delete. Can't delete the last record -->
    <template v-slot:body-cell-delete="propz"
              v-if="isEditMode() && (props.accounts.length > 1)" >
      <QTd>
        <QBtn dense flat round color="blue" field="edit" :icon="matDelete"
              @click="emit('onDelete', propz.row)"/>
      </QTd>
    </template>
    </TTable>
</template>
