<script lang="ts" setup>
import {reactive} from "vue";
import {useDialogPluginComponent, QForm, QInput} from "quasar";

import TDialog from "@/components/TDialog.vue";
import {IDashboardRecord} from '@/data/AppState'

// inject handlers
const { dialogRef, onDialogOK, onDialogCancel } = useDialogPluginComponent();


const props = defineProps<{
    title: string,
    dashboard?: IDashboardRecord,
}>()

// editDash is a copy the dashboard being edited or empty on add
const editDash = reactive<IDashboardRecord>(
    props.dashboard ? {...props.dashboard} : {}
);

const emits = defineEmits( [
    // REQUIRED; need to specify some events that your
    // component will emit through useDialogPluginComponent()
    ...useDialogPluginComponent.emits,
]);

const handleSubmit = (ev:any) =>{
  console.log("AppAddPageDialog.handleSubmit: ", editDash)
  onDialogOK(editDash)
 };

</script>

<template>
  <TDialog 
      ref="dialogRef" 
      :title="props.title"
      @onClosed="onDialogCancel"
      @onSubmit="handleSubmit"
      showCancel showOk
      :okDisabled="(editDash.label==='')"
  >
    <QForm @submit="handleSubmit" class="q-gutter-md" style="min-width: 350px">
          <QInput v-model="editDash.label"
                  no-error-icon
                  autofocus filled required lazy-rules
                  id="pageName" type="text"
                  label="Page name"
                  :hint="'Name of the dashboard as shown on the Tabs (' + props.dashboard?.label"
                  :rules="[()=>editDash.label !== ''||'Please provide a name']"
                  stack-label
          >
          </QInput>
    </QForm>

  </TDialog>
</template>
