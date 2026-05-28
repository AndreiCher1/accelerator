/**
 * Типы для работы с задачами (соответствуют DTO бекенда:
 * backend/internal/features/tasks/transport/dto/dto.go).
 */

/** Все возможные статусы из domains/task.go */
export type TaskStatus =
  | "processing_upload"
  | "pending_denoise"
  | "processing_denoise"
  | "pending_transcribe"
  | "processing_transcribe"
  | "pending_diarize"
  | "processing_diarize"
  | "pending_summarize"
  | "processing_summarize"
  | "done"
  | "error_upload"
  | "error_denoise"
  | "error_transcribe"
  | "error_diarize"
  | "error_summarize"
  | (string & {}); // допускаем любые другие, чтобы не падать на новых статусах

/** Сводное состояние для UI */
export type TaskUiStatus = "processing" | "pending" | "done" | "error" | "unknown";

/** Полное представление задачи (ResponseTaskDTO) */
export interface TaskDto {
  task_id: string;
  user_id: string;
  group_id: string;

  task_name: string;
  description: string;
  meeting_date: string;
  pattern_id: string;

  status: TaskStatus;
  /** json.RawMessage на бекенде — структура зависит от воркеров */
  result: unknown;

  original_filename: string;
  duration_seconds: number;

  created_at: string;
  updated_at: string;
  started_at: string;
  completed_at: string;

  change_flag: boolean;
}

/** Тело multipart-поля `data` при загрузке */
export interface UploadTaskData {
  task_name: string;
  description: string;
  meeting_date: string;
  pattern_id: string;
}

/** Ответ на upload (ResponseUploadDTO) */
export interface UploadTaskResponse {
  task_id: string;
  status: TaskStatus;
  task_name: string;
  description: string;
  meeting_date: string;
  pattern_id: string;

  original_filename: string;
  file_type: string;
  created_at: string;

  change_flag: boolean;
}

/** Ответ на /tasks/{taskID}/status (ResponseCheckTaskDTO) */
export interface TaskStatusResponse {
  status: TaskStatus;
  /** true — идёт обработка, false — задача в очереди */
  is_process: boolean;
  /** Сколько задач в очереди перед текущей (только если is_process === false) */
  in_the_queue_before: number;
  /** Примерное оставшееся время процесса в минутах (только если is_process === true) */
  approximate_lead_time_process: number;
}

/** Пагинация в ответе списка (PaginationResponseDTO) */
export interface TasksPagination {
  page: number;
  limit: number;
  total: number;
}

/** Ответ на /tasks/group/{groupID} (AllTasksResponseDTO) */
export interface TasksListResponse {
  tasks: TaskDto[];
  pagination: TasksPagination;
}

/** Тело для PUT /tasks/{taskID} (EditTaskRequestDTO) */
export interface EditTaskRequest {
  task_name?: string;
  description?: string;
  meeting_date?: string;
}
