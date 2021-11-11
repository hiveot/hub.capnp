<script  lang="ts">
import {defineComponent, reactive} from "vue";
import {ElForm, ElFormItem, ElInput, ElButton, ElCheckbox} from 'element-plus';
import {hubAuth} from '../store/HubAuth';

export default defineComponent({
  // name: "LoginView",
  components: {ElForm, ElFormItem, ElInput, ElButton, ElCheckbox},
  props: {
    title: {
      type: String,
      default: "WoST Login"
    },
  },

  // emits: ['onLogin'], // {login:String, password:String}

  setup(props, context){
    const data = reactive({
      busyLoggingIn: false,
      loginEmail: "",
      password: "",
      rememberMe: false,
    })
    const handleLoginButtonClick = function(ev:any){
      ev.preventDefault();
      data.busyLoggingIn = true;
      console.log("Submitting 'onLogin' with user: "+data.loginEmail)
      // context.emit('onLogin', {login: data.loginEmail, password: data.password});
      hubAuth.login(data.loginEmail, data.password, data.rememberMe)
    }
    return {
      handleLoginButtonClick,
      data,
    }
  }
})
</script>

<template >
<div className="container">
  <h3>{{title}}</h3>
  <ElForm  class="form"
    ref="form"
    labelPosition="left"
    labelWidth="120px"
    :model="data"
    :disabled="data.busyLoggingIn"
    size="medium"
  >
    <ElFormItem label="Login Email" prop="loginEmail" required>
      <ElInput
        v-model="data.loginEmail" 
        placeholder="Your login email"  
        maxlength="100" 
      />
    </ElFormItem>
    <ElFormItem label="Password" prop="password"  required
    >
      <ElInput  
        v-model="data.password" 
        placeholder="Password"  
        type="password" 
        minlength="3"
        />
    </ElFormItem>
    <ElFormItem style="text-align: left">
      <ElCheckbox
        label="Remember Me"
        v-model="data.rememberMe"
      />
    </ElFormItem>
    <div style="display: flex; justify-content: flex-end;">
        <ElButton
          :disabled="data.busyLoggingIn||data.loginEmail==''||data.password==''"
          @click="handleLoginButtonClick" round type="primary">
        Login
        </ElButton>
      </div>
  </ElForm>
  </div>
</template>

<style scoped>
.container {
  display: flex;
  width: 100%;
  flex-direction: column;
  align-items: center;
}
.form {
    width:98%;
    max-width: 500px;
    display: flex;
    flex-direction: column;

    /* border: 1px solid var(--el-border-color-base); */
    padding: 40px;
    padding-bottom: 20px;
  }
</style>