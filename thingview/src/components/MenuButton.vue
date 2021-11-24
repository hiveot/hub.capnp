<script lang="ts" setup>
import {QBtn, QMenu, QList, QItem, QItemLabel, QItemSection} from "quasar";

// Menu item description
export interface IMenuItem {
  id?:string
  label?:string
  icon?: string
  to?: string
  separator?: boolean
}

// Dropdown menu properties
export interface IMenu {
  label?: string
  icon?: string  // mdi name
  items: Array<IMenuItem>
}
const props = withDefaults(defineProps<IMenu>(),{
  icon: "mdi-menu"
})

const emit = defineEmits(['onMenuSelect']) // IMenuItem

const handleMenuSelect = (item:IMenuItem) => {
  console.log('MenuButton: onMenuSelect: ', item.label);
  emit("onMenuSelect", item);
}
</script>


<template>
  <QBtn style="margin-right: 10px" flat
         :label="props.label" :icon="props.icon">
    <QMenu auto-close>
        <QList style="min-width: 150px;">
          <template  v-for="item in props.items">

           <QSeparator v-if='item.separator'/>
           <QItem v-else clickable :to="item.to"
                   @click='handleMenuSelect(item)'
           >
             <QItemSection v-if='item.icon !== ""' avatar>
                 <QIcon :name="item.icon"/>
             </QItemSection>

             <QItemSection v-if='item.label !== ""'>
               <QItemLabel>{{item.label}}</QItemLabel>
             </QItemSection>
           </QItem>
          </template>
        </QList>
      </QMenu>
  </QBtn>
</template>
