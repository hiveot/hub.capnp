<script lang="ts" setup>
/**
 *  Closing dialog, parent has to revert the route with router.go(-1)
 */

import { RouterView } from 'vue-router';

/**
 * This component does not have any props
 */
const props = defineProps()

</script>


<!--RouterView to display a dialog on top of a view
    Intended to enable routes to a dialog
  -->
<template>
<div>
  <RouterView></RouterView>
  <RouterView name="dialog" ></RouterView>
</div>
</template>

