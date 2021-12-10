<script lang="ts" setup>
// import {reactive} from "vue";

// import TDialog from '@/components/TDialog.vue'
import TDialog from '@/components/TDialog.vue'

const version = "0.3 alpha";
const licenseURL = "https://github.com/wostzone/wostzone.github.io/blob/81f41a781dfd37d82a5e37bf9ac905199161ee85/LICENSE"
const wostSiteURL = "https://wostzone.github.io/"
import {mdiClose} from "@quasar/extras/mdi-v6";

interface IProps {
  visible: boolean
}
const props = withDefaults(
    defineProps<IProps>(),
    {visible: false}
)

const emit = defineEmits(['onClosed'])

console.log("ShowAbout: visible=",props.visible)

const handleCloseDialog = () => {
  console.log("handleCloseDialog")
  emit("onClosed");
}
</script>


<template>
<TDialog
    :visible="props.visible"
  title="About ThingView"
  @onClosed="handleCloseDialog"
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



