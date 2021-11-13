
<template>
<ElDialog :model-value="visible" 
  center show-close close-on-press-escape
  @closed='emit("onClosed")'
  title="About ThingView"
  >
<div style="display:flex; flex-direction: column; align-items: center;">
  <div style="display:flex; flex-direction: column; max-width: 400px; ">
    <div style="display: flex; width:100%">
      <div style="flex: .4; ; align-self:center">
          <img style="width:100px; padding:20px" src="@/assets/logo.png" />
      </div>
      <div style="flex:1; width:100%; text-align: start;" >
        <h1>ThingView</h1>
        <p>WoST Hub Viewer</p>
        <p>Version {{version}}</p>
        <p>By the <a :href="wostSiteURL" target="_blank">WoST Zone</a></p>
      </div>  
    </div>

    <div style="display: flex; flex-direction: row; justify-content: space-between; padding-top: 20px">
      <a :href="licenseURL" target="_blank">License Information</a>
      <a :href="wostSiteURL" target="_blank">About WoST</a>
    </div>
  </div>
</div>
</ElDialog>
</template>


<script lang="ts">
import {defineComponent, reactive} from "vue";
import { ElDialog } from "element-plus";
import { boolean } from "@storybook/addon-knobs";

export default defineComponent({
  components: {ElDialog},
  props: {
    visible:Boolean
  },
  emits: ['onClosed'],

  setup(props, {emit}) {
    const version = "0.1 alpha";
    const licenseURL = "https://github.com/wostzone/wostzone.github.io/blob/81f41a781dfd37d82a5e37bf9ac905199161ee85/LICENSE"
    const wostSiteURL = "https://wostzone.github.io/"

console.log("ShowAbout: visible=",props.visible)
    return {emit, version, licenseURL,  wostSiteURL};
  }
})

</script>

