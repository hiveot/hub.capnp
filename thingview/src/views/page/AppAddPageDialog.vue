<script lang="ts" setup>
import {ref} from "vue";
import TDialog from "@/components/TDialog.vue";
import {QDialog, QBtn, QCard, QBar, QSpace, QCardActions, QCardSection, QForm, QInput} from "quasar";
import {mdiClose} from "@quasar/extras/mdi-v6";

const props = defineProps({
    visible:Boolean,
  })

const emit = defineEmits({
    'onAdd': String,
    'onClosed': null,
    })

const pageName = ref("");

const handleSubmit = (ev:any) =>{
  console.log("AppAddPageDialog.handleSubmit: ", pageName.value)
  emit("onAdd", pageName.value);
  pageName.value = "";
  emit("onClosed");
 };
console.log("AddPageDialog: visible=",props.visible);

const handleCancel = (ev:any) => {
  console.log("AppAddPageDialog.cancelled")
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
    <QForm @submit="handleSubmit" class="q-gutter-md" style="min-width: 350px">
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
