<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ArrowLeft, Clock, Download, Edit, Pause, Play, Trash2 } from "@lucide/vue";
import Header from "../components/layout/Header.vue";
import Card from "../components/ui/Card.vue";
import Button from "../components/ui/Button.vue";
import Input from "../components/ui/Input.vue";
import Textarea from "../components/ui/Textarea.vue";
import Modal from "../components/ui/Modal.vue";
import Spinner from "../components/ui/Spinner.vue";
import FormError from "../components/ui/FormError.vue";
import DatePicker from "../components/ui/DatePicker.vue";
import { extractApiErrorMessage, patternsService } from "../api";
import { useTasksStore } from "../stores/tasks";
import { downloadTextFile } from "../utils/download";
import {
  formatDuration,
  formatHms,
  formatMeetingDate,
  isDone,
  isError as isErrorStatus,
} from "../utils/taskStatus";
import { renderMarkdown } from "../utils/markdown";
import { speakerPillClass } from "../utils/speakerColor";
import {
  isResultEmpty,
  normalizeTaskResult,
  type NormalizedTaskResult,
} from "../utils/taskResult";

const route = useRoute();
const router = useRouter();
const tasksStore = useTasksStore();

const isLoading = ref(true);
const error = ref<string | null>(null);

const showEditModal = ref(false);
const showDeleteModal = ref(false);
const editError = ref<string | null>(null);
const deleteError = ref<string | null>(null);

const editForm = ref({
  taskName: "",
  description: "",
  meetingDate: "",
});

const patternName = ref<string | null>(null);

const task = computed(() => tasksStore.currentTask);

const normalizedResult = computed<NormalizedTaskResult>(() =>
  normalizeTaskResult(task.value?.result),
);

const hasResult = computed(() => !isResultEmpty(task.value?.result));

const renderedSummary = computed(() => renderMarkdown(normalizedResult.value.summary));

const speakerCount = computed(() => {
  const set = new Set(normalizedResult.value.transcript.map((e) => e.speaker));
  return set.size;
});

const formattedMeetingDate = computed(() => formatMeetingDate(task.value?.meeting_date));
const formattedDurationLabel = computed(() =>
  task.value && task.value.duration_seconds > 0
    ? formatDuration(task.value.duration_seconds)
    : null,
);
const speakersLabel = computed(() => {
  const n = speakerCount.value;
  if (!n) return null;
  const word = n === 1 ? "спикер" : n >= 2 && n <= 4 ? "спикера" : "спикеров";
  return `${n} ${word}`;
});

// ---- эмулируемый плеер: бекенд пока не отдаёт URL аудио ----
const isPlaying = ref(false);
const currentTime = ref(0);
let playbackTimer: ReturnType<typeof setInterval> | null = null;

const totalSeconds = computed(() => task.value?.duration_seconds ?? 0);
const formattedCurrent = computed(() => formatHms(currentTime.value));
const formattedTotal = computed(() => formatHms(totalSeconds.value));
const progressPercent = computed(() => {
  if (totalSeconds.value <= 0) return 0;
  return Math.min(100, (currentTime.value / totalSeconds.value) * 100);
});

const togglePlayback = () => {
  if (totalSeconds.value <= 0) return;
  isPlaying.value = !isPlaying.value;
  if (isPlaying.value) {
    if (playbackTimer !== null) clearInterval(playbackTimer);
    playbackTimer = setInterval(() => {
      currentTime.value = Math.min(totalSeconds.value, currentTime.value + 1);
      if (currentTime.value >= totalSeconds.value) {
        isPlaying.value = false;
        if (playbackTimer !== null) {
          clearInterval(playbackTimer);
          playbackTimer = null;
        }
      }
    }, 1000);
  } else if (playbackTimer !== null) {
    clearInterval(playbackTimer);
    playbackTimer = null;
  }
};

const handleScrub = (event: MouseEvent) => {
  if (totalSeconds.value <= 0) return;
  const target = event.currentTarget as HTMLDivElement;
  const rect = target.getBoundingClientRect();
  const ratio = Math.min(1, Math.max(0, (event.clientX - rect.left) / rect.width));
  currentTime.value = Math.round(ratio * totalSeconds.value);
};

