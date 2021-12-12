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


</script>

<template>

<TDialog :visible="props.visible" 
  :title="props.td.description"
      @onClosed="emit('onClosed')"
      showCancel
  >
  <ThingDetailsView :td="props.td" />

</TDialog>

</template>
