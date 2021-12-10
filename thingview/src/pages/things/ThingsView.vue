<script lang="ts" setup>

import {QCard, QCardSection} from "quasar"

import ds from "@/data/td/DirectoryStore"
import ThingsTable from "./ThingsTable.vue"

const things = ds.GetAll()

</script>


<template>
  <div>An overview of things...

  <QCard class="my-card" flat>
    <QCardSection class="">
      <ThingsTable  :things="things"
                    title="Hub Things"
                    style="width: 100%"
      >
      </ThingsTable>
    </QCardSection>

  </QCard>
  </div>
</template>
