<script lang="ts" setup>
import {QBtn, QDialog, QCard, QCardSection, QCardActions, QBar, QSpace } from "quasar";

import {matClose} from "@quasar/extras/material-icons";

interface IProps{
  // Dialog is visible on/off
  visible: boolean
  // Dialog title
  title?: string
  // show the Cancel button
  showCancel?: boolean
  // show the Ok button
  showOk?: boolean
  // disable the Ok button
  okDisabled?: boolean
  // optional override of Ok button label
  okLabel?: string
}

const props = withDefaults(
    defineProps<IProps>(),
    {
      visible: false,
      showCancel: false,
      showOk: false,
      okDisabled: false,
      okLabel: "Ok",
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

</script>

<!--Dialog component with title and Ok/Cancel buttons with standardized dialog configuration
 -->
<template>
  <QDialog  :model-value="props.visible"
            @hide='handleCancel'
    >
    <QCard>

<!--  dialog title with close button -->
      <QBar>
        <div class="text-h6">{{title}}</div>
        <QSpace/>
        <QBtn :icon="matClose" flat dense v-close-popup/>
      </QBar>

<!--  default Slot for the dialog content-->
      <QCardSection>
        <slot></slot>
      </QCardSection>

<!--  optional override of footer-->
      <QCardActions>
        <slot name="footer"></slot>
      </QCardActions>

<!--  default Cancel/OK footer buttons-->
      <QCardActions v-if="(props.showCancel || props.showOk)" align="right">
          <QBtn v-if="props.showCancel"
            label="Cancel" flat
            @click="handleCancel"
          />
          <QBtn v-if="props.showOk"
            label="Ok"
            :disabled="props.okDisabled"
            color="primary"
            class="q-ml-sm"
            @click="handleSubmit"
          />
      </QCardActions>
    </QCard>
  </QDialog>
</template>
