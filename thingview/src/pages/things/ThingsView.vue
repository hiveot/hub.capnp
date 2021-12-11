<script lang="ts" setup>

import {useQuasar} from "quasar"
import {QCard, QCardSection, QIcon} from "quasar"
import {matLensBlur} from "@quasar/extras/material-icons"

import ds from "@/data/td/DirectoryStore"
import ThingsTable from "./ThingsTable.vue"
import { ThingTD } from "@/data/td/ThingTD"
import ThingDetails from "./ThingDetails.vue"

// const things = (ds.all: TDCollection):Array<ThingTD>  {
//     return Array.from(ds.index.values())
// }
// const $q = useQuasar()

const handleViewDetails = (row:ThingTD) => {
//  $q.dialog({
//     title:"Details of Thing: "+row.description,
//     cancel:true,
//     // component: ThingDetails,
//     componentProps: {td:row}
//   })
}


</script>


<template>

  <QCard class="q-pa-md" flat>
    <!-- Header -->
    <QCardSection horizontal>
      <QIcon :name="matLensBlur" left size="2rem" />
      <div class="text-h6">An overview of known Things</div>
    </QCardSection>

    <!-- List of all things -->
    <QCardSection class="text-h6">
      <ThingsTable  :things="ds.all"
                    title="Hub Things"
                    style="width: 100%"
                    @on-view-details="handleViewDetails"
      >
      </ThingsTable>
    </QCardSection>

  </QCard>
</template>
