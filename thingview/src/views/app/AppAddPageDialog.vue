<script lang="ts">
import {defineComponent, ref,reactive} from "vue";
import { ElDialog, ElInput, ElForm, ElFormItem } from "element-plus";

export default defineComponent({
  components: {ElDialog, ElInput, ElForm, ElFormItem},
  props: {
    visible:Boolean,
  },
  emits: {
    'onAdd': String,
    'onClosed': null,
    },
  setup(props, {emit}) {
    const pageName = ref("");

    const handleSubmit = () =>{
      emit("onAdd", pageName.value);
      pageName.value = "";
      emit("onClosed");
     };
    console.log("AddPageDialog: visible=",props.visible);
    return {emit, pageName, handleSubmit};
  }
})
</script>

<template>
  <ElDialog :model-value="visible"
            center show-close close-on-press-escape
            @closed='emit("onClosed")'
            title="Add A New Page"
  >
    <div style="display:flex; flex-direction: column; align-items: center;">
      <ElForm @submit.prevent>
        <ElFormItem label="Page Name" required >
          <ElInput v-model="pageName" placeholder="new page" label="names"/>
        </ElFormItem>
      </ElForm>
    </div>
    <template #footer>
      <div className="dialog-footer" style="text-align: right;">
        <el-button @click="emit('onClosed')">Cancel</el-button>
        <el-button @click="handleSubmit" type="primary" :disabled="(pageName=='')" >Confirm</el-button>
      </div>
    </template>
  </ElDialog>
</template>
