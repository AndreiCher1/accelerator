<script setup lang="ts">
import { reactive, ref } from "vue";
import { useAuthStore } from "../stores/auth";
import Input from "../components/ui/Input.vue";
import Button from "../components/ui/Button.vue";
import { CircleX } from "@lucide/vue";

const authStore = useAuthStore();
const localError = ref<string | null>(null);

const form = reactive({
  password: "",
  confirmPassword: "",
});

const handleSubmit = async () => {
  localError.value = null;

  if (form.password.length < 8) {
    localError.value = "Минимальная длина пароля — 8 символов";
    return;
  }

  if (form.password !== form.confirmPassword) {
    localError.value = "Пароли не совпадают";
    return;
  }

  try {
    await authStore.changeTempPassword(form.password);
  } catch {
    localError.value = authStore.error;
  }
};
</script>

<template>
  <main class="flex min-h-screen w-full items-center justify-center bg-white p-6 dark:bg-dark">
    <div class="w-full max-w-md">
      <div class="mb-6">
        <h1 class="mb-2 text-2xl font-semibold text-gray-900 dark:text-white">Смена пароля</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400">
          Для продолжения работы задайте постоянный пароль. Временный пароль больше не будет действовать.
        </p>
      </div>

      <div
        v-if="localError || authStore.error"
        class="mb-4 flex items-center gap-2 rounded-xl bg-red-50 p-3 text-sm text-red-600 dark:bg-red-500/10 dark:text-red-400"
      >
        <CircleX :size="16" />
        {{ localError || authStore.error }}
      </div>

      <form class="space-y-4" @submit.prevent="handleSubmit">
        <Input
          v-model="form.password"
          label="Новый пароль"
          type="password"
          placeholder="Не менее 8 символов"
          required
        />
        <Input
          v-model="form.confirmPassword"
          label="Подтверждение пароля"
          type="password"
          placeholder="Повторите пароль"
          required
        />
        <Button :is-loading="authStore.isLoading" type="submit">Сохранить пароль</Button>
      </form>
    </div>
  </main>
</template>
