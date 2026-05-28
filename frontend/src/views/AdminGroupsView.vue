<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useRoute } from "vue-router";
import { Plus, Search, Edit, Trash2, Users } from "@lucide/vue";
import Header from "../components/layout/Header.vue";
import Card from "../components/ui/Card.vue";
import Button from "../components/ui/Button.vue";
import Input from "../components/ui/Input.vue";
import Select from "../components/ui/Select.vue";
import Modal from "../components/ui/Modal.vue";
import Spinner from "../components/ui/Spinner.vue";
import Badge from "../components/ui/Badge.vue";
import FormError from "../components/ui/FormError.vue";
import { useGroupsStore, groupColorClass, groupPrefix } from "../stores/groups";
import { useUsersStore } from "../stores/users";
import { extractApiErrorMessage } from "../api";
import { getInitials } from "../utils/initials.ts"
import { useAutoRefresh } from "../composables/useAutoRefresh";

const route = useRoute();
const groupsStore = useGroupsStore();
const usersStore = useUsersStore();

const searchQuery = ref("");
const showCreateModal = ref(false);
const showEditModal = ref(false);
const showDeleteModal = ref(false);
const showMembersModal = ref(false);
const selectedGroupId = ref<string | null>(null);

const createError = ref<string | null>(null);
const editError = ref<string | null>(null);
const deleteError = ref<string | null>(null);
const membersError = ref<string | null>(null);

const createForm = ref({
  name: "",
  description: "",
  ownerId: "",
});

const editForm = ref({
  name: "",
  description: "",
  ownerId: "",
});

const filteredGroups = computed(() => {
  if (!searchQuery.value) return groupsStore.groups;

  const query = searchQuery.value.toLowerCase();
  return groupsStore.groups.filter(
    (group) =>
      group.name.toLowerCase().includes(query) ||
      group.description.toLowerCase().includes(query),
  );
});

const selectedGroup = computed(() => {
  if (!selectedGroupId.value) return null;
  return groupsStore.groups.find((g) => g.group_id === selectedGroupId.value) || null;
});

const availableUsers = computed(() => {
  return usersStore.users.filter((u) => u.role === "user" || u.role === "admin");
});

const ownerOptions = computed(() =>
  availableUsers.value.map((user) => ({
    value: user.user_id,
    label: user.full_name,
  })),
);

const addMemberSelection = ref("");

const memberAddOptions = computed(() => {
  const members = groupsStore.activeGroupDetails?.members ?? [];
  return availableUsers.value
    .filter((user) => !members.some((member) => member.user_id === user.user_id))
    .map((user) => ({
      value: user.user_id,
      label: user.full_name,
    }));
});

const onAddMemberSelect = async (userId: string) => {
  if (!userId) return;
  await handleAddMember(userId);
  addMemberSelection.value = "";
};

const handleCreateGroup = async () => {
  createError.value = null;

  if (!createForm.value.name.trim()) {
    createError.value = "Укажите название группы";
    return;
  }
  if (!createForm.value.ownerId) {
    createError.value = "Выберите владельца группы";
    return;
  }

  try {
    await groupsStore.createGroup({
      name: createForm.value.name.trim(),
      description: createForm.value.description.trim(),
      owner_id: createForm.value.ownerId,
    });

    showCreateModal.value = false;
    createForm.value = {
      name: "",
      description: "",
      ownerId: "",
    };
  } catch (error) {
    console.error("Failed to create group:", error);
    createError.value = extractApiErrorMessage(error, "Не удалось создать группу");
  }
};

const handleEditGroup = async () => {
  if (!selectedGroupId.value) return;

  editError.value = null;

  if (!editForm.value.name.trim()) {
    editError.value = "Укажите название группы";
    return;
  }
  if (!editForm.value.ownerId) {
    editError.value = "Выберите владельца группы";
    return;
  }

  try {
    await groupsStore.updateGroup(selectedGroupId.value, {
      name: editForm.value.name.trim(),
      description: editForm.value.description.trim(),
      owner_id: editForm.value.ownerId,
    });

    showEditModal.value = false;
    selectedGroupId.value = null;
  } catch (error) {
    console.error("Failed to update group:", error);
    editError.value = extractApiErrorMessage(error, "Не удалось обновить группу");
  }
};

const handleDeleteGroup = async () => {
  if (!selectedGroupId.value) return;

  deleteError.value = null;

  try {
    await groupsStore.deleteGroup(selectedGroupId.value);
    showDeleteModal.value = false;
    selectedGroupId.value = null;
  } catch (error) {
    console.error("Failed to delete group:", error);
    deleteError.value = extractApiErrorMessage(error, "Не удалось удалить группу");
  }
};

