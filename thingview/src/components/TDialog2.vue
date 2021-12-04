<!--<script lang="ts" setup>-->
<!--import {QBtn, QDialog, QCard, QCardSection, QCardActions, QBar, QSpace } from "quasar";-->
<!--import { useDialogPluginComponent } from 'quasar'-->
<!--const { dialogRef, onDialogHide, onDialogOK, onDialogCancel } = useDialogPluginComponent()-->

<!--import {matClose} from "@quasar/extras/material-icons";-->

<!--interface IProps{-->
<!--  // Dialog title-->
<!--  title?: string-->
<!--  // show the Cancel button-->
<!--  showCancel?: boolean-->
<!--  // show the Ok button-->
<!--  showOk?: boolean-->
<!--  // disable the Ok button-->
<!--  okDisabled?: boolean-->
<!--  // optional override of Ok button label-->
<!--  okLabel?: string-->
<!--}-->

<!--const props = withDefaults(-->
<!--    defineProps<IProps>(),-->
<!--    {-->
<!--      showCancel: false,-->
<!--      showOk: false,-->
<!--      okDisabled: false,-->
<!--      okLabel: "Ok",-->
<!--    }-->
<!--)-->

<!--// https://quasar.dev/quasar-plugins/dialog#predefined-->
<!--const emit = defineEmits([-->
<!--  ...useDialogPluginComponent.emits,-->
<!--])-->

<!--</script>-->

<!--&lt;!&ndash;Dialog component with title and Ok/Cancel buttons with standardized dialog configuration-->
<!-- &ndash;&gt;-->
<!--<template>-->
<!--  <QDialog ref="dialogRef"-->
<!--      @hide='onDialogHide'-->
<!--    >-->
<!--    <QCard>-->

<!--&lt;!&ndash;  dialog title with close button &ndash;&gt;-->
<!--      <QBar>-->
<!--&lt;!&ndash;        <div class="text-h6">{{title}}</div>&ndash;&gt;-->
<!--        <QSpace/>-->
<!--&lt;!&ndash;        <QBtn :icon="matClose" flat dense v-close-popup/>&ndash;&gt;-->
<!--      </QBar>-->

<!--&lt;!&ndash;  default Slot for the dialog content&ndash;&gt;-->
<!--      <QCardSection>-->
<!--        <slot></slot>-->
<!--      </QCardSection>-->

<!--&lt;!&ndash;  optional override of footer&ndash;&gt;-->
<!--      <QCardActions>-->
<!--        <slot name="footer"></slot>-->
<!--      </QCardActions>-->

<!--&lt;!&ndash;  default Cancel/OK footer buttons&ndash;&gt;-->
<!--      <QCardActions v-if="(props.showCancel || props.showOk)" align="right">-->
<!--          <QBtn v-if="props.showCancel"-->
<!--            label="Cancel" flat-->
<!--            @click="onDialogCancel"-->
<!--          />-->
<!--          <QBtn v-if="props.showOk"-->
<!--            label="Ok"-->
<!--            :disabled="props.okDisabled"-->
<!--            color="primary"-->
<!--            class="q-ml-sm"-->
<!--            @click="onDialogOK"-->
<!--          />-->
<!--      </QCardActions>-->
<!--    </QCard>-->
<!--  </QDialog>-->
<!--</template>-->
