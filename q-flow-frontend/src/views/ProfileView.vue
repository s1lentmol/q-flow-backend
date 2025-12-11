<template>
  <main>
    <ProfilePanel :errors="errors" :linkToken="linkToken" @create-link="createLinkToken" />
  </main>
</template>

<script setup>
import { reactive, ref } from "vue";
import ProfilePanel from "../components/ProfilePanel.vue";
import { api } from "../api";

const errors = reactive({ contact: "" });
const linkToken = ref(null);

async function createLinkToken(username) {
  errors.contact = "";
  try {
    linkToken.value = await api("/profile/contact/link", "POST", {});
  } catch (e) {
    errors.contact = e.message;
  }
}
</script>
