<script lang="ts" setup>

import {QBtn, QTooltip} from 'quasar'

/**
 * TButton wrapper 
 * This standardizes attributes and includes a tooltip
 */
const props = defineProps<{
  color?: string
  icon?: any
  dense?: boolean
  label?: string
  tooltip?: string
  round?: boolean
  flat?: boolean
}>()

const emits = defineEmits(["click"])


</script>


<template>
<QBtn :color="props.color"
      :label="props.label"
      :icon="props.icon" 
      :round="round" 
      :flat="flat" 
      :dense="dense"
      @click="emits('click')"
      >
  <QTooltip   v-if="props.tooltip"
    class="tooltip" 
    :delay="1500"
  >{{props.tooltip}}</QTooltip>
</QBtn>

</template>

<style>

/* the default tooltip is tiny */
.tooltip {
  font-size: inherit !important;
  font-style: italic;
}

</style>