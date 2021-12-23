<script lang="ts" setup>
import {reactive} from "vue";
import {useDialogPluginComponent, QForm, QInput} from "quasar";

import TDialog from "@/components/TDialog.vue";
import {DashboardDefinition, DashboardTile} from "@/data/dashboard/DashboardStore";

// inject handlers
const { dialogRef, onDialogOK } = useDialogPluginComponent();


const props = defineProps<{
  title: string,
  tile?: DashboardTile,
}>()

// editTile is a copy the tile being edited or empty on add
const editTile = reactive<DashboardTile>(
    props.tile ? {...props.tile} : new DashboardTile()
);

const emits = defineEmits( [
  // REQUIRED; need to specify some events that your
  // component will emit through useDialogPluginComponent()
  ...useDialogPluginComponent.emits,
]);

const handleSubmit = () =>{
  console.log("EditTileDialog.handleSubmit: ", editTile)
  onDialogOK(editTile)
};

</script>

<template>
  <TDialog
      ref="dialogRef"
      :title="props.title"
      @onSubmit="handleSubmit"
      showCancel showOk
  >
    <QForm @submit="handleSubmit" class="q-gutter-md" style="min-width: 350px">
      <QInput v-model="editTile.title"
              autofocus filled required
              id="title" type="text"
              label="Title"
              hint="Title of this tile"
              :rules="[()=>editTile.title !== ''||'Please provide a title']"
              stack-label
      />
    </QForm>

  </TDialog>
</template>
