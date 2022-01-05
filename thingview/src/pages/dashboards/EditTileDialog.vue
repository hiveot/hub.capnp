<script lang="ts" setup>
import {ref, reactive} from "vue";
import {cloneDeep as _cloneDeep, remove as _remove} from 'lodash-es'
import {useDialogPluginComponent, useQuasar, QForm, QInput, QSelect} from "quasar";
import {matAdd} from "@quasar/extras/material-icons";

import TDialog from "@/components/TDialog.vue";
import TButton from "@/components/TButton.vue";

const $q = useQuasar()

import { DashboardTileConfig, IDashboardTileItem, TileTypeCard, TileTypeImage } from "@/data/dashboard/DashboardStore";
import thingStore from '@/data/td/ThingStore'
import SelectTilePropertyDialog from "./SelectTilePropertyDialog.vue";
import ThingPropertiesTable from "../ThingPropertiesTable.vue";
import TileItemsTable from "./TileItemsTable.vue";

// inject handlers
const { dialogRef, onDialogOK, onDialogCancel } = useDialogPluginComponent();

const props = defineProps<{
  title: string,
  tile?: DashboardTileConfig,
}>()

const formRef = ref()

// editTile is a copy the tile being edited or empty on add
const editTile:DashboardTileConfig = reactive<DashboardTileConfig>(
    props.tile ? _cloneDeep(props.tile) : new DashboardTileConfig()
);

const emits = defineEmits( [
  // REQUIRED; need to specify some events that your
  // component will emit through useDialogPluginComponent()
  ...useDialogPluginComponent.emits,
]);


// popup an output selection dialog to add an output to the tile
const handleAddTileProperty = () => {
  console.info("EditTileDialog.handleAddTileProperty: Showing the add tile property dialog")
  $q.dialog({
     component: SelectTilePropertyDialog,
     componentProps: {
       tile: editTile,
       thingStore: thingStore,
      },
  // }).onOk( ({thingID, propID})=>{
  }).onOk( (props)=>{
    let thingID=props.thingID
    let propID=props.propID
    // Add a new view property to the tile
    console.log("EditTileDialog.handleAddTileProperty: props:", props)
    editTile.items.push({thingID:thingID, propertyID:propID})
    })
}

// Remove the tile item from the list of items
const handleRemoveTileItem = (tileItem:IDashboardTileItem) => {
  console.info("EditTileDialog.handleRemoveTileItem. item=",tileItem)
  _remove(editTile.items, (item) => (item.propertyID === tileItem.propertyID && (item.thingID === item.thingID)))
}

// Submit the updated Tile
const handleSubmit = () =>{
  console.log("EditTileDialog.handleSubmit: ", editTile)
  // put focus on invalid component
  formRef.value.validate(true)
      .then((isValid:boolean)=>{
        if (isValid) {
          console.info("EditTileDialog.handleSubmit tile is valid")
          onDialogOK(editTile)
        } else {
          console.info("EditTileDialog.handleSubmit tile is not valid")
        }
      })
};

</script>

<template>
  <TDialog
      ref="dialogRef"
      :title="props.title"
      @onClose="onDialogCancel"
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
               map-options  emit-value
               :options="[
                  {label:'Card', value:TileTypeCard},
                  {label:'Image', value:TileTypeImage}]"
               :rules="[(val:any)=> 
                    (!!val && (val.length > 0)) || 
                        'please select a valid type: '+val
                        ]"

               label="Type of tile"
      />
      <p>Properties 
        <TButton round flat 
                 :icon="matAdd" 
                 tooltip="Add property to tile"
                 @click="handleAddTileProperty"
        />
      </p>
      <TileItemsTable v-if="editTile.items"
        :tileItems="editTile.items"
        :thingStore="thingStore"
        dense flat
        edit-mode
        @on-remove-tile-item="handleRemoveTileItem"  
        />
      <div v-else>Missing dashboard tile items</div>
    </QForm>

  </TDialog>
</template>

<style scoped>
</style>


<style>

/** follow app/dialog font size rules */
.q-field {
  font-size: inherit !important;
}
</style>