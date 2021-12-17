<script setup lang="ts">

import {TDProperty, ThingTD} from "@/data/td/ThingTD";
import TTable, {ITableCol} from "@/components/TTable.vue";

const props= defineProps<{td:ThingTD}>()

// Convert the attributes into an array for display
const getThingAttributes = (td: ThingTD): Array<TDProperty> => {
  let res = Array<TDProperty>()
  if (!!td && !!td.properties) {
    for (let [key, val] of Object.entries(td.properties)) {
      if (!val.writable) {
        res.push(val)
      }
    }
  }
  return res
}

// columns to display properties
const attributesColumns = <Array<ITableCol>>[
  {name: "title", label: "Attributes", field:"title", align:"left",
    sortable:true},
  {name: "value", label: "Value", field:"value", align:"left",
    style:"max-width:200px; overflow-x: auto"
  },
  {name: "unit", label: "Unit", field:"unit", align:"left",
    sortable: true,
  },
]

</script>

<template>
  <TTable row-key="id"
          :columns="attributesColumns"
          :rows="getThingAttributes(props.td)"
          no-data-label="No attributes available"
  />

</template>