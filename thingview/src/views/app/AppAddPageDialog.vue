<script lang="ts" setup>
import {ref} from "vue";
import {QForm, QInput} from "quasar";

import {mdiInformation} from '@quasar/extras/mdi-v6'
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
  <Dialog
      :visible="props.visible"
      title="New Dashboard Page"
      @onClosed="emit('onClosed')"
      @onSubmit="handleSubmit"
      showCancel showOk
      :okDisabled="(pageName==='')"
  >
    <QForm @submit="handleSubmit" class="q-gutter-md" style="min-width: 350px">
          <QInput v-model="pageName"
                  no-error-icon
                  autofocus filled required lazy-rules
                  id="pageName" type="text"
                  label="Page name"
                  hint="Name of the dashboard as shown on the Tabs"
                  :rules="[()=>pageName !== ''||'Please provide a name']"
                  stack-label
          >
          </QInput>
    </QForm>

<!--    <template v-slot:footer>-->
<!--      <Button @click="emit('onClosed',$event)">Cancel</Button>-->
<!--      <Button @click="handleSubmit"  color="primary"  :disabled="(pageName==='')" >Confirm</Button>-->
<!--    </template>-->
  </Dialog>
</template>
