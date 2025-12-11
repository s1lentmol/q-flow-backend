<template>
  <div>
    <header>
      <h1>Q-Flow</h1>
      <div class="user">
        <nav class="nav">
          <RouterLink to="/queues">Очереди</RouterLink>
          <RouterLink to="/profile">Профиль</RouterLink>
        </nav>
        <span v-if="isAuthed">В сети</span>
        <span v-else>Гость</span>
        <button class="btn text" v-if="isAuthed" @click="logout">Выйти</button>
        <RouterLink class="btn secondary" v-else to="/login">Войти</RouterLink>
      </div>
    </header>
    <RouterView />
  </div>
</template>

<script setup>
import { computed } from "vue";
import { useRouter, RouterView, RouterLink } from "vue-router";
import { setToken, isAuthed as sessionAuthed } from "./api";

const router = useRouter();
const isAuthed = computed(() => sessionAuthed.value);

function logout() {
  setToken("");
  router.push({ name: "login" });
}
</script>
