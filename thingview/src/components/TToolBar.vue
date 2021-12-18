<script lang="ts" setup>


</script>


<template>

<div class="ttoolbar">
  <slot></slot>

</div>
</template>


<style>

.ttoolbar {
  width: 100%;
  display: flex;
  flex-direction: row;
  /* justify-content: space-around; */
  align-items: center;
  gap: 10px;

}

</style>