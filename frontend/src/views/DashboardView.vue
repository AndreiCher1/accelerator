<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { ChevronDown, Plus, Search } from "@lucide/vue";
import Header from "../components/layout/Header.vue";
import TaskCard from "../components/features/TaskCard.vue";
import Pagination from "../components/ui/Pagination.vue";
import Button from "../components/ui/Button.vue";
import Input from "../components/ui/Input.vue";
import Select from "../components/ui/Select.vue";
import Spinner from "../components/ui/Spinner.vue";
import Dropdown from "../components/ui/Dropdown.vue";
import Modal from "../components/ui/Modal.vue";
import FormError from "../components/ui/FormError.vue";
import { extractApiErrorMessage } from "../api";
import { useTasksStore } from "../stores/tasks";
import { useGroupsStore } from "../stores/groups";
import { toUiStatus } from "../utils/taskStatus";
import { useAutoRefresh } from "../composables/useAutoRefresh";

const router = useRouter();
const tasksStore = useTasksStore();
const groupsStore = useGroupsStore();

const searchQuery = ref("");
const selectedDate = ref<"" | "today" | "week" | "month">("");
const selectedStatus = ref<"" | "done" | "processing" | "pending" | "error">("");
const currentPage = ref(1);
const itemsPerPage = 12;

const deleteTaskId = ref<string | null>(null);
const deleteError = ref<string | null>(null);
const showDeleteModal = ref(false);

const dateOptions = [
  { value: "", label: "Дата" },
  { value: "today", label: "За сегодня" },
  { value: "week", label: "За эту неделю" },
  { value: "month", label: "За этот месяц" },
];

const statusOptions = [
  { value: "", label: "Статус" },
  { value: "done", label: "Готово" },
  { value: "processing", label: "В процессе" },
  { value: "pending", label: "В очереди" },
  { value: "error", label: "Ошибка" },
];

function inDateRange(iso: string | undefined): boolean {
  if (!selectedDate.value || !iso) return true;
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return true;

  const now = new Date();
  switch (selectedDate.value) {
    case "today":
      return date.toDateString() === now.toDateString();
    case "week": {
      const weekAgo = new Date(now);
      weekAgo.setDate(now.getDate() - 7);
      return date >= weekAgo && date <= now;
    }
    case "month": {
      const monthAgo = new Date(now);
      monthAgo.setMonth(now.getMonth() - 1);
      return date >= monthAgo && date <= now;
    }
    default:
      return true;
  }
}

const filteredTasks = computed(() => {
  let tasks = tasksStore.tasks;

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase();
    tasks = tasks.filter(
      (task) =>
        (task.task_name || "").toLowerCase().includes(query) ||
        (task.original_filename || "").toLowerCase().includes(query),
    );
  }

  if (selectedStatus.value) {
    tasks = tasks.filter((task) => toUiStatus(task.status) === selectedStatus.value);
  }

  if (selectedDate.value) {
    tasks = tasks.filter((task) => inDateRange(task.meeting_date || task.created_at));
  }

  return tasks;
});

const paginatedTasks = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage;
  return filteredTasks.value.slice(start, start + itemsPerPage);
});

const totalPages = computed(() =>
  Math.max(1, Math.ceil(filteredTasks.value.length / itemsPerPage)),
);

const handlePageChange = (page: number) => {
  currentPage.value = page;
  window.scrollTo({ top: 0, behavior: "smooth" });
};

const handleCreateTask = () => {
  router.push({ name: "RecordCreate" });
};

const loadGroupTasks = async (groupId: string | null | undefined) => {
  if (!groupId) {
    tasksStore.tasks = [];
    return;
  }
  await tasksStore.fetchTasks(groupId, 1, 0);
};

const handleGroupSelect = async (groupId: string) => {
  await groupsStore.selectGroup(groupId);
  await loadGroupTasks(groupId);
};

const requestDelete = (taskId: string) => {
  deleteTaskId.value = taskId;
  deleteError.value = null;
  showDeleteModal.value = true;
};

const confirmDelete = async () => {
  if (!deleteTaskId.value) return;
  deleteError.value = null;
  try {
    await tasksStore.deleteTask(deleteTaskId.value);
    showDeleteModal.value = false;
    deleteTaskId.value = null;
  } catch (err: unknown) {
    deleteError.value = extractApiErrorMessage(err, "Не удалось удалить задачу");
  }
};

watch([searchQuery, selectedDate, selectedStatus], () => {
  currentPage.value = 1;
});

watch(
  () => groupsStore.activeGroupId,
  async (newGroupId) => {
    currentPage.value = 1;
    await loadGroupTasks(newGroupId);
  },
);

onMounted(async () => {
  try {
    await groupsStore.fetchGroups({ force: true });
  } catch {
    // если группы не подгрузились — список будет пустой, ошибка покажется через store.error
  }
  await loadGroupTasks(groupsStore.activeGroupId);
});

