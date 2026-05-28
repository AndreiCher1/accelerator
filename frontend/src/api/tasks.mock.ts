/**
 * Тестовый «бекенд» для задач: реализует тот же контракт, что и `tasksService`,
 * но работает полностью в памяти браузера. Полезен для прототипирования вёрстки
 * без поднятого Go-сервиса.
 *
 * Включение: VITE_USE_MOCK_TASKS=true в .env (требуется перезапуск Vite).
 *
 * Поведение прогрессии: каждый вызов getStatus() при включённом моке
 * автоматически продвигает задачу по конвейеру, как только истечёт небольшой
 * таймер. Так UI обработки/очереди легко проверить.
 */

import { ApiError } from "./api.types";
import type {
  EditTaskRequest,
  TaskDto,
  TaskStatus,
  TaskStatusResponse,
  TasksListResponse,
  UploadTaskData,
  UploadTaskResponse,
} from "./tasks.types";

export const USE_MOCK_TASKS =
  String(import.meta.env.VITE_USE_MOCK_TASKS ?? "").toLowerCase() === "true";

// ----- параметры эмуляции прогресса (в миллисекундах) -----
const PENDING_DURATION_MS = 4_000;
const PROCESSING_DURATION_MS = 12_000;

// ----- порядок прохождения этапов конвейера -----
const PIPELINE: TaskStatus[] = [
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

interface MockTask extends TaskDto {
  /** Когда задача перешла в текущий статус (для эмуляции прогрессии) */
  _stageStartedAt: number;
}

const STORE: MockTask[] = [];

// ============================== ТЕСТОВЫЕ ДАННЫЕ ==============================

const MOCK_SUMMARY = `# I. Хронологический обзор обсуждения

**1. Вступительное слово и ретроспектива прошлого квартала.** Собрание открыл [Должность, ФИО], представив детальный отчёт по итогам завершённого квартала. Были озвучены итоговые цифры по выручке, процент выполнения плана и ключевые показатели конверсии. Участники обсудили основные успехи (закрытие крупных сделок, выход на новые рынки) и проанализировали причины отставания по отдельным направлениям.

**2. Презентация глобальной стратегии на новый квартал.** Руководство транслировало вектор развития компании на предстоящие три месяца. Основной упор в обсуждении был сделан на том, как глобальные бизнес-цели (например, увеличение доли рынка, запуск нового продукта) напрямую зависят от эффективности отдела продаж.

**3. Сессия формирования OKR (Objectives and Key Results).** Команда перешла к разработке квартальных целей (Objectives) и ключевых результатов (Key Results). Обсуждение велось в формате мозгового штурма с последующей фильтрацией идей. Были предложены и скорректированы приоритетные направления: фокус на LTV текущих клиентов, снижение цикла сделки и увеличение среднего чека.

**4. Распределение ресурсов и тактическое планирование.** Участники собрания обсудили инструменты, необходимые для достижения поставленных целей. Были подняты вопросы нехватки маркетинговых материалов по новым продуктам и необходимости обновления скриптов продаж. Согласован формат еженедельных статус-встреч для трекинга прогресса.

# II. Принятые решения

1. **Утвердить итоги прошлого квартала**: признать работу отдела (удовлетворительной / успешной), зафиксировать процент выполнения общего плана продаж на уровне **[X]%**.
2. **Утвердить командные OKR на следующий квартал:**
   - Цель *(Objective)* 1: **Увеличить проникновение в Enterprise-сегмент.**
     - Ключевые результаты *(KR)*: **[Метрика 1], [Метрика 2]**.
   - Цель *(Objective)* 2: **Повысить эффективность обработки входящих лидов.**
     - Ключевые результаты *(KR)*: **[Метрика 1], [Метрика 2]**.
3. **Распределить зоны ответственности**: назначить руководителей групп [ФИО] ответственными за декомпозицию командных OKR до уровня индивидуальных планов сотрудников в срок до **[Дата]**.
4. **Обновить инструментарий**: поручить отделу обучения/маркетинга актуализировать скрипты продаж и презентационные материалы к **[Дата]**.
5. **Внедрить систему контроля**: утвердить проведение коротких (15 минут) еженедельных синхронизаций каждую **[День недели]** для мониторинга движения по метрикам OKR.`;

const MOCK_TRANSCRIPT = [
  {
    speaker: "СПИКЕР №1",
    timestamp: "00:00:00",
    text: "Так, коллеги, добрый день всем. Начинаем наш квартальный синхрон. Сегодня обсуждаем планирование на третий квартал, наши цели и OKR. Сначала кратко по итогам второго. Мы закрыли его на два процента… ну, чуть-чуть не дотянули до таргета. Э-э, Дмитрий, выведи, пожалуйста, данные по воронке на экран.",
  },
  {
    speaker: "СПИКЕР №2",
    timestamp: "00:01:57",
    text: "Так, коллеги, добрый день всем. Начинаем наш квартальный синхрон. Сегодня обсуждаем планирование на третий квартал, наши цели и OKR. Сначала кратко по итогам второго. Мы закрыли его на два процента… ну, чуть-чуть не дотянули до таргета.",
  },
  {
    speaker: "СПИКЕР №3",
    timestamp: "00:02:15",
    text: "Так, коллеги, добрый день всем. Начинаем наш квартальный синхрон. Сегодня обсуждаем планирование на третий квартал, наши цели и OKR. Сначала кратко по итогам второго. Мы закрыли его на два процента… ну, чуть-чуть не дотянули до таргета.",
  },
  {
    speaker: "СПИКЕР №1",
    timestamp: "00:03:02",
    text: "Так, коллеги, добрый день всем. Начинаем наш квартальный синхрон. Сегодня обсуждаем планирование на третий квартал, наши цели и OKR. Сначала кратко по итогам второго. Мы закрыли его на два процента… ну, чуть-чуть не дотянули до таргета.",
  },
];

function nowMinusDays(days: number): string {
  const d = new Date();
  d.setDate(d.getDate() - days);
  return d.toISOString();
}

function uuidv4(): string {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return crypto.randomUUID();
  }
  return "mock-" + Math.random().toString(36).slice(2, 10);
}

