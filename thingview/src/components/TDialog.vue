<script lang="ts" setup>
import {QBtn, QDialog, QCard, QCardActions, QCardSection, QSeparator, QScrollArea, QSpace, QToolbar, QToolbarTitle, } from "quasar";
// import TButton from '@/components/TButton.vue'

import {matClose} from "@quasar/extras/material-icons";

interface IProps{
  // disable the Ok button
  okDisabled?: boolean
  // optional override of Cancel button label
  cancelLabel?: string
  // optional override of Close button label
  closeLabel?: string
  // Maximum width of dialog, default 100vw
  maxWidth?: string
  // optional override of Ok button label
  okLabel?: string
  // show the Cancel button
  showCancel?: boolean
  // show the Close button
  showClose?: boolean
  // show the Ok button
  showOk?: boolean
  // Dialog title
  title?: string
  // Dialog is visible on/off
  visible: boolean
  // Width of dialog, eg 80%
  width?: string
}

const props = withDefaults(
    defineProps<IProps>(),
    {
      cancelLabel: "Cancel",
      closeLabel: "Close",
      okDisabled: false,
      okLabel: "Ok",
      showCancel: false,
      showClose: false,
      showOk: false,
      visible: false,
      maxWidth: "100vw"
    }
)

const emit = defineEmits(['onClosed', 'onSubmit'])

const handleSubmit = (ev:any) => {
  console.log("handleSubmit emit onSubmit")
  emit('onSubmit')
}

const handleCancel = () => {
  console.log("cancel dialog")
  emit('onClosed')
}

const maxWidth = "50vw"

</script>

<!--Dialog component with title and Ok/Cancel buttons with standardized dialog configuration
 -->
<template>
  <QDialog  :model-value="props.visible"
            @hide='handleCancel'

    >
<!--    maxWidth must be set for width to work-->
    <QCard class="column" :style="{height:'100%', width: props.width, maxWidth: props.maxWidth}" >

<!--  dialog title with close button -->
      <QToolbar  class="text-primary bg-grey-3" style="height: 40px">
        <QToolbarTitle>{{props.title}}</QToolbarTitle>
        <QBtn :icon="matClose" flat dense v-close-popup/>
      </QToolbar>

<!--  default Slot for the dialog content-->
      <QCardSection style="width:100%;display: flex;overflow: auto" class="col" >
        <slot></slot>
      </QCardSection>

<!--  optional override of footer-->
      <QCardActions>
        <slot name="footer"></slot>
      </QCardActions>

<!--  default Cancel/OK footer buttons-->
      <q-separator />
      <QCardSection v-if="(props.showCancel || props.showOk || props.showClose)" 
        class="row q-pa-xs bg-grey-3 text-primary"
        >
          <QBtn v-if="props.showCancel"
            label="Cancel"
            :label="props.cancelLabel"
            @click="handleCancel"
          />
          <QSpace/>
          <QBtn v-if="props.showClose" flat
            :label="props.closeLabel"
            label="Close"
            @click="handleCancel"
          />
          <QBtn v-if="props.showOk" 
            :label="props.okLabel"
            :disabled="props.okDisabled"
            color="primary"
            @click="handleSubmit"
          />
      </QCardSection>
    </QCard>
  </QDialog>
</template>

<style>
.tdialogcontent {
  height: 100%;
  overflow-y: auto;
}
</style>