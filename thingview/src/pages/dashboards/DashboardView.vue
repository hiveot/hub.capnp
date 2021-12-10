<script lang="ts" setup>
import {reactive} from "vue";
import ds from '@/data/dashboard/DashboardStore'

// Dashboard view shows the dashboard with the given name 
const props = defineProps({
    name: String
})

const data = reactive({
  myPage: props.name ? props.name : "default dashboard",
})

const dsp = ds.getDashboard(props.name)

</script>

<template>
  <h4>Dashboard name={{props.name}}, verbose={{data.myPage}} </h4>
</template>
