<script lang="ts" setup>

import {h} from 'vue'

import TButton from '@/components/TButton.vue'
import { IDashboardTileItem } from '@/data/dashboard/DashboardStore';
import { ThingStore } from '@/data/td/ThingStore';
import { matRemove } from '@quasar/extras/material-icons';
import { TDProperty } from '@/data/td/ThingTD';
import { ISimpleTableColumn } from '@/components/TSimpleTable.vue';
import TSimpleTable from '../../components/TSimpleTable.vue';


/**
 * Table that shows a list of dashboard tile items
 */
const props = defineProps<{
  /**
   * Reduce padding to compact layout
   */
  dense?: boolean
  /**
   * In edit mode show the remove item button as first column
   */
  editMode?: boolean
  /**
   * Hide the border box
   */
  flat?: boolean
  /**
   * Grow row content to use the available height
   */
  grow?: boolean
  /**
   * Hide the table border
   */
  noBorder?: boolean
  /**
   * Hide the header
   */
  noHeader?: boolean
  /**
   * The rows to display
   */
  tileItems:IDashboardTileItem[]
  /**
   * Lookup item values from this store
   */
  thingStore: ThingStore
}>()

const emits = defineEmits(["onRemoveTileItem"])

// Look up the property from the thing ID and property ID
const getThingProperty = (thingID:string, propID:string):TDProperty|undefined => {
  let td = props.thingStore.GetThingTDById(thingID)
  let prop = td?.properties[propID]
  return prop
}



// tile item to display
interface IThingTileItem {
  key: string,
  item: IDashboardTileItem,
  prop: TDProperty
}


/**
 * Return the list of Thing tile items from a dashboard tile item list
 */
const getThingTileItems = (items:IDashboardTileItem[]|undefined): 
  IThingTileItem[] => {

  let itemAndProps: IThingTileItem[] = []
  if (items) {
    items.forEach(item=>{
      let tdProp =  getThingProperty(item.thingID, item.propertyID)
      if (tdProp) {
        itemAndProps.push({
          key: item.thingID+"."+item.propertyID,
          item: item, 
          prop: tdProp
        })
      }
    })
  }
  return itemAndProps
}

const getThingPropValue = (item:IThingTileItem):string => {
  if (!item || !item.prop) {
    return "Missing value"
  }
  let valueStr = item.prop.value + " " + (item.prop?.unit ? item.prop?.unit:"")
  return valueStr
}

/**
 * User clicked the 'remove' button while in edit mode
 */
const handleRemove = (row:any) => {
  console.log("TileItemsTable.handleRemove: row=", row, "row.row",row.row)
  emits("onRemoveTileItem", row.item)
}

/**
 * Thing properties table columns 
 *   display items from IThingTileItem 
 */
const getColumns = (editMode:boolean|undefined):ISimpleTableColumn[] => {
  return [ {
    // show the 'remove' button only in edit mode
      title: "", 
      width:"35px", 
      field:"remove", 
      align:'center', 
      hidden: !props.editMode,
      component:(row:IThingTileItem)=>h(TButton,{
          icon:matRemove, round:true, dense:true, flat:true, height:'10px', 
          style: "min-width: 1.5em",
          tooltip:"Remove property from tile",
          onClick: ()=>handleRemove(row),
        }),
    }, {
      // Show property title. TODO: use label field
      title: "Name", 
      field: "prop.title", 
      align: 'left'
    }, {
      // show value and unit
      title: "Value", 
      field: "prop.value",
      component: (row:IThingTileItem)=>h('span',
          null, getThingPropValue(row))
    }
  ]
}
</script>


<template>
    <TSimpleTable 
        :columns="getColumns(props.editMode)"
        :rows="getThingTileItems(props.tileItems)"
        :flat="props.flat"
        :no-border="props.noBorder"
        :no-header="props.noHeader"
        :grow="props.grow"
        :dense="props.dense"
        empty-text="Please add tile properties..."
    />
</template>

