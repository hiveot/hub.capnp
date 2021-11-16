<script lang="ts">
// exports must be in a separate script block https://github.com/vuejs/rfcs/pull/227
import {defineComponent} from "vue";
export const MenuAbout = "about";
export const MenuAddPage = "addPage";
export const MenuEditMode = "editMode";
export const MenuSettings = "settings"
export default defineComponent({});
</script>

<script lang="ts" setup>

interface IAppMenu {
  editMode: boolean;
  pages: Array<string>;
}
const props = withDefaults(
    defineProps<IAppMenu>(),
    {
      editMode: false,
      pages: ()=>["Overview"],
    })

const emit = defineEmits(['onMenuSelect'])

const handleMenuSelect = (name:string) => {
  console.log('AppMenu: onMenuSelect: ', name);
  emit("onMenuSelect", name);
}

const menuItems = [{
  label: "Add Page...",
  icon: "mdi-add"
}, {
  label: "Edit Mode",
  icon: "",
}, {
  label: "Connections...",
  icon: "mdi-link"
},{
  label: "About...",
  icon: "mdi-about"
}]

</script>


<template>
  <q-btn style="margin-right: 10px" flat icon="mdi-menu">
  <q-menu auto-close>
    <q-list style="min-width: 150px">
      <!-- Allow page select from the menu in case tabs are hidden -->
      <q-item v-for="page in props.pages"
              clickable
        @click='handleMenuSelect(page)'
      >
        <q-item-section avatar>
          <q-icon name="mdi-book-open-page-variant"/>
        </q-item-section>
        <q-item-section>
          <q-item-label>{{page}}</q-item-label>
        </q-item-section>
      </q-item>

      <q-separator/>
      <!-- Add a page -->
      <q-item clickable @click='handleMenuSelect(MenuAddPage)'>
        <q-item-section avatar>
          <q-icon name="mdi-plus" />
        </q-item-section>
        <q-item-section no-wrap>
          Add Page...
        </q-item-section>
      </q-item>

      <!-- Toggle edit mode -->
      <q-item clickable
        @click='handleMenuSelect(MenuEditMode)'>
        <q-item-section avatar>
          <q-icon
              :name='(props.editMode)?"mdi-checkbox-marked-outline":"mdi-checkbox-blank-outline"'
              />
        </q-item-section>
        <q-item-section>
            <q-item-label>Edit Mode</q-item-label>
        </q-item-section>
      </q-item>

        <!-- Show settings menu -->
      <q-item clickable
              @click='handleMenuSelect(MenuSettings)'>
        <q-item-section avatar>
          <q-icon name="mdi-lan-connect"/>
        </q-item-section>
        <q-item-section>
          <q-item-label>Connections</q-item-label>
        </q-item-section>
      </q-item>

        <!-- Show About dialog -->
        <q-item clickable  @click='handleMenuSelect(MenuAbout)'>
          <q-item-section avatar>
            <q-img src="@/assets/logo.png" style="width:22px; height:22px"/>
          </q-item-section>
          <q-item-section>
            About...
          </q-item-section>

        </q-item>

      </q-list>
  </q-menu>
  </q-btn>
</template>

