<script lang="ts" setup>

import {h, VNode } from 'vue';
import {get as _get} from 'lodash-es'
import {QMarkupTable, QTooltip} from "quasar";

// Table Column Definition
export interface ISimpleTableColumn {
  /**
   * Column alignment. Default is left
   */
  align?: "left" | "center" | "right"
  /**
   * Custom component or render function to show
   */
  component?: any
  /**
   * field name in row data to display
   */
  field: string
  /**
   * Column is sortable (TODO)
   */
  sortable?: boolean
  /**
   * column header title
   */
  title?: string
  /**
   * show/hide column
   */
  hidden?: boolean,
  /**
   * Field width
   */
  width?: string
}

/**
 * Simple lightweight table for displaying data
 * This is intended for small simple tables and does not do any pagination, sorting, or filtering
 */
const props = defineProps<{
  /**
   * Column definition
   */
  columns: ISimpleTableColumn[]
  /**
   * Compact the cells
   */
  dense?: boolean
  /**
   * Empty text
   */
  emptyText?: string
  /**
   * Flat design, eg no border box shadow but with border
   */
  flat?: boolean
  /**
   * Grow row height to fill its container 
   */
  grow?: boolean
  /**
   * Hide the table outside border
   */
  noBorder?: boolean
  /**
   * Hide the table header
   */
  noHeader?: boolean
  /**
   * Hide the table cell lines
   */
  noLines?: boolean
  /**
   * Hide the row strips
   */
  noStripes?: boolean
  /** 
   * Rows with objects to show. rows must have an id field
   */
  rows: any[]
  
}>()

const emit = defineEmits(['onRowSelect'])

// filter hidden columns
const getVisibleColumns = (columns:ISimpleTableColumn[]):ISimpleTableColumn[] => {
  return columns.filter( (col)=>!col.hidden)
}

</script>

<template>
  <QMarkupTable 
    :dense="props.dense" 
    :flat="props.flat" 
    :class="
      (props.grow ? 't-simple-table-grow':'') +
      (props.noLines ? ' no-table-lines':' table-lines')+
      (props.noBorder ? ' no-border' : ' with-border')
      "
  >
    <thead v-if="!props.noHeader">
      <tr key="header" class="header-row" >
        <th v-for="column in getVisibleColumns(props.columns)" 
          :key="column.field" 
          :style="{
              textAlign: column.align? column.align:'left',
              width:column.width
            }"
          >
          {{column.title}}
        </th>
      </tr>
    </thead>

    <tbody style="height:100%; width: 100%;">
      <tr v-for="row in props.rows" 
        :key="row.key"
        :class="props.noStripes ? '' : 'with-stripes'"
        @click="()=>{emit('onRowSelect', row)}"
      >
        <td v-for="column in getVisibleColumns(props.columns)"
            :key="column.field"
            :style="{
              textAlign: column.align? column.align:'left',
              width:column.width
              }"
          >
          <span v-if="column.component">
            <component :is="column.component" v-bind="row" />
          </span>
          <span v-else>
            {{_get(row, column.field, "field '"+column.field+"' not found")}}
            <!-- <QTooltip>hello</QTooltip> -->
          </span>
        </td>
      </tr>
      <tr v-if="props.rows.length===0">
        <td :colspan='3'>{{props.emptyText||"No data"}}</td>
      </tr>
    </tbody>
    <!-- </table> -->
  </QMarkupTable>
</template>

<style >

/* cells fill the table and do not cause a scrollbar; overflow td with ellipsis
 */
.q-table {
  table-layout: fixed;
}

/* make QTable dense a bit denser */
.q-table--dense .q-table td {
  padding: 4px 4px !important;
  /* text-overflow: ellipsis; */
}
.q-table--dense .q-table th:first-child,
.q-table--dense .q-table td:first-child {
  padding: 4px 4px !important;
}
/* use the same default font size as elsewhere */
.q-table tbody td {
  font-size: inherit !important;
  overflow: hidden;
  text-overflow: ellipsis;
  /* padding: 0; */
}
/* use the same default font color as elsewhere */
.q-table__card {
  color: inherit !important;
}
/* use the heavier font */ 
.q-table th {
  font-weight: bold !important;
  font-size: 0.8rem !important;
}


/* grow table in available space as used in dashboard widgets */
.t-simple-table-grow  {
  height: 100%;
}
.t-simple-table-grow table {
  height: 100%;
}

.header-row   {
  background-color: rgb(224, 229, 230);
  text-transform: uppercase;
  font-weight: 800;
}

/** stripes rows */
.with-stripes:nth-child(even) td {
  background: #f6f7ff;
}

.table-row   {
  margin: 0;
  padding: 0;
}

.with-border {
  border: 1px solid lightgray
}
.no-border {
  border: none
}

/** header divider lines are white in grey background except first column */
/* .table-lines table thead tr th:nth-child { */
.table-lines table thead tr th:not(:first-child) {
  border-left: 1px solid #ffffff !important;
 }
.table-lines table tbody tr td {
  border-top: none; /*1px solid #eeeded !important;*/
  border-left: 1px solid #eeeded !important;
  border-right: 1px solid #eeeded !important;
  border-bottom: 1px solid #eeeded !important; 
  /* border-right: 1px solid #eeeded;*/
 }
.no-table-lines  td {
  border-style: none !important;
  /* box-shadow: none; */
  /* border-left: 1px solid #eeeded; */
  /* border-right: 1px solid #eeeded; */
 }


</style>