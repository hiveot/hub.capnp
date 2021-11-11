<script lang="ts">

import {defineComponent} from "vue";

import AccountsTable from '@/components/AccountsTable.vue'

import hubAccountStore, {IAccountStore} from '@/store/HubAccountStore'

export default defineComponent({
  components: {AccountsTable},
  emits: [ 
    ],
  props: {
  },

  setup(props, {emit}) {
    return {hubAccountStore}
  }
})

</script>


<template>
<AccountsTable :value="hubAccountStore.GetAccounts()" />
</template>
