<script lang="ts" setup>

import {QTable} from "quasar";
// simple table with sticky header style, inline scrollbar, no pagination
import {ref} from "vue";

const props = defineProps<{
  columns: Array<ITableCol>
  rows: any
  rowKey: string
}>()

// types for columns (fixes a typescript error on align)
export interface ITableCol {
  name:string,
  label:string,
  field:string,
  align?:"left"|"right"|"center",
  sortable?:boolean
  style?:string
}

const pagination = ref({
  rowsPerPage: 0
})
</script>


<template>

  <QTable
      :row-key="props.rowKey"
      :columns="props.columns"
      :rows="props.rows"
      class="simpleTableStyle"
      table-header-style="background:lightgray"
      dense
      striped
      :rows-per-page-options="[0]"
      virtual-scroll
      :pagination="pagination"
  >

  </QTable>

</template>

<style>
.simpleTableStyle {
  height: 100%;
}
.simpleTableStyle thead th {
  position: sticky;
  z-index: 1;
  top: 0;
}
.q-table__bottom {
  background-color: lightgray;
  height: 30px !important;
  min-height: 20px !important;
}
</style>