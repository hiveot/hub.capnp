<script lang="ts" setup>

// Wrapper around the QTable for showing a list of Things
// QTable slots are available to the parent
import {ref} from 'vue'
import  {ThingTD} from "@/data/td/ThingTD";
import {QBtn, QIcon, QToolbar, QTable, QTd, QToggle, QToolbarTitle, QTableProps} from "quasar";
import {matVisibility} from "@quasar/extras/material-icons"


// Accounts table API
interface IThingsTable {
  // Collection of things
  things: Array<ThingTD>
  // optional title to display above the table
  title?: string
}
const props = defineProps<IThingsTable>()


const emit = defineEmits([
  'onViewDetails'
])

// table columns
interface IThingCol {
  name: string
  label: string
  field: string
  required?: boolean
  align?: "left"| "right" | "center" | undefined
  style?: string
  format?: (val:any, row:any)=>any
}

// The column field name should match the TD field names
const columns: Array<IThingCol> = [
  // {name: "edit", label: "", field: "edit", align:"center"},
 // Use large width to minimize the other columns
  {name: "id", label: "Thing ID", field:"id" , align:"left", required:true, style:"width:400px",
    },
  {name: "desc", label: "Description", field:"description" , align:"left", required:true,
    },
  {name: "type", label: "Device Type", field:"@type", align:"left", required:true, },
  // {name: "pub", label: "Publisher", field:"pub", align:"left", required:true, },
  {name: "details", label: "Detail", field:"", align:"center", required:true, },

]

const visibleColumns = ref([ 'name', 'type', 'pub' ])
</script>


<template>
  <QTable :rows="props.things"
          :columns="columns"
          :visible-columns="visibleColumns"
          hide-pagination
          row-key="id"
          hide-selected-banner
          table-header-class="text-primary text-h5"
  >
    <!-- export the slots-->
    <template v-for="(index, name:string|number) in $slots" v-slot:[name]>
      <slot :name="name" />
    </template>

    <!-- Header style: large and primary -->
    <template #header-cell="props">
      <q-th style="font-size:1.1rem;" :props="props">{{props.col.label}}</q-th>
    </template>

    <!-- button for viewing the Thing TD -->
    <template v-slot:body-cell-details="propz">
      <QTd>
        <QBtn dense flat round color="blue" field="edit" :icon="matVisibility"
              @click="emit('onViewDetails', propz.row)"/>
      </QTd>
    </template>
  </QTable>
</template>
