<template>
  <main>
    <section class="panel">
      <div class="row" style="justify-content: space-between; align-items: center">
        <div>
          <h2>Очередь</h2>
          <p class="muted">ID: {{ queueId }}</p>
        </div>
        <div class="row">
          <input style="width:180px" v-model="groupCode" placeholder="Код группы" />
          <button class="btn secondary" @click="loadQueue">Обновить</button>
        </div>
      </div>
    </section>
    <QueueDetails
      :activeQueue="activeQueue"
      @join="joinQueue"
      @leave="leaveQueue"
      @advance="advanceQueue"
      @remove="removeParticipant"
      @add="addParticipant"
      @update="updateQueue"
    />
  </main>
</template>

<script setup>
import { reactive, ref, watch, onMounted } from "vue";
import QueueDetails from "../components/QueueDetails.vue";
import { api } from "../api";
import { appState } from "../state";

const props = defineProps({
  queueId: { type: Number, required: true },
  group: { type: String, default: "" },
});

const errors = reactive({ general: "" });
const activeQueue = ref(null);
const groupCode = ref(props.group || appState.groupCode);

watch(
  () => groupCode.value,
  (val) => {
    appState.groupCode = val;
  },
);

async function loadQueue() {
  if (!groupCode.value) return;
  try {
    activeQueue.value = await api(`/queues/${props.queueId}?group=${encodeURIComponent(groupCode.value)}`);
  } catch (e) {
    errors.general = e.message;
  }
}

async function updateQueue(payload) {
  if (!activeQueue.value) return;
  try {
    const dto = await api(`/queues/${props.queueId}`, "PUT", {
      group_code: groupCode.value,
      title: payload.title,
      description: payload.description,
    });
    activeQueue.value.queue = dto;
    await loadQueue();
  } catch (e) {
    errors.general = e.message;
  }
}

async function joinQueue() {
  if (!activeQueue.value) return;
  try {
    await api(`/queues/${props.queueId}/join`, "POST", {
      group_code: groupCode.value,
    });
    await loadQueue();
  } catch (e) {
    errors.general = e.message;
  }
}

async function leaveQueue() {
  if (!activeQueue.value) return;
  try {
    await api(`/queues/${props.queueId}/leave`, "POST", {
      group_code: groupCode.value,
    });
    await loadQueue();
  } catch (e) {
    errors.general = e.message;
  }
}

async function advanceQueue() {
  if (!activeQueue.value) return;
  try {
    await api(`/queues/${props.queueId}/advance`, "POST", {
      group_code: groupCode.value,
    });
    await loadQueue();
  } catch (e) {
    errors.general = e.message;
  }
}

async function removeParticipant(userId) {
  if (!activeQueue.value) return;
  try {
    await api(`/queues/${props.queueId}/remove`, "POST", {
      group_code: groupCode.value,
      user_id: userId,
    });
    await loadQueue();
  } catch (e) {
    errors.general = e.message;
  }
}

async function addParticipant({ userId }) {
  if (!activeQueue.value) return;
  try {
    await api(`/queues/${props.queueId}/add`, "POST", {
      group_code: groupCode.value,
      user_id: userId,
    });
    await loadQueue();
  } catch (e) {
    errors.general = e.message;
  }
}

onMounted(loadQueue);
</script>
