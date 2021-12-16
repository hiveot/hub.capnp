<script lang="ts" setup>

import {ThingTD, TDProperty, TDAction, TDEvent} from '@/data/td/ThingTD';
import { QCard, QCardSection, QTable } from 'quasar';

// Thing Details View
const props = defineProps<{td:ThingTD}>()


// columns to display action
const actionColumns = [
  {name: "title", label: "Action", field:"title", align:"left", sortable:true},
  {name: "description", label: "Description", field:"description", align:"left"},
]
// columns to display properties
const attributesColumns = [
  {name: "title", label: "Attributes", field:"title", align:"left", sortable:true},
  {name: "value", label: "Value", field:"value", align:"left",
    style:"max-width:400px; overflow: auto"},
  {name: "unit", label: "Unit", field:"unit", align:"left"},
]
// columns to display configuration
const configurationColumns = [
  {name: "title", label: "Configuration", field:"title", align:"left", sortable:true},
  {name: "value", label: "Value", field:"value", align:"left",
    style:"max-width:400px; overflow: auto"},
  {name: "type", label: "Type", field:"type", align:"left"},
  {name: "default", label: "Default", field:"default", align:"left"},
  {name: "unit", label: "Unit", field:"unit", align:"left"},
]

// columns to display events (outputs)
const eventColumns = [
  {name: "name", label: "Event", field:"name", align:"left", sortable:true},
  {name: "params", label: "Parameters", field:"params", align:"left"},
]
// Convert the actions map into an array for display
const getThingActions = (td: ThingTD): Array<TDAction> => {
  let res = Array<TDAction>()
  if (!td || !td.actions) {
    return Array<TDAction>()
  }
  for (let [key, val] of Object.entries(td.actions)) {
    res.push(val)
  }
  return res
}

// Convert the attributes into an array for display
const getThingAttributes = (td: ThingTD): Array<TDProperty> => {
  let res = Array<TDProperty>()
  if (!td || !td.properties) {
    console.error("Missing TD or TD without properties")
    return Array<TDProperty>()
  }
  for (let [key, val] of Object.entries(td.properties)) {
    if (!val.writable) {
      res.push(val)
    }
  }
  return res
}

// Convert the writable properties into an array for display
const getThingConfiguration = (td: ThingTD): Array<TDProperty> => {
  let res = Array<TDProperty>()
  if (!td || !td.properties) {
    return Array<TDProperty>()
  }
  for (let [key, val] of Object.entries(td.properties)) {
    if (val.writable) {
      res.push(val)
    }
  }
  return res
}

//
const getThingEvents = (td: ThingTD): Array<TDEvent> => {
  let res = Array<TDEvent>()
  if (!!td && !!td.events) {
    for (let [key, val] of Object.entries(td.actions)) {
      res.push(val)
    }
  }
  return res
}

</script>

<template>

  <QCard flat style="width: 100%">
    Attributes
    <QCardSection>
      <QTable row-key="id" dense striped
              :columns="attributesColumns"
              :rows="getThingAttributes(props.td)"
              :rows-per-page-options="[0]"
              table-header-style="background:lightgray"
              :visible-columns="['title', 'value', 'type', 'unit']"
      >
      </QTable>
    </QCardSection>

    Configuration
    <QCardSection title="Thing Configuration">
      <QTable row-key="id" dense striped
              :columns="configurationColumns"
              :rows="getThingConfiguration(props.td)"
              :rows-per-page-options="[0]"
              table-header-style="background:lightgray"
      >
      </QTable>
    </QCardSection>

    Events
    <QCardSection title="Thing Events">
      <QTable row-key="id" dense striped
              :columns="eventColumns"
              :rows="getThingEvents(props.td)"
              :rows-per-page-options="[0]"
              table-header-style="background:lightgray"
              no-data-label="No events available"
      ></QTable>
    </QCardSection>

    Actions
    <QCardSection title="Thing Actions">
      <QTable row-key="id" dense striped
              :columns="actionColumns"
              :rows="getThingActions(props.td)"
              :rows-per-page-options="[0]"
              table-header-style="background:lightgray"
              no-data-label="No actions available"
      >
      </QTable>
    </QCardSection>



  </QCard>
</template>
