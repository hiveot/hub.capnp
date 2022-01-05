<script lang="ts" setup>

import { reactive } from 'vue';
import { useDialogPluginComponent, QCard, QInput, QList, QExpansionItem } from 'quasar';

import TDialog from '@/components/TDialog.vue';
import {DashboardTileConfig} from "@/data/dashboard/DashboardStore";
import { TDProperty, ThingTD } from '@/data/td/ThingTD';
import { ThingStore } from '@/data/td/ThingStore';
import TSimpleTable, { ISimpleTableColumn } from '@/components/TSimpleTable.vue';

// inject handlers
const { dialogRef, onDialogOK, onDialogCancel } = useDialogPluginComponent();

/**
 *  This dialog shows a selection of things and their properties for adding to a dashboard tile
 */
const props = defineProps<{
  tile: DashboardTileConfig
  thingStore: ThingStore
}>()

const emits = defineEmits( [
  // REQUIRED; need to specify some events that your
  // component will emit through useDialogPluginComponent()
  ...useDialogPluginComponent.emits,
]);


const data = reactive({
  selectedThing: new ThingTD(),
  selectedProp: TDProperty,
  searchInput: "",
})

const handleRowSelect = (thing:ThingTD, keyProp:{key:string, prop:TDProperty}) => {
  console.log("SelectTilePropertyDialog.handleRowSelect: Selected property '%s' from thing '%s'",
   keyProp.key, thing.description)
  onDialogOK({thingID: thing.id, propID: keyProp.key})
}



/**
 * Get a list of all properties of all tiles in the thing store
 */
const getAllThings = ():Array<ThingTD> => {
  let tdList = props.thingStore.all
  return tdList
}

/**
 * Table columns from the tile item rows: [{key:string, prop:TDProperty}]
 */
const propertyItemColumns:ISimpleTableColumn[] = [
  {title: "Name", field: "prop.title"},
  {title: "Value", field: "prop.value"}
]


</script>

<template>

<TDialog  ref="dialogRef"
  title="Select Property to Add"
  @onClose="onDialogCancel"
  showClose
>
  <QCard>
      <QInput label="Search" v-model="data.searchInput"/>
      <QList>
        <QExpansionItem v-for="td in getAllThings()"
          :label="td.description + ' ('+td.id+')'" 
          group="tdgroup"
          >
          <div style="padding-left: 2em; width:99%">
            <TSimpleTable dense
                :columns="propertyItemColumns"
                :rows="ThingTD.GetThingProperties(td)"
                :emptyText="'Thing \''+td.description+'\' has no properties'"
                @on-row-select="(keyProp)=>{handleRowSelect(td, keyProp)}"
            />
          </div>
        </QExpansionItem>
      </QList>

  </QCard>
</TDialog>

</template>