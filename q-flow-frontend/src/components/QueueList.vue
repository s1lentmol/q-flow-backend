<template>
  <section class="panel">
    <h2>Очереди группы</h2>
    <div class="row">
      <input style="flex:1" :value="groupCode" @input="onGroup($event.target.value)" placeholder="Код группы" />
      <button class="btn secondary" @click="$emit('refresh')">Обновить</button>
    </div>
    <div class="list">
      <div class="item" v-for="q in queues" :key="q.id">
        <div class="inline">
          <strong>{{ q.title }}</strong>
          <span class="muted">{{ mapMode(q.mode) }}</span>
          <span class="muted" v-if="q.status === 'archived'">архив</span>
        </div>
        <p class="muted">{{ q.description }}</p>
        <div class="row">
          <button class="btn secondary" @click="$emit('open', q.id)">Открыть</button>
          <button class="btn danger" @click="$emit('archive', q.id)">Архивировать</button>
          <button class="btn danger" @click="$emit('delete', q.id)">Удалить</button>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup>
const props = defineProps({
  queues: { type: Array, default: () => [] },
  groupCode: { type: String, default: "" },
});

const emit = defineEmits(["update:groupCode", "refresh", "open", "archive", "delete"]);

function onGroup(val) {
  emit("update:groupCode", val);
}

function mapMode(mode) {
  switch (mode) {
    case "live":
      return "живая";
    case "managed":
      return "управляемая";
    case "random":
      return "случайная";
    case "slots":
      return "слоты";
    default:
      return "";
  }
}
</script>
