import { defineStore } from "pinia";
import { computed, ref } from "vue";
import {
  ApiError,
  extractApiErrorMessage,
  tasksService,
  type EditTaskRequest,
  type TaskDto,
  type TaskStatusResponse,
  type UploadTaskData,
  type UploadTaskResponse,
} from "../api";
import { isDone, isError as isErrorStatus } from "../utils/taskStatus";

const STATUS_POLL_INTERVAL_MS = 5000;

interface ListState {
  groupId: string;
  page: number;
  total: number;
}

export const useTasksStore = defineStore("tasks", () => {
  // ---------- список задач (для DashboardView) ----------
  const tasks = ref<TaskDto[]>([]);
  const listState = ref<ListState | null>(null);
  const isLoading = ref(false);
  const error = ref<string | null>(null);

  // ---------- состояние одной задачи (для деталей) ----------
  const currentTask = ref<TaskDto | null>(null);
  const currentStatus = ref<TaskStatusResponse | null>(null);
  const isCurrentLoading = ref(false);
  const currentError = ref<string | null>(null);

  // ---------- мутации ----------
  const isUploading = ref(false);
  const isMutating = ref(false);

  // ---------- поллинг статуса ----------
  let pollTimer: ReturnType<typeof setTimeout> | null = null;
  let pollAbort = false;

  const tasksByGroup = computed(() => {
    if (!listState.value) return tasks.value;
    return tasks.value.filter((t) => t.group_id === listState.value!.groupId);
  });

  function stopPolling(): void {
    pollAbort = true;
    if (pollTimer !== null) {
      clearTimeout(pollTimer);
      pollTimer = null;
    }
  }

  async function fetchTasks(groupId: string, page = 1, limit = 0): Promise<void> {
    isLoading.value = true;
    error.value = null;

    try {
      const data = await tasksService.listByGroup(groupId, page, limit);
      tasks.value = data.tasks ?? [];
      listState.value = {
        groupId,
        page: data.pagination?.page ?? page,
        total: data.pagination?.total ?? tasks.value.length,
      };
    } catch (err: unknown) {
      tasks.value = [];
      if (err instanceof ApiError && err.isNotFound) {
        error.value = "Группа не найдена или недоступна";
      } else if (err instanceof ApiError && err.isForbidden) {
        error.value = "Недостаточно прав для просмотра задач";
      } else {
        error.value = extractApiErrorMessage(err, "Не удалось загрузить список задач");
      }
    } finally {
      isLoading.value = false;
    }
  }

  async function uploadTask(
    groupId: string,
    audio: File,
    data: UploadTaskData,
  ): Promise<UploadTaskResponse> {
    isUploading.value = true;
    error.value = null;

    try {
      const response = await tasksService.upload(groupId, audio, data);

      // Локально вставим заглушку списка, чтобы карточка появилась сразу.
      const placeholder: TaskDto = {
        task_id: response.task_id,
        user_id: "",
        group_id: groupId,
        task_name: response.task_name,
        description: response.description,
        meeting_date: response.meeting_date,
        pattern_id: response.pattern_id,
        status: response.status,
        result: null,
        original_filename: response.original_filename,
        duration_seconds: 0,
        created_at: response.created_at,
        updated_at: response.created_at,
        started_at: "",
        completed_at: "",
        change_flag: response.change_flag,
      };

      const idx = tasks.value.findIndex((t) => t.task_id === response.task_id);
      if (idx >= 0) {
        tasks.value[idx] = { ...tasks.value[idx], ...placeholder };
      } else {
        tasks.value = [placeholder, ...tasks.value];
      }

      return response;
    } catch (err: unknown) {
      error.value = extractApiErrorMessage(err, "Не удалось загрузить запись");
      throw err;
    } finally {
      isUploading.value = false;
    }
  }

  async function fetchTask(taskId: string): Promise<TaskDto> {
    isCurrentLoading.value = true;
    currentError.value = null;

    try {
      const task = await tasksService.getById(taskId);
      currentTask.value = task;
      return task;
    } catch (err: unknown) {
      currentTask.value = null;
      if (err instanceof ApiError && err.isNotFound) {
        currentError.value = "Задача не найдена";
      } else {
        currentError.value = extractApiErrorMessage(err, "Не удалось загрузить задачу");
      }
      throw err;
    } finally {
      isCurrentLoading.value = false;
    }
  }

  async function fetchStatus(taskId: string): Promise<TaskStatusResponse> {
    const status = await tasksService.getStatus(taskId);
    currentStatus.value = status;

    // Синхронизируем статус и в списке, и в текущей задаче (без полной перезагрузки).
    const idx = tasks.value.findIndex((t) => t.task_id === taskId);
    if (idx >= 0 && tasks.value[idx].status !== status.status) {
      tasks.value[idx] = { ...tasks.value[idx], status: status.status };
    }
    if (currentTask.value && currentTask.value.task_id === taskId) {
      currentTask.value = { ...currentTask.value, status: status.status };
    }

    return status;
  }

  /**
   * Поллит статус задачи. Когда придёт `done`, дозагружает полный объект
   * (с результатом) через GET /tasks/{taskID} и вызывает `onDone`.
   * При ошибке/завершении просто перестаёт опрашивать.
   */
  function pollStatus(
    taskId: string,
    options: {
      onUpdate?: (status: TaskStatusResponse) => void;
      onDone?: (task: TaskDto) => void;
      onError?: (err: unknown) => void;
      intervalMs?: number;
    } = {},
  ): () => void {
    stopPolling();
    pollAbort = false;
    const interval = options.intervalMs ?? STATUS_POLL_INTERVAL_MS;

    const tick = async (): Promise<void> => {
      if (pollAbort) return;
      try {
        const status = await fetchStatus(taskId);
        if (pollAbort) return;
        options.onUpdate?.(status);

        if (isDone(status.status)) {
          const task = await fetchTask(taskId);
          if (pollAbort) return;
          options.onDone?.(task);
          return;
        }

        if (isErrorStatus(status.status)) {
          options.onError?.(new Error("Обработка завершилась с ошибкой"));
          return;
        }

        pollTimer = setTimeout(() => void tick(), interval);
      } catch (err) {
        if (pollAbort) return;
        options.onError?.(err);
      }
    };

    void tick();
    return stopPolling;
  }

  async function updateTask(taskId: string, payload: EditTaskRequest): Promise<TaskDto> {
    isMutating.value = true;
    try {
      const updated = await tasksService.update(taskId, payload);

      const idx = tasks.value.findIndex((t) => t.task_id === taskId);
      if (idx >= 0) tasks.value[idx] = updated;
      if (currentTask.value && currentTask.value.task_id === taskId) {
        currentTask.value = updated;
      }
      return updated;
    } finally {
      isMutating.value = false;
    }
  }

  async function deleteTask(taskId: string): Promise<void> {
    isMutating.value = true;
    try {
      await tasksService.remove(taskId);
      tasks.value = tasks.value.filter((t) => t.task_id !== taskId);
      if (currentTask.value && currentTask.value.task_id === taskId) {
        currentTask.value = null;
      }
    } finally {
      isMutating.value = false;
    }
  }

  function reset(): void {
    stopPolling();
    tasks.value = [];
    listState.value = null;
    currentTask.value = null;
    currentStatus.value = null;
    error.value = null;
    currentError.value = null;
  }

  return {
    // state
    tasks,
    tasksByGroup,
    listState,
    isLoading,
    error,
    currentTask,
    currentStatus,
    isCurrentLoading,
    currentError,
    isUploading,
    isMutating,
    // actions
    fetchTasks,
    fetchTask,
    fetchStatus,
    pollStatus,
    stopPolling,
    uploadTask,
    updateTask,
    deleteTask,
    reset,
  };
});
