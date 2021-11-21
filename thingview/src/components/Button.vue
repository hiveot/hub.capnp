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
</script>


<template>
  <div>
  <QBtn  :disabled="props.disabled" :color="(props.primary? 'primary':'default')"
          :label="props.label" :icon="props.icon" :flat="props.flat">
    <QTooltip v-if="props.tooltip !=='' ">{{props.tooltip}}</QTooltip>
  </QBtn>
  </div>
</template>
