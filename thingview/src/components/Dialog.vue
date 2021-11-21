<script lang="ts" setup>
import {QDialog, QCard, QCardSection, QCardActions, QBar } from "quasar";

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

</script>

<!--QDialog is too basic so make our own component based on it-->
<template>
  <QDialog  :model-value="props.visible"
    @hide='emit("onClosed")'
    >
    <QCard>
      <QBar>
        <div class="text-h6">{{title}}</div>
        <QSpace/>
        <QBtn icon="mdi-close" flat dense v-close-popup/>
      </QBar>
      <QCardSection>
        <slot></slot>
      </QCardSection>
      <QCardActions>
        <slot name="footer"></slot>
      </QCardActions>
      <QCardActions v-if="(props.showCancel || props.showOk)" align="right">
          <Button label="Cancel" flat
                  v-if="props.showCancel" @click="emit('onClosed')"/>
          <Button label="Ok" :disabled="props.okDisabled"
                  primary class="q-ml-sm"
                  v-if="props.showOk" @click="emit('onSubmit')"/>
      </QCardActions>
    </QCard>
  </QDialog>
</template>