const handleAddMember = async (userId: string) => {
  if (!selectedGroupId.value) return;

  membersError.value = null;

  try {
    await groupsStore.addMember(selectedGroupId.value, userId);
  } catch (error) {
    console.error("Failed to add member:", error);
    membersError.value = extractApiErrorMessage(error, "Не удалось добавить участника");
  }
};

const handleRemoveMember = async (userId: string) => {
  if (!selectedGroupId.value) return;

  membersError.value = null;

  try {
    await groupsStore.removeMember(selectedGroupId.value, userId);
  } catch (error) {
    console.error("Failed to remove member:", error);
    membersError.value = extractApiErrorMessage(error, "Не удалось удалить участника");
  }
};

const openEditModal = (groupId: string) => {
  const group = groupsStore.groups.find((g) => g.group_id === groupId);
  if (group) {
    selectedGroupId.value = groupId;
    editError.value = null;
    editForm.value = {
      name: group.name,
      description: group.description,
      ownerId: group.owner_id,
    };
    showEditModal.value = true;
  }
};

const openDeleteModal = (groupId: string) => {
  selectedGroupId.value = groupId;
  deleteError.value = null;
  showDeleteModal.value = true;
};

const openMembersModal = async (groupId: string) => {
  selectedGroupId.value = groupId;
  membersError.value = null;
  await groupsStore.fetchGroupMembers(groupId);
  showMembersModal.value = true;
};

onMounted(async () => {
  await groupsStore.fetchGroups({ force: true });
  await usersStore.fetchUsers({ page: 1, limit: 100, force: true });

  // Если передан query-параметр edit — открываем модалку редактирования
  const editGroupId = route.query.edit as string | undefined;
  if (editGroupId) {
    openEditModal(editGroupId);
  }

  // Если передан query-параметр members — открываем модалку участников
  const membersGroupId = route.query.members as string | undefined;
  if (membersGroupId) {
    await openMembersModal(membersGroupId);
  }
});

