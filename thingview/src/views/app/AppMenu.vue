<script lang="ts">
// exports must be in a separate script block https://github.com/vuejs/rfcs/pull/227
import { defineComponent} from "vue";
export const MenuAbout = "about";
export const MenuAddPage = "addPage";
export const MenuEditMode = "editMode";
export const MenuAccounts = "accounts"
export default defineComponent({});
</script>

<script lang="ts" setup>

import MenuButton, {IMenuItem} from "@/components/MenuButton.vue";
import {
  matMenu,
  matInfo,
  matLink,
  matAdd,
  matCheckBox,
  matCheckBoxOutlineBlank
} from "@quasar/extras/material-icons";

interface IAppMenu {
  editMode: boolean;
  pages: Array<IMenuItem>;
}
const props = withDefaults(
    defineProps<IAppMenu>(),
    {
      editMode: false,
      pages: ()=>[],
    })

const emit = defineEmits(['onMenuSelect']) // IMenuItem

// defineExpose({MenuAbout, MenuAddPage, MenuEditMode, MenuSettings})

const handleMenuSelect = (item:IMenuItem) => {
  console.log('AppMenu: onMenuSelect: ', item);
  emit("onMenuSelect", item);
}


const getMenuItems = (pages: Array<IMenuItem>, editMode:boolean): Array<IMenuItem> => {
  return  [...pages, {
    separator: true,
  }, {
    label: "Add Page...",
    icon: matAdd,
    id: MenuAddPage,
  }, {
    label: "Edit Mode",
    icon: editMode ? matCheckBox : matCheckBoxOutlineBlank,
    id: MenuEditMode,
  }, {
    id: MenuAccounts,
    label: "Accounts...",
    icon: matLink,
    to: "/accounts",
  },{
    id: MenuAbout,
    label: "About...",
    icon: matInfo,
  }]
}



</script>


<template>
  <MenuButton :icon="matMenu"
            :items="getMenuItems(props.pages, props.editMode)"
   @onMenuSelect='handleMenuSelect'
  />

</template>

