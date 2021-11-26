<script lang="ts" setup>
import {reactive} from "vue";

// Dashboard view.
const props = defineProps({
    page: String,
})

const data = reactive({
  myPage: props.page ? props.page : "default page",
})

</script>

<template>
  <h4>Dashboard page={{props.page}}, verbose={{data.myPage}} </h4>
</template>