const safeFileBase = (name: string | undefined, fallback: string) => {
  const base = (name ?? "").replace(/[<>:"/\\|?*]/g, "").trim();
  return base.length > 0 ? base : fallback;
};

const handleDownloadSummary = () => {
  if (!task.value || !normalizedResult.value.summary) return;
  const filename = `${safeFileBase(task.value.task_name, "konspekt")}.md`;
  downloadTextFile(filename, normalizedResult.value.summary, "text/markdown;charset=utf-8");
};

const handleDownloadTranscript = () => {
  if (!task.value || normalizedResult.value.transcript.length === 0) return;
  const text = normalizedResult.value.transcript
    .map((e) => `[${e.timestamp}] ${e.speaker}\n${e.text}`)
    .join("\n\n");
  const filename = `${safeFileBase(task.value.task_name, "stenogramma")}.txt`;
  downloadTextFile(filename, text);
};

const openEditModal = () => {
  if (!task.value) return;
  const meetingDate = task.value.meeting_date ? new Date(task.value.meeting_date) : null;
  const dateInput =
    meetingDate && !Number.isNaN(meetingDate.getTime())
      ? meetingDate.toISOString().slice(0, 10)
      : "";

  editForm.value = {
    taskName: task.value.task_name,
    description: task.value.description,
    meetingDate: dateInput,
  };
  editError.value = null;
  showEditModal.value = true;
};

const handleSaveEdit = async () => {
  if (!task.value) return;
  editError.value = null;

  const payload: Record<string, string | undefined> = {};
  if (editForm.value.taskName.trim() !== task.value.task_name) {
    payload.task_name = editForm.value.taskName.trim();
  }
  if (editForm.value.description.trim() !== task.value.description) {
    payload.description = editForm.value.description.trim();
  }
  if (editForm.value.meetingDate) {
    const newIso = new Date(`${editForm.value.meetingDate}T12:00:00`).toISOString();
    if (newIso !== task.value.meeting_date) {
      payload.meeting_date = newIso;
    }
  }

  if (Object.keys(payload).length === 0) {
    showEditModal.value = false;
    return;
  }

  try {
    await tasksStore.updateTask(task.value.task_id, payload);
    showEditModal.value = false;
  } catch (err: unknown) {
    editError.value = extractApiErrorMessage(err, "Не удалось сохранить изменения");
  }
};

const handleDelete = async () => {
  if (!task.value) return;
  deleteError.value = null;
  try {
    await tasksStore.deleteTask(task.value.task_id);
    showDeleteModal.value = false;
    router.push({ name: "Dashboard" });
  } catch (err: unknown) {
    deleteError.value = extractApiErrorMessage(err, "Не удалось удалить запись");
  }
};

onMounted(async () => {
  try {
    const fetched = await tasksStore.fetchTask(route.params.id as string);

    if (isErrorStatus(fetched.status)) {
      error.value = "Обработка завершилась с ошибкой";
      isLoading.value = false;
      return;
    }

    if (!isDone(fetched.status)) {
      void router.replace({
        name: "RecordProcessingDetails",
        params: { id: fetched.task_id },
      });
      return;
    }

    isLoading.value = false;

    if (fetched.pattern_id) {
      try {
        const pattern = await patternsService.getPatternById(fetched.pattern_id);
        patternName.value = pattern.name;
      } catch {
        // если шаблон удалён или недоступен — просто не показываем
      }
    }
  } catch (err: unknown) {
    error.value =
      tasksStore.currentError ??
      (err instanceof Error ? err.message : "Не удалось загрузить запись");
    isLoading.value = false;
  }
});
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-dark">
    <Header max-width="max-w-[1800px]" />

    <main class="mx-auto max-w-[1800px] px-4 py-6 sm:px-6 lg:px-8">
      <div v-if="isLoading" class="flex items-center justify-center py-12">
        <Spinner size="lg" class="text-blue-600 dark:text-white" />
      </div>

      <div
        v-else-if="error"
        class="rounded-lg border border-red-200 bg-red-50 p-6 text-center dark:border-red-800 dark:bg-red-900/20"
      >
        <p class="text-red-600 dark:text-red-400">{{ error }}</p>
      </div>

      <template v-else-if="task">
        <button
          type="button"
          class="mb-4 flex items-center gap-1.5 text-sm font-medium text-gray-600 transition-colors hover:text-gray-900 dark:text-gray-400 dark:hover:text-white"
          @click="router.push({ name: 'Dashboard' })"
        >
          <ArrowLeft :size="16" />
          Главная
        </button>

        <!-- Шапка: заголовок + плеер + действия -->
        <div class="mb-3 flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div class="min-w-0 flex-1">
            <h1 class="truncate text-2xl font-bold text-gray-900 dark:text-white">
              {{ task.task_name || "Без названия" }}
            </h1>
          </div>

          <div class="flex items-center gap-3">
            <button
              type="button"
              :disabled="totalSeconds <= 0"
              class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-full bg-blue-600 text-white transition-colors hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50 dark:bg-white dark:text-dark dark:hover:bg-gray-200"
              :title="isPlaying ? 'Пауза' : 'Воспроизвести'"
              @click="togglePlayback"
            >
              <Play v-if="!isPlaying" :size="18" />
              <Pause v-else :size="18" />
            </button>
            <span
              class="flex h-10 w-10 items-center justify-center rounded-full border border-gray-200 text-gray-500 dark:border-dark-border dark:text-gray-400"
            >
              <Clock :size="16" />
            </span>

            <div v-if="task.change_flag" class="ml-1 flex items-center gap-2">
              <Button variant="outline" size="sm" @click="openEditModal">
                <Edit :size="16" />
                Редактировать
              </Button>
              <Button variant="outline" size="sm" @click="showDeleteModal = true">
                <Trash2 :size="16" />
                Удалить
              </Button>
            </div>
          </div>
        </div>

        <!-- Метаданные и шкала -->
        <div class="mb-6 flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
          <p class="text-sm text-gray-500 dark:text-gray-400">
            {{ formattedMeetingDate }}
            <template v-if="formattedDurationLabel"> • {{ formattedDurationLabel }}</template>
            <template v-if="speakersLabel"> • {{ speakersLabel }}</template>
          </p>

          <div class="flex items-center gap-3 lg:w-2/5">
            <span
              class="font-mono text-xs tabular-nums text-gray-500 dark:text-gray-400"
            >
              {{ formattedCurrent }}
            </span>
            <div
              class="relative h-1.5 flex-1 cursor-pointer rounded-full bg-gray-200 dark:bg-dark-elevated"
              @click="handleScrub"
            >
              <div
                class="absolute inset-y-0 left-0 rounded-full bg-blue-500 dark:bg-white"
                :style="{ width: `${progressPercent}%` }"
              />
              <div
                class="absolute top-1/2 -translate-x-1/2 -translate-y-1/2 rounded-full bg-blue-500 transition-all dark:bg-white"
                :class="totalSeconds > 0 ? 'h-3 w-3 opacity-100' : 'h-0 w-0 opacity-0'"
                :style="{ left: `${progressPercent}%` }"
              />
            </div>
            <span
              class="font-mono text-xs tabular-nums text-gray-500 dark:text-gray-400"
            >
              {{ formattedTotal }}
            </span>
          </div>
        </div>

        <p
          v-if="task.description"
          class="mb-6 whitespace-pre-line text-sm leading-relaxed text-gray-700 dark:text-gray-300"
        >
          {{ task.description }}
        </p>

        <div
          v-if="!hasResult"
          class="rounded-lg border border-gray-200 bg-white p-12 text-center dark:border-dark-border dark:bg-dark-card"
        >
          <p class="text-gray-500 dark:text-gray-400">
            Результат обработки пока недоступен.
          </p>
        </div>

        <div v-else class="grid gap-6 lg:grid-cols-2">
          <!-- Конспект -->
          <Card padding="lg" class="h-fit">
            <div class="mb-2 flex items-start justify-between gap-3">
              <div>
                <h2 class="text-base font-semibold text-gray-900 dark:text-white">Конспект</h2>
                <p v-if="patternName" class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">
                  Шаблон: {{ patternName }}
                </p>
              </div>
              <Button
                variant="outline"
                size="sm"
                :disabled="!normalizedResult.summary"
                @click="handleDownloadSummary"
              >
                <Download :size="14" />
                Скачать
              </Button>
            </div>

            <div
              v-if="normalizedResult.summary"
              class="summary-content text-sm text-gray-800 dark:text-gray-200"
              v-html="renderedSummary"
            />
            <p v-else class="text-sm text-gray-500 dark:text-gray-400">
              Конспект ещё не сформирован.
            </p>
          </Card>

          <!-- Стенограмма -->
          <Card padding="lg" class="h-fit">
            <div class="mb-4 flex items-center justify-between">
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">Стенограмма</h2>
              <Button
                variant="outline"
                size="sm"
                :disabled="normalizedResult.transcript.length === 0"
                @click="handleDownloadTranscript"
              >
                <Download :size="14" />
                Скачать
              </Button>
            </div>

            <div v-if="normalizedResult.transcript.length > 0" class="space-y-4">
              <div
                v-for="(entry, index) in normalizedResult.transcript"
                :key="index"
                class="space-y-2"
              >
                <div class="flex items-center justify-between gap-3">
                  <span
                    :class="[
                      'inline-flex items-center rounded-full px-3 py-0.5 text-xs font-semibold uppercase tracking-wide',
                      speakerPillClass(entry.speaker),
                    ]"
                  >
                    {{ entry.speaker }}
                  </span>
                  <span
                    v-if="entry.timestamp"
                    class="font-mono text-xs tabular-nums text-gray-400 dark:text-gray-500"
                  >
                    {{ entry.timestamp }}
                  </span>
                </div>
                <p class="text-sm leading-relaxed text-gray-700 dark:text-gray-300">
                  {{ entry.text }}
                </p>
              </div>
            </div>
            <p v-else class="text-sm text-gray-500 dark:text-gray-400">
              Стенограмма недоступна.
            </p>
          </Card>
        </div>
      </template>
    </main>

    <Modal v-model="showEditModal" title="Редактировать запись" size="md">
      <form @submit.prevent="handleSaveEdit" class="space-y-4">
        <FormError :message="editError" />
        <Input v-model="editForm.taskName" label="Название" />
        <Textarea v-model="editForm.description" label="Описание" :rows="4" />
        <DatePicker v-model="editForm.meetingDate" label="Дата встречи" />
      </form>

      <template #footer="{ close }">
        <Button variant="outline" @click="close">Отмена</Button>
        <Button @click="handleSaveEdit" :is-loading="tasksStore.isMutating">Сохранить</Button>
      </template>
    </Modal>

    <Modal v-model="showDeleteModal" title="Удалить запись?" size="sm">
      <div class="space-y-4">
        <FormError :message="deleteError" />
        <p class="text-gray-700 dark:text-gray-300">
          Вы уверены, что хотите удалить эту запись? Действие нельзя отменить.
        </p>
      </div>

      <template #footer="{ close }">
        <Button variant="outline" @click="close">Отмена</Button>
        <Button @click="handleDelete" :is-loading="tasksStore.isMutating">Удалить</Button>
      </template>
    </Modal>
  </div>
</template>

<style scoped>
.summary-content :deep(h1) {
  font-size: 1.05rem;
  font-weight: 700;
  margin: 1.25rem 0 0.5rem;
  color: inherit;
}
.summary-content :deep(h2) {
  font-size: 0.95rem;
  font-weight: 700;
  margin: 1rem 0 0.5rem;
}
.summary-content :deep(h3) {
  font-size: 0.9rem;
  font-weight: 600;
  margin: 0.75rem 0 0.4rem;
}
.summary-content :deep(p) {
  margin: 0.5rem 0;
  line-height: 1.55;
}
.summary-content :deep(ul),
.summary-content :deep(ol) {
  margin: 0.4rem 0 0.6rem 1.25rem;
  line-height: 1.55;
}
.summary-content :deep(ul) {
  list-style: disc;
}
.summary-content :deep(ol) {
  list-style: decimal;
}
.summary-content :deep(li) {
  margin: 0.2rem 0;
}
.summary-content :deep(li > ul),
.summary-content :deep(li > ol) {
  margin-top: 0.2rem;
  margin-bottom: 0.2rem;
}
.summary-content :deep(strong) {
  font-weight: 700;
}
.summary-content :deep(em) {
  font-style: italic;
}
.summary-content :deep(code) {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  background: rgba(0, 0, 0, 0.05);
  padding: 0.05rem 0.35rem;
  border-radius: 0.25rem;
  font-size: 0.85em;
}
.dark .summary-content :deep(code) {
  background: rgba(255, 255, 255, 0.08);
}
.summary-content > :first-child {
  margin-top: 0;
}
.summary-content > :last-child {
  margin-bottom: 0;
}
</style>
