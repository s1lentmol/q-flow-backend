<template>
  <section class="panel" v-if="activeQueue">
    <div class="row" style="justify-content: space-between; align-items: flex-start">
      <div>
        <h2>Очередь: {{ activeQueue.queue.title }}</h2>
        <p class="muted">{{ activeQueue.queue.description }}</p>
        <div class="inline muted">
          <span v-if="modeLabel">{{ modeLabel }}</span>
          <span v-if="activeQueue.queue.status === 'archived'">архив</span>
        </div>
      </div>
      <div class="row">
        <button class="btn secondary" @click="$emit('advance')">Продвинуть</button>
        <button class="btn secondary" @click="$emit('leave')">Покинуть</button>
      </div>
    </div>
    <div class="row" style="margin-bottom:10px">
      <button class="btn secondary" @click="emitJoin">Вступить</button>
    </div>
    <div class="row" style="margin-bottom:10px" v-if="isManaged">
      <input style="flex:1" v-model.number="manualUserId" placeholder="User ID для ручного добавления" />
      <button class="btn secondary" @click="emitAdd">Добавить участника</button>
    </div>
    <div class="divider"></div>
    <h3>Обновить очередь</h3>
    <div class="row">
      <input style="flex:1" v-model="updateForm.title" placeholder="Новое название" />
      <input style="flex:1" v-model="updateForm.description" placeholder="Новое описание" />
      <button class="btn secondary" @click="emitUpdate">Сохранить</button>
    </div>
    <div class="divider"></div>
    <h3>Участники</h3>
    <div class="list">
      <div class="item inline" v-for="p in activeQueue.participants" :key="p.id">
        <strong>#{{ p.position }}</strong>
        <span class="muted">{{ p.full_name || p.name || ('Пользователь ' + p.user_id) }}</span>
        <button class="btn text" @click="$emit('remove', p.user_id)">Удалить</button>
      </div>
    </div>
  </section>
</template>

<script setup>
import { reactive, ref, watch, computed } from "vue";

const props = defineProps({
  activeQueue: { type: Object, default: null },
});

const emit = defineEmits(["join", "leave", "advance", "remove", "add", "update"]);

const manualUserId = ref("");
const updateForm = reactive({ title: "", description: "" });
const modeLabel = ref("");
const isManaged = computed(() => props.activeQueue?.queue?.mode === "managed");

watch(
  () => props.activeQueue,
  (val) => {
    if (val?.queue) {
      updateForm.title = val.queue.title || "";
      updateForm.description = val.queue.description || "";
      modeLabel.value = mapMode(val.queue.mode);
    }
  },
  { immediate: true },
);

function emitJoin() {
  emit("join");
}

function emitAdd() {
  if (!manualUserId.value) return;
  emit("add", { userId: Number(manualUserId.value) });
  manualUserId.value = "";
}

function emitUpdate() {
  emit("update", { ...updateForm });
}

function mapMode(mode) {
  switch (mode) {
    case "live":
      return "живая очередь";
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
