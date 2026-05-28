<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { Plus, Save, Trash2, ArrowUp, ArrowDown } from "@lucide/vue";
import Header from "../components/layout/Header.vue";
import Card from "../components/ui/Card.vue";
import Select from "../components/ui/Select.vue";
import Input from "../components/ui/Input.vue";
import Textarea from "../components/ui/Textarea.vue";
import Button from "../components/ui/Button.vue";
import Modal from "../components/ui/Modal.vue";
import Spinner from "../components/ui/Spinner.vue";
import FormError from "../components/ui/FormError.vue";
import { usePatternsStore, buildCreatePayload, buildUpdatePayload, patternSaveErrorMessage } from "../stores/patterns";
import { useGroupsStore } from "../stores/groups";
import { extractApiErrorMessage } from "../api";
import type { PatternDto } from "../api";
import { useAutoRefresh } from "../composables/useAutoRefresh";

interface Section {
  id: number;
  title: string;
  prompt: string;
}

interface LocalDraft {
  id: string;
  name: string;
  description: string;
  groupId: string;
  isDraft: true;
}

const patternsStore = usePatternsStore();
const groupsStore = useGroupsStore();

const isInitialLoading = ref(true);
const activeTab = ref<"global" | "group">("global");
const selectedPatternId = ref<string | null>(null);
const isEditing = ref(false);
const showDeleteModal = ref(false);
const localDrafts = ref<LocalDraft[]>([]);
let draftCounter = 0;

const saveError = ref<string | null>(null);
const deleteError = ref<string | null>(null);

const form = ref({
  name: "",
  description: "",
  details: "Оптимально",
  style: "Разговорный",
  sections: [] as Section[],
  groupId: "",
});

const detailsOptions = [
  { value: "Оптимально", label: "Оптимально" },
  { value: "Кратко", label: "Кратко" },
  { value: "Подробно", label: "Подробно" },
];

const styleOptions = [
  { value: "Разговорный", label: "Разговорный" },
  { value: "Официальный", label: "Официальный" },
  { value: "Технический", label: "Технический" },
];

const groupOptions = computed(() => {
  return groupsStore.groups.map((g) => ({
    value: g.group_id,
    label: g.name,
  }));
});

const displayedPatterns = computed((): Array<PatternDto & { isDraft?: true }> => {
  const currentGroupId = form.value.groupId;
  
  if (activeTab.value === "global") {
    const globalDrafts = localDrafts.value
      .filter((d) => !d.groupId)
      .map((d) => ({
        pattern_id: d.id,
        name: d.name,
        description: d.description,
        summary_prompt: "",
        additional_prompt: null,
        group_id: "",
        created_at: "",
        change_flag: false,
        isDraft: true as const,
      }));
    const globalPatterns = patternsStore.globalPatterns ?? [];
    return [...globalDrafts, ...globalPatterns];
  }
  
  const groupDrafts = localDrafts.value
    .filter((d) => d.groupId === currentGroupId)
    .map((d) => ({
      pattern_id: d.id,
      name: d.name,
      description: d.description,
      summary_prompt: "",
      additional_prompt: null,
      group_id: d.groupId,
      created_at: "",
      change_flag: false,
      isDraft: true as const,
    }));
  
  const groupLocalPatterns = patternsStore.groupLocalPatterns ?? [];
  const groupPatterns = groupLocalPatterns.filter(
    (p) => p.group_id === currentGroupId
  );
  
  return [...groupDrafts, ...groupPatterns];
});

const generatedPrompt = computed(() => {
  let prompt = `Ты — ассистент, который делает summary встречи по транскрипту.\n\n`;
  prompt += `Стиль: ${form.value.style}.\n`;
  prompt += `Детализация: ${form.value.details}.\n`;
  prompt += `Форматирование:\n- Используй маркированные списки, где уместно.\n- Выделяй имена участников жирным.\n\n`;

  if (form.value.sections.length > 0) {
    prompt += `Секции для анализа:\n`;
    form.value.sections.forEach((sec) => {
      if (sec.title || sec.prompt) {
        prompt += `[${sec.title || "Без названия"}]\n${sec.prompt || "..."}\n\n`;
      }
    });
  }

  return prompt.trim();
});

