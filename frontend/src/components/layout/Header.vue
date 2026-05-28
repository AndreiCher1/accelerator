<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";
import {
  Home,
  LayoutTemplate,
  SquareUserRound,
  UsersRound,
  Moon,
  SunDim,
  LogOut,
} from "@lucide/vue";
import { useTheme } from "../../composables/useTheme";
import { useAuthStore } from "../../stores/auth";

interface Props {
  /**
   * Tailwind max-width класс для внутреннего контейнера шапки.
   * Должен совпадать с шириной основного контента страницы,
   * чтобы логотип/навигация/действия были выровнены по нему.
   * Пример: "max-w-7xl", "max-w-4xl", "max-w-[1800px]".
   */
  maxWidth?: string;
}

const props = withDefaults(defineProps<Props>(), {
  maxWidth: "max-w-7xl",
});

const route = useRoute();
const { isDark, toggleTheme } = useTheme();
const authStore = useAuthStore();

const isAdmin = computed(() => {
  return authStore.role === "admin" || authStore.role === "creator";
});

const containerClass = computed(
  () =>
    `mx-auto flex h-16 w-full ${props.maxWidth} items-center justify-between px-4 sm:px-6 lg:px-8`,
);

const handleLogout = () => {
  authStore.logout();
};
</script>

<template>
  <header
    class="shrink-0 border-b border-gray-200 bg-white transition-colors dark:border-dark-border dark:bg-dark"
  >
    <div :class="containerClass">
      <router-link to="/dashboard" class="flex items-center">
        <img src="../../assets/images/logo.svg" alt="Логотип" class="h-8" />
      </router-link>

      <nav
        v-if="isAdmin"
        class="flex items-center gap-1 rounded-xl border border-gray-100 bg-gray-50/50 p-1 dark:border-dark-border dark:bg-dark-card"
      >
        <router-link
          to="/dashboard"
          :class="[
            'flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-colors',
            route.path === '/dashboard'
              ? 'bg-blue-600 text-white dark:bg-white dark:text-dark'
              : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-dark-elevated dark:hover:text-white',
          ]"
        >
          <Home :size="16" />
          Главная
        </router-link>

        <router-link
          to="/admin/patterns"
          :class="[
            'flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-colors',
            route.path === '/admin/patterns'
              ? 'bg-blue-600 text-white dark:bg-white dark:text-dark'
              : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-dark-elevated dark:hover:text-white',
          ]"
        >
          <LayoutTemplate :size="16" />
          Шаблоны
        </router-link>

        <router-link
          to="/admin/users"
          :class="[
            'flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-colors',
            route.path === '/admin/users'
              ? 'bg-blue-600 text-white dark:bg-white dark:text-dark'
              : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-dark-elevated dark:hover:text-white',
          ]"
        >
          <SquareUserRound :size="16" />
          Пользователи
        </router-link>

        <router-link
          to="/admin/groups"
          :class="[
            'flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-colors',
            route.path === '/admin/groups'
              ? 'bg-blue-600 text-white dark:bg-white dark:text-dark'
              : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-dark-elevated dark:hover:text-white',
          ]"
        >
          <UsersRound :size="16" />
          Группы
        </router-link>
      </nav>

      <!-- Right Actions -->
      <div class="flex items-center gap-3">
        <!-- Theme Toggle -->
        <button
          @click="toggleTheme"
          :class="
            isDark
              ? 'border-dark-border bg-dark-card'
              : 'border-gray-200 bg-gray-100'
          "
          class="relative flex h-8 w-14 cursor-pointer items-center rounded-full border p-1 transition-colors duration-300 focus:outline-none"
        >
          <div
            :class="
              isDark
                ? 'translate-x-6 bg-white'
                : 'translate-x-0 bg-white'
            "
            class="flex h-6 w-6 transform items-center justify-center rounded-full shadow-sm transition-transform duration-300"
          >
            <Moon
              v-if="isDark"
              :size="14"
              color="#1a1a1a"
              :stroke-width="2.5"
            />
            <SunDim
              v-else
              :size="14"
              color="#6a7282"
              :stroke-width="2.5"
            />
          </div>
        </button>

        <!-- Logout Button -->
        <button
          @click="handleLogout"
          class="flex h-8 w-8 items-center justify-center rounded-full text-gray-600 transition-colors hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-dark-elevated dark:hover:text-white"
        >
          <LogOut :size="18" />
        </button>
      </div>
    </div>
  </header>
</template>
