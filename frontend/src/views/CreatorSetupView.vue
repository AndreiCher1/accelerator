<script setup lang="ts">
import { reactive, ref } from "vue";
import { useRouter } from "vue-router";
import { CircleX } from "@lucide/vue";
import Input from "../components/ui/Input.vue";
import Button from "../components/ui/Button.vue";
import { extractApiErrorMessage, userService } from "../api";

const router = useRouter();

const form = reactive({
  login: "",
  fullName: "",
  position: "",
  password: "",
  confirmPassword: "",
});

const isLoading = ref(false);
const error = ref<string | null>(null);
const success = ref(false);

const handleSubmit = async () => {
  error.value = null;

  if (form.password.length < 8) {
    error.value = "Минимальная длина пароля — 8 символов";
    return;
  }

  if (form.password !== form.confirmPassword) {
    error.value = "Пароли не совпадают";
    return;
  }

  isLoading.value = true;

  try {
    await userService.registerCreator({
      login: form.login.trim(),
      full_name: form.fullName.trim(),
      position: form.position.trim(),
      password: form.password,
    });
    success.value = true;
  } catch (err: unknown) {
    error.value = extractApiErrorMessage(err, "Не удалось зарегистрировать создателя");
  } finally {
    isLoading.value = false;
  }
};
</script>

<template>
  <main class="flex min-h-screen w-full bg-white dark:bg-dark">
    <div class="flex flex-1 flex-col items-center justify-center p-6 sm:p-12">
      <div class="w-full max-w-md">
        <div class="mb-6">
          <h1 class="mb-2 text-2xl font-semibold text-gray-900 dark:text-white">
            Первичная настройка
          </h1>
          <p class="text-sm text-gray-500 dark:text-gray-400">
            Создайте учётную запись создателя платформы. Доступно только при первом запуске с пустой
            базой пользователей.
          </p>
        </div>

        <div
          v-if="success"
          class="space-y-4 rounded-xl border border-green-200 bg-green-50 p-4 text-sm text-green-800 dark:border-green-800 dark:bg-green-900/20 dark:text-green-300"
        >
          <p>Учётная запись создателя успешно создана. Войдите с указанным email и паролем.</p>
          <Button type="button" @click="router.push({ name: 'Login' })">Перейти к входу</Button>
        </div>

        <template v-else>
          <div
            v-if="error"
            class="mb-4 flex items-center gap-2 rounded-xl bg-red-50 p-3 text-sm text-red-600 dark:bg-red-500/10 dark:text-red-400"
          >
            <CircleX :size="16" />
            {{ error }}
          </div>

          <form class="space-y-4" @submit.prevent="handleSubmit">
            <Input v-model="form.login" label="Email" type="email" placeholder="admin@company.ru" required />
            <Input v-model="form.fullName" label="ФИО" placeholder="Иванов Иван Иванович" required />
            <Input v-model="form.position" label="Должность" placeholder="Руководитель" required />
            <Input
              v-model="form.password"
              label="Пароль"
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
            <Button :is-loading="isLoading" type="submit">Создать создателя</Button>
            <button
              type="button"
              class="w-full text-center text-sm text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
              @click="router.push({ name: 'Login' })"
            >
              Уже есть аккаунт? Войти
            </button>
          </form>
        </template>
      </div>
    </div>
  </main>
</template>
