<script lang="ts" setup>
import {ref} from "vue";
import TDialog from "@/components/TDialog.vue";
import {QForm, QInput} from "quasar";

import {mdiInformation} from '@quasar/extras/mdi-v6'
const props = defineProps({
    visible:Boolean,
  })

const emit = defineEmits({
    'onAdd': String,
    'onClosed': null,
    })

const pageName = ref("");

const handleSubmit = () =>{
  console.log("AppAddPageDialog.handleSubmit: ", pageName.value)
  emit("onAdd", pageName.value);
  pageName.value = "";
  debugger
  emit("onClosed");
 };
console.log("AddPageDialog: visible=",props.visible);

const handleCancel = () => {
  console.log("AppAddPageDialog.cancelled")
  debugger
  emit("onClosed");
}

</script>

<template>
  <TDialog
      :visible="props.visible"
      title="New Dashboard Page"
      @onClosed="handleCancel"
      @onSubmit="handleSubmit"
      showCancel showOk
      :okDisabled="(pageName==='')"
  >
    <QForm class="q-gutter-md" style="min-width: 350px">
          <QInput v-model="pageName"
                  no-error-icon
                  autofocus filled required lazy-rules
                  id="pageName" type="text"
                  label="Page name"
                  hint="Name of the dashboard as shown on the Tabs"
                  :rules="[()=>pageName !== ''||'Please provide a name']"
                  stack-label
          >
          </QInput>
    </QForm>

  </TDialog>
</template>
