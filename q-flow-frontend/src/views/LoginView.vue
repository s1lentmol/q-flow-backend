<template>
  <main>
    <AuthPanel :errors="errors" @register="register" @login="login" />
  </main>
</template>

<script setup>
import { reactive } from "vue";
import { useRouter } from "vue-router";
import AuthPanel from "../components/AuthPanel.vue";
import { api, setToken } from "../api";

const router = useRouter();
const errors = reactive({ register: "", login: "" });

async function register(payload) {
  errors.register = "";
  try {
    await api("/auth/register", "POST", payload);
    await login({ email: payload.email, password: payload.password });
  } catch (e) {
    errors.register = e.message;
  }
}

async function login(payload) {
  errors.login = "";
  try {
    const data = await api("/auth/login", "POST", payload);
    setToken(data.token);
    router.push({ name: "queues" });
  } catch (e) {
    errors.login = e.message;
  }
}
</script>
