<script lang="ts">
// exports must be in a separate script block https://github.com/vuejs/rfcs/pull/227
import { defineComponent} from "vue";
export const MenuAbout = "about";
export const MenuAddPage = "addPage";
export const MenuEditMode = "editMode";
export const MenuSettings = "settings"
export default defineComponent({});
</script>

<script lang="ts" setup>
import Button from 'primevue/button';
import Menu from 'primevue/menu';
import IconMenu from '~icons/mdi/menu'
import {ref} from "vue";

interface IAppMenu {
  editMode: boolean;
  pages: Array<{label:string, icon?: string, to?: string}>;
}
const props = withDefaults(
    defineProps<IAppMenu>(),
    {
      editMode: false,
      pages: ()=>[{label:"Overview"}],
    })

const emit = defineEmits(['onMenuSelect'])

defineExpose({MenuAbout, MenuAddPage, MenuEditMode, MenuSettings})

// reference to the component
const themenu = ref<any>(null);

const handleMenuSelect = (name:string) => {
  console.log('AppMenu: onMenuSelect: ', name);
  emit("onMenuSelect", name);
}

const menuItems = [...props.pages, {
  separator: true,
}, {
  label: "Add Page...",
  icon: "mdi-add",
  command: ()=>handleMenuSelect(MenuAddPage),
}, {
  label: "Edit Mode",
  icon: "",
  command: ()=>handleMenuSelect(MenuEditMode),
}, {
  label: "Connections...",
  icon: "mdi-link",
  to: "/accounts",
  command: ()=>handleMenuSelect(MenuSettings),
},{
  label: "About...",
  icon: "mdi-about",
  command: ()=>handleMenuSelect(MenuAbout),
}]

const toggle = (event:any) => {
  themenu.value.toggle(event);
}

</script>


<template>
  <Button style="margin-right: 10px"
          @click="toggle" class="p-button-text">
    <IconMenu/>
  </Button>
  <Menu ref="themenu" :model="menuItems" :popup="true"/>

</template>