const parseSectionsFromPrompt = (additionalPrompt: any): Section[] => {
  const sections: Section[] = [];
  
  if (additionalPrompt && typeof additionalPrompt === "object") {
    if (Array.isArray(additionalPrompt)) {
      additionalPrompt.forEach((item, index) => {
        sections.push({
          id: index + 1,
          title: item.title || `Секция ${index + 1}`,
          prompt: item.prompt || "",
        });
      });
    } else {
      Object.entries(additionalPrompt).forEach(([key, value], index) => {
        sections.push({
          id: index + 1,
          title: key,
          prompt: String(value),
        });
      });
    }
  }
  
  return sections;
};

const buildAdditionalPromptFromSections = (): any => {
  if (form.value.sections.length === 0) return null;
  
  return form.value.sections.map((section) => ({
    title: section.title,
    prompt: section.prompt,
  }));
};

const loadPatternToForm = (pattern: PatternDto & { isDraft?: true }) => {
  form.value.name = pattern.name;
  form.value.description = pattern.description;
  form.value.groupId = pattern.group_id || "";
  
  if (pattern.isDraft) {
    form.value.details = "Оптимально";
    form.value.style = "Разговорный";
    form.value.sections = [];
    isEditing.value = false;
    return;
  }
  
  // Парсим промпт для извлечения стиля и детализации
  const summaryPrompt = pattern.summary_prompt;
  if (summaryPrompt.includes("Стиль:")) {
    const styleMatch = summaryPrompt.match(/Стиль:\s*(\w+)/);
    if (styleMatch) form.value.style = styleMatch[1];
  }
  if (summaryPrompt.includes("Детализация:")) {
    const detailsMatch = summaryPrompt.match(/Детализация:\s*(\w+)/);
    if (detailsMatch) form.value.details = detailsMatch[1];
  }
  
  form.value.sections = parseSectionsFromPrompt(pattern.additional_prompt);
  isEditing.value = true;
};

const resetForm = () => {
  form.value = {
    name: "",
    description: "",
    details: "Оптимально",
    style: "Разговорный",
    sections: [],
    groupId: activeTab.value === "group" && groupsStore.groups.length > 0 
      ? groupsStore.groups[0].group_id 
      : "",
  };
  selectedPatternId.value = null;
  isEditing.value = false;
};

const clearDrafts = () => {
  localDrafts.value = [];
};

const handleNewPattern = () => {
  const draftId = `draft-${++draftCounter}`;
  const groupId = activeTab.value === "group" ? form.value.groupId : "";
  
  const draft: LocalDraft = {
    id: draftId,
    name: "Новый шаблон",
    description: "",
    groupId,
    isDraft: true,
  };
  
  localDrafts.value.push(draft);
  selectedPatternId.value = draftId;
};

const addSection = () => {
  const newId = form.value.sections.length > 0 
    ? Math.max(...form.value.sections.map((s) => s.id)) + 1 
    : 1;
  form.value.sections.push({ id: newId, title: "", prompt: "" });
};

const removeSection = (index: number) => {
  form.value.sections.splice(index, 1);
};

const moveSectionUp = (index: number) => {
  if (index > 0) {
    const temp = form.value.sections[index];
    form.value.sections[index] = form.value.sections[index - 1];
    form.value.sections[index - 1] = temp;
  }
};

const moveSectionDown = (index: number) => {
  if (index < form.value.sections.length - 1) {
    const temp = form.value.sections[index];
    form.value.sections[index] = form.value.sections[index + 1];
    form.value.sections[index + 1] = temp;
  }
};

