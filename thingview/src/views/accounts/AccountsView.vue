<script lang="ts" setup>

import {reactive,unref} from "vue";
import {useQuasar} from "quasar";
import {QCard, QCardSection, QToolbar, QToolbarTitle} from "quasar";
import {matAdd, matAssignmentInd} from "@quasar/extras/material-icons";

import AccountsTable from './AccountsTable.vue'
import accountStore, {AccountRecord} from '@/data/AccountStore'
import appState from '@/data/AppState'
import EditAccountDialog from "@/views/accounts/EditAccountDialog.vue";

// const $q = useQuasar()

//
const data = reactive({
  selectedRow : AccountRecord,
  showAddDialog: false,
  showEditDialog: false,
  // record for editing. This is mutable
  editRecord: new AccountRecord(),
})

const accounts = accountStore.GetAccounts()

const handleStartAdd = () => {
  console.log("handleStartAdd")
  data.showAddDialog = !data.showAddDialog
}

const handleStartEdit = (record: AccountRecord) => {
  console.log("handleStartEdit. record=", record)
  data.editRecord = record
  data.showEditDialog = !data.showEditDialog
}
const handleSubmitEdit = (record: AccountRecord) => {
  console.log("handleSubmitEdit")
  accountStore.Update(record)
  data.showEditDialog = false
}
const handleCancelEdit = () => {
  data.showEditDialog = false
}

const handleStartDelete = (record: AccountRecord) => {
  console.log("handleStartDelete")
  // todo: ask for confirmation
  // let $q = useQuasar()
  // $q.dialog({
  //   title:"Delete Account?",
  //   message:"Please confirm delete account "+record.name,
  //   ok:true, cancel:true,
  // }).onOk(payload => {
  //   accountStore.Remove(record.id)
  // })
}

// toggle the enabled
const handleToggleEnabled = (record: AccountRecord) => {
  console.log("handleOnToggleEnabled")
  accountStore.SetEnabled(record.id, !record.enabled)
}

</script>


<template>
  <div>
  <EditAccountDialog :visible="data.showEditDialog"
                     :account="data.editRecord"
                     @onSubmit="handleSubmitEdit"
                     @onClosed="handleCancelEdit"
  />
  <QCard class="my-card" flat>
    <QCardSection class="">
      <AccountsTable :accounts="accounts"
                     title="Hub Accounts"
                     style="width: 100%"
                     :editMode="appState.State().editMode"
                     @onEdit="handleStartEdit"
                     @onDelete="handleStartDelete"
                     @onToggleEnabled="handleToggleEnabled"
      >
        // Add account button
        <template v-slot:top>
          <QToolbar>
            <QIcon :name="matAssignmentInd" size="28px"/>
            <QToolbarTitle shrink>Hub Accounts</QToolbarTitle>
            <QBtn v-if="appState.State().editMode"
                size="sm" round color="primary" :icon="matAdd"
              @click="handleStartAdd"
            />
          </QToolbar>
        </template>

      </AccountsTable>
    </QCardSection>

  </QCard>
  </div>
</template>
