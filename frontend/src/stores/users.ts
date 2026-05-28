import { defineStore } from "pinia";
import { computed, ref } from "vue";
import {
  ApiError,
  extractApiErrorMessage,
  userService,
  type AssignableUserRole,
  type EditUserRequest,
  type Pagination,
  type RegisterUserRequest,
  type RegisterUserResponse,
  type ResetPasswordResponse,
  type UserDto,
} from "../api";

const CACHE_TTL_MS = 5 * 60 * 1000;

interface UsersCacheEntry {
  users: UserDto[];
  pagination: Pagination;
  fetchedAt: number;
}

function cacheKey(page: number, limit: number): string {
  return `${page}:${limit}`;
}

export const useUsersStore = defineStore("users", () => {
  const users = ref<UserDto[]>([]);
  const pagination = ref<Pagination | null>(null);
  const isLoading = ref(false);
  const isMutating = ref(false);
  const error = ref<string | null>(null);
  const forbidden = ref(false);

  const cache = ref<Map<string, UsersCacheEntry>>(new Map());
  const currentPage = ref(1);
  const currentLimit = ref(50);

  const total = computed(() => pagination.value?.total ?? 0);
  const regularUsers = computed(() => users.value.filter((u) => u.role === "user"));
  const adminUsers = computed(() => users.value.filter((u) => u.role === "admin"));
  const isCacheValid = computed(() => {
    const entry = cache.value.get(cacheKey(currentPage.value, currentLimit.value));
    if (!entry) return false;
    return Date.now() - entry.fetchedAt < CACHE_TTL_MS;
  });

  function applyCacheEntry(entry: UsersCacheEntry): void {
    users.value = entry.users;
    pagination.value = entry.pagination;
  }

  function invalidateCache(): void {
    cache.value.clear();
  }

  async function fetchUsers(options?: {
    page?: number;
    limit?: number;
    force?: boolean;
  }): Promise<void> {
    const page = options?.page ?? currentPage.value;
    const limit = options?.limit ?? currentLimit.value;
    const force = options?.force ?? false;
    const key = cacheKey(page, limit);

    currentPage.value = page;
    currentLimit.value = limit;

    const cached = cache.value.get(key);
    if (!force && cached && Date.now() - cached.fetchedAt < CACHE_TTL_MS) {
      applyCacheEntry(cached);
      error.value = null;
      forbidden.value = false;
      return;
    }

    isLoading.value = true;
    error.value = null;
    forbidden.value = false;

    try {
      const data = await userService.getUsers({ page, limit });
      const entry: UsersCacheEntry = {
        users: data.users,
        pagination: data.pagination,
        fetchedAt: Date.now(),
      };
      cache.value.set(key, entry);
      applyCacheEntry(entry);
    } catch (err: unknown) {
      if (err instanceof ApiError && err.isForbidden) {
        forbidden.value = true;
        error.value = "Недостаточно прав для просмотра пользователей";
      } else {
        error.value = extractApiErrorMessage(err, "Не удалось загрузить пользователей");
      }
      throw err;
    } finally {
      isLoading.value = false;
    }
  }

  async function registerUser(payload: RegisterUserRequest): Promise<RegisterUserResponse> {
    isMutating.value = true;
    try {
      const created = await userService.registerUser(payload);
      invalidateCache();
      await fetchUsers({ force: true });
      return created;
    } finally {
      isMutating.value = false;
    }
  }

  async function createUser(payload: RegisterUserRequest): Promise<RegisterUserResponse> {
    return registerUser(payload);
  }

  async function updateUser(userId: string, payload: EditUserRequest): Promise<UserDto> {
    isMutating.value = true;
    try {
      const updated = await userService.updateUser(userId, payload);
      invalidateCache();
      await fetchUsers({ force: true });
      return updated;
    } finally {
      isMutating.value = false;
    }
  }

  async function resetPassword(userId: string): Promise<ResetPasswordResponse> {
    isMutating.value = true;
    try {
      return await userService.resetPassword(userId);
    } finally {
      isMutating.value = false;
    }
  }

  async function deleteUser(userId: string): Promise<void> {
    isMutating.value = true;
    try {
      await userService.deleteUser(userId);
      invalidateCache();
      await fetchUsers({ force: true });
    } finally {
      isMutating.value = false;
    }
  }

  function roleLabel(role: string): string {
    switch (role) {
      case "creator":
        return "Креатор";
      case "admin":
        return "Админ";
      case "user":
        return "Пользователь";
      default:
        return role;
    }
  }

  function uiRoleToApi(roleLabel: string): AssignableUserRole {
    return roleLabel === "Администратор" ? "admin" : "user";
  }

  function apiRoleToUi(role: string): string {
    return role === "admin" ? "Администратор" : "Пользователь";
  }

  return {
    users,
    pagination,
    total,
    regularUsers,
    adminUsers,
    isLoading,
    isMutating,
    error,
    forbidden,
    currentPage,
    currentLimit,
    isCacheValid,
    fetchUsers,
    invalidateCache,
    registerUser,
    createUser,
    updateUser,
    resetPassword,
    deleteUser,
    roleLabel,
    uiRoleToApi,
    apiRoleToUi,
  };
});
