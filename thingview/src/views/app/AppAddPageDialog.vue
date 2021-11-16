<script lang="ts" setup>
import {ref} from "vue";

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

</script>

<template>
  <q-dialog :model-value="visible"
            center show-close close-on-press-escape
            @closed='emit("onClosed", $event)'
  >
    <q-card>
      <q-card-section>
        Add A New Page
      </q-card-section>
    <div style="display:flex; flex-direction: column; align-items: center;">
      <q-form @submit.prevent>
        <q-item label="Page Name" required >
          <q-input v-model="pageName" placeholder="new page" label="names"/>
        </q-item>
      </q-form>
    </div>

    <q-card-actions align="right">
      <q-btn @click="emit('onClosed',$event)">Cancel</q-btn>
      <q-btn @click="handleSubmit"  color="primary"  :disabled="(pageName==='')" >Confirm</q-btn>
    </q-card-actions>
    </q-card>
  </q-dialog>
</template>
