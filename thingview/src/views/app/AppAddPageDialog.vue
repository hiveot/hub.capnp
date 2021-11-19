<script lang="ts" setup>
import {ref} from "vue";
import Dialog from 'primevue/dialog';
import Button from 'primevue/button';
import InputText from 'primevue/inputtext';

const props = defineProps({
    visible:Boolean,
  })

const emit = defineEmits({
    'onAdd': String,
    'onClosed': null,
    })

const pageName = ref("");

const handleSubmit = () =>{
  emit("onAdd", pageName.value);
  pageName.value = "";
  emit("onClosed");
 };
console.log("AddPageDialog: visible=",props.visible);

const handleCloseDialog = (ev:any) => {
  console.log("handleCloseDialog, ev: ", ev)
  if (ev === false) {
    emit("onClosed");
  }
}
</script>

<template>
  <Dialog :visible="props.visible"
          header="Add A New Page"
          modal closable dismissableMask closeOnEscape showHeader draggable
          @update:visible="handleCloseDialog"
  >
    <div style="display:flex; flex-direction: column; align-items: center;">
      <div class="p-fluid">
        <div class="p-field p-grid">
          <label for="pageName">Name</label>
          <InputText id="pageName" type="text" v-model="pageName"
                     placeholder="New Page"/>
        </div>
      </div>
    </div>

    <template #footer>
      <Button @click="emit('onClosed',$event)">Cancel</Button>
      <Button @click="handleSubmit"  color="primary"  :disabled="(pageName==='')" >Confirm</Button>
    </template>
  </Dialog>
</template>