function emptyDate(): string {
  return "0001-01-01T00:00:00Z";
}

function makeTask(overrides: Partial<MockTask> & { task_id: string; status: TaskStatus }): MockTask {
  const now = Date.now();
  const created = overrides.created_at ?? nowMinusDays(2);
  return {
    user_id: "mock-user",
    group_id: "mock-group",
    task_name: "Без названия",
    description: "",
    meeting_date: created,
    pattern_id: "mock-pattern",
    result: null,
    original_filename: "record.mp3",
    duration_seconds: 0,
    created_at: created,
    updated_at: created,
    started_at: emptyDate(),
    completed_at: emptyDate(),
    change_flag: false,
    _stageStartedAt: now,
    ...overrides,
  } as MockTask;
}

function buildInitialTasks(): MockTask[] {
  const now = Date.now();
  return [
    // ------- Готовые задачи с разными результатами -------
    makeTask({
      task_id: "mock-done-1",
      status: "done",
      task_name: "Квартальное планирование: цели и OKR отдела продаж",
      description: "Запись еженедельного совещания отдела продаж.",
      original_filename: "sales_q2.mp3",
      meeting_date: nowMinusDays(1),
      created_at: nowMinusDays(1),
      duration_seconds: 1860,
      completed_at: new Date(now - 60 * 60 * 1000).toISOString(),
      change_flag: true,
      result: {
        summary: MOCK_SUMMARY,
        transcript: MOCK_TRANSCRIPT,
      },
    }),
    makeTask({
      task_id: "mock-done-2",
      status: "done",
      task_name: "Дизайн-ревью нового онбординга",
      description: "Обсуждение прототипов с продуктовой командой.",
      original_filename: "onboarding_review.wav",
      meeting_date: nowMinusDays(3),
      created_at: nowMinusDays(3),
      duration_seconds: 2700,
      change_flag: true,
      result: {
        summary:
          "Команда согласовала направление онбординга: добавим интерактивный туториал и упростим первый шаг регистрации. Открытый вопрос — формат welcome-сообщения.",
      },
    }),
    makeTask({
      task_id: "mock-done-3",
      status: "done",
      task_name: "Интервью с кандидатом — Senior Backend",
      description: "Технический скрин по Go и распределённым системам.",
      original_filename: "interview_senior_backend.m4a",
      meeting_date: nowMinusDays(5),
      created_at: nowMinusDays(5),
      duration_seconds: 3600,
      change_flag: true,
      result: {
        transcript: MOCK_TRANSCRIPT.slice(0, 2),
      },
    }),
    makeTask({
      task_id: "mock-done-4",
      status: "done",
      task_name: "Ретроспектива спринта №14",
      description: "Что зашло, что пошло не так, что улучшить.",
      original_filename: "retro_sprint_14.ogg",
      meeting_date: nowMinusDays(7),
      created_at: nowMinusDays(7),
      duration_seconds: 1500,
      change_flag: true,
      result: {},
    }),
    makeTask({
      task_id: "mock-done-5",
      status: "done",
      task_name: "Обзор инцидента — кластер БД",
      description: "Postmortem по сбою 18 числа.",
      original_filename: "postmortem_db.flac",
      meeting_date: nowMinusDays(10),
      created_at: nowMinusDays(10),
      duration_seconds: 2200,
      change_flag: true,
      result: { summary: MOCK_SUMMARY },
    }),

    // ------- Идёт обработка -------
    makeTask({
      task_id: "mock-proc-1",
      status: "processing_transcribe",
      task_name: "Звонок с клиентом — продление контракта",
      description: "Переговоры о продлении на 12 месяцев.",
      original_filename: "client_call.mp3",
      duration_seconds: 1800,
      created_at: new Date(now - 5 * 60 * 1000).toISOString(),
      _stageStartedAt: now - 8_000,
    }),
    makeTask({
      task_id: "mock-proc-2",
      status: "processing_summarize",
      task_name: "Ежемесячный all-hands",
      description: "Итоги месяца и планы на следующий.",
      original_filename: "all_hands.wav",
      duration_seconds: 4200,
      created_at: new Date(now - 12 * 60 * 1000).toISOString(),
      _stageStartedAt: now - 4_000,
    }),

    // ------- Ожидание в очереди -------
    makeTask({
      task_id: "mock-pend-1",
      status: "pending_transcribe",
      task_name: "Брифинг по новому проекту",
      description: "Стартовое обсуждение скоупа и сроков.",
      original_filename: "brief.aac",
      duration_seconds: 2400,
      created_at: new Date(now - 30 * 60 * 1000).toISOString(),
      _stageStartedAt: now - 1_500,
    }),
    makeTask({
      task_id: "mock-pend-2",
      status: "pending_summarize",
      task_name: "Звонок поддержки — эскалация",
      description: "Разбор инцидента по тикету #4821.",
      original_filename: "support_call.ogg",
      duration_seconds: 900,
      created_at: new Date(now - 45 * 60 * 1000).toISOString(),
      _stageStartedAt: now - 1_500,
    }),

    // ------- Ошибка -------
    makeTask({
      task_id: "mock-error-1",
      status: "error_transcribe",
      task_name: "Тестовая запись — повреждённое аудио",
      description: "Демонстрация состояния ошибки.",
      original_filename: "broken.mp3",
      duration_seconds: 0,
      created_at: nowMinusDays(2),
    }),
  ];
}

