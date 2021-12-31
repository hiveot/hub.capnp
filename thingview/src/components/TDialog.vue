<script lang="ts" setup>

// import {ref} from 'vue'
import {useDialogPluginComponent, QBtn, QDialog, QCard, QCardActions, QCardSection, QSeparator, QScrollArea, QSpace, QToolbar, QToolbarTitle, } from "quasar";


// import TButton from '@/components/TButton.vue'
import {matClose} from "@quasar/extras/material-icons";

// This is REQUIRED; Need to inject these (from useDialogPluginComponent() call)
// into the vue scope for the vue html template
const { dialogRef, onDialogHide, onDialogOK, onDialogCancel } = useDialogPluginComponent()

/**
 *  Standardized Dialog that handles header, footer, inner scrollbars 
 */
interface IDialog {
  /** disable the Ok button */
  okDisabled?: boolean
  /** optional override of Cancel button label */
  cancelLabel?: string
  /** optional override of Close button label */
  closeLabel?: string
  /** dialog height, %, px, vh, ... */
  height?: string,
  /** Minimum dialog height */
  minHeight?: string,
  /** Min width of dialog, default 400px */
  minWidth?: string
  /** Maximum width of dialog, default 100vw */
  maxWidth?: string
  /** optional override of Ok button label */
  okLabel?: string
  /** show the Cancel button */
  showCancel?: boolean
  /** show the Close button */
  showClose?: boolean
  /** show the Ok button */
  showOk?: boolean
  /** Dialog title */
  title?: string
  /** Dialog is visible on/off */
  visible?: boolean
  /** Width of dialog, eg 80% */
  width?: string
}

const props = withDefaults(
    defineProps<IDialog>(),
    {
      cancelLabel: "Cancel",
      closeLabel: "Close",
      minHeight: "20%",
      okDisabled: false,
      okLabel: "Ok",
      maxWidth: "100vw",
      minWidth: "300px",
      showCancel: false,
      showClose: false,
      showOk: false,
      visible: true,
      width: "60%"
    }
)

const emits = defineEmits( [
    'onClosed', 'onSubmit', 'hide',
    // REQUIRED; need to specify some events that your
    // component will emit through useDialogPluginComponent()
    ...useDialogPluginComponent.emits,
]);

/**
 * Send the 'submit' event. The parent component must validate and include the result
 * in a call to onDialogOK(result) from 'useDialogPluginComponent'
 */
const handleSubmit = () => {
  console.debug("TDialog.handleSubmit emit onSubmit")
  emits('onSubmit')
  // to be done by parent: onDialogOK()
}

// Notify listeners this dialog is closed
const handleClose = () => {
  console.debug("TDialog. Closing Dialog")
  emits('onClosed')
  onDialogHide()
}

// show and hide are for use by $q.dialog. Pass it on the the QDialog child
const hide = () => {
  dialogRef.value?.hide()
}

// show and hide are for use by $q.dialog. Pass it on the the QDialog child
const show = () => {
  dialogRef.value?.show()
}

// Export show and hide
defineExpose({show, hide})

</script>

<!--Dialog component with title and Ok/Cancel buttons with standardized dialog configuration
 -->
<template>
  <QDialog ref="dialogRef"
           :model-value="props.visible"
           @hide='handleClose'

    >
<!--    maxWidth must be set for width to work-->
    <QCard :style="{
      display: 'flex', flexDirection: 'column',
      height: props.height,
      minHeight: props.minHeight,
      minWidth: props.minWidth,
      maxWidth: props.maxWidth,
      width: props.width,
    }">

      <!--  dialog title with close button -->
      <QToolbar  class="text-primary bg-grey-3">
        <QToolbarTitle>{{props.title}}</QToolbarTitle>
        <QBtn :icon="matClose" flat dense v-close-popup/>
      </QToolbar>

<!--       Slot for the dialog content-->
<!--       To keep vertical scrolling within the slot, use flex column with overflow-->
      <QCardSection
          style="height: 100%; display:flex; flex-direction:column; overflow: auto"
      >
        <slot></slot>
      </QCardSection>

<!--&lt;!&ndash;  optional override of footer&ndash;&gt;-->
<!--      <QCardActions >-->
<!--        <slot name="footer"></slot>-->
<!--      </QCardActions>-->

<!--  default Cancel/OK footer buttons-->
      <q-separator />
      <QCardSection style="height:50px" v-if="(props.showCancel || props.showOk || props.showClose)"
        class="row q-pa-xs bg-grey-3 text-primary"
        >
          <QBtn v-if="props.showCancel"
            label="Cancel"
            :label="props.cancelLabel"
            @click="onDialogCancel"
          />
          <QSpace/>
          <QBtn v-if="props.showClose" flat
            :label="props.closeLabel"
            label="Close"
            @click="handleClose"
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