useAutoRefresh(async () => {
  try {
    await groupsStore.fetchGroups({ force: true });
  } catch {
    // ok
  }
  await loadGroupTasks(groupsStore.activeGroupId);
});
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-dark">
    <Header max-width="max-w-[1200px]" />

    <main class="mx-auto max-w-[1200px] px-4 py-8 sm:px-6 lg:px-8">
      <!-- Page Header -->
      <div class="mb-6 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <p class="text-sm text-gray-500 dark:text-gray-400">Список задач</p>
          <h1 class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">
            Группа: {{ groupsStore.activeGroup?.name || "Пусто" }}
          </h1>
        </div>

        <Dropdown
          align="right"
          :empty="groupsStore.groups.length === 0"
          empty-text="Нет доступных групп"
        >
          <template #trigger>
            <button
              class="flex items-center gap-2 rounded-lg border border-gray-200 bg-white px-4 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-50 dark:border-dark-border dark:bg-dark-card dark:text-gray-200 dark:hover:bg-dark-elevated"
            >
              {{ groupsStore.activeGroup?.name || "Выберите группу" }}
              <ChevronDown :size="16" class="text-gray-400" />
            </button>
          </template>

          <template #content="{ close }">
            <div class="max-h-64 overflow-y-auto">
              <button
                v-for="group in groupsStore.groups"
                :key="group.group_id"
                class="flex w-full items-center px-4 py-2 text-left text-sm text-gray-700 transition-colors hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-elevated"
                :class="{
                  'bg-blue-50 text-blue-600 dark:bg-white/10 dark:text-white':
                    groupsStore.activeGroupId === group.group_id,
                }"
                @click="
                  () => {
                    handleGroupSelect(group.group_id);
                    close();
                  }
                "
              >
                {{ group.name }}
              </button>
            </div>
          </template>
        </Dropdown>
      </div>

      <!-- Filters -->
      <div class="mb-6 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div class="flex flex-1 flex-col gap-3 sm:flex-row sm:items-end">
          <div class="relative flex-1 sm:max-w-xs">
            <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
              <Search :size="18" class="text-gray-400" />
            </div>
            <Input v-model="searchQuery" placeholder="Найти запись" class="pl-10" />
          </div>

          <Select v-model="selectedDate" :options="dateOptions" class="sm:w-48" />
          <Select v-model="selectedStatus" :options="statusOptions" class="sm:w-48" />
        </div>

        <Button @click="handleCreateTask" :disabled="!groupsStore.activeGroupId">
          <Plus :size="18" />
          Создать запись
        </Button>
      </div>

      <!-- Loading -->
      <div v-if="tasksStore.isLoading" class="flex items-center justify-center py-12">
        <Spinner size="lg" class="text-blue-600 dark:text-white" />
      </div>

      <!-- Error -->
      <div
        v-else-if="tasksStore.error"
        class="rounded-lg border border-red-200 bg-red-50 p-6 text-center dark:border-red-800 dark:bg-red-900/20"
      >
        <p class="text-red-600 dark:text-red-400">{{ tasksStore.error }}</p>
      </div>

      <!-- Empty -->
      <div
        v-else-if="filteredTasks.length === 0"
        class="rounded-lg border border-gray-200 bg-white p-12 text-center dark:border-dark-border dark:bg-dark-card"
      >
        <p class="text-gray-500 dark:text-gray-400">
          {{
            searchQuery || selectedStatus || selectedDate
              ? "Записи не найдены"
              : groupsStore.activeGroupId
                ? "В этой группе пока нет записей"
                : "Выберите группу, чтобы увидеть записи"
          }}
        </p>
      </div>

      <!-- Tasks -->
      <div v-else class="space-y-4">
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-1">
          <TaskCard
            v-for="task in paginatedTasks"
            :key="task.task_id"
            :task="task"
            @delete="requestDelete"
          />
        </div>

        <div v-if="totalPages > 1" class="mt-6">
          <Pagination
            :current-page="currentPage"
            :total-pages="totalPages"
            :total-items="filteredTasks.length"
            :items-per-page="itemsPerPage"
            @page-change="handlePageChange"
          />
        </div>
      </div>
    </main>

    <Modal v-model="showDeleteModal" title="Удалить запись?" size="sm">
      <div class="space-y-4">
        <FormError :message="deleteError" />
        <p class="text-gray-700 dark:text-gray-300">
          Вы уверены, что хотите удалить запись? Это действие нельзя отменить.
        </p>
      </div>

      <template #footer="{ close }">
        <Button variant="outline" @click="close">Отмена</Button>
        <Button @click="confirmDelete" :is-loading="tasksStore.isMutating">Удалить</Button>
      </template>
    </Modal>
  </div>
</template>