const handleSave = async () => {
  saveError.value = null;

  if (!form.value.name.trim()) {
    saveError.value = "Укажите название шаблона";
    return;
  }

  if (activeTab.value === "group" && !form.value.groupId) {
    saveError.value = "Выберите группу для группового шаблона";
    return;
  }

  try {
    const additionalPrompt = buildAdditionalPromptFromSections();
    const isDraft = selectedPatternId.value?.startsWith("draft-");
    
    if (isEditing.value && selectedPatternId.value && !isDraft) {
      // Обновление существующего шаблона
      const payload = buildUpdatePayload({
        patternId: selectedPatternId.value,
        name: form.value.name,
        description: form.value.description,
        summaryPrompt: generatedPrompt.value,
        additionalPromptJson: JSON.stringify(additionalPrompt),
        scope: activeTab.value,
        groupId: form.value.groupId,
      });
      
      await patternsStore.updatePattern(selectedPatternId.value, payload, activeTab.value);
    } else {
      // Создание нового шаблона
      const payload = buildCreatePayload({
        patternId: "",
        name: form.value.name,
        description: form.value.description,
        summaryPrompt: generatedPrompt.value,
        additionalPromptJson: JSON.stringify(additionalPrompt),
        scope: activeTab.value,
        groupId: form.value.groupId,
      });
      
      const created = await patternsStore.createPattern(payload, activeTab.value);
      
      // Удаляем черновик, если сохраняли из него
      if (isDraft && selectedPatternId.value) {
        localDrafts.value = localDrafts.value.filter((d) => d.id !== selectedPatternId.value);
      }
      
      selectedPatternId.value = created.pattern_id;
      isEditing.value = true;
    }
  } catch (error) {
    console.error("Failed to save pattern:", error);
    saveError.value = patternSaveErrorMessage(error);
  }
};

const handleDelete = async () => {
  if (!selectedPatternId.value) return;

  deleteError.value = null;
  
  const isDraft = selectedPatternId.value.startsWith("draft-");
  
  if (isDraft) {
    // Удаляем черновик локально
    localDrafts.value = localDrafts.value.filter((d) => d.id !== selectedPatternId.value);
    showDeleteModal.value = false;
    selectFirstPatternOrCreate();
    return;
  }
  
  try {
    await patternsStore.deletePattern(selectedPatternId.value, activeTab.value);
    showDeleteModal.value = false;
    selectFirstPatternOrCreate();
  } catch (error) {
    console.error("Failed to delete pattern:", error);
    deleteError.value = extractApiErrorMessage(error, "Не удалось удалить шаблон");
  }
};

const selectFirstPatternOrCreate = () => {
  const patterns = displayedPatterns.value;
  if (patterns.length > 0) {
    selectedPatternId.value = patterns[0].pattern_id;
  } else {
    handleNewPattern();
  }
};

const handleTabChange = async (tab: "global" | "group") => {
  // Очищаем черновики при смене вкладки
  clearDrafts();
  
  activeTab.value = tab;
  resetForm();
  
  if (tab === "global") {
    await patternsStore.fetchGlobalPatterns();
  } else if (groupsStore.groups.length > 0) {
    const firstGroupId = groupsStore.groups[0].group_id;
    form.value.groupId = firstGroupId;
    await patternsStore.fetchGroupPatterns(firstGroupId);
  }
  
  selectFirstPatternOrCreate();
};

const handleGroupChange = async (groupId: string) => {
  // Очищаем черновики при смене группы
  clearDrafts();
  
  form.value.groupId = groupId;
  await patternsStore.fetchGroupPatterns(groupId);
  resetForm();
  form.value.groupId = groupId;
  selectFirstPatternOrCreate();
};

watch(selectedPatternId, (newId) => {
  saveError.value = null;
  if (newId) {
    const pattern = displayedPatterns.value.find((p) => p.pattern_id === newId);
    if (pattern) {
      loadPatternToForm(pattern);
    }
  }
});

// Инициализация groupId при загрузке групп
watch(
  () => groupsStore.groups,
  (groups) => {
    if (groups.length > 0 && !form.value.groupId) {
      form.value.groupId = groups[0].group_id;
    }
  },
  { immediate: true }
);

onMounted(async () => {
  try {
    await groupsStore.fetchGroups({ force: true });
    await patternsStore.fetchGlobalPatterns({ force: true });
  } catch (error) {
    console.error("Failed to load initial data:", error);
  } finally {
    isInitialLoading.value = false;
  }
  
  selectFirstPatternOrCreate();
});

