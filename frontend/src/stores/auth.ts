import { defineStore } from "pinia";
import { ref } from "vue";
import {
  authService,
  clearStoredTokens,
  extractApiErrorMessage,
  getStoredAccessToken,
  getStoredRefreshToken,
  setStoredTokens,
} from "../api";
import { registerLogoutHandler, registerRoleSetter } from "../api/auth-bridge";
import router from "../router";

const REQUIRES_PASSWORD_CHANGE_KEY = "requires_password_change";

function readRequiresPasswordChange(): boolean {
  return sessionStorage.getItem(REQUIRES_PASSWORD_CHANGE_KEY) === "1";
}

function persistRequiresPasswordChange(value: boolean): void {
  if (value) {
    sessionStorage.setItem(REQUIRES_PASSWORD_CHANGE_KEY, "1");
  } else {
    sessionStorage.removeItem(REQUIRES_PASSWORD_CHANGE_KEY);
  }
}

/**
 * DEV-only: позволяет переопределить роль пользователя через VITE_DEV_ROLE в .env.
 * Используется только для тестирования UI админских вкладок. Бекенд продолжит
 * проверять реальную роль из JWT.
 */
function devRoleOverride(): string | null {
  const raw = (import.meta.env.VITE_DEV_ROLE ?? "").toString().trim().toLowerCase();
  if (raw === "") return null;
  if (raw !== "creator" && raw !== "admin" && raw !== "user") return null;
  return raw;
}

export const useAuthStore = defineStore("auth", () => {
  const accessToken = ref<string | null>(getStoredAccessToken());
  const refreshToken = ref<string | null>(getStoredRefreshToken());
  // Роль НЕ хранится в localStorage/sessionStorage. Источник истины — ответ
  // /auth/login и /auth/refresh. Здесь это просто переменная в памяти модуля,
  // недоступная пользователю через DevTools-хранилища браузера.
  const role = ref<string | null>(devRoleOverride());
  const requiresPasswordChange = ref(readRequiresPasswordChange());
  // Признак того, что начальный refresh при загрузке приложения завершился.
  const isBootstrapped = ref(false);

  const isLoading = ref(false);
  const error = ref<string | null>(null);

  const setTokens = (access: string, refresh: string, rememberMe = true) => {
    accessToken.value = access;
    refreshToken.value = refresh;
    setStoredTokens(access, refresh, rememberMe);
  };

  const setRole = (nextRole: string | null | undefined) => {
    const finalRole = devRoleOverride() ?? nextRole ?? null;
    role.value = finalRole;
  };

  const setRequiresPasswordChange = (value: boolean) => {
    requiresPasswordChange.value = value;
    persistRequiresPasswordChange(value);
  };

  const login = async (email: string, pass: string, rememberMe: boolean) => {
    isLoading.value = true;
    error.value = null;

    try {
      const data = await authService.login({ login: email, password: pass });
      const refresh = data.refresh_token ?? "";
      setTokens(data.access_token, refresh, rememberMe);
      setRole(data.user_role);

      if (data.temporary_password) {
        setRequiresPasswordChange(true);
        await router.push({ name: "ChangePassword" });
        return;
      }

      setRequiresPasswordChange(false);
      const redirectPath = router.currentRoute.value.query.redirect as string | undefined;
      await router.push(redirectPath || { name: "Dashboard" });
    } catch (err: unknown) {
      error.value = extractApiErrorMessage(err, "Неправильные данные для авторизации");
      throw err;
    } finally {
      isLoading.value = false;
    }
  };

  const changeTempPassword = async (password: string) => {
    isLoading.value = true;
    error.value = null;

    try {
      await authService.changeTempPassword({ password });
      setRequiresPasswordChange(false);
      await router.push({ name: "Dashboard" });
    } catch (err: unknown) {
      error.value = extractApiErrorMessage(err, "Не удалось сменить пароль");
      throw err;
    } finally {
      isLoading.value = false;
    }
  };

  /**
   * Пытается восстановить сессию при старте приложения.
   * Роль не хранится в браузере, поэтому мы вытаскиваем её из ответа /auth/refresh.
   */
  const bootstrap = async () => {
    if (isBootstrapped.value) return;
    const storedRefresh = getStoredRefreshToken();
    if (!storedRefresh) {
      isBootstrapped.value = true;
      return;
    }

    try {
      const data = await authService.refresh({ refresh_token: storedRefresh });
      const refresh = data.refresh_token ?? storedRefresh;
      // Сохраняем токены в том же хранилище, в котором они уже лежали.
      const persistent = localStorage.getItem("access_token") !== null
        || (getStoredAccessToken() === null && localStorage.getItem("refresh_token") !== null);
      setTokens(data.access_token, refresh, persistent);
      setRole(data.user_role);
    } catch {
      // Невалидный refresh — чистим всё и отправляем на логин.
      accessToken.value = null;
      refreshToken.value = null;
      role.value = devRoleOverride();
      clearStoredTokens();
    } finally {
      isBootstrapped.value = true;
    }
  };

  const logout = () => {
    accessToken.value = null;
    refreshToken.value = null;
    role.value = devRoleOverride();
    error.value = null;
    setRequiresPasswordChange(false);
    clearStoredTokens();
    void router.push({ name: "Login" });
  };

  // Пробрасываем сеттер роли и обработчик выхода в axios-интерсептор,
  // чтобы /auth/refresh внутри интерсептора мог обновлять роль в памяти.
  registerRoleSetter((nextRole) => {
    setRole(nextRole);
  });
  registerLogoutHandler(() => {
    accessToken.value = null;
    refreshToken.value = null;
    role.value = devRoleOverride();
    setRequiresPasswordChange(false);
  });

  return {
    accessToken,
    refreshToken,
    role,
    requiresPasswordChange,
    isBootstrapped,
    isLoading,
    error,
    setTokens,
    setRole,
    setRequiresPasswordChange,
    bootstrap,
    login,
    changeTempPassword,
    logout,
  };
});
