<script  lang="ts">
import {defineComponent, reactive} from "vue";
import AppView from '@/views/app/AppView.vue';

export default defineComponent({
  components: {AppView},
})

</script>

<template>
  <AppView/>
</template>

<style>
#app {
  font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  /*color: darkblue;*/
  font-size: large;
  margin:0;
}
</style>
