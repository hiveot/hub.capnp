<script lang="ts" setup>

import { h } from 'vue';
import { TDProperty, ThingTD } from '@/data/td/ThingTD';
import TSimpleTable, { ISimpleTableColumn } from '@/components/TSimpleTable.vue';


const props = defineProps<{
  /**
   * The thing description document whose properties to show
   */
  td: ThingTD

  /**
   * List of properties to show. Default is the properties list from the td.
   * This allows for filtering of properties, eg show configuration.
   */
  propList?: TDProperty[]

  /**
   * Enable selecting a property. This shows a pointer cursor on the property name
   */
  enablePropSelect?: boolean
}>()

/**
 * Optional event when enabled
 */
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
    title: "Name", 
    field: "tdProperty.title",
    // maxWidth: "0",
    // width: "50%",
    component: (row:any) => h('span', 
      { 
        style: (props.enablePropSelect ? 'cursor:pointer':''), 
        onClick: ()=>handleThingPropertySelect(props.td, row),
      }, 
      {default: ()=>row.tdProperty.title}
    ),
  },
  {
    title: "Value", 
    // maxWidth: "0",
    // width: "50%",
    field: "tdProperty.value",
    component: (row:any)=>h('span', {}, 
        { default: ()=>getThingPropValue(row.tdProperty) }
      )
  },
  {title: "Type", field:"tdProperty.type", align:"left",
    // maxWidth: "0",
    width: "80px",
    sortable:true
  },
  {title: "Default", field:"tdProperty.default", align:"left",
    width: "70px",
    // maxWidth: "70px",
  },
]

</script>


<template>

 <TSimpleTable 
  dense
  :columns="propertyItemColumns"
  :rows="propList? propList : ThingTD.GetThingProperties(td)"
  :emptyText="'Thing \''+td.description+'\' has no properties'"
/>

</template>