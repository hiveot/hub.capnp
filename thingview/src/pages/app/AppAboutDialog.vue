<script lang="ts" setup>

// import TDialog from '@/components/TDialog.vue'
import TDialog from '@/components/TDialog.vue'
import {useDialogPluginComponent } from 'quasar'
const version = "0.3 alpha";
const licenseURL = "https://github.com/wostzone/wostzone.github.io/blob/81f41a781dfd37d82a5e37bf9ac905199161ee85/LICENSE"
const wostSiteURL = "https://wostzone.github.io/"

interface IProps {
  visible?: boolean
}
const props = withDefaults(
    defineProps<IProps>(),
    {visible: true}
)
const { dialogRef, onDialogCancel } = useDialogPluginComponent();

const emit = defineEmits(['onClosed',
    // REQUIRED; need to specify some events that your
    // component will emit through useDialogPluginComponent()
    ...useDialogPluginComponent.emits,   
]);


</script>


<template>
<TDialog ref="dialogRef"
  :visible="props.visible"
  title="About ThingView"
  show-close
  @onClosed="onDialogCancel"
  >

    <div style="display: flex; width:100%">
      <div style="flex: .4; ; align-self:center">
          <img style="width:240px; padding:20px"
               src="@/assets/logo.svg"
               alt="logo"
          />
      </div>
      <div style="flex:1; width:100%; text-align: start; padding: 10px" >
        <h4>ThingView</h4>
        <p>WoST Hub Viewer presents information and status of Things that are connected to the Hub.</p>
        <p>Version {{version}}</p>
        <p>By the <a :href="wostSiteURL" target="_blank">WoST Zone</a></p>
      </div>  
    </div>

  <template #footer>
    <div style="width:100%; display: flex; flex-direction: row; justify-content: space-between; padding-top: 20px">
      <a :href="licenseURL" target="_blank">License Information</a>
      <a :href="wostSiteURL" target="_blank">About WoST</a>
    </div>
  </template>
</TDialog>
</template>