useAutoRefresh(async () => {
  await groupsStore.fetchGroups({ force: true });
  await usersStore.fetchUsers({
    page: usersStore.currentPage,
    limit: usersStore.currentLimit,
    force: true,
  });
  if (showMembersModal.value && selectedGroupId.value) {
    await groupsStore.fetchGroupMembers(selectedGroupId.value, { force: true });
  }
});
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-dark">
    <Header max-width="max-w-[1800px]" />

    <main class="mx-auto max-w-[1800px] px-4 py-8 sm:px-6 lg:px-8">
      <!-- Page Header -->
      <div class="mb-6 flex items-center justify-between">
        <div>
          <p class="text-sm text-gray-500 dark:text-gray-400">Управление группами</p>
          <h1 class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">Группы</h1>
        </div>
        <Button @click="createError = null; showCreateModal = true">
          <Plus :size="18" />
          Создать группу
        </Button>
      </div>

      <!-- Search -->
      <div class="mb-6">
        <div class="relative max-w-xs">
          <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
            <Search :size="18" class="text-gray-400" />
          </div>
          <Input v-model="searchQuery" placeholder="Поиск групп..." class="pl-10" />
        </div>
      </div>

      <!-- Groups Grid -->
      <div v-if="groupsStore.isLoading" class="flex items-center justify-center py-12">
        <Spinner size="lg" class="text-blue-600 dark:text-white" />
      </div>

      <div
        v-else-if="filteredGroups.length === 0"
        class="rounded-lg border border-gray-200 bg-white p-12 text-center dark:border-dark-border dark:bg-dark-card"
      >
        <p class="text-gray-500 dark:text-gray-400">
          {{ searchQuery ? "Группы не найдены" : "Нет групп" }}
        </p>
      </div>

      <div v-else class="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
        <Card
          v-for="group in filteredGroups"
          :key="group.group_id"
          padding="lg"
          hover
          class="cursor-pointer"
        >
          <div class="flex items-start justify-between">
            <div class="flex items-start gap-3">
              <div
                :class="[
                  'flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-lg text-lg font-bold',
                  groupColorClass(group.group_id),
                ]"
              >
                {{ groupPrefix(group.name) }}
              </div>
              <div class="flex-1 min-w-0">
                <h3 class="truncate font-semibold text-gray-900 dark:text-white">
                  {{ group.name }}
                </h3>
                <p class="mt-1 text-sm text-gray-500 dark:text-gray-400 line-clamp-2">
                  {{ group.description }}
                </p>
                <div class="mt-3 flex items-center gap-2">
                  <Badge variant="default" size="sm">
                    <Users :size="12" />
                    {{ group.member_count }} участников
                  </Badge>
                </div>
              </div>
            </div>
          </div>

          <div class="mt-4 flex items-center gap-2 border-t border-gray-200 pt-4 dark:border-dark-border">
            <Button
              variant="outline"
              size="sm"
              class="flex-1"
              @click="openMembersModal(group.group_id)"
            >
              <Users :size="16" />
              Участники
            </Button>
            <Button
              variant="outline"
              size="sm"
              @click="openEditModal(group.group_id)"
            >
              <Edit :size="16" />
            </Button>
            <Button
              variant="outline"
              size="sm"
              @click="openDeleteModal(group.group_id)"
            >
              <Trash2 :size="16" />
            </Button>
          </div>
        </Card>
      </div>
    </main>

    <!-- Create Group Modal -->
    <Modal v-model="showCreateModal" title="Создать группу" size="md">
      <form @submit.prevent="handleCreateGroup" class="space-y-4">
        <FormError :message="createError" />
        <div>
          <Input v-model="createForm.name" label="Название" placeholder="Название группы" />
        </div>

        <div>
          <Input v-model="createForm.description" label="Описание" placeholder="Описание группы" />
        </div>

        <div>
          <Select
            v-model="createForm.ownerId"
            label="Владелец"
            :options="ownerOptions"
            placeholder="Выберите владельца"
          />
        </div>
      </form>

      <template #footer="{ close }">
        <Button variant="outline" @click="close">Отмена</Button>
        <Button @click="handleCreateGroup" :is-loading="groupsStore.isMutating">Создать</Button>
      </template>
    </Modal>

    <!-- Edit Group Modal -->
    <Modal v-model="showEditModal" title="Редактировать группу" size="md">
      <form @submit.prevent="handleEditGroup" class="space-y-4">
        <FormError :message="editError" />
        <div>
          <Input v-model="editForm.name" label="Название" />
        </div>

        <div>
          <Input v-model="editForm.description" label="Описание" />
        </div>

        <div>
          <Select
            v-model="editForm.ownerId"
            label="Владелец"
            :options="ownerOptions"
            placeholder="Выберите владельца"
          />
        </div>
      </form>

      <template #footer="{ close }">
        <Button variant="outline" @click="close">Отмена</Button>
        <Button @click="handleEditGroup" :is-loading="groupsStore.isMutating">Сохранить</Button>
      </template>
    </Modal>

    <!-- Delete Group Modal -->
    <Modal v-model="showDeleteModal" title="Удалить группу?" size="sm">
      <div class="space-y-4">
        <FormError :message="deleteError" />
        <p class="text-gray-700 dark:text-gray-300">
          Вы уверены, что хотите удалить эту группу? Это действие нельзя отменить.
        </p>
      </div>

      <template #footer="{ close }">
        <Button variant="outline" @click="close">Отмена</Button>
        <Button @click="handleDeleteGroup" :is-loading="groupsStore.isMutating">Удалить</Button>
      </template>
    </Modal>

    <!-- Members Modal -->
    <Modal v-model="showMembersModal" :title="`Участники: ${selectedGroup?.name}`" size="lg">
      <div v-if="groupsStore.isMembersLoading" class="flex items-center justify-center py-8">
        <Spinner size="md" class="text-blue-600 dark:text-white" />
      </div>

      <div v-else-if="groupsStore.activeGroupDetails" class="space-y-4">
        <FormError :message="membersError" />
        <!-- Add Member -->
        <Select
          :model-value="addMemberSelection"
          label="Добавить участника"
          :options="memberAddOptions"
          placeholder="Выберите пользователя"
          @update:model-value="onAddMemberSelect"
        />

        <!-- Members List -->
        <div class="space-y-2">
          <div
            v-for="member in groupsStore.activeGroupDetails.members"
            :key="member.user_id"
            class="flex items-center justify-between rounded-lg border border-gray-200 p-3 dark:border-dark-border"
          >
            <div class="flex items-center gap-3">
              <div
                class="flex h-10 w-10 items-center justify-center rounded-full bg-blue-100 text-sm font-medium text-blue-600 dark:bg-white/10 dark:text-white"
              >
                {{ getInitials(member.full_name)  }}
              </div>
              <div>
                <p class="font-medium text-gray-900 dark:text-white">
                  {{ member.full_name }}
                </p>
                <p class="text-sm text-gray-500 dark:text-gray-400">
                  {{ member.position }}
                </p>
              </div>
            </div>
            <Button
              variant="outline"
              size="sm"
              @click="handleRemoveMember(member.user_id)"
            >
              <Trash2 :size="16" />
            </Button>
          </div>
        </div>
      </div>

      <template #footer="{ close }">
        <Button @click="close">Закрыть</Button>
      </template>
    </Modal>
  </div>
</template>
