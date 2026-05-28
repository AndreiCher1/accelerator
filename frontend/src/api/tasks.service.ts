import { api } from "./api";
import { cleanPayload } from "./utils";
import { mockTasksService, USE_MOCK_TASKS } from "./tasks.mock";
import type {
  EditTaskRequest,
  TaskDto,
  TaskStatusResponse,
  TasksListResponse,
  UploadTaskData,
  UploadTaskResponse,
} from "./tasks.types";

const TASKS = "/api/v1/tasks";

const realTasksService = {
  /**
   * POST /api/v1/tasks/upload/{groupID}
   * Multipart: сначала JSON-поле `data`, затем файл `audio` — порядок важен для бекенда.
   */
  upload(
    groupId: string,
    audio: File,
    data: UploadTaskData,
  ): Promise<UploadTaskResponse> {
    const formData = new FormData();
    formData.append("data", JSON.stringify(data));
    formData.append("audio", audio);

    return api
      .post<UploadTaskResponse>(`${TASKS}/upload/${groupId}`, formData, {
        headers: { "Content-Type": "multipart/form-data" },
      })
      .then((r) => r.data);
  },

  /**
   * GET /api/v1/tasks/{groupID}/all — список задач в группе с пагинацией.
   * `limit = 0` означает «без пагинации» (бекенд игнорирует limit/offset
   * при limit <= 0). Отрицательные значения слать нельзя — серверная
   * валидация `number` режет их с ответом «Ошибка во входных данных».
   */
  listByGroup(
    groupId: string,
    page = 1,
    limit = 0,
  ): Promise<TasksListResponse> {
    return api
      .get<TasksListResponse>(`${TASKS}/${groupId}/all`, {
        params: { page: String(page), limit: String(limit) },
      })
      .then((r) => r.data);
  },

  /** GET /api/v1/tasks/{taskID} — полная информация о задаче (включая result, если done) */
  getById(taskId: string): Promise<TaskDto> {
    return api.get<TaskDto>(`${TASKS}/${taskId}`).then((r) => r.data);
  },

  /** GET /api/v1/tasks/{taskID}/status — короткий ответ для поллинга прогресса */
  getStatus(taskId: string): Promise<TaskStatusResponse> {
    return api
      .get<TaskStatusResponse>(`${TASKS}/${taskId}/status`)
      .then((r) => r.data);
  },

  /** PUT /api/v1/tasks/{taskID} — изменение метаданных (только если status === done) */
  update(taskId: string, payload: EditTaskRequest): Promise<TaskDto> {
    const body = cleanPayload(payload, { stripEmptyStrings: true });
    return api.put<TaskDto>(`${TASKS}/${taskId}`, body).then((r) => r.data);
  },

  /** DELETE /api/v1/tasks/{taskID} */
  remove(taskId: string): Promise<void> {
    return api.delete<void>(`${TASKS}/${taskId}`).then(() => undefined);
  },
};

/**
 * При VITE_USE_MOCK_TASKS=true все запросы к задачам идут в in-memory мок,
 * что удобно для тестирования вёрстки без бекенда. По умолчанию используется
 * настоящий API.
 */
export const tasksService = USE_MOCK_TASKS ? mockTasksService : realTasksService;

if (USE_MOCK_TASKS) {
  // Заметно в DevTools, чтобы случайно не подумать, что бекенд жив.
  console.info("[tasks] Используется in-memory мок (VITE_USE_MOCK_TASKS=true)");
}
