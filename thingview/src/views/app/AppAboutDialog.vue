<script lang="ts" setup>
// import {reactive} from "vue";
import Dialog from 'primevue/dialog';

const version = "0.2 alpha";
const licenseURL = "https://github.com/wostzone/wostzone.github.io/blob/81f41a781dfd37d82a5e37bf9ac905199161ee85/LICENSE"
const wostSiteURL = "https://wostzone.github.io/"

interface IProps {
  visible: boolean
}
const props = withDefaults(
    defineProps<IProps>(),
    {visible: false}
)

const emit = defineEmits(['onClosed'])

console.log("ShowAbout: visible=",props.visible)
const handleCloseDialog = (ev:any) => {
  console.log("handleCloseDialog, ev: ", ev)
  if (ev === false) {
    emit("onClosed");
  }
}
</script>


<template>
<Dialog :visible="props.visible"
  header="About ThingView"
  modal closable dismissableMask closeOnEscape showHeader draggable
  @update:visible="handleCloseDialog"
  >
<div style="display:flex; flex-direction: column; align-items: center;">
  <div style="display:flex; flex-direction: column; max-width: 400px; ">
    <div style="display: flex; width:100%">
      <div style="flex: .4; ; align-self:center">
          <img style="width:100px; padding:20px"
               src="@/assets/logo.png"
               alt="logo"
          />
      </div>
      <div style="flex:1; width:100%; text-align: start;" >
        <h1>ThingView</h1>
        <p>WoST Hub Viewer</p>
        <p>Version {{version}}</p>
        <p>By the <a :href="wostSiteURL" target="_blank">WoST Zone</a></p>
      </div>  
    </div>

  </div>
</div>
  <template #footer>
    <div style="display: flex; flex-direction: row; justify-content: space-between; padding-top: 20px">
      <a :href="licenseURL" target="_blank">License Information</a>
      <a :href="wostSiteURL" target="_blank">About WoST</a>
    </div>
  </template>
</Dialog>
</template>