function ensureSeeded(): void {
  if (STORE.length === 0) {
    STORE.push(...buildInitialTasks());
  }
}

// ============================== ПРОГРЕССИЯ ==============================

function advanceIfDue(task: MockTask): void {
  const idx = PIPELINE.indexOf(task.status);
  if (idx === -1 || task.status === "done") return;

  const isProc = task.status.startsWith("processing_");
  const threshold = isProc ? PROCESSING_DURATION_MS : PENDING_DURATION_MS;
  const elapsed = Date.now() - task._stageStartedAt;

  if (elapsed < threshold) return;

  const nextStatus = PIPELINE[idx + 1];
  task.status = nextStatus;
  task.updated_at = new Date().toISOString();
  task._stageStartedAt = Date.now();

  if (nextStatus === "done") {
    task.completed_at = task.updated_at;
    task.change_flag = true;
    if (!task.result || (typeof task.result === "object" && task.result !== null && Object.keys(task.result).length === 0)) {
      task.result = { summary: MOCK_SUMMARY, transcript: MOCK_TRANSCRIPT };
    }
  }
}

function tickAll(): void {
  for (const task of STORE) {
    advanceIfDue(task);
  }
}

function findById(taskId: string): MockTask | undefined {
  return STORE.find((t) => t.task_id === taskId);
}

function toDto(task: MockTask): TaskDto {
  // Возвращаем копию без приватного поля _stageStartedAt.
  const { _stageStartedAt: _omit, ...dto } = task;
  void _omit;
  return dto;
}

function delay<T>(value: T, ms = 200): Promise<T> {
  return new Promise((resolve) => setTimeout(() => resolve(value), ms));
}

// ============================== ПУБЛИЧНЫЙ СЕРВИС ==============================

