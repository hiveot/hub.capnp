<script setup lang="ts">

import {ref} from 'vue';
import {TDProperty, ThingTD} from "@/data/td/ThingTD";
import TTable, {ITableCol} from "@/components/TTable.vue";

const props= defineProps<{td:ThingTD}>()


// Convert the writable properties into an array for display
const getThingConfiguration = (td: ThingTD): Array<TDProperty> => {
  let res = Array<TDProperty>()
  if (!!td && !!td.properties) {
    for (let [key, val] of Object.entries(td.properties)) {
      if (val.writable) {
        res.push(val)
      }
    }
  }
  return res
}


// columns to display configuration
const configurationColumns = <Array<ITableCol>>[
  {name: "title", label: "Configuration", field:"title", align:"left",
    sortable:true},
  {name: "value", label: "Value", field:"value", align:"left",
    style:"max-width:400px; overflow: auto"},
  {name: "type", label: "Type", field:"type", align:"left",
    sortable:true},
  {name: "default", label: "Default", field:"default", align:"left"
  },
  {name: "unit", label: "Unit", field:"unit", align:"left"
  },
]

</script>

<template>

  <TTable row-key="id"
          :columns="configurationColumns"
          :rows="getThingConfiguration(props.td)"
          no-data-label="No configuration available"
  />

</template>

