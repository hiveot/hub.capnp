<script setup lang="ts">

import {QTable} from 'quasar';
import {TDEvent, ThingTD} from "@/data/td/ThingTD";
import TTable, {ITableCol} from "@/components/TTable.vue";

const props= defineProps<{td:ThingTD}>()

const getThingEvents = (td: ThingTD): Array<TDEvent> => {
  let res = Array<TDEvent>()
  if (!!td && !!td.events) {
    for (let [key, val] of Object.entries(td.actions)) {
      res.push(val)
    }
  }
  return res
}
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
          :rows="getThingEvents(props.td)"
          no-data-label="No events available"
  />
</template>