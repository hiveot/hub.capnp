<script setup lang="ts">

import {TDAction, ThingTD} from "@/data/td/ThingTD";
import TTable, {ITableCol} from "@/components/TTable.vue";

const props= defineProps<{td:ThingTD}>()

// columns to display action
const actionColumns = <Array<ITableCol>>[
  {name: "title", label: "Action", field:"title", align:"left",
    sortable:true},
  {name: "description", label: "Description", field:"description", align:"left"
  },
]

</script>



<template>

  <TTable row-key="id"
          :columns="actionColumns"
          :rows="ThingTD.GetThingActions(props.td)"
          no-data-label="No actions available"
  />

</template>