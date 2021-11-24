<script lang="ts" setup>

import AccountsTable from '@/components/AccountsTable.vue'
import hubAccountStore, {AccountRecord} from '@/store/HubAccountStore'
import {reactive} from "vue";
import {QCard, QCardSection, QToolbar, QToolbarTitle} from "quasar";
import {mdiAccountPlus, mdiDelete, mdiAccountEdit} from "@quasar/extras/mdi-v6";
import appState from '@/store/AppState'

// Accounts view API
interface IAccountsView {
  // Allow editing of accounts
  editMode?: boolean
}
const props = defineProps<IAccountsView>()

const data = reactive({
  selectedRow : AccountRecord,
  accounts : hubAccountStore.GetAccounts(),
  showAddDialog: false,
  showEditDialog: false,
})

const handleStartAdd = () => {
  console.log("handleStartAdd")
  data.showAddDialog = !data.showAddDialog
}

const handleEdit = (record: AccountRecord) => {
  data.showEditDialog = !data.showEditDialog
  console.log("handleEdit")
}

const handleDelete = (record: AccountRecord) => {
  console.log("handleDelete")
  // todo: ask for confirmation
  hubAccountStore.Remove(record.name)
}

// toggle the enabled
const handleToggleEnabled = (record: AccountRecord) => {
  console.log("handleOnToggleEnabled")
  hubAccountStore.SetEnabled(record.name, !record.enabled)
}

</script>


<template>
  <QCard class="my-card" flat>
    <QCardSection class="">
      <AccountsTable :accounts="data.accounts"
                     title="Hub Accounts"
                     style="width: 100%"
                     :editMode="appState.State().editMode"
                     @onEdit="handleEdit"
                     @onDelete="handleDelete"
                     @onToggleEnabled="handleToggleEnabled"
      >
        // Add account button
        <template v-slot:top>
          <QToolbar>
            <QToolbarTitle  shrink>Hub Accounts</QToolbarTitle>
            <QBtn size="sm" round color="primary" :icon="mdiAccountPlus"
              @click="handleStartAdd"
            />
          </QToolbar>
        </template>

      </AccountsTable>
    </QCardSection>

  </QCard>

</template>
