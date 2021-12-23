<script lang="ts" setup>
import {reactive} from "vue";
import {useDialogPluginComponent, QDialog, QForm, QInput} from "quasar";

import TDialog from "@/components/TDialog.vue";
import {DashboardDefinition} from '@/data/dashboard/DashboardStore'

// inject handlers
const { dialogRef, onDialogOK, onDialogCancel } = useDialogPluginComponent();

// EditDashboardDialog properties
const props = defineProps<{
    title: string,
    dashboard?: DashboardDefinition,
}>()

// editDash is a copy the dashboard being edited or empty on add
const editDash = reactive<DashboardDefinition>(
    props.dashboard ? {...props.dashboard} : new DashboardDefinition()
);

const emits = defineEmits( [
    // REQUIRED; need to specify some events that your
    // component will emit through useDialogPluginComponent()
    ...useDialogPluginComponent.emits,
]);

const handleSubmit = () =>{
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
      :okDisabled="(editDash.name==='')"
  >
    <QForm @submit="handleSubmit" class="q-gutter-md" style="min-width: 350px">
          <QInput v-model="editDash.name"
                  no-error-icon
                  autofocus filled required lazy-rules
                  label="Dashboard name"
                  :rules="[()=>editDash.name !== ''||'Please provide a name']"
                  stack-label
          />
    </QForm>

  </TDialog>
</template>
