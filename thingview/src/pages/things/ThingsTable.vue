<script lang="ts" setup>

// Wrapper around the QTable for showing a list of Things
// QTable slots are available to the parent
import {ref} from 'vue'
import {date} from 'quasar'
import  {ThingTD} from "@/data/td/ThingTD";
import {QBtn, QIcon, QToolbar, QTable, QTd, QToggle, QToolbarTitle, QTableProps} from "quasar";
import TTable, {ITableCol} from '@/components/TTable.vue'
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

// The column field name should match the TD field names
const columns: Array<ITableCol> = [
  // {name: "edit", label: "", field: "edit", align:"center"},
 // Use large width to minimize the other columns
  {name: "id", label: "Thing ID", field:"id" , align:"left", style:"width:400px",
    sortable:true,
    },
  {name: "deviceID", label: "Device ID", field:"deviceID" , align:"left",
    sortable:true,
    },
  {name: "publisherID", label: "Publisher", field:"publisher" , align:"left",
    sortable:true,
    },
  {name: "deviceType", label: "Device Type", field:"deviceType" , align:"left",
    sortable:true,
    },
  {name: "desc", label: "Description", field:"description" , align:"left",
    sortable:true,
    },
  {name: "type", label: "@Type", field:"@type", align:"left", 
    sortable:true,
    },
  {name: "created", label: "Created", field:"created", align:"left", 
    format: (val, row) => getDateText(val),
    sortable:true,
    },
  // {name: "pub", label: "Publisher", field:"pub", align:"left",  },
  {name: "details", label: "Detail", field:"", align:"center",  
    sortable:true,
    },
]
// Convert iso9601 date format to text representation 
const getDateText = (iso:string): string => {
  let timeStamp = new Date(iso)
  // return date.formatDate(timeStamp, "ddd Do MMM YYYY HH:mm:ss (Z)")
  return date.formatDate(timeStamp, "ddd YYYY-MM-DD HH:mm:ss (Z)")
}

const visibleColumns = ['id', 'deviceID', 'publisherID', 'deviceType', 'desc', 'type', 'created', 'details']

</script>


<template>
  <TTable :rows="props.things"
          :columns="columns"
          :visible-columns="visibleColumns"
          row-key="id"
  >
    <!-- Header style: large and primary -->
    <!-- <template #header-cell="propz">
      <q-th style="font-size:1.1rem;" :props="props">{{propz.col.label}}</q-th>
    </template> -->

    <!-- button for viewing the Thing TD -->
    <template v-slot:body-cell-details="propz">
      <QTd>
        <!-- <QBtn dense flat round color="blue" field="edit" :icon="matVisibility"
              @click="emit('onViewDetails', propz.row)"/> -->
        <QBtn dense flat round :icon="matVisibility"
              color="primary"
              @click="emit('onViewDetails', propz.row)"/>
      </QTd>
    </template>
  </TTable>
</template>
