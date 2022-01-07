<script lang="ts" setup>
import { RouterView } from 'vue-router';

// const props = defineProps()

</script>


<!--RouterView to enable routes to a dialog
    See the route for things/:thingID
  -->
<template>
<div style="width: 100%; height:100%;">
  <!-- Main view -->
  <RouterView></RouterView>
  <!-- Dialog view -->
  <RouterView name="dialog" ></RouterView>
</div>
</template>

