<script lang="ts" setup>

import {h, VNode, VNodeArrayChildren} from 'vue'

import TButton from '@/components/TButton.vue'
import { IDashboardTileItem } from '@/data/dashboard/DashboardStore';
import { ThingStore } from '@/data/td/ThingStore';
import { matRemove } from '@quasar/extras/material-icons';
import { TDProperty, ThingTD } from '@/data/td/ThingTD';
import { ISimpleTableColumn } from '@/components/TSimpleTable.vue';
import TSimpleTable from '../../components/TSimpleTable.vue';
import { QTooltip } from 'quasar';
import { PropNameDeviceType, PropNameName } from '@/data/td/Vocabulary';
import { get as _get} from 'lodash-es'


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

// tile item to display
interface IThingTileItem {
  key: string,
  item: IDashboardTileItem,
  td: ThingTD,
  tdProp: TDProperty,
}

/**
 * Return the list of Thing tile items from a dashboard tile item list
 */
const getThingTileItems = (items:IDashboardTileItem[]|undefined): 
  IThingTileItem[] => {

  let itemAndProps: IThingTileItem[] = []
  if (items) {
    items.forEach(item=>{
      let td = props.thingStore.GetThingTDById(item.thingID)
      let tdProp = td?.properties[item.propertyID]
      if (td && tdProp) {
        itemAndProps.push({
          key: item.thingID+"."+item.propertyID,
          item: item, 
          td: td,
          tdProp: tdProp,
        })
      }
    })
  }
  return itemAndProps
}

const getThingPropValue = (item:IThingTileItem):string => {
  if (!item || !item.tdProp) {
    return "Missing value"
  }
  let valueStr = item.tdProp.value + " " + (item.tdProp?.unit ? item.tdProp?.unit:"")
  return valueStr
}

/**
 * Return the property name for presentation in the following order
 *  1. DashboardTileItem's label     (which is an override)
 *  2. thing's name property   (yes, not the property name)
 *  3. thing's description
 *  4. property title
 *  
 * This uses the thing's name or description if name is not configured
 * A tooltip shows more info
 */
const getThingPropName = (tileItem:IThingTileItem):VNode => {
  // 1. default use the tile item's label override
  let thingName = tileItem.item.label
  // 2. use thing's name property  (not property name)
  if (!thingName) {
    let tdProps = tileItem.td.properties
    let thingNameProp = _get(tdProps, PropNameName)
    if (thingNameProp) {
      thingName = thingNameProp.value
    }
  }
  // 3. use thing's description (if no name property is defined)
  if (!thingName) {
    thingName = tileItem.td?.description
    if (tileItem.td.deviceType) {
      thingName += " (" + tileItem.td.deviceType + ")"
    }
  }
  // 4. finally, use property 'title' 
  if (!thingName) {
    thingName = tileItem.tdProp.title
  }

  let comp = h('span',
              {'style': 'width:"100%"'},
              [ tileItem.tdProp.title,
                h(QTooltip, {
                    style: 'font-size:inherit',
                  }, ()=>thingName
                ),
              ]
  )
  return comp
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
      // maxWidth: "35px",
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
      // Show property name. TODO: use label field
      title: "Property Name", 
      field: "tdProp.title", 
      // width: "%", 
      // maxWidth: "0",
      component: (row:IThingTileItem)=>h('span',
         {'style': 'width:"100%"'},
         getThingPropName(row)
        //  [ row.tdProp.title,
        //    h(QTooltip, {
        //      style: 'font-size:inherit',
        //      }, ()=>row.item.thingID
        //    ),
        //  ]
      ),
      align: 'left'
    }, {
      // show value and unit
      title: "Value", 
      field: "tdProp.value",
      // width: "50%",
      // maxWidth: "0",
      component: (row:IThingTileItem)=>h('span', {}, 
        { default: ()=>getThingPropValue(row) }
      )
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

