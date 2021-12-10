<script  lang="ts" setup>
import {reactive} from "vue";
import {hubAuth} from '@/data/HubAuth';
import {QInput, QCheckbox} from "quasar";

const props = defineProps({
  title: {
    type: String,
    default: "WoST Login"
  },
})

  // emits: ['onLogin'], // {login:String, password:String}
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

</script>

<template >
<div class="container">
  <h3>{{title}}</h3>

  <div class="p-fluid p-formgrid p-grid">

    <div  class="p-field p-grid">
      <label for="loginEmail" class="p-col-fixed">Login Email</label>
      <div class="p-col">
        <QInput v-model="data.loginEmail" id="loginEmail"
                type="text" placeholder="Your login email"
        label="email"/>
      </div>
    </div>

    <div  class="p-field p-grid">
      <label for="loginPassword" class="p-col-fixed">Password</label>
      <div class="p-col">
        <QInput v-model="data.password"
                  id="loginPassword" type="password"
                  placeholder="Your login password"/>
      </div>
    </div>


    <div style="text-align: left">
      <QCheckbox
        label="Remember Me"
        v-model="data.rememberMe"
      />
    </div>
    <div style="display: flex; justify-content: flex-end;">
        <QBtn color="primary"
          :disabled="data.busyLoggingIn||data.loginEmail===''||data.password===''"
          @click="handleLoginButtonClick" >
        Login
        </QBtn>
      </div>
  </div>
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