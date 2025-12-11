import { reactive, computed } from "vue";

const defaultBase = window.localStorage.getItem("qflow_api") || "http://localhost:8080";

export const session = reactive({
  token: window.localStorage.getItem("qflow_token") || "",
  baseUrl: defaultBase,
});

export const isAuthed = computed(() => !!session.token);

export function setToken(token) {
  session.token = token || "";
  if (token) {
    window.localStorage.setItem("qflow_token", token);
  } else {
    window.localStorage.removeItem("qflow_token");
  }
}

export function setBaseUrl(url) {
  session.baseUrl = url;
  window.localStorage.setItem("qflow_api", url);
}

export async function api(path, method = "GET", body) {
  const res = await fetch(session.baseUrl + path, {
    method,
    headers: {
      "Content-Type": "application/json",
      ...(session.token ? { Authorization: `Bearer ${session.token}` } : {}),
    },
    body: body ? JSON.stringify(body) : undefined,
  });

  const text = await res.text();
  const payload = text ? safeJson(text) : {};

  if (!res.ok) {
    const errMsg = payload?.error || res.statusText || "Request failed";
    throw new Error(errMsg);
  }

  return payload?.data ?? payload ?? null;
}

function safeJson(text) {
  try {
    return JSON.parse(text);
  } catch (_) {
    return null;
  }
}
