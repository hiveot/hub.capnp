<script lang="ts" setup>

import {DashboardTile} from "@/data/dashboard/DashboardStore";
import {QCard, QCardSection, QBtn, QToolbar, QToolbarTitle} from "quasar";
import {matMenu} from "@quasar/extras/material-icons";

/**
 * Dashboard Grid Tile
 * Wrapper around grid-item to handle positions
 */
const props = defineProps<{
  id: string
  tile: DashboardTile
}>()

</script>

<template>
   <QCard  style="display:flex; flex-direction:column; width:100%; height:100%">

     <span v-if="props.tile">
       <!--  title with menu -->
       <QToolbar  class="tile-header-bar not-draggable-area" dense>
         <QToolbarTitle  class="tile-header text-weight-bold ">{{props.tile.title}}</QToolbarTitle>
         <QBtn   :icon="matMenu" flat dense class="tile-header"/>
       </QToolbar>

       <!--       Slot for the widget content-->
       <!--       To keep vertical scrolling within the slot, use flex column with overflow-->
       <QCardSection
           style="height: 100%; display:flex; flex-direction:column; overflow: auto"
       >
         <slot>
           Item: {{props.tile.title}}
         </slot>
       </QCardSection>

     </span>
     <span v-else>
       Tile with id {{props.id}} not provided
     </span>

   </QCard>

</template>

<style scoped>

.tile-header-bar {
  background-color: lightgray;
  min-height: 20px;
  padding: 0;
}
.tile-header {
  font-size: 1.2em;
  min-height: 20px;
  padding: 4px;
  /*height: 1rem;*/
}
</style>