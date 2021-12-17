<script lang="ts" setup>

import {defineProps} from "vue";
import TDialog from '@/components/TDialog.vue'
import ThingDetailsView from './ThingDetailsView.vue'
import {ThingTD} from "@/data/td/ThingTD";

const props = defineProps<{
  td: ThingTD,
  visible: boolean,
}>()

const emit = defineEmits(["onClosed"])

const getHeight = (td: ThingTD):string => {
  // the height should ideally accomodate the tallest view. However we don't know what that is.
  // so, just estimate based on nr of properties.
  let count = Object.keys(td.properties).length
  let height = count/24 * 100;
  return height.toString() + '%'
}

</script>

<template>

<TDialog :visible="props.visible" 
        :title="props.td.description"
        @onClosed="emit('onClosed')"
        showClose
        :height="getHeight(props.td)"
         minHeight="40%"
         minWidth="600px"
        >

  <ThingDetailsView :td="props.td"/>

</TDialog>

</template>
