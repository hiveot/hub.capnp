<script lang="ts" setup>
import {reactive, ref, watchEffect} from "vue";
import TDialog from "@/components/TDialog.vue";
import {QCardSection, QForm, QInput, QToggle} from "quasar";
import accountStore, {AccountRecord} from "@/data/accounts/AccountStore";
import {useRouter} from "vue-router";
import connectionManager from "@/data/accounts/ConnectionManager";

/**
 * View/Edit account details
 */
const props = defineProps<{
  /**
   * The account to view or edit
   */
  account: AccountRecord,
  /**
   * When invoked via router, navigate to this path on close
   */
  returnTo?: string|object,
}>()

let data = reactive({
  editRecord: props.account ? {...props.account} : new AccountRecord(),
  password: "",
})

// replace the edit record with a copy of the account that was updated in props
watchEffect(()=>{
  Object.assign(data.editRecord, {...props.account})
})

const emit = defineEmits({
  'onSubmit': String, // AccountRecord
  'onClosed': null,
})

/**
 * Save the new account in the store and submit an event with the new record
 * and close the dialog
 */
const handleSubmit = () =>{
  console.log("EditAccountDialog.handleSubmit: ", data.editRecord)
  accountStore.Update(data.editRecord)
  if (data.editRecord.enabled) {
    // re-authenticate
    connectionManager.Authenticate(data.editRecord, data.password)
        .then(()=>{
          // if authentication succeeds
          connectionManager.Connect(data.editRecord)
        })
  }
  emit('onSubmit', data.editRecord)
  handleClose()
};

const router = useRouter()
const handleClose = () => {
  console.log("EditAccountDialog.handleClose")
  if (props.returnTo) {
    router.push(props.returnTo)
  }
  emit('onClosed')
}

</script>

<template>
  <TDialog
      :title="props.account ? 'Edit Account' : 'Add Account'"
      @onClosed="handleClose"
      @onSubmit="handleSubmit"
      showCancel showOk
      :okDisabled="(data.editRecord.name==='')"
  >
    <QForm @submit="handleSubmit"
           class="q-gutter-auto "
           style="min-width: 350px"
           autofocus
    >
      <QInput v-model="data.editRecord.name" autofocus dense
              id="accountName" type="text"
              label="Account name"
              aria-required="true"
              :rules="[()=>data.editRecord.name !== ''||'Please provide a name']"
      />
      <QInput v-model="data.editRecord.loginName" dense
              id="loginName" type="text"
              label="Login name"
              :rules="[()=>data.editRecord.name !== ''||'Please provide a name']"
      />
      <QInput v-model="data.password"
              id="password" type="password" dense
              label="Password (will not be stored)"
              :rules="[()=>data.editRecord.name !== ''||'Please provide credentials']"
      />
      <QInput v-model="data.editRecord.address"
              id="address" type="text" borderless
              label="Address"
      />
      <QCardSection class="q-pa-none q-ml-md">
      <QInput v-model="data.editRecord.authPort"
              id="authPort" type="number" filled
              label="Authentication port (default 8881)"
      />
      <QInput v-model="data.editRecord.directoryPort"
              id="dirPort" type="number" filled
              label="Directory Service port (default 8886)"
      />
      <QInput v-model="data.editRecord.mqttPort"
              id="mqttPort" type="number" filled
              label="MQTT Broker Websocket Port (default 8885)"
      />
      </QCardSection>
      <QToggle v-model="data.editRecord.enabled"
              id="enabled" type="boolean" class="q-mt-md q-mb-none q-pb-none"
              label="Enabled"
      />
    </QForm>

  </TDialog>
</template>
