<script lang="ts" setup>

import {reactive, ref} from "vue";
import {DashboardTile} from "@/data/dashboard/DashboardStore";
import {GridItem} from 'vue3-grid-layout'
import {QCard} from "quasar";

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
   <QCard  style="display:flex; flex-direction:column; width:100%; height:100%" title="Hello">
     <div v-if="props.tile">
      Item: {{props.tile.title}}
     </div>
     <div v-else>
       Tile with id {{props.id}} not provided
     </div>
   </QCard>

</template>