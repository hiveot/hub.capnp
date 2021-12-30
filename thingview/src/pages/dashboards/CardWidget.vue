<script lang="ts" setup>

import {QCard, QCardSection} from "quasar";
import {DashboardTileConfig, IDashboardTileItem} from "@/data/dashboard/DashboardStore";
import SimpleTable, {ISimpleTableColumn} from '@/components/SimpleTable.vue'
import {ThingStore} from "@/data/td/ThingStore";
import {TDProperty} from "@/data/td/ThingTD";
import { ref } from "vue";

const props= defineProps<{
  config:DashboardTileConfig
  ts: ThingStore
}>()

interface IDisplayItem {
  label: string
  property?: TDProperty
}

/**
 * Get array of thing attributes to display on the tile
 * This returns an array of items: [{thingID, thingAttr}]
 */
const getThingsAttributes = (config:DashboardTileConfig): IDisplayItem[] => {
  let res = new Array<IDisplayItem>()
  config.items.forEach( (item: IDashboardTileItem) => {
    let td =props.ts.GetThingTDById(item.thingID)
    let tdProp:TDProperty|undefined = td?.properties.get(item.propertyID)
    let di:IDisplayItem = {label: item.propertyID, property:tdProp}
    res.push(di)
  })
  return res
}

/**
 * Column definition when multiple items are displayed in a table
 */
const itemColumns:ISimpleTableColumn[] = [
  {title:"label", field:"property.title"}
]

const item0 = ref(props.config?.items?.[0])
console.info("CardWidget. props.config=", props.config)

</script>

<template>
  <QCard>
    <QCardSection>
      <div>{{config}} </div>
      </QCardSection>
    <QCardSection>
      <SimpleTable
          v-if="props.config.items?.length>1"
          :hideHeader="true"
          dense
          :columns="itemColumns"
          :rows="getThingsAttributes(props.config)"
      />
      <div v-else>
        {{item0?.label}}
      </div>
    </QCardSection>
  </QCard>
</template>