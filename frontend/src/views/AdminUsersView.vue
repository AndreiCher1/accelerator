<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { useRouter } from "vue-router";
import { Plus, Search, Edit, Trash2, RefreshCw, Copy, Check } from "@lucide/vue";
import Header from "../components/layout/Header.vue";
import Card from "../components/ui/Card.vue";
import Button from "../components/ui/Button.vue";
import Input from "../components/ui/Input.vue";
import Select from "../components/ui/Select.vue";
import Modal from "../components/ui/Modal.vue";
import Spinner from "../components/ui/Spinner.vue";
import Pagination from "../components/ui/Pagination.vue";
import Badge from "../components/ui/Badge.vue";
import FormError from "../components/ui/FormError.vue";
import { useUsersStore } from "../stores/users";
import { useGroupsStore, groupPrefix } from "../stores/groups";
import { extractApiErrorMessage, groupService } from "../api";
import { getInitials } from "../utils/initials.ts"
import { useAutoRefresh } from "../composables/useAutoRefresh";

const router = useRouter();
const usersStore = useUsersStore();
const groupsStore = useGroupsStore();

const OPAQUE_GROUP_COLORS = [
  "bg-blue-100 text-blue-600 dark:bg-blue-900 dark:text-blue-300",
  "bg-orange-100 text-orange-600 dark:bg-orange-900 dark:text-orange-300",
  "bg-green-100 text-green-600 dark:bg-green-900 dark:text-green-300",
  "bg-purple-100 text-purple-600 dark:bg-purple-900 dark:text-purple-300",
  "bg-pink-100 text-pink-600 dark:bg-pink-900 dark:text-pink-300",
  "bg-red-100 text-red-600 dark:bg-red-900 dark:text-red-300",
] as const;

function opaqueGroupColor(groupId: string): string {
  let hash = 0;
  for (let i = 0; i < groupId.length; i++) {
    hash = (hash << 5) - hash + groupId.charCodeAt(i);
    hash |= 0;
  }
  return OPAQUE_GROUP_COLORS[Math.abs(hash) % OPAQUE_GROUP_COLORS.length];
}

function navigateToGroup(groupId: string) {
  router.push({ name: "AdminGroups", query: { members: groupId } });
}

// Маппинг userId -> список групп, в которых состоит пользователь
const userGroupsMap = ref<Map<string, { group_id: string; name: string }[]>>(new Map());

function getUserGroups(userId: string): { group_id: string; name: string }[] {
  return userGroupsMap.value.get(userId) ?? [];
}

async function loadUserGroupsMap(force = false) {
  await groupsStore.fetchGroups({ force });
  const groups = groupsStore.groups;

  const map = new Map<string, { group_id: string; name: string }[]>();

  // Загружаем участников каждой группы
  const results = await Promise.allSettled(
    groups.map((g) => groupService.getGroupMembers(g.group_id))
  );

  results.forEach((result, index) => {
    if (result.status === "fulfilled") {
      const groupData = result.value;
      const groupInfo = { group_id: groups[index].group_id, name: groups[index].name };
      for (const member of groupData.members) {
        const existing = map.get(member.user_id) ?? [];
        existing.push(groupInfo);
        map.set(member.user_id, existing);
      }
    }
  });

  userGroupsMap.value = map;
}

const searchQuery = ref("");
const selectedRole = ref("");
const currentPage = ref(1);
const itemsPerPage = 10;

const showCreateModal = ref(false);
const showEditModal = ref(false);
const showDeleteModal = ref(false);
const showResetPasswordModal = ref(false);
const selectedUserId = ref<string | null>(null);
const newPassword = ref<string | null>(null);
const createdUserPassword = ref<string | null>(null);

const createError = ref<string | null>(null);
const editError = ref<string | null>(null);
const deleteError = ref<string | null>(null);
const resetPasswordError = ref<string | null>(null);

const createForm = ref({
  login: "",
  fullName: "",
  position: "",
  role: "user" as "user" | "admin",
});

const editForm = ref({
  login: "",
  fullName: "",
  position: "",
  role: "user" as "user" | "admin",
});

const roleOptions = [
  { value: "", label: "Все роли" },
  { value: "user", label: "Пользователь" },
  { value: "admin", label: "Администратор" },
];

const roleOptionsForForm = [
  { value: "user", label: "Пользователь" },
  { value: "admin", label: "Администратор" },
];

