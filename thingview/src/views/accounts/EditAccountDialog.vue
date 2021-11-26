<script lang="ts" setup>
import {ref, watchEffect} from "vue";
import TDialog from "@/components/TDialog.vue";
import {QForm, QInput} from "quasar";
import {AccountRecord} from "@/data/AccountStore";

// since EditAccountDialog is not recreated when re-used, the account is not updated
//
interface IEditAccountDialog {
  account: AccountRecord
  visible: boolean
}
const props = defineProps<IEditAccountDialog>()
let editRecord = ref(new AccountRecord())

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
    <QForm @submit="handleSubmit" class="q-gutter-md" style="min-width: 350px">
      <QInput v-model="editRecord.name" autofocus
              id="accountName" type="text"
              label="Account name"
              :rules="[()=>editRecord.name !== ''||'Please provide a name']"
      />
      <QInput v-model="editRecord.mqttPort"
              id="mqttPort" type="number"
              label="MQTT port"
      />
    </QForm>

  </TDialog>
</template>
