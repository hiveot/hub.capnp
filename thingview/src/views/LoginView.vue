<script  lang="ts" setup>
import {reactive} from "vue";
import {hubAuth} from '../store/HubAuth';

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
  <q-form  class="form"
    ref="form"
    labelPosition="left"
    labelWidth="120px"
    :model="data"
    :disabled="data.busyLoggingIn"
    size="medium"
  >
    <q-item label="Login Email" prop="loginEmail" required>
      <q-input
        v-model="data.loginEmail" 
        placeholder="Your login email"  
        maxlength="100" 
      />
    </q-item>
    <q-item label="Password" prop="password"  required
    >
      <q-input
        v-model="data.password" 
        placeholder="Password"  
        type="password" 
        minlength="3"
        />
    </q-item>
    <q-item style="text-align: left">
      <q-checkbox
        label="Remember Me"
        v-model="data.rememberMe"
      />
    </q-item>
    <div style="display: flex; justify-content: flex-end;">
        <q-btn round color="primary"
          :disabled="data.busyLoggingIn||data.loginEmail===''||data.password===''"
          @click="handleLoginButtonClick" >
        Login
        </q-btn>
      </div>
  </q-form>
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