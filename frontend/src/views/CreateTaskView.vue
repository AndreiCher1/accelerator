<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { ArrowLeft, Upload, Paperclip } from "@lucide/vue";
import Header from "../components/layout/Header.vue";
import Card from "../components/ui/Card.vue";
import Select from "../components/ui/Select.vue";
import Input from "../components/ui/Input.vue";
import DatePicker from "../components/ui/DatePicker.vue";
import Textarea from "../components/ui/Textarea.vue";
import Button from "../components/ui/Button.vue";
import FormError from "../components/ui/FormError.vue";
import { extractApiErrorMessage } from "../api";
import { useGroupsStore } from "../stores/groups";
import { usePatternsStore } from "../stores/patterns";
import { useTasksStore } from "../stores/tasks";
import {
  fieldControlClass,
  fieldLabelClass,
} from "../components/ui/fieldStyles";
import { useAutoRefresh } from "../composables/useAutoRefresh";

const router = useRouter();
const groupsStore = useGroupsStore();
const patternsStore = usePatternsStore();
const tasksStore = useTasksStore();

const form = ref({
  groupId: "",
  meetingDate: "",
  taskName: "",
  description: "",
  patternId: "",
  file: null as File | null,
});

const error = ref<string | null>(null);
const fileInputRef = ref<HTMLInputElement | null>(null);

const groupOptions = computed(() =>
  groupsStore.groups.map((g) => ({ value: g.group_id, label: g.name })),
);

const patternOptions = computed(() => {
  if (!form.value.groupId) return [];
  const patterns = [
    ...patternsStore.groupGlobalPatterns,
    ...patternsStore.groupLocalPatterns,
  ];
  return patterns.map((p) => ({ value: p.pattern_id, label: p.name }));
});

const fileName = computed(() =>
  form.value.file ? form.value.file.name : "Не выбран ни один файл",
);

const isFormValid = computed(
  () =>
    !!form.value.groupId &&
    form.value.taskName.trim().length > 0 &&
    form.value.description.trim().length > 0 &&
    !!form.value.meetingDate &&
    !!form.value.patternId &&
    !!form.value.file,
);

const triggerFileInput = () => {
  fileInputRef.value?.click();
};

const handleFileChange = (event: Event) => {
  const target = event.target as HTMLInputElement;
  form.value.file = target.files && target.files.length > 0 ? target.files[0] : null;
};

const handleGroupChange = async (groupId: string) => {
  form.value.groupId = groupId;
  form.value.patternId = "";
  if (groupId) {
    try {
      await patternsStore.fetchGroupPatterns(groupId);
    } catch {
      // ошибка отобразится через store
    }
  }
};

const handleSubmit = async () => {
  if (!isFormValid.value) return;
  error.value = null;

  try {
    const meetingDate = new Date(`${form.value.meetingDate}T12:00:00`).toISOString();

    const response = await tasksStore.uploadTask(form.value.groupId, form.value.file!, {
      task_name: form.value.taskName.trim(),
      description: form.value.description.trim(),
      meeting_date: meetingDate,
      pattern_id: form.value.patternId,
    });

    await router.push({
      name: "RecordProcessingDetails",
      params: { id: response.task_id },
    });
  } catch (err: unknown) {
    error.value = extractApiErrorMessage(err, "Не удалось создать запись");
  }
};

onMounted(async () => {
  try {
    await groupsStore.fetchGroups({ force: true });
  } catch {
    return;
  }

  const initial = groupsStore.activeGroupId ?? groupsStore.groups[0]?.group_id ?? "";
  if (initial) {
    form.value.groupId = initial;
    try {
      await patternsStore.fetchGroupPatterns(initial, { force: true });
    } catch {
      // ok
    }
  }
});

useAutoRefresh(async () => {
  try {
    await groupsStore.fetchGroups({ force: true });
    if (form.value.groupId) {
      await patternsStore.fetchGroupPatterns(form.value.groupId, { force: true });
    }
  } catch {
    // ok
  }
});
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-dark">
    <Header max-width="max-w-[1200px]" />

    <main class="mx-auto max-w-[1200px] px-4 py-8 sm:px-6 lg:px-8">
      <button
        type="button"
        class="mb-4 flex items-center gap-1.5 text-sm font-medium text-gray-600 transition-colors hover:text-gray-900 dark:text-gray-400 dark:hover:text-white"
        @click="router.push({ name: 'Dashboard' })"
      >
        <ArrowLeft :size="16" />
        Главная
      </button>

      <div class="mb-6">
        <p class="text-sm text-gray-500 dark:text-gray-400">Создание записи</p>
        <h1 class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">Создание записи</h1>
      </div>

      <Card padding="lg">
        <form @submit.prevent="handleSubmit" class="space-y-6">
          <FormError :message="error" />

          <div class="grid gap-6 sm:grid-cols-2">
            <Select
              label="Группа"
              :model-value="form.groupId"
              :options="groupOptions"
              placeholder="Выберите группу"
              @update:model-value="handleGroupChange"
            />

            <DatePicker v-model="form.meetingDate" label="Дата встречи" />
          </div>

          <Input
            v-model="form.taskName"
            label="Название записи"
            placeholder='Например: "Совещание №1"'
          />

          <Textarea
            v-model="form.description"
            label="Описание"
            :rows="4"
            placeholder='Например: "продуктовое совещание"'
          />

          <Select
            v-model="form.patternId"
            label="Шаблон конспекта"
            :options="patternOptions"
            :disabled="!form.groupId || patternOptions.length === 0"
            placeholder="Выберите шаблон"
          />

          <div class="flex w-full flex-col gap-1.5">
            <label :class="fieldLabelClass">Аудиофайл</label>
            <input
              ref="fileInputRef"
              type="file"
              class="hidden"
              accept="audio/*"
              @change="handleFileChange"
            />
            <div
              :class="[
                fieldControlClass(),
                'flex h-10 items-center gap-0 overflow-hidden !p-0',
              ]"
            >
              <button
                type="button"
                class="inline-flex h-full shrink-0 items-center gap-2 rounded-l-lg bg-blue-600 px-4 text-sm font-medium text-white transition-colors hover:bg-blue-700 active:scale-[0.98] dark:bg-white dark:text-dark dark:hover:bg-gray-200"
                @click="triggerFileInput"
              >
                <Paperclip :size="15" class="opacity-90" />
                Выбор файла
              </button>
              <span
                class="flex-1 truncate px-4 text-sm"
                :class="
                  form.file
                    ? 'text-gray-900 dark:text-white'
                    : 'text-gray-400 dark:text-gray-500'
                "
              >
                {{ fileName }}
              </span>
            </div>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              Поддерживаемые форматы: MP3, WAV, OGG, AAC, FLAC.
            </p>
          </div>

          <div class="flex justify-end gap-3 pt-4">
            <Button type="submit" :disabled="!isFormValid" :is-loading="tasksStore.isUploading">
              <Upload :size="18" />
              Создать запись
            </Button>
          </div>
        </form>
      </Card>
    </main>
  </div>
</template>
