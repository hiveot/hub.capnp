<script lang="ts" setup>

import { reactive } from 'vue';
import { useDialogPluginComponent, QCard, QInput, QList, QExpansionItem } from 'quasar';

import TDialog from '@/components/TDialog.vue';
import { TDProperty, ThingTD } from '@/data/td/ThingTD';
import ThingPropertiesTable from './ThingPropertiesTable.vue';

// inject handlers
const { dialogRef, onDialogOK, onDialogCancel } = useDialogPluginComponent();

/**
 *  This dialog shows a selection of things and their properties for adding to a dashboard tile
 */
const props = defineProps<{
  /**
   * The things to select from
   */
  things: ThingTD[]
}>()

const emits = defineEmits( [
  // REQUIRED; need to specify some events that your
  // component will emit through useDialogPluginComponent()
  ...useDialogPluginComponent.emits,
]);

const data = reactive({
  // selectedThing: new ThingTD(),
  // selectedProp: TDProperty,
  searchInput: "",
})

/**
 * Get a list of things to display
 */
const getAllThings = ():Array<ThingTD> => {
  let tdList = props.things
  return tdList
}


const handleThingPropertySelect = (td:ThingTD, propID:string, tdProp:TDProperty)=>{
  onDialogOK({td:td, thingID:td.id, propID:propID, tdProp:tdProp})
}


</script>

<template>

<TDialog  ref="dialogRef"
  title="Select Thing Property to Add"
  @onClose="onDialogCancel"
  showClose
>
  <QCard>
      <!-- Search to reduce the amount of Things to select from -->
      <QInput label="Search" v-model="data.searchInput"/>
      <QList>
        <!-- Accordion to select a Thing and view its properties -->
        <QExpansionItem v-for="td in getAllThings()"
          :label="td.publisher + (td.description ? (' - ' + td.description) : '')" 
          :label-lines="1"
          group="tdgroup"
          :caption-lines="1"
          switch-toggle-sides
          :expand-separator="false"
          :content-inset-level="0.5"
        >
          <div style="width:99%">
            <p style="font-size: small; font-style: italic;">ID: {{td.id}}</p>
            
            <ThingPropertiesTable :td="td"
            @onThingPropertySelect="handleThingPropertySelect"
            />

          </div>
        </QExpansionItem>
      </QList>

  </QCard>
</TDialog>

</template>