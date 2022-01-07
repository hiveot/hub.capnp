<script lang="ts" setup>

// import {useQuasar} from "quasar"
import {reactive, ref} from 'vue'
import {QCard, QCardSection, QIcon} from "quasar"
import {matLensBlur} from "@quasar/extras/material-icons"

import ds from "@/data/td/ThingStore"
// import ThingsTable from "./ThingsTable.vue"
import { ThingTD } from "@/data/td/ThingTD"
// import ThingDetails from "./ThingDetailsDialog.vue"
import ThingsTable2 from "./ThingsTable2.vue"

// const things = (ds.all: TDCollection):Array<ThingTD>  {
//     return Array.from(ds.index.values())
// }
// const $q = useQuasar()


const tdToShow = ref(new ThingTD())
const data =reactive({
  showDetails: false,
  tdToShow: new ThingTD()
})

const handleViewDetails = (row:ThingTD) => {
  console.log("ThingsView.handleViewDetails showing the TD details. TD=",row)
  tdToShow.value = row
  data.tdToShow = row
  data.showDetails = !data.showDetails
}

</script>


<template>

<!--  <ThingDetails :visible="data.showDetails"-->
<!--    :td="tdToShow"-->
<!--    @on-closed="data.showDetails=false"-->
<!--  />-->

  <QCard class="q-pa-md" flat>
    <!-- Header -->
    <QCardSection horizontal>
      <QIcon :name="matLensBlur" left size="2rem" />
      <div class="text-h6">An overview of known Things</div>
    </QCardSection>

    <!-- List of all things -->
    <QCardSection class="text-h6">
      <!-- <ThingsTable  :things="ds.all" dense
                    title="Hub Things"
                    style="width: 100%"
                    @on-view-details="handleViewDetails"
      >
      </ThingsTable> -->
      <ThingsTable2  :things="ds.all" dense
                    title="Hub Things"
                    style="width: 100%"
                    @on-view-details="handleViewDetails"
      />
    </QCardSection>

  </QCard>
</template>