const filteredUsers = computed(() => {
  let users = usersStore.users;

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase();
    users = users.filter(
      (user) =>
        user.full_name.toLowerCase().includes(query) ||
        user.login.toLowerCase().includes(query) ||
        user.position.toLowerCase().includes(query),
    );
  }

  if (selectedRole.value) {
    users = users.filter((user) => user.role === selectedRole.value);
  }

  return users;
});

const paginatedUsers = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage;
  const end = start + itemsPerPage;
  return filteredUsers.value.slice(start, end);
});

const totalPages = computed(() => {
  return Math.ceil(filteredUsers.value.length / itemsPerPage);
});

const getRoleBadgeVariant = (role: string) => {
  switch (role) {
    case "creator":
      return "info";
    case "admin":
      return "warning";
    default:
      return "default";
  }
};

const getRoleLabel = (role: string) => {
  switch (role) {
    case "creator":
      return "Создатель";
    case "admin":
      return "Администратор";
    default:
      return "Пользователь";
  }
};

const handlePageChange = (page: number) => {
  currentPage.value = page;
};

watch([searchQuery, selectedRole], () => {
  currentPage.value = 1;
});

const handleCreateUser = async () => {
  createError.value = null;

  if (!createForm.value.login.trim()) {
    createError.value = "Укажите email";
    return;
  }
  if (!createForm.value.fullName.trim()) {
    createError.value = "Укажите ФИО";
    return;
  }
  if (!createForm.value.position.trim()) {
    createError.value = "Укажите должность";
    return;
  }

  try {
    const response = await usersStore.createUser({
      login: createForm.value.login.trim(),
      full_name: createForm.value.fullName.trim(),
      position: createForm.value.position.trim(),
      role: createForm.value.role,
    });
    
    // Сохраняем сгенерированный пароль для отображения
    createdUserPassword.value = response.password;
  } catch (error) {
    console.error("Failed to create user:", error);
    createError.value = extractApiErrorMessage(error, "Не удалось создать пользователя");
  }
};

const closeCreateModal = () => {
  showCreateModal.value = false;
  createdUserPassword.value = null;
  createError.value = null;
  createForm.value = {
    login: "",
    fullName: "",
    position: "",
    role: "user",
  };
};

const copyPassword = async () => {
  if (createdUserPassword.value) {
    await navigator.clipboard.writeText(createdUserPassword.value);
  }
};

const passwordCopied = ref(false);

const handleCopyPassword = async () => {
  await copyPassword();
  passwordCopied.value = true;
  setTimeout(() => {
    passwordCopied.value = false;
  }, 2000);
};

const handleEditUser = async () => {
  if (!selectedUserId.value) return;

  editError.value = null;

  if (!editForm.value.login.trim()) {
    editError.value = "Укажите email";
    return;
  }
  if (!editForm.value.fullName.trim()) {
    editError.value = "Укажите ФИО";
    return;
  }
  if (!editForm.value.position.trim()) {
    editError.value = "Укажите должность";
    return;
  }

  try {
    await usersStore.updateUser(selectedUserId.value, {
      login: editForm.value.login.trim(),
      full_name: editForm.value.fullName.trim(),
      position: editForm.value.position.trim(),
      role: editForm.value.role,
    });
    
    showEditModal.value = false;
    selectedUserId.value = null;
  } catch (error) {
    console.error("Failed to update user:", error);
    editError.value = extractApiErrorMessage(error, "Не удалось обновить пользователя");
  }
};

const handleDeleteUser = async () => {
  if (!selectedUserId.value) return;

  deleteError.value = null;

  try {
    await usersStore.deleteUser(selectedUserId.value);
    showDeleteModal.value = false;
    selectedUserId.value = null;
  } catch (error) {
    console.error("Failed to delete user:", error);
    deleteError.value = extractApiErrorMessage(error, "Не удалось удалить пользователя");
  }
};

const handleResetPassword = async () => {
  if (!selectedUserId.value) return;

  resetPasswordError.value = null;

  try {
    const response = await usersStore.resetPassword(selectedUserId.value);
    newPassword.value = response.password;
  } catch (error) {
    console.error("Failed to reset password:", error);
    resetPasswordError.value = extractApiErrorMessage(error, "Не удалось сбросить пароль");
  }
};

