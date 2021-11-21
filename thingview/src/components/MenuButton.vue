<script lang="ts" setup>
import Button from "@/components/Button.vue"

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
  <q-btn style="margin-right: 10px" flat
         :label="props.label" :icon="props.icon">
    <q-menu auto-close>
        <q-list style="min-width: 150px" dense>
          <template  v-for="item in props.items">

           <q-separator v-if='item.separator'/>
           <q-item v-else clickable :to="item.to"
                   @click='handleMenuSelect(item)'
           >
             <q-item-section v-if='item.icon !== ""' avatar>
                 <QIcon :name="item.icon"/>
             </q-item-section>

             <q-item-section v-if='item.label !== ""'>
               <q-item-label>{{item.label}}</q-item-label>
             </q-item-section>
           </q-item>
          </template>
        </q-list>
      </q-menu>
  </q-btn>
</template>
