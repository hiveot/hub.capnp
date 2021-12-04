<!--<script lang="ts" setup>-->
<!--import {ref, watchEffect} from "vue";-->
<!--import TDialog2 from "@/components/TDialog.vue";-->
<!--import {QForm, QInput} from "quasar";-->
<!--import {AccountRecord} from "@/data/AccountStore";-->

<!--// since EditAccountDialog is not recreated when re-used, the account is not updated-->
<!--//-->
<!--interface IEditAccountDialog {-->
<!--  account: AccountRecord-->
<!--}-->
<!--const props = defineProps<IEditAccountDialog>()-->
<!--let editRecord = ref(new AccountRecord())-->

<!--// replace the edit record when it is updated in props-->
<!--watchEffect(()=>{-->
<!--  Object.assign(editRecord.value, props.account)-->
<!--})-->

<!--</script>-->

<!--<template>-->
<!--  <TDialog2-->
<!--  :visible="true"-->
<!--      title="Edit Account"-->
<!--      :okDisabled="(editRecord.name==='')"-->
<!--  :showOk="true"-->
<!--  :showCancel="true"-->
<!--  >-->
<!--    <QForm class="q-gutter-md" style="min-width: 350px">-->
<!--      <QInput v-model="editRecord.name" autofocus-->
<!--              id="accountName" type="text"-->
<!--              label="Account name"-->
<!--              :rules="[()=>editRecord.name !== ''||'Please provide a name']"-->
<!--      />-->
<!--      <QInput v-model="editRecord.mqttPort"-->
<!--              id="mqttPort" type="number"-->
<!--              label="MQTT port"-->
<!--      />-->
<!--    </QForm>-->

<!--  </TDialog2>-->
<!--</template>-->