const openEditModal = (userId: string) => {
  const user = usersStore.users.find((u) => u.user_id === userId);
  if (user) {
    selectedUserId.value = userId;
    editError.value = null;
    editForm.value = {
      login: user.login,
      fullName: user.full_name,
      position: user.position,
      role: user.role === "creator" ? "admin" : (user.role as "user" | "admin"),
    };
    showEditModal.value = true;
  }
};

const openDeleteModal = (userId: string) => {
  selectedUserId.value = userId;
  deleteError.value = null;
  showDeleteModal.value = true;
};

const openResetPasswordModal = (userId: string) => {
  selectedUserId.value = userId;
  newPassword.value = null;
  resetPasswordError.value = null;
  showResetPasswordModal.value = true;
};

onMounted(async () => {
  await usersStore.fetchUsers({ page: 1, limit: 100, force: true });
  await loadUserGroupsMap(true);
});

useAutoRefresh(async () => {
  await usersStore.fetchUsers({
    page: usersStore.currentPage,
    limit: usersStore.currentLimit,
    force: true,
  });
  await loadUserGroupsMap(true);
});
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-dark">
    <Header max-width="max-w-[1800px]" />

    <main class="mx-auto max-w-[1800px] px-4 py-8 sm:px-6 lg:px-8">
      <!-- Page Header -->
      <div class="mb-6 flex items-center justify-between">
        <div>
          <p class="text-sm text-gray-500 dark:text-gray-400">
            Управление пользователями
          </p>
          <h1 class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">
            Пользователи
          </h1>
        </div>
        <Button @click="createError = null; showCreateModal = true">
          <Plus :size="18" />
          Добавить пользователя
        </Button>
      </div>

      <!-- Filters -->
      <div class="mb-6 flex flex-col gap-4 sm:flex-row sm:items-center">
        <div class="relative flex-1 sm:max-w-xs">
          <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
            <Search :size="18" class="text-gray-400" />
          </div>
          <Input v-model="searchQuery" placeholder="Поиск..." class="pl-10" />
        </div>

        <Select v-model="selectedRole" :options="roleOptions" class="sm:w-48" />
      </div>

      <!-- Users Table -->
      <Card padding="none">
        <div v-if="usersStore.isLoading" class="flex items-center justify-center py-12">
          <Spinner size="lg" class="text-blue-600 dark:text-white" />
        </div>

        <div v-else-if="paginatedUsers.length === 0" class="py-12 text-center">
          <p class="text-gray-500 dark:text-gray-400">
            {{ searchQuery || selectedRole ? "Пользователи не найдены" : "Нет пользователей" }}
          </p>
        </div>

        <div v-else>
          <table class="w-full">
            <thead class="border-b border-gray-200 bg-gray-50 dark:border-dark-border dark:bg-dark-elevated">
              <tr>
                <th
                  class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400"
                >
                  Пользователь
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400"
                >
                  Email
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400"
                >
                  Должность
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400"
                >
                  Роль
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400"
                >
                  Группы
                </th>
                <th
                  class="px-6 py-3 text-right text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400"
                >
                  Действия
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200 bg-white dark:divide-gray-700 dark:bg-dark-card">
              <tr
                v-for="user in paginatedUsers"
                :key="user.user_id"
                class="transition-colors hover:bg-gray-50 dark:hover:bg-dark-elevated"
              >
                <td class="whitespace-nowrap px-6 py-4">
                  <div class="flex items-center">
                    <div
                      class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-full bg-blue-100 text-sm font-medium text-blue-600 dark:bg-white/10 dark:text-white"
                    >
                      {{ getInitials(user.full_name) }}
                    </div>
                    <div class="ml-4">
                      <div class="text-sm font-medium text-gray-900 dark:text-white">
                        {{ user.full_name }}
                      </div>
                    </div>
                  </div>
                </td>
                <td class="whitespace-nowrap px-6 py-4 text-sm text-gray-500 dark:text-gray-400">
                  {{ user.login }}
                </td>
                <td class="whitespace-nowrap px-6 py-4 text-sm text-gray-500 dark:text-gray-400">
                  {{ user.position }}
                </td>
                <td class="whitespace-nowrap px-6 py-4">
                  <Badge :variant="getRoleBadgeVariant(user.role)" size="sm">
                    {{ getRoleLabel(user.role) }}
                  </Badge>
                </td>
                <td class="px-6 py-4">
                  <div class="flex items-center -space-x-2">
                    <template v-if="getUserGroups(user.user_id).length === 0">
                      <span class="text-sm text-gray-400 dark:text-gray-500">—</span>
                    </template>
                    <template v-else>
                      <div
                        v-for="group in getUserGroups(user.user_id).slice(0, 2)"
                        :key="group.group_id"
                        class="group/avatar relative cursor-pointer"
                        @click="navigateToGroup(group.group_id)"
                      >
                        <div
                          :class="[
                            'flex h-8 w-8 items-center justify-center rounded-full border-2 border-white text-xs font-medium dark:border-dark-card',
                            opaqueGroupColor(group.group_id),
                          ]"
                        >
                          {{ groupPrefix(group.name) }}
                        </div>
                        <div
                          class="pointer-events-none absolute bottom-full left-1/2 z-50 mb-2 -translate-x-1/2 whitespace-nowrap rounded-md bg-white px-2 py-1 text-xs text-gray-900 opacity-0 shadow-lg ring-1 ring-gray-200 transition-opacity group-hover/avatar:opacity-100 dark:bg-dark-elevated dark:text-gray-100 dark:ring-dark-border"
                        >
                          {{ group.name }}
                        </div>
                      </div>
                      <div
                        v-if="getUserGroups(user.user_id).length > 2"
                        class="group/overflow relative"
                      >
                        <div
                          class="flex h-8 w-8 items-center justify-center rounded-full border-2 border-white bg-gray-200 text-xs font-medium text-gray-600 dark:border-dark-card dark:bg-gray-700 dark:text-gray-200"
                        >
                          +{{ getUserGroups(user.user_id).length - 2 }}
                        </div>
                        <div
                          class="invisible absolute bottom-full left-1/2 z-50 mb-2 -translate-x-1/2 rounded-lg bg-white p-2 opacity-0 shadow-lg ring-1 ring-gray-200 transition-all group-hover/overflow:visible group-hover/overflow:opacity-100 dark:bg-dark-elevated dark:ring-dark-border"
                        >
                          <div
                            v-for="group in getUserGroups(user.user_id).slice(2)"
                            :key="group.group_id"
                            class="flex cursor-pointer items-center gap-2 whitespace-nowrap rounded-md px-2 py-1.5 transition-colors hover:bg-gray-100 dark:hover:bg-dark-card"
                            @click.stop="navigateToGroup(group.group_id)"
                          >
                            <div
                              :class="[
                                'flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full text-[10px] font-medium',
                                opaqueGroupColor(group.group_id),
                              ]"
                            >
                              {{ groupPrefix(group.name) }}
                            </div>
                            <span class="text-xs text-gray-700 dark:text-gray-200">{{ group.name }}</span>
                          </div>
                        </div>
                      </div>
                    </template>
                  </div>
                </td>
                <td class="whitespace-nowrap px-6 py-4 text-right text-sm font-medium">
                  <div class="flex items-center justify-end gap-2">
                    <button
                      class="text-blue-600 transition-colors hover:text-blue-900 dark:text-gray-400 dark:hover:text-white"
                      @click="openEditModal(user.user_id)"
                    >
                      <Edit :size="18" />
                    </button>
                    <button
                      class="text-orange-600 transition-colors hover:text-orange-900 dark:text-orange-400 dark:hover:text-orange-300"
                      @click="openResetPasswordModal(user.user_id)"
                    >
                      <RefreshCw :size="18" />
                    </button>
                    <button
                      class="text-red-600 transition-colors hover:text-red-900 dark:text-red-400 dark:hover:text-red-300"
                      @click="openDeleteModal(user.user_id)"
                    >
                      <Trash2 :size="18" />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination -->
        <div v-if="totalPages > 1" class="border-t border-gray-200 p-4 dark:border-dark-border">
          <Pagination
            :current-page="currentPage"
            :total-pages="totalPages"
            :total-items="filteredUsers.length"
            :items-per-page="itemsPerPage"
            @page-change="handlePageChange"
          />
        </div>
      </Card>
    </main>

    <!-- Create User Modal -->
    <Modal v-model="showCreateModal" title="Добавить пользователя" size="md" :closeOnClickOutside="false">
      <form class="space-y-4">
        <FormError :message="createError" />
        <Input v-model="createForm.login" label="Email" type="email" placeholder="user@example.com" />
        <Input v-model="createForm.fullName" label="ФИО" placeholder="Иванов Иван Иванович" />
        <Input v-model="createForm.position" label="Должность" placeholder="Менеджер" />
        <Select v-model="createForm.role" label="Роль" :options="roleOptionsForForm" />
      </form>
      
      <div v-if="createdUserPassword" class="mt-4 space-y-4 rounded-lg border border-green-200 bg-green-50 p-4 dark:border-green-700 dark:bg-green-900/20">
        <p class="text-gray-700 dark:text-gray-300">
          Пользователь успешно создан. Временный пароль сгенерирован автоматически.
        </p>
        <p class="text-sm text-gray-500 dark:text-gray-400">
          Скопируйте пароль и передайте пользователю. После первого входа пользователь должен будет сменить пароль.
        </p>
        <div class="flex items-center gap-2">
          <div
            class="flex-1 rounded-lg border border-gray-200 bg-gray-50 p-4 font-mono text-sm dark:border-dark-border dark:bg-dark-elevated"
          >
            {{ createdUserPassword }}
          </div>
          <button
            @click="handleCopyPassword"
            class="flex h-11 w-11 items-center justify-center rounded-lg border border-gray-200 text-gray-600 transition-colors hover:bg-gray-100 dark:border-dark-border dark:text-gray-400 dark:hover:bg-dark-elevated"
            :title="passwordCopied ? 'Скопировано' : 'Копировать'"
          >
            <Check v-if="passwordCopied" :size="18" class="text-green-500" />
            <Copy v-else :size="18" />
          </button>
        </div>
      </div>

      <template #footer>
        <Button variant="outline" @click="closeCreateModal">Отмена</Button>
        <Button @click="handleCreateUser" :is-loading="usersStore.isMutating" :disabled="!!createdUserPassword">
          {{ createdUserPassword ? 'Создано' : 'Создать' }}
        </Button>
      </template>
    </Modal>

    <!-- Edit User Modal -->
    <Modal v-model="showEditModal" title="Редактировать пользователя" size="md">
      <form @submit.prevent="handleEditUser" class="space-y-4">
        <FormError :message="editError" />
        <Input v-model="editForm.login" label="Email" type="email" />
        <Input v-model="editForm.fullName" label="ФИО" />
        <Input v-model="editForm.position" label="Должность" />
        <Select v-model="editForm.role" label="Роль" :options="roleOptionsForForm" />
      </form>

      <template #footer="{ close }">
        <Button variant="outline" @click="close">Отмена</Button>
        <Button @click="handleEditUser" :is-loading="usersStore.isMutating">Сохранить</Button>
      </template>
    </Modal>

    <!-- Delete User Modal -->
    <Modal v-model="showDeleteModal" title="Удалить пользователя?" size="sm">
      <div class="space-y-4">
        <FormError :message="deleteError" />
        <p class="text-gray-700 dark:text-gray-300">
          Вы уверены, что хотите удалить этого пользователя? Это действие нельзя отменить.
        </p>
      </div>

      <template #footer="{ close }">
        <Button variant="outline" @click="close">Отмена</Button>
        <Button @click="handleDeleteUser" :is-loading="usersStore.isMutating">Удалить</Button>
      </template>
    </Modal>

    <!-- Reset Password Modal -->
    <Modal v-model="showResetPasswordModal" title="Сброс пароля" size="sm">
      <div v-if="!newPassword" class="space-y-4">
        <FormError :message="resetPasswordError" />
        <p class="text-gray-700 dark:text-gray-300">
          Вы уверены, что хотите сбросить пароль для этого пользователя?
        </p>
      </div>
      <div v-else class="space-y-4">
        <p class="text-gray-700 dark:text-gray-300">
          Новый пароль сгенерирован. Скопируйте его и передайте пользователю.
        </p>
        <div
          class="rounded-lg border border-gray-200 bg-gray-50 p-4 font-mono text-sm dark:border-dark-border dark:bg-dark-elevated"
        >
          {{ newPassword }}
        </div>
      </div>

      <template #footer="{ close }">
        <Button v-if="!newPassword" variant="outline" @click="close">Отмена</Button>
        <Button v-if="!newPassword" @click="handleResetPassword" :is-loading="usersStore.isMutating">Сбросить</Button>
        <Button v-else @click="close">Закрыть</Button>
      </template>
    </Modal>
  </div>
</template>
