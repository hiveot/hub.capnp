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
  mdiMenu,
  mdiInformation,
  mdiLink,
  mdiPlus,
  mdiCheckboxOutline,
  mdiCheckboxBlankOutline
} from "@quasar/extras/mdi-v6";

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

const handleMenuSelect = (name:any) => {
  console.log('AppMenu: onMenuSelect: ', name);
  emit("onMenuSelect", name);
}


const getMenuItems = (pages: Array<IMenuItem>, editMode:boolean): Array<IMenuItem> => {
  return  [...pages, {
    separator: true,
  }, {
    label: "Add Page...",
    icon: mdiPlus,
    id: MenuAddPage,
  }, {
    label: "Edit Mode",
    icon: editMode ? mdiCheckboxOutline : mdiCheckboxBlankOutline,
    id: MenuEditMode,
  }, {
    id: MenuAccounts,
    label: "Accounts...",
    icon: mdiLink,
    to: "/accounts",
  },{
    id: MenuAbout,
    label: "About...",
    icon: mdiInformation,
  }]
}



</script>


<template>
  <MenuButton :icon="mdiMenu"
            :items="getMenuItems(props.pages, props.editMode)"
   @onMenuSelect='handleMenuSelect'
  />

</template>

