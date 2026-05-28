import { createApp } from "vue";
import { createPinia } from "pinia";
import App from "./App.vue";
import router from "./router";
import "./style.css";
import { useAuthStore } from "./stores/auth";

const pinia = createPinia();
const app = createApp(App).use(pinia).use(router);

// До монтирования пытаемся восстановить сессию: если в storage есть refresh-токен,
// делаем /auth/refresh, чтобы получить актуальную роль (она НЕ хранится в браузере).
const authStore = useAuthStore();
authStore
  .bootstrap()
  .catch(() => {
    /* ошибки уже обработаны внутри bootstrap */
  })
  .finally(() => {
    app.mount("#app");
  });
