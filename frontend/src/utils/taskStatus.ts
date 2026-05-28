import type { TaskStatus, TaskUiStatus } from "../api/tasks.types";

/**
 * Все статусы конвейера в порядке прохождения.
 * Соответствует domains/task.go (StatusProcessingUpload → StatusDone).
 */
export const PIPELINE_STATUSES: TaskStatus[] = [
  "processing_upload",
  "pending_denoise",
  "processing_denoise",
  "pending_transcribe",
  "processing_transcribe",
  "pending_diarize",
  "processing_diarize",
  "pending_summarize",
  "processing_summarize",
  "done",
];

const PROCESSING_STATUSES = new Set<TaskStatus>([
  "processing_upload",
  "processing_denoise",
  "processing_transcribe",
  "processing_diarize",
  "processing_summarize",
]);

const PENDING_STATUSES = new Set<TaskStatus>([
  "pending_denoise",
  "pending_transcribe",
  "pending_diarize",
  "pending_summarize",
]);

const ERROR_STATUSES = new Set<TaskStatus>([
  "error_upload",
  "error_denoise",
  "error_transcribe",
  "error_diarize",
  "error_summarize",
]);

export function isProcessing(status: TaskStatus): boolean {
  return PROCESSING_STATUSES.has(status);
}

export function isPending(status: TaskStatus): boolean {
  return PENDING_STATUSES.has(status);
}

export function isError(status: TaskStatus): boolean {
  return ERROR_STATUSES.has(status);
}

export function isDone(status: TaskStatus): boolean {
  return status === "done";
}

/** Сгруппированный статус для UI (карточка задачи, фильтры, переходы) */
export function toUiStatus(status: TaskStatus): TaskUiStatus {
  if (isDone(status)) return "done";
  if (isError(status)) return "error";
  if (isProcessing(status)) return "processing";
  if (isPending(status)) return "pending";
  return "unknown";
}

/** Подпись для бейджа карточки */
export function statusLabel(status: TaskStatus): string {
  switch (toUiStatus(status)) {
    case "done":
      return "Готово";
    case "processing":
      return "В процессе";
    case "pending":
      return "В очереди";
    case "error":
      return "Ошибка";
    default:
      return "Неизвестно";
  }
}

/** Локализованное название текущего этапа (для деталей обработки) */
export function stageLabel(status: TaskStatus): string {
  switch (status) {
    case "processing_upload":
      return "Загрузка файла";
    case "pending_denoise":
      return "Ожидание шумоподавления";
    case "processing_denoise":
      return "Шумоподавление";
    case "pending_transcribe":
      return "Ожидание транскрибации";
    case "processing_transcribe":
      return "Транскрибация";
    case "pending_diarize":
      return "Ожидание диаризации";
    case "processing_diarize":
      return "Диаризация";
    case "pending_summarize":
      return "Ожидание конспектирования";
    case "processing_summarize":
      return "Конспектирование";
    case "done":
      return "Готово";
    default:
      if (isError(status)) return "Ошибка обработки";
      return status;
  }
}

/** "12 апр, 2026" из ISO-строки. Возвращает исходное значение, если не парсится. */
export function formatMeetingDate(iso?: string | null): string {
  if (!iso) return "—";
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return iso;
  return date.toLocaleDateString("ru-RU", {
    day: "numeric",
    month: "short",
    year: "numeric",
  });
}

/** "01:23:45" из секунд */
export function formatHms(seconds: number): string {
  const safe = Math.max(0, Math.floor(seconds || 0));
  const h = Math.floor(safe / 3600);
  const m = Math.floor((safe % 3600) / 60);
  const s = safe % 60;
  return `${String(h).padStart(2, "0")}:${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
}

/** "5 мин" / "1 ч 12 мин" */
export function formatDuration(seconds: number): string {
  const safe = Math.max(0, Math.floor(seconds || 0));
  const h = Math.floor(safe / 3600);
  const m = Math.floor((safe % 3600) / 60);
  if (h > 0) return `${h} ч ${m} мин`;
  return `${m} мин`;
}

/** "~12 мин" из минут */
export function formatMinutes(minutes: number): string {
  const safe = Math.max(0, Math.round(minutes || 0));
  if (safe < 60) return `~${safe} мин`;
  const h = Math.floor(safe / 60);
  const m = safe % 60;
  return m === 0 ? `~${h} ч` : `~${h} ч ${m} мин`;
}
