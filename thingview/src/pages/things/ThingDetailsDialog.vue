<script lang="ts" setup>

import TDialog from '@/components/TDialog.vue'
import ThingDetailsView from './ThingDetailsView.vue'
import {ThingTD} from "@/data/td/ThingTD";
import {useRouter} from "vue-router";

const props = defineProps<{
  /** Thing TD to view */
  td?: ThingTD,
  /** Route to go to on close */
  to: string|object,
}>()
const emit = defineEmits(["onClosed"])

const router = useRouter()
const handleClosed = (ev:any) => {
  if (props.to) {
    router.push(props.to)
  }
}

const getHeight = (td: ThingTD|undefined):string => {
  if (!td) {
    return "100px"
  }
  // the height should ideally accommodate the tallest view. However we don't know what that is.
  // so, just estimate based on nr of attributes and configuration.
  let attrCount = ThingTD.GetThingAttributes(td).length
  let configCount = ThingTD.GetThingConfiguration(td).length
  // row height is approx 29px + estimated header and footer size around 300px
  let height = Math.max(attrCount, configCount)*29+340;
  return height.toString() + 'px'
}

</script>

<template>

<TDialog :visible="true"
        :title="(!!props.td) ? props.td.description: ''"
        @onClosed="handleClosed"
        showClose
        :height="getHeight(props.td)"
         minHeight="40%"
         minWidth="600px"
        >

  <ThingDetailsView v-if="props.td" :td="props.td"/>

</TDialog>

</template>
