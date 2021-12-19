<script setup lang="ts">

import {TDProperty, ThingTD} from "@/data/td/ThingTD";
import TTable, {ITableCol} from "@/components/TTable.vue";

const props= defineProps<{td:ThingTD}>()


// columns to display properties
const attributesColumns = <Array<ITableCol>>[
  {name: "title", label: "Attributes", field:"title", align:"left",
    sortable:true},
  {name: "value", label: "Value", field:"value", align:"left",
    style:"max-width:200px; overflow-x: auto"
  },
  {name: "unit", label: "Unit", field:"unit", align:"left",
    style:"width:20px", sortable: true,
  },
]

</script>

<template>
  <TTable row-key="id"
          :columns="attributesColumns"
          :rows="ThingTD.GetThingAttributes(props.td)"
          no-data-label="No attributes available"
  />

</template>