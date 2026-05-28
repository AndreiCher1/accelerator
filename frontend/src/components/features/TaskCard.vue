<script setup lang="ts">
import { computed } from "vue";
import { useRouter } from "vue-router";
import {
  AlertCircle,
  CheckCircle2,
  Clock,
  FileAudio,
  Loader2,
  Trash2,
} from "@lucide/vue";
import Card from "../ui/Card.vue";
import Badge from "../ui/Badge.vue";
import type { TaskDto } from "../../api/tasks.types";
import {
  formatDuration,
  formatMeetingDate,
  isDone,
  statusLabel,
  toUiStatus,
} from "../../utils/taskStatus";

interface Props {
  task: TaskDto;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  delete: [taskId: string];
}>();
const router = useRouter();

const ui = computed(() => toUiStatus(props.task.status));

const statusConfig = computed(() => {
  switch (ui.value) {
    case "done":
      return { variant: "success" as const, icon: CheckCircle2 };
    case "processing":
      return { variant: "warning" as const, icon: Loader2 };
    case "pending":
      return { variant: "info" as const, icon: Clock };
    case "error":
      return { variant: "error" as const, icon: AlertCircle };
    default:
      return { variant: "default" as const, icon: AlertCircle };
  }
});

const formattedMeetingDate = computed(() => formatMeetingDate(props.task.meeting_date));
const formattedDurationLabel = computed(() => formatDuration(props.task.duration_seconds));

const handleClick = () => {
  if (ui.value === "error") return;
  if (isDone(props.task.status)) {
    router.push({ name: "RecordDetails", params: { id: props.task.task_id } });
  } else {
    router.push({ name: "RecordProcessingDetails", params: { id: props.task.task_id } });
  }
};

const handleDelete = (event: MouseEvent) => {
  event.stopPropagation();
  emit("delete", props.task.task_id);
};
</script>

<template>
  <Card
    padding="none"
    :hover="ui !== 'error'"
    :class="[
      'overflow-hidden transition-all',
      ui === 'error' ? 'cursor-not-allowed opacity-60' : 'cursor-pointer',
    ]"
    @click="handleClick"
  >
    <div class="flex items-start gap-4 p-4">
      <div
        :class="[
          'flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-lg',
          {
            'bg-green-100 text-green-600 dark:bg-green-500/20 dark:text-green-400': ui === 'done',
            'bg-yellow-100 text-yellow-600 dark:bg-yellow-500/20 dark:text-yellow-400':
              ui === 'processing',
            'bg-blue-100 text-blue-600 dark:bg-white/10 dark:text-gray-300': ui === 'pending',
            'bg-red-100 text-red-600 dark:bg-red-500/20 dark:text-red-400': ui === 'error',
            'bg-gray-100 text-gray-500 dark:bg-dark-elevated dark:text-gray-400':
              ui === 'unknown',
          },
        ]"
      >
        <component
          :is="statusConfig.icon"
          :size="24"
          :class="{ 'animate-spin': ui === 'processing' }"
        />
      </div>

      <div class="min-w-0 flex-1">
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0 flex-1">
            <h3 class="truncate text-base font-semibold text-gray-900 dark:text-white">
              {{ task.task_name || "Без названия" }}
            </h3>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
              {{ formattedMeetingDate }}
              <template v-if="task.duration_seconds > 0"> • {{ formattedDurationLabel }}</template>
            </p>
          </div>

          <div class="flex items-center gap-2">
            <Badge :variant="statusConfig.variant" size="sm">
              {{ statusLabel(task.status) }}
            </Badge>
            <button
              v-if="task.change_flag"
              type="button"
              class="rounded-lg p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-red-600 dark:hover:bg-dark-elevated dark:hover:text-red-400"
              title="Удалить"
              @click="handleDelete"
            >
              <Trash2 :size="16" />
            </button>
          </div>
        </div>

        <div
          class="mt-3 flex flex-wrap items-center gap-4 text-xs text-gray-500 dark:text-gray-400"
        >
          <div class="flex items-center gap-1.5">
            <FileAudio :size="14" />
            <span class="truncate">{{ task.original_filename }}</span>
          </div>
        </div>
      </div>
    </div>
  </Card>
</template>
