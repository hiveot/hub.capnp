<script lang="ts" setup>
import {ref, watchEffect} from "vue";
import TDialog from "@/components/TDialog.vue";
import {QCard, QCardSection, QForm, QInput, QToggle} from "quasar";
import {AccountRecord} from "@/data/AccountStore";

// since EditAccountDialog is not recreated when re-used, the account is not updated
//
interface IEditAccountDialog {
  account: AccountRecord
  visible: boolean
}
const props = defineProps<IEditAccountDialog>()

let editRecord = ref(new AccountRecord())
let password = ref("")

// replace the edit record when it is updated in props
watchEffect(()=>{
  Object.assign(editRecord.value, props.account)
})

const emit = defineEmits({
  'onSubmit': String, // AccountRecord
  'onClosed': null,
})


const handleSubmit = (ev:any) =>{
  console.log("EditAccountDialog.handleSubmit: ", editRecord)
  emit("onSubmit", editRecord.value);
  emit("onClosed");
};

const handleClose = () => {
  console.log("EditAccountDialog.handleClose")
  emit("onClosed");
}

</script>

<template>
  <TDialog
      :visible="props.visible"
      title="Edit Account"
      @onClosed="handleClose"
      @onSubmit="handleSubmit"
      showCancel showOk
      :okDisabled="(editRecord.name==='')"
  >
    <QForm @submit="handleSubmit"
           class="q-gutter-auto "
           style="min-width: 350px"
    >
      <QInput v-model="editRecord.name" autofocus dense
              id="accountName" type="text"
              label="Account name"
              :rules="[()=>editRecord.name !== ''||'Please provide a name']"
      />
      <QInput v-model="editRecord.loginName" dense
              id="loginName" type="text"
              label="Login name"
              :rules="[()=>editRecord.name !== ''||'Please provide a name']"
      />
      <QInput v-model="password"
              id="password" type="password" dense
              label="Password (will not be stored)"
              :rules="[()=>editRecord.name !== ''||'Please provide credentials']"
      />
      <QInput v-model="editRecord.address"
              id="address" type="text" borderless
              label="Address"
      />
      <QCardSection class="q-pa-none q-ml-md">
      <QInput v-model="editRecord.authPort"
              id="authPort" type="number" filled
              label="Authentication port (default 8881)"
      />
      <QInput v-model="editRecord.directoryPort"
              id="dirPort" type="number" filled
              label="Directory Service port (default 8886)"
      />
      <QInput v-model="editRecord.mqttPort"
              id="mqttPort" type="number" filled
              label="MQTT Broker Websocket Port (default 8885)"
      />
      </QCardSection>
      <QToggle v-model="editRecord.enabled"
              id="enabled" type="boolean" class="q-mt-md q-mb-none q-pb-none"
              label="Enabled"
      />
    </QForm>

  </TDialog>
</template>
