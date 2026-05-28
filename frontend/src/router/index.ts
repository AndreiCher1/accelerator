import { createRouter, createWebHistory } from "vue-router";
import type { RouteRecordRaw } from "vue-router";
import { useAuthStore } from "../stores/auth";

const routes: Array<RouteRecordRaw> = [
  {
    path: "/",
    name: "Home",
    redirect: () => {
      const authStore = useAuthStore();

      if (authStore.accessToken) {
        if (authStore.requiresPasswordChange) {
          return { name: "ChangePassword" };
        }
        return { name: "Dashboard" };
      }
      return { name: "Login" };
    },
  },
  {
    path: "/login",
    name: "Login",
    component: () => import("../views/LoginView.vue"),
    meta: {
      title: "Авторизация",
      guestOnly: true,
    },
  },
  {
    path: "/setup",
    name: "CreatorSetup",
    component: () => import("../views/CreatorSetupView.vue"),
    meta: {
      title: "Первичная настройка",
      guestOnly: true,
    },
  },
  {
    path: "/change-password",
    name: "ChangePassword",
    component: () => import("../views/ChangePasswordView.vue"),
    meta: {
      title: "Смена пароля",
      requiresAuth: true,
      requiresPasswordChange: true,
    },
  },
  {
    path: "/dashboard",
    name: "Dashboard",
    component: () => import("../views/DashboardView.vue"),
    meta: {
      title: "Главная",
      requiresAuth: true,
    },
  },
  {
    path: "/templates",
    name: "Templates",
    redirect: { name: "AdminPatterns" },
  },
  {
    path: "/admin/patterns",
    name: "AdminPatterns",
    component: () => import("../views/AdminPatternsView.vue"),
    meta: {
      title: "Управление шаблонами",
      requiresAuth: true,
      requiresAdmin: true,
    },
  },
  {
    path: "/admin/users",
    name: "AdminUsers",
    component: () => import("../views/AdminUsersView.vue"),
    meta: {
      title: "Управление пользователями",
      requiresAuth: true,
      requiresAdmin: true,
    },
  },
  {
    path: "/admin/groups",
    name: "AdminGroups",
    component: () => import("../views/AdminGroupsView.vue"),
    meta: {
      title: "Управление группами",
      requiresAuth: true,
      requiresAdmin: true,
    },
  },
  {
    path: "/records/create",
    name: "RecordCreate",
    component: () => import("../views/CreateTaskView.vue"),
    meta: {
      title: "Создание записи",
      requiresAuth: true,
    },
  },
  {
    path: "/records/:id/processing",
    name: "RecordProcessingDetails",
    component: () => import("../views/ProcessingTaskDetailsView.vue"),
    meta: {
      title: "Детали конспекта",
      requiresAuth: true,
    },
  },
  {
    path: "/records/:id",
    name: "RecordDetails",
    component: () => import("../views/ReadyTaskDetailsView.vue"),
    meta: {
      title: "Детали конспекта",
      requiresAuth: true,
    },
  },
  {
    path: "/:pathMatch(.*)*",
    name: "NotFound",
    component: () => import("../views/DashboardView.vue"),
    meta: {
      title: "Страница не найдена",
    },
  },
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
  scrollBehavior(_to, _from, savedPosition) {
    if (savedPosition) {
      return savedPosition;
    } else {
      return { top: 0 };
    }
  },
});

router.beforeEach(async (to, _from, next) => {
  document.title = (to.meta.title as string) || "Платформа протоколирования";

  const authStore = useAuthStore();
  // Дожидаемся первой попытки восстановить сессию, чтобы роль из /auth/refresh
  // успела попасть в стор до проверки requiresAdmin.
  if (!authStore.isBootstrapped) {
    await authStore.bootstrap();
  }

  const isAuthenticated = !!authStore.accessToken;
  const requiresAuth = to.matched.some((record) => record.meta.requiresAuth);
  const guestOnly = to.matched.some((record) => record.meta.guestOnly);
  const isChangePasswordRoute = to.name === "ChangePassword";

  if (requiresAuth && !isAuthenticated) {
    next({ name: "Login", query: { redirect: to.fullPath } });
    return;
  }

  if (isAuthenticated && authStore.requiresPasswordChange) {
    if (!isChangePasswordRoute) {
      next({ name: "ChangePassword" });
      return;
    }
    next();
    return;
  }

  if (isAuthenticated && isChangePasswordRoute && !authStore.requiresPasswordChange) {
    next({ name: "Dashboard" });
    return;
  }

  if (guestOnly && isAuthenticated) {
    next({ name: "Dashboard" });
    return;
  }

  const requiresAdmin = to.matched.some((record) => record.meta.requiresAdmin);
  if (requiresAdmin) {
    const userRole = authStore.role;
    if (userRole !== "admin" && userRole !== "creator") {
      next({ name: "Dashboard" });
      return;
    }
  }

  next();
});

export default router;
