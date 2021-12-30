<script lang="ts" setup>

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

</script>

<template>
  <QMarkupTable>
    <thead>
      <tr>
        <th v-for="header in props.columns">
          {{header.title}}
        </th>
      </tr>
    </thead>

    <tbody>
      <tr v-for="row in props.rows">
        <td v-for="column in props.columns">
          {{row[column.field]}}
        </td>
      </tr>
    </tbody>
  </QMarkupTable>
</template>
