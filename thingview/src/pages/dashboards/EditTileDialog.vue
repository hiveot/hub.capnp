<script lang="ts" setup>
import {h, ref, reactive} from "vue";
import {useDialogPluginComponent, QBtn, QForm, QIcon, QInput, QSelect} from "quasar";
import {matAdd, matRemove} from "@quasar/extras/material-icons";


import TDialog from "@/components/TDialog.vue";
import {
  DashboardDefinition, DashboardTileConfig, IDashboardTileItem,
  TileTypeCard, TileTypeImage
} from "@/data/dashboard/DashboardStore";
import SimpleTable, { ISimpleTableColumn } from "@/components/TSimpleTable.vue";
import TSimpleTable from "@/components/TSimpleTable.vue";

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

const propertyItemColumns:ISimpleTableColumn[] = [
  {title: "", width:"40px", field:"remove", 
    component:()=>h(QBtn,{icon:matRemove, round:true, flat:true})
  }, // remove
  {title: "Name", field: "name"},
  {title: "Value", field: "value"}
]

/**
 * Return the list of Thing properties and their value to display in a table
 */
const getThingsProperties = (items:IDashboardTileItem[]|undefined):{}[] => {
  let itemAndValues: {}[] = []
  if (items) {
    itemAndValues.push({name:"test item 1", value:"21"})
  }
  return itemAndValues
}

</script>

<template>
  <TDialog
      ref="dialogRef"
      width="400px"
      :title="props.title"
      @onSubmit="handleSubmit"
      showOk
      showCancel
  >
    <QForm @submit="handleSubmit"
           ref="formRef"
           >
      <QInput v-model="editTile.title"
              :autocomplete="TileTypeCard"
              autofocus  required
              id="title" type="text"
              label="Title"
              :rules="[()=>editTile.title !== ''||'Please provide a title']"
              stack-label
      />
      <QSelect v-model="editTile.type"
               :options="[
                  {label:'Card', value:TileTypeCard},
                  {label:'Image', value:TileTypeImage}]"
               :rules="[(val:any)=> (!!val && !!val.label && (val.label.length > 0)) || 'please select a valid type']"

               label="Type of tile"
      />
      <p>Properties <QBtn round flat :icon="matAdd"/></p>
      <TSimpleTable class="propTable"
          :columns="propertyItemColumns"
          :rows="getThingsProperties(props.tile?.items)"
      />
    </QForm>

  </TDialog>
</template>

<style scoped>
.prop-table > .thead {
  background-color: lightgray;
}
</style>