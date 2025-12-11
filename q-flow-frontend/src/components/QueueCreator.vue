<template>
  <section class="panel">
    <h2>Создать очередь</h2>
    <form @submit.prevent="emitCreate">
      <div>
        <label>Название</label>
        <input v-model="form.title" required />
      </div>
      <div>
        <label>Описание</label>
        <textarea v-model="form.description"></textarea>
      </div>
      <div>
        <label>Режим</label>
        <select v-model="form.mode" required>
          <option value="live">Живая</option>
          <option value="managed">Управляемая</option>
          <option value="random">Случайная</option>
        </select>
      </div>
      <div>
        <label>Код группы</label>
        <input v-model="form.group_code" required />
      </div>
      <button class="btn" type="submit">Создать</button>
    </form>
    <p class="error" v-if="error">{{ error }}</p>
  </section>
</template>

<script setup>
import { reactive, watch } from "vue";

const props = defineProps({
  error: { type: String, default: "" },
});

const emits = defineEmits(["create"]);

const form = reactive({
  title: "",
  description: "",
  mode: "live",
  group_code: "",
});

watch(
  () => props.error,
  () => {},
);

function emitCreate() {
  emits("create", { ...form });
}
</script>
