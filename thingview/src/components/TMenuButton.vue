<script lang="ts" setup>
import {QBtn, QMenu, QIcon,QSeparator, QList, QItem, QItemLabel, QItemSection} from "quasar";

/** Menu item description */
export interface IMenuItem {
  /** unique id of menu item */
  id?:string

  /** label to display on the menu */
  label?:string

  /** optional icon to display with the menu */
  icon?: string
   
  /** route path, or object with {routename:name} */
  to?: string  | Object

  /** this is a separator */
  separator?: boolean

  /** Item is disabled */
  disabled?: boolean
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

const emit = defineEmits(['onMenuAction']) // IMenuItem

const handleMenuAction = (item:IMenuItem) => {
  console.debug('MenuButton: handleMenuAction: ', item.label);
  emit("onMenuAction", item);
}
</script>


<template>
  <QBtn flat
         :label="props.label" 
         :icon="props.icon">
    <QMenu auto-close>
        <QList style="min-width: 150px">
          <template  v-for="item in props.items">

           <QSeparator v-if='item.separator'/>
           
           <QItem v-else 
            :to="item.to"
            :disable="item.disabled"
            clickable @click='handleMenuAction(item)'
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
