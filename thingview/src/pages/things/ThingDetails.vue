<script lang="ts" setup>

import { ThingTD } from '@/data/td/ThingTD';
import { useQuasar, QCardSection, QTable } from 'quasar';

const props = defineProps({
  td: ThingTD
})



const columns: Array<Object> = [
  {name: "Property", label: "Property", field:""},
  {name: "Value", label: "Value", field:""}
]

</script>

<template>

<TDialog>
  <QCardSection title="Thing Attributes">
    <QTable title="Details of {{props.td.['@type']: props.td.description}}"
    >

    </QTable>
  </QCardSection>

<QCardSection title="Thing Inputs">
    </QCardSection>
    
<QCardSection title="Thing Outputs">
    </QCardSection>
    
<QCardSection title="Thing Configuration">
</QCardSection>


</TDialog>

</template>