<script lang="ts">
import {defineComponent, reactive} from "vue";
import { ElDialog, ElInput } from "element-plus";
import { boolean } from "@storybook/addon-knobs";

export default defineComponent({
  components: {ElDialog, ElInput},
  props: {
    visible:Boolean
  },
  emits: ['onClosed'],

  setup(props, {emit}) {

  console.log("AddPageDialog: visible=",props.visible)
    return {emit};
  }
})

</script>


<template>
<ElDialog :model-value="visible" 
  center show-close close-on-press-escape
  @closed='emit("onClosed")'
  >
  <div style="display:flex; flex-direction: column; align-items: center;">
  <p>Add Page</p>
    <ElInput label="name"/>

  </div>
</ElDialog>
</template>
