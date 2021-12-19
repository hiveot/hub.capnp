<script lang="ts" setup>

import {QTable} from "quasar";
import {ref} from "vue";

// Simple table with sticky header style, inline scrollbar, no pagination
// All QTable slots are available to the parent, most importantly:
//  header-cell-[name]  
//  body-cell-[name]
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
  format?: (val:any, row:any)=>string,
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
      table-header-class="ttable-header text-primary"
      table-header-style="background-color: lightgrey"
      dense
      striped
      :rows-per-page-options="[0]"
      virtual-scroll
      :pagination="pagination"
  >
   <!-- export the slots -->
    <template v-for="(index, name:string|number) in $slots" v-slot:[name]="props">
      <!-- pass all props to the named slot -->
      <slot :name="name" v-bind="props" > </slot>
    </template>
   
   <!-- setheader -->
   <template #header-cell="propz">
      <q-th :props="propz" >
        <span>{{propz.col.label}}</span>
      </q-th>
    </template> 
  </QTable>

</template>

<style>
.simpleTableStyle {
  height: 100%;
}

/* Table header style */
/* .ttable-header { */
  /* position: sticky; */
  /* z-index: 1; */
  /* top: 0; */
  /* color: blue; */
  
  /* font-size: 1.1rem !important; */
  /* background-color: lightgray; */
/* } */


/* Table header style */
.simpleTableStyle thead th {
  /* Sticky header for keeping the header in place when scrolling. This used to be needed
   * but for some reason it now works without. Leaving this here as a reminder in case  
   * it stops working again.
   */
  /* position: sticky;
  z-index: 1;
  top: 0; */

  /* background-color: lightgray; */

  /* font-size only seem to have effect here, not in table-header-class, nor in table-header-style */
  font-size:1.1rem;
} 

.q-table__bottom {
  /* background-color: lightgrey; */
  height: 30px !important;
  min-height: 20px !important;
}
</style>