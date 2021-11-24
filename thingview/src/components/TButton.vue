<script lang="ts" setup>
import {QBtn, QTooltip} from "quasar";
interface IButton {
  label?: string
  icon?: string
  flat?: boolean
  tooltip?: string
  disabled?: boolean
  primary?: boolean
}
 const props = defineProps<IButton>()

const emit = defineEmits(["click"])
</script>


<template>
  <QBtn  :disabled="props.disabled"
         :color="(props.primary? 'primary':undefined)"
         :flat="props.flat"
         :label="props.label"
         :icon="props.icon"
  >
    <slot></slot>
    <QTooltip v-if="props.tooltip">{{props.tooltip}}</QTooltip>
  </QBtn>
</template>
