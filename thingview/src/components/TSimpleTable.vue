<script lang="ts" setup>

import {h, VNode } from 'vue';
import {get as _get} from 'lodash-es'
import {QMarkupTable} from "quasar";

// Table Column Definition
export interface ISimpleTableColumn {
  /**
   * column header title
   */
  title: string
  /**
   * field name in row data to display
   */
  field: string
  /**
   * Field width
   */
  width?: string
  /**
   * Column alignment. Default is left
   */
  align?: "left" | "center" | "right"
  /**
   * Component to show
   */
  component?: any
}

/**
 * Simple lightweight table for displaying data
 * This is intended for small simple tables and does not do any pagination, sorting, or filtering
 */
const props = defineProps<{
  rows: any[]
  columns: ISimpleTableColumn[]

  /**
   * Hide the table header
   */
  hideHeader?: boolean
  

  /**
   * Flat design, eg no shadow
   */
  flat?: boolean

  /**
   * Compact the cells
   */
  dense?: boolean
}>()

const test1 = (row:any):VNode => {
  return h('h1', {}, "test")
}

</script>

<template>
  <QMarkupTable>
    <thead>
      <tr key="header">
        <th :key="column.field" v-for="column in props.columns" 
          :style="{textAlign: column.align?column.align:'left', width:column.width}"
          >
          {{column.title}}
        </th>
      </tr>
    </thead>

    <tbody>
      <tr :key="row.id" v-for="row in props.rows">
        <td :key="column.field" v-for="column in props.columns">
          <span v-if="column.component">
            <component :is="column.component" row/>
          </span>
          <span v-else>
            {{_get(row, column.field, "")}}
            <!-- {{row[column.field]}} -->
          </span>
        </td>
      </tr>
    </tbody>
  </QMarkupTable>
</template>
