<template>
  <section class="panel">
    <h2>Telegram</h2>
    <p class="muted">Получите ссылку /start для привязки Telegram.</p>
    <button class="btn secondary" @click="$emit('create-link')">Получить ссылку</button>
    <p class="muted" v-if="linkToken?.link">Ссылка: <a :href="linkToken.link" target="_blank">{{ linkToken.link }}</a></p>
    <p class="error" v-if="errors.contact">{{ errors.contact }}</p>
  </section>
</template>

<script setup>
defineProps({
  errors: { type: Object, default: () => ({}) },
  linkToken: { type: Object, default: null },
});

defineEmits(["create-link"]);
</script>
