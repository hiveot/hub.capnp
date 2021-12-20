<script lang="ts" setup>

import {QTable} from "quasar";
import {ref} from "vue";

// Simple table with sticky header style, inline scrollbar, no pagination
// All QTable slots are available to the parent, most importantly:
//  header-cell-[name]  
//  body-cell-[name]
const props = defineProps<{
  // columns to display
  columns: Array<ITableCol>
  // the collection to display
  rows: any
  // unique ID of each row, typically 'id' 
  rowKey: string
  // dense rows
  dense?: boolean
  // striped rows
  striped?: boolean
}>()

// types for columns (fixes a typescript error on align)
export interface ITableCol {
  // Name of column for show/hide
  name:string,
  // Column header label
  label:string,
  // Data field in rows to display
  field:string,
  // Optional field conversion 
  format?: (val:any, row:any)=>string,
  // Align the field content left, right or center
  align?:"left"|"right"|"center",
  // Column can be sorted
  sortable?:boolean
  // Optional style for display of the column data
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
      :dense="props.dense"
      :class="'simpleTableStyle ' + ((props.striped)? 'tableStriped' : '')"
      :pagination="pagination"
      :rows="props.rows"
      :rows-per-page-options="[0]"
      table-header-class="ttable-header text-primary"
      table-header-style="background-color: lightgrey"
      virtual-scroll
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

/* Table header style */
.tableStriped tr:nth-child(even) {
  background-color: rgb(240, 242, 245);
}


.q-table__bottom {
  /* background-color: lightgrey; */
  height: 30px !important;
  min-height: 20px !important;
}
</style>