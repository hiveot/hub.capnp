<script lang="ts" setup>

import { h } from 'vue';
import { QItem } from 'quasar';
import { TDProperty, ThingTD } from '@/data/td/ThingTD';
import TSimpleTable, { ISimpleTableColumn } from '@/components/TSimpleTable.vue';


const props = defineProps<{
  /**
   * The thing description document whose properties to show
   */
  td: ThingTD
}>()


const emits = defineEmits(["onThingPropertySelect"])

const getThingPropValue = (tdProp:TDProperty):string => {
  if (!tdProp) {
    return "Missing value"
  }
  let valueStr = tdProp.value + " " + (tdProp?.unit ? tdProp?.unit:"")
  return valueStr
}

/**
 * Select property 
 */
const handleThingPropertySelect = (td:any,propInfo:any)=>{
  console.log("ThingPropertiesTable.handleThingPropertySelect, \
      thingID=%s, propID=%s, thingProperty", td.id, propInfo.propID, propInfo.tdProperty)
  emits("onThingPropertySelect", td, propInfo.propID, propInfo.tdProperty)
}

/**
 * Table columns from the tile item rows: [{propID:string, tdProperty:TDProperty}]
 */
const propertyItemColumns:ISimpleTableColumn[] = [
  {
    title: "Property Name", 
    field: "",
    component: (row:any) => h(QItem, 
      { 
        dense: true,
        clickable: true,
        style: "padding:0",
        // style: 'cursor:pointer', 
        onClick: ()=>handleThingPropertySelect(props.td, row),
      }, 
      {default: ()=>row.tdProperty.title}
    ),
  },
  {
    title: "Value", 
    field: "tdProperty.value",
      component: (row:any)=>h('span', {}, 
        { default: ()=>getThingPropValue(row.tdProperty) }
      )
  },
  // {
  //   title: "key", 
  //   field: "propID",
  // }
]

</script>


<template>

 <TSimpleTable dense
                :columns="propertyItemColumns"
                :rows="ThingTD.GetThingProperties(td)"
                :emptyText="'Thing \''+td.description+'\' has no properties'"
            />

</template>