export const mockTasksService = {
  upload(
    groupId: string,
    audio: File,
    data: UploadTaskData,
  ): Promise<UploadTaskResponse> {
    ensureSeeded();
    tickAll();

    const taskId = uuidv4();
    const createdAt = new Date().toISOString();
    const task = makeTask({
      task_id: taskId,
      group_id: groupId,
      user_id: "mock-user",
      task_name: data.task_name,
      description: data.description,
      meeting_date: data.meeting_date,
      pattern_id: data.pattern_id,
      original_filename: audio.name,
      duration_seconds: 0,
      status: "processing_upload",
      created_at: createdAt,
      updated_at: createdAt,
      change_flag: false,
    });
    STORE.unshift(task);

    const response: UploadTaskResponse = {
      task_id: task.task_id,
      status: task.status,
      task_name: task.task_name,
      description: task.description,
      meeting_date: task.meeting_date,
      pattern_id: task.pattern_id,
      original_filename: task.original_filename,
      file_type: audio.type || "audio/mpeg",
      created_at: task.created_at,
      change_flag: task.change_flag,
    };
    return delay(response);
  },

  listByGroup(
    groupId: string,
    page = 1,
    limit = 0,
  ): Promise<TasksListResponse> {
    ensureSeeded();
    tickAll();

    // Подменяем group_id на запрошенный, чтобы тестовые задачи появлялись
    // в любой реальной группе пользователя.
    const all = STORE.map((t) => toDto({ ...t, group_id: groupId }));
    const total = all.length;
    let slice = all;
    if (limit > 0) {
      const start = (page - 1) * limit;
      slice = all.slice(start, start + limit);
    }

    return delay({
      tasks: slice,
      pagination: { page, limit, total },
    });
  },

  getById(taskId: string): Promise<TaskDto> {
    ensureSeeded();
    tickAll();
    const task = findById(taskId);
    if (!task) {
      return Promise.reject(new ApiError("Задача не найдена", 404));
    }
    return delay(toDto(task));
  },

  getStatus(taskId: string): Promise<TaskStatusResponse> {
    ensureSeeded();
    tickAll();
    const task = findById(taskId);
    if (!task) {
      return Promise.reject(new ApiError("Задача не найдена", 404));
    }

    const isProc = task.status.startsWith("processing_");
    const isPending = task.status.startsWith("pending_");

    let inQueue = 0;
    let leadTime = 0;

    if (isPending) {
      inQueue = STORE.filter(
        (t) => t.status === task.status && t.task_id !== task.task_id,
      ).length;
      // Плюс немного «синтетических» — чтобы было что показать.
      inQueue = Math.max(inQueue, 2);
    } else if (isProc) {
      const remainingMs = Math.max(
        0,
        PROCESSING_DURATION_MS - (Date.now() - task._stageStartedAt),
      );
      leadTime = Math.max(1, Math.round(remainingMs / 60_000));
      // Делаем числа выглядящими реалистичнее: показываем целые минуты.
      if (leadTime < 1) leadTime = 1;
    }

    return delay({
      status: task.status,
      is_process: isProc,
      in_the_queue_before: inQueue,
      approximate_lead_time_process: leadTime,
    });
  },

  update(taskId: string, payload: EditTaskRequest): Promise<TaskDto> {
    ensureSeeded();
    const task = findById(taskId);
    if (!task) {
      return Promise.reject(new ApiError("Задача не найдена", 404));
    }
    if (task.status !== "done") {
      return Promise.reject(
        new ApiError("Нельзя изменять задачу, которая находится в обработке", 400),
      );
    }
    if (payload.task_name !== undefined) task.task_name = payload.task_name;
    if (payload.description !== undefined) task.description = payload.description;
    if (payload.meeting_date !== undefined) task.meeting_date = payload.meeting_date;
    task.updated_at = new Date().toISOString();
    return delay(toDto(task));
  },

  remove(taskId: string): Promise<void> {
    ensureSeeded();
    const idx = STORE.findIndex((t) => t.task_id === taskId);
    if (idx === -1) {
      return Promise.reject(new ApiError("Задача не найдена", 404));
    }
    if (STORE[idx].status !== "done" && !STORE[idx].status.startsWith("error_")) {
      return Promise.reject(
        new ApiError("Нельзя удалить задачу, которая находится в обработке", 400),
      );
    }
    STORE.splice(idx, 1);
    return delay(undefined);
  },
};

export type TasksServiceLike = typeof mockTasksService;
