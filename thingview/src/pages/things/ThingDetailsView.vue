<script lang="ts" setup>

import { ThingTD, TDProperty } from '@/data/td/ThingTD';
import { QCard, QCardSection, QTable } from 'quasar';

// Thing Details View
const props = defineProps<{td:ThingTD}>()

// table columns todo, move to table component
interface ICol {
  name: string
  label: string
  field: string
  required?: boolean
  align?: "left"| "right" | "center" | undefined
  style?: string
  headerStyle?: string
  sortable?: boolean
  format?: (val:any, row:any)=>any
}
const columns: Array<ICol> = [
  {name: "title", label: "Name", field:"title", align:"left", sortable:true},
  {name: "value", label: "Value", field:"value", align:"left",
    style:"max-width:130px",
    headerStyle:"max-width:30%; backgroudColor:green"},
  {name: "unit", label: "Unit", field:"unit", align:"left"}
]


// Convert the properties map into an array for display
const getThingAttributes = (td: ThingTD): Array<TDProperty> => {
  let res = Array<TDProperty>()
  if (!td || !td.properties) {
    console.error("Missing TD or TD without properties")
    return Array<TDProperty>()
  }
  for (let [key, val] of Object.entries(td.properties)) {
    res.push(val)
  }
  return res
}

</script>

<template>

  <QCard flat>
    Attributes
    <QCardSection>
      <QTable row-key="id" dense striped
              :columns="columns"
              :rows="getThingAttributes(props.td)"
              :rows-per-page-options="[0]"
              table-header-style="background:lightgray"
      >
      </QTable>
    </QCardSection>

    Inputs
    <QCardSection title="Thing Inputs">
    </QCardSection>

    Outputs
    <QCardSection title="Thing Outputs">
    </QCardSection>

    Configuration
    <QCardSection title="Thing Configuration">
    </QCardSection>

  </QCard>
</template>