useAutoRefresh(async () => {
  try {
    await groupsStore.fetchGroups({ force: true });
    if (activeTab.value === "global") {
      await patternsStore.fetchGlobalPatterns({ force: true });
    } else if (form.value.groupId) {
      await patternsStore.fetchGroupPatterns(form.value.groupId, { force: true });
    }
  } catch (error) {
    console.error("Failed to refresh patterns:", error);
  }
});
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-dark">
    <Header max-width="max-w-[1800px]" />

    <!-- Initial Loading State -->
    <div v-if="isInitialLoading" class="flex min-h-[50vh] items-center justify-center">
      <Spinner size="lg" class="text-blue-600 dark:text-white" />
    </div>

    <main v-else class="mx-auto max-w-[1800px] px-4 py-8 sm:px-6 lg:px-8">
      <!-- Page Header -->
      <div class="mb-6 flex items-center justify-between">
        <div>
          <p class="text-sm text-gray-500 dark:text-gray-400">Шаблоны</p>
          <h1 class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">
            Конструктор шаблонов
          </h1>
        </div>
        <Button @click="handleNewPattern">
          <Plus :size="18" />
          Новый шаблон
        </Button>
      </div>

      <!-- Tabs -->
      <div class="mb-6 flex gap-2">
        <button
          :class="[
            'rounded-lg px-4 py-2 text-sm font-medium transition-colors',
            activeTab === 'global'
              ? 'bg-blue-600 text-white dark:bg-white dark:text-dark'
              : 'bg-white text-gray-700 hover:bg-gray-50 dark:bg-dark-card dark:text-gray-300 dark:hover:bg-dark-elevated cursor-pointer',
          ]"
          @click="handleTabChange('global')"
        >
          Глобальные
        </button>
        <button
          :class="[
            'rounded-lg px-4 py-2 text-sm font-medium transition-colors',
            activeTab === 'group'
              ? 'bg-blue-600 text-white dark:bg-white dark:text-dark'
              : 'bg-white text-gray-700 hover:bg-gray-50 dark:bg-dark-card dark:text-gray-300 dark:hover:bg-dark-elevated cursor-pointer',
          ]"
          @click="handleTabChange('group')"
        >
          Групповые
        </button>
      </div>

      <!-- Group Selector for Group Tab -->
      <div v-if="activeTab === 'group'" class="mb-6">
        <Select
          :model-value="form.groupId"
          :options="groupOptions"
          @update:model-value="handleGroupChange"
        />
      </div>

      <!-- Content Grid -->
      <div class="grid gap-6 lg:grid-cols-12">
        <!-- Patterns List -->
        <Card padding="none" class="lg:col-span-4 xl:col-span-3">
          <div class="border-b border-gray-200 p-4 dark:border-dark-border">
            <h2 class="text-sm font-semibold text-gray-900 dark:text-white">Список</h2>
          </div>

          <div class="max-h-[600px] space-y-2 overflow-y-auto p-4">
            <div v-if="patternsStore.isLoading" class="flex justify-center py-8">
              <Spinner size="md" class="text-blue-600 dark:text-white" />
            </div>

            <div
              v-else-if="displayedPatterns.length === 0"
              class="py-8 text-center text-sm text-gray-500 dark:text-gray-400"
            >
              Нет шаблонов
            </div>

            <button
              v-else
              v-for="pattern in displayedPatterns"
              :key="pattern.pattern_id"
              :class="[
                'w-full rounded-lg p-3 text-left transition-colors',
                selectedPatternId === pattern.pattern_id
                  ? 'bg-blue-50 dark:bg-white/10'
                  : 'hover:bg-gray-50 dark:hover:bg-dark-elevated',
                pattern.isDraft ? 'border-2 border-dashed border-blue-300 dark:border-gray-500' : '',
              ]"
              @click="selectedPatternId = pattern.pattern_id"
            >
              <h3 class="font-medium text-gray-900 dark:text-white">
                {{ pattern.name }}
                <span v-if="pattern.isDraft" class="ml-2 text-xs text-blue-500 dark:text-gray-400">(черновик)</span>
              </h3>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ pattern.description || "Описание отсутствует" }}
              </p>
            </button>
          </div>
        </Card>

        <!-- Pattern Editor -->
        <Card padding="none" class="lg:col-span-8 xl:col-span-9">
          <div
            class="flex items-center justify-between border-b border-gray-200 p-6 dark:border-dark-border"
          >
            <div>
              <p class="text-xs text-gray-500 dark:text-gray-400">Редактирование</p>
              <h2 class="mt-1 text-xl font-bold text-gray-900 dark:text-white">
                {{ form.name || "Новый шаблон" }}
              </h2>
            </div>
            <div class="flex items-center gap-3">
              <Button @click="handleSave" :disabled="!form.name.trim()">
                <Save :size="16" />
                Сохранить
              </Button>
              <Button
                v-if="isEditing"
                variant="outline"
                @click="deleteError = null; showDeleteModal = true"
              >
                <Trash2 :size="16" />
                Удалить
              </Button>
            </div>
          </div>

          <div class="max-h-[600px] space-y-6 overflow-y-auto p-6">
            <FormError :message="saveError" />

            <!-- Basic Info -->
            <div class="grid gap-6 sm:grid-cols-2">
              <Input v-model="form.name" label="Название" placeholder="Введите название" />
              <Input v-model="form.description" label="Описание" placeholder="Введите описание" />
              <Select v-model="form.details" label="Детализация" :options="detailsOptions" />
              <Select v-model="form.style" label="Стиль" :options="styleOptions" />
            </div>

            <div class="h-px bg-gray-200 dark:bg-dark-border" />

            <!-- Sections -->
            <div>
              <div class="mb-4 flex items-center justify-between">
                <h3 class="text-base font-semibold text-gray-900 dark:text-white">
                  Секции
                </h3>
                <Button variant="outline" size="sm" @click="addSection">
                  <Plus :size="16" />
                  Добавить секцию
                </Button>
              </div>

              <div class="space-y-4">
                <Card
                  v-for="(section, index) in form.sections"
                  :key="section.id"
                  padding="md"
                >
                  <div class="mb-3 flex items-center gap-3">
                    <div
                      class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-gray-100 text-sm font-bold text-gray-600 dark:bg-dark-elevated dark:text-gray-300"
                    >
                      {{ index + 1 }}
                    </div>
                    <Input
                      v-model="section.title"
                      placeholder="Название секции"
                      class="flex-1 p-1.5"
                    />
                    <div class="flex items-center gap-1">
                      <button
                        :disabled="index === form.sections.length - 1"
                        class="rounded border border-gray-300 p-2 text-gray-400 transition-colors hover:bg-gray-50 hover:text-gray-700 disabled:opacity-50 dark:border-dark-border dark:hover:bg-dark-elevated"
                        @click="moveSectionDown(index)"
                      >
                        <ArrowDown :size="16" />
                      </button>
                      <button
                        :disabled="index === 0"
                        class="rounded border border-gray-300 p-2 text-gray-400 transition-colors hover:bg-gray-50 hover:text-gray-700 disabled:opacity-50 dark:border-dark-border dark:hover:bg-dark-elevated"
                        @click="moveSectionUp(index)"
                      >
                        <ArrowUp :size="16" />
                      </button>
                      <button
                        class="rounded border border-gray-300 p-2 text-red-400 transition-colors hover:border-red-200 hover:bg-red-50 hover:text-red-600 dark:border-dark-border dark:hover:bg-red-900/20"
                        @click="removeSection(index)"
                      >
                        <Trash2 :size="16" />
                      </button>
                    </div>
                  </div>
                  <Textarea
                    v-model="section.prompt"
                    :rows="3"
                    placeholder="Какой результат хотите получить от ИИ"
                  />
                </Card>

                <div
                  v-if="form.sections.length === 0"
                  class="py-8 text-center text-sm text-gray-500 dark:text-gray-400"
                >
                  Секции пока не добавлены. Нажмите «Добавить секцию», чтобы начать.
                </div>
              </div>
            </div>

            <!-- Generated Prompt -->
            <Card padding="none">
              <div
                class="border-b border-gray-200 bg-gray-50 px-6 py-4 dark:border-dark-border dark:bg-dark-elevated"
              >
                <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
                  Получившийся промпт
                </h3>
              </div>
              <div class="p-6">
                <pre
                  class="whitespace-pre-wrap font-mono text-sm leading-relaxed text-gray-700 dark:text-gray-300"
                >{{ generatedPrompt }}</pre>
              </div>
            </Card>
          </div>
        </Card>
      </div>
    </main>

    <!-- Delete Confirmation Modal -->
    <Modal v-model="showDeleteModal" title="Удалить шаблон?" size="sm">
      <div class="space-y-4">
        <FormError :message="deleteError" />
        <p class="text-gray-700 dark:text-gray-300">
          Вы уверены, что хотите удалить этот шаблон? Это действие нельзя отменить.
        </p>
      </div>

      <template #footer="{ close }">
        <Button variant="outline" @click="close">Отмена</Button>
        <Button variant="primary" @click="handleDelete" :is-loading="patternsStore.isMutating">Удалить</Button>
      </template>
    </Modal>
  </div>
</template>
