import { createRouter, createWebHistory } from "vue-router";
import LoginView from "./views/LoginView.vue";
import ProfileView from "./views/ProfileView.vue";
import QueuesView from "./views/QueuesView.vue";
import QueueDetailsView from "./views/QueueDetailsView.vue";
import { session } from "./api";

const routes = [
  { path: "/login", name: "login", component: LoginView },
  {
    path: "/profile",
    name: "profile",
    component: ProfileView,
    meta: { requiresAuth: true },
  },
  {
    path: "/queues",
    name: "queues",
    component: QueuesView,
    meta: { requiresAuth: true },
  },
  {
    path: "/queues/:id",
    name: "queueDetails",
    component: QueueDetailsView,
    props: (route) => ({
      queueId: Number(route.params.id),
      group: route.query.group || "",
    }),
    meta: { requiresAuth: true },
  },
  { path: "/", redirect: "/queues" },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach((to, from, next) => {
  if (to.meta.requiresAuth && !session.token) {
    next({ name: "login" });
  } else {
    next();
  }
});

export default router;
