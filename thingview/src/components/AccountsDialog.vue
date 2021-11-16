<script lang="ts" setup>

import AccountsTable from '@/components/AccountsTable.vue'
import hubAccountStore from '@/store/HubAccountStore'


</script>


<template>
  <AccountsTable :accounts="hubAccountStore.GetAccounts()" />
</template>
