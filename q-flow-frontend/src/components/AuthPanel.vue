<template>
  <section class="panel">
    <h2>{{ mode === 'register' ? 'Регистрация' : 'Вход' }}</h2>
    <div class="row" style="margin-bottom:8px">
      <button
        :class="['btn', mode === 'login' ? 'secondary' : 'text']"
        @click="mode = 'login'"
      >Вход</button>
      <button
        :class="['btn', mode === 'register' ? 'secondary' : 'text']"
        @click="mode = 'register'"
      >Регистрация</button>
    </div>
    <form v-if="mode === 'register'" @submit.prevent="$emit('register', registerForm)">
      <div>
        <label>Email</label>
        <input v-model="registerForm.email" type="email" required />
      </div>
      <div>
        <label>ФИО</label>
        <input v-model="registerForm.full_name" required />
      </div>
      <div>
        <label>Пароль</label>
        <input v-model="registerForm.password" type="password" required />
      </div>
      <button class="btn" type="submit">Зарегистрироваться</button>
    </form>
    <form v-else @submit.prevent="$emit('login', loginForm)">
      <div>
        <label>Email</label>
        <input v-model="loginForm.email" type="email" required />
      </div>
      <div>
        <label>Пароль</label>
        <input v-model="loginForm.password" type="password" required />
      </div>
      <button class="btn" type="submit">Войти</button>
    </form>
    <p class="error" v-if="errors.register && mode === 'register'">{{ errors.register }}</p>
    <p class="error" v-if="errors.login && mode === 'login'">{{ errors.login }}</p>
  </section>
</template>

<script setup>
import { reactive, ref } from "vue";

defineProps({
  errors: { type: Object, default: () => ({}) },
});

const registerForm = reactive({ email: "", password: "", full_name: "" });
const loginForm = reactive({ email: "", password: "" });
const mode = ref("login");
</script>
