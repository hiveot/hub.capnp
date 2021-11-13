<script lang="ts">
import {defineComponent} from "vue";

import { ElButton, ElDropdown, ElDropdownMenu, ElDropdownItem } from "element-plus";
import MdiMenu from '@iconify/icons-mdi/menu'
import { addIcon } from '@iconify/vue/dist/offline';

export const MenuAbout = "about";
export const MenuAddPage = "addPage";
export const MenuEditMode = "editMode";
export const MenuSettings = "settings"


export default defineComponent({
  components: {ElButton, ElDropdown, ElDropdownMenu, ElDropdownItem: ElDropdownItem},

  emits: ['onMenuSelect'],

  props: {
    editMode: Boolean,
    pages: {
      // https://forum.vuejs.org/t/vue-typescript-problem-with-component-props-array-type-declaration/29478/7
      type: Array as () => Array<any>,
      default: ()=>[""],
    },
  },

  setup(props, {emit}) {
    addIcon('mdi:menu', MdiMenu);

    const handleMenuSelect = (name:string) => {
      console.log('AppMenu: onMenuSelect: ', name);
      emit("onMenuSelect", name);
    }
    return {handleMenuSelect, MenuAbout, MenuAddPage, MenuEditMode, MenuSettings};
  }
})
</script>


<template>
  <ElDropdown  trigger="click" >
    <button  style="margin-right: 10px" className='buttonHover'>
      <v-icon height="24" icon="mdi:menu" />
    </button>
    <template #dropdown>
      <ElDropdownMenu>
        <!-- Allow page select from the menu in case tabs are hidden -->
        <ElDropdownItem v-for="page in pages" 
          @click='handleMenuSelect(page)'
        >
          {{page}}
        </ElDropdownItem>

        <ElDropdownItem divided/>
        <!-- Add a page -->
        <ElDropdownItem 
          icon="el-icon-plus" 
          @click='handleMenuSelect(MenuAddPage)'>
          Add Page...
        </ElDropdownItem>

        <!-- Toggle edit mode -->
        <ElDropdownItem :icon='(editMode)?"el-icon-check":"el-icon-edit"' 
          @click='handleMenuSelect(MenuEditMode)'>
          Edit Mode
        </ElDropdownItem>

        <!-- Show settings menu -->
        <ElDropdownItem icon="el-icon-setting" 
            @click='handleMenuSelect(MenuSettings)'>
            Settings...
        </ElDropdownItem>

        <!-- Show About dialog -->
        <ElDropdownItem 
        @click='handleMenuSelect(MenuAbout)'>
          About...
        </ElDropdownItem>

      </ElDropdownMenu>
    </template>
  </ElDropdown>
</template>

