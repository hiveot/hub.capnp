<script setup lang="ts">

import {ThingTD} from "@/data/td/ThingTD";
import TSimpleTable, { ISimpleTableColumn } from "@/components/TSimpleTable.vue";

/**
 * Show a table of all properties of a thing TD
 */
const props= defineProps<{
  td:ThingTD,
  // onlyConfiguration?: boolean, 
  // noConfiguration?: boolean,
  }>()


// columns to display properties
// A row contains {key:string, prop:TDProperty}
const tdPropColumns = <Array<ISimpleTableColumn>>[
  {field:"prop.title", title: "Property", align:"left", width: "50px",
    //style:"max-width:300px",  
    sortable:true
  },

  {field:"prop.value", title: "Value", align:"left",
    style:"max-width:200px; overflow-x: auto"
  },

  {field:"prop.unit", title: "Unit", align:"left", width: "20px", sortable: true },
  
]

</script>

<template>
  <TSimpleTable row-key="id"
          :columns="tdPropColumns"
          :rows="ThingTD.GetThingProperties(props.td)"
          dense
          empty-text="No properties available"

  />

</template>