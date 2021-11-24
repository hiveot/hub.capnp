<script lang="ts" setup>
import {QBtn, QDialog, QCard, QCardSection, QCardActions, QBar } from "quasar";
import TButton from '@/components/TButton.vue'

import {mdiClose} from "@quasar/extras/mdi-v6";

interface IProps{
  visible: boolean
  title?: string
  showCancel?: boolean
  showOk?: boolean
  okDisabled?: boolean
}

const props = withDefaults(
    defineProps<IProps>(),
    {
      visible: false,
      showCancel: false,
      showOk: false,
      okDisabled: false,
    }
)

const emit = defineEmits(['onClosed', 'onSubmit'])

const handleSubmit = (ev:any) => {
  console.log("handleSubmit emit onSubmit")
  debugger
  emit('onSubmit')
}

const handleCancel = () => {
  console.log("cancel dialog")
  emit('onClosed')
}

</script>

<!--QDialog is too basic so make our own component based on it-->
<template>
  <QDialog  :model-value="props.visible"
            @hide='handleCancel'
    >
    <QCard>
      <QBar>
        <div class="text-h6">{{title}}</div>
        <QSpace/>
        <QBtn :icon="mdiClose" flat dense v-close-popup/>
      </QBar>
      <QCardSection>
        <slot></slot>
      </QCardSection>

      <QCardActions>
        <slot name="footer"></slot>
      </QCardActions>

      <QCardActions v-if="(props.showCancel || props.showOk)" align="right">
          <TButton v-if="props.showCancel"
                  label="Cancel" flat
                  @click="handleCancel"/>
          <TButton v-if="props.showOk"
                  label="Ok"
                  :disabled="props.okDisabled"
                  primary class="q-ml-sm"
                  @click="handleSubmit"/>
      </QCardActions>
    </QCard>
  </QDialog>
</template>
