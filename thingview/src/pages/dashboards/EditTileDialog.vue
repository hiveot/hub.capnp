<script lang="ts" setup>
import {ref, reactive} from "vue";
import {useDialogPluginComponent, QBtn, QForm, QInput, QSpace, QSelect} from "quasar";

import TDialog from "@/components/TDialog.vue";
import {DashboardDefinition, DashboardTileConfig,
  TileTypeCard, TileTypeImage} from "@/data/dashboard/DashboardStore";

// inject handlers
const { dialogRef, onDialogOK, onDialogCancel } = useDialogPluginComponent();

const props = defineProps<{
  title: string,
  tile?: DashboardTileConfig,
}>()

const formRef = ref()

// editTile is a copy the tile being edited or empty on add
const editTile:DashboardTileConfig = reactive<DashboardTileConfig>(
    props.tile ? {...props.tile} : new DashboardTileConfig()
);

const emits = defineEmits( [
  // REQUIRED; need to specify some events that your
  // component will emit through useDialogPluginComponent()
  ...useDialogPluginComponent.emits,
]);

const handleSubmit = () =>{
  console.log("EditTileDialog.handleSubmit: ", editTile)
  // put focus on invalid component
  formRef.value.validate(true)
      .then((isValid:boolean)=>{
        if (isValid) {
          console.info("EditTileDialog.handleSubmit:",isValid)
          onDialogOK(editTile)
        } else {
          console.info("EditTileDialog.handleSubmit invalid")
        }
      })
};

</script>

<template>
  <TDialog
      ref="dialogRef"
      :title="props.title"
      @onSubmit="handleSubmit"
      showOk
      showCancel
  >
    <QForm @submit="handleSubmit"
           ref="formRef"
           class="q-gutter-md" style="min-width: 350px">
      <QInput v-model="editTile.title"
              :autocomplete="TileTypeCard"
              autofocus filled required
              id="title" type="text"
              label="Title"
              :rules="[()=>editTile.title !== ''||'Please provide a title']"
              stack-label
      />
      <QSelect v-model="editTile.type"
               :options="[
                  {label:'Card', value:TileTypeCard},
                  {label:'Image', value:TileTypeImage}]"
               :rules="[val=> (!!val && !!val.label && (val.label.length > 0)) || 'please select a valid type']"
               options-dense
               menu-shrink
               filled
               label="Type of tile"
      />
<!--      <QBtn label="Cancel"-->
<!--            @click="onDialogCancel"-->
<!--      />-->
<!--      <QSpace/>-->
<!--      <QBtn label="Save"-->
<!--            color="primary"-->
<!--            type="submit"-->
<!--      />-->
    </QForm>

  </TDialog>
</template>
