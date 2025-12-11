<template>
  <main>
    <QueueList
      :queues="queues"
      v-model:groupCode="groupCode"
      @refresh="fetchQueues"
      @open="openQueue"
      @archive="archiveQueue"
      @delete="deleteQueue"
    />
    <QueueCreator :error="errors.create" @create="createQueue" />
  </main>
</template>

<script setup>
import { reactive, ref, watch } from "vue";
import { useRouter } from "vue-router";
import QueueList from "../components/QueueList.vue";
import QueueCreator from "../components/QueueCreator.vue";
import { api } from "../api";
import { appState } from "../state";

const router = useRouter();
const errors = reactive({ create: "", general: "" });
const queues = ref([]);
const groupCode = ref(appState.groupCode);

watch(
  () => groupCode.value,
  (val) => {
    appState.groupCode = val;
  },
);

async function fetchQueues() {
  if (!groupCode.value) return;
  try {
    queues.value = await api(`/queues?group=${encodeURIComponent(groupCode.value)}`);
  } catch (e) {
    errors.general = e.message;
  }
}

async function openQueue(id) {
  router.push({ name: "queueDetails", params: { id }, query: { group: groupCode.value } });
}

async function createQueue(payload) {
  errors.create = "";
  try {
    const dto = await api("/queues", "POST", payload);
    queues.value.unshift(dto);
  } catch (e) {
    errors.create = e.message;
  }
}

async function archiveQueue(id) {
  try {
    await api(`/queues/${id}/archive`, "POST", { group_code: groupCode.value });
    await fetchQueues();
  } catch (e) {
    errors.general = e.message;
  }
}

async function deleteQueue(id) {
  try {
    await api(`/queues/${id}`, "DELETE", { group_code: groupCode.value });
    queues.value = queues.value.filter((q) => q.id !== id);
  } catch (e) {
    errors.general = e.message;
  }
}
</script>
