<script lang="ts">
import {defineComponent, reactive} from "vue";

// Dashboard view.
export default defineComponent({
  props: {
    page: String,
  },
  components: {  },
  setup(props) {
    const data = reactive({
      myPage: props.page ? props.page : "default page",
    })
    return { data }
  }
})

</script>

<template>
  <h1>Dashboard page={{page}}, verbose={{data.myPage}} </h1>
</template>
