<script setup lang="ts">

import {TDEvent, ThingTD} from "@/data/td/ThingTD";
import TTable, {ITableCol} from "@/components/TTable.vue";

const props= defineProps<{td:ThingTD}>()

// columns to display events (outputs)
const eventColumns = <Array<ITableCol>>[
  {name: "name", label: "Event", field:"name", align:"left",
    sortable:true},
  {name: "params", label: "Parameters", field:"params", align: "left"
  },
]
</script>



<template>
  <TTable row-key="id"
          :columns="eventColumns"
          :rows="ThingTD.GetThingEvents(props.td)"
          no-data-label="No events available"
  />
</template>