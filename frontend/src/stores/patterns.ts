import { defineStore } from "pinia";
import { ref } from "vue";
import {
  ApiError,
  extractApiErrorMessage,
  patternsService,
  type CreatePatternRequest,
  type EditPatternRequest,
  type GroupPatternsResponse,
  type PatternDto,
  type PatternFormValues,
  type PatternScope,
} from "../api";
import {
  formatAdditionalPrompt,
  isGlobalPattern,
  parseAdditionalPrompt,
} from "../api/patterns.utils";

const CACHE_TTL_MS = 5 * 60 * 1000;

interface GlobalCacheEntry {
  patterns: PatternDto[];
  fetchedAt: number;
}

interface GroupPatternsCacheEntry {
  global_patterns: GroupPatternsResponse["global_patterns"];
  group_patterns: GroupPatternsResponse["group_patterns"];
  fetchedAt: number;
}

export function patternFormFromDto(pattern: PatternDto): PatternFormValues {
  const scope: PatternScope = isGlobalPattern(pattern) ? "global" : "group";
  return {
    patternId: pattern.pattern_id,
    name: pattern.name,
    description: pattern.description,
    summaryPrompt: pattern.summary_prompt,
    additionalPromptJson: formatAdditionalPrompt(pattern.additional_prompt),
    scope,
    groupId: pattern.group_id ?? "",
  };
}

export function emptyPatternForm(defaultGroupId = "", defaultScope: PatternScope = "global"): PatternFormValues {
  return {
    patternId: "",
    name: "",
    description: "",
    summaryPrompt: "",
    additionalPromptJson: "",
    scope: defaultScope,
    groupId: defaultGroupId,
  };
}

export function buildCreatePayload(form: PatternFormValues): CreatePatternRequest {
  const additional_prompt = parseAdditionalPrompt(form.additionalPromptJson);
  const payload: CreatePatternRequest = {
    name: form.name.trim(),
    description: form.description.trim(),
    summary_prompt: form.summaryPrompt.trim(),
  };

  if (additional_prompt !== undefined) {
    payload.additional_prompt = additional_prompt;
  }

  if (form.scope === "group") {
    payload.group_id = form.groupId;
  }

  return payload;
}

export function buildUpdatePayload(form: PatternFormValues): EditPatternRequest {
  const additional_prompt = parseAdditionalPrompt(form.additionalPromptJson);
  const payload: EditPatternRequest = {
    name: form.name.trim(),
    description: form.description.trim(),
    summary_prompt: form.summaryPrompt.trim(),
  };

  if (additional_prompt !== undefined) {
    payload.additional_prompt = additional_prompt;
  }

  return payload;
}

export function patternSaveErrorMessage(err: unknown): string {
  if (err instanceof ApiError) {
    if (err.isForbidden) return err.message || "Недостаточно прав для изменения шаблона";
    if (err.isNotFound) return err.message || "Шаблон не найден";
    if (err.status === 400) return err.message || "Ошибка во входных данных";
  }

  if (err instanceof Error && err.message.includes("JSON")) {
    return err.message;
  }

  return extractApiErrorMessage(err, "Не удалось сохранить шаблон");
}

export const usePatternsStore = defineStore("patterns", () => {
  const globalPatterns = ref<PatternDto[]>([]);
  const groupGlobalPatterns = ref<PatternDto[]>([]);
  const groupLocalPatterns = ref<PatternDto[]>([]);

  const globalCache = ref<GlobalCacheEntry | null>(null);
  const groupPatternsCache = ref<Map<string, GroupPatternsCacheEntry>>(new Map());

  const activeGroupId = ref<string | null>(null);
  const isLoading = ref(false);
  const isMutating = ref(false);
  const error = ref<string | null>(null);
  const forbidden = ref(false);

  function invalidateCache(): void {
    globalCache.value = null;
    groupPatternsCache.value.clear();
  }

  function applyGlobalEntry(entry: GlobalCacheEntry): void {
    globalPatterns.value = entry.patterns ?? [];
  }

  function applyGroupEntry(entry: GroupPatternsCacheEntry): void {
    groupGlobalPatterns.value = entry.global_patterns ?? [];
    groupLocalPatterns.value = entry.group_patterns ?? [];
  }

  async function fetchGlobalPatterns(options?: { force?: boolean }): Promise<void> {
    const force = options?.force ?? false;

    if (!force && globalCache.value && Date.now() - globalCache.value.fetchedAt < CACHE_TTL_MS) {
      applyGlobalEntry(globalCache.value);
      error.value = null;
      forbidden.value = false;
      return;
    }

    isLoading.value = true;
    error.value = null;
    forbidden.value = false;

    try {
      const data = await patternsService.fetchGlobalPatterns();
      const entry: GlobalCacheEntry = {
        patterns: data.global_patterns ?? [],
        fetchedAt: Date.now(),
      };
      globalCache.value = entry;
      applyGlobalEntry(entry);
    } catch (err: unknown) {
      if (err instanceof ApiError && err.isForbidden) {
        forbidden.value = true;
        error.value = "Недостаточно прав для просмотра шаблонов";
      } else {
        error.value = extractApiErrorMessage(err, "Не удалось загрузить глобальные шаблоны");
      }
      throw err;
    } finally {
      isLoading.value = false;
    }
  }

  async function fetchGroupPatterns(groupId: string, options?: { force?: boolean }): Promise<void> {
    const force = options?.force ?? false;
    activeGroupId.value = groupId;

    const cached = groupPatternsCache.value.get(groupId);
    if (!force && cached && Date.now() - cached.fetchedAt < CACHE_TTL_MS) {
      applyGroupEntry(cached);
      error.value = null;
      forbidden.value = false;
      return;
    }

    isLoading.value = true;
    error.value = null;
    forbidden.value = false;

    try {
      const data = await patternsService.fetchGroupPatterns(groupId);
      const entry: GroupPatternsCacheEntry = {
        global_patterns: data.global_patterns ?? [],
        group_patterns: data.group_patterns ?? [],
        fetchedAt: Date.now(),
      };
      groupPatternsCache.value.set(groupId, entry);
      applyGroupEntry(entry);
    } catch (err: unknown) {
      groupGlobalPatterns.value = [];
      groupLocalPatterns.value = [];
      if (err instanceof ApiError && err.isForbidden) {
        forbidden.value = true;
        error.value = "Недостаточно прав для просмотра шаблонов группы";
      } else if (err instanceof ApiError && err.isNotFound) {
        error.value = "Группа не найдена или недоступна";
      } else {
        error.value = extractApiErrorMessage(err, "Не удалось загрузить шаблоны группы");
      }
      throw err;
    } finally {
      isLoading.value = false;
    }
  }

  async function refreshActiveTab(activeTab: "global" | "group"): Promise<void> {
    if (activeTab === "global") {
      await fetchGlobalPatterns({ force: true });
      return;
    }
    if (activeGroupId.value) {
      await fetchGroupPatterns(activeGroupId.value, { force: true });
    }
  }

  async function createPattern(
    payload: CreatePatternRequest,
    activeTab: "global" | "group",
  ): Promise<PatternDto> {
    isMutating.value = true;
    try {
      const created = await patternsService.createPattern(payload);
      invalidateCache();
      await refreshActiveTab(activeTab);
      if (activeTab === "group" && payload.group_id) {
        activeGroupId.value = payload.group_id;
      }
      return created;
    } finally {
      isMutating.value = false;
    }
  }

  async function updatePattern(
    patternId: string,
    payload: EditPatternRequest,
    activeTab: "global" | "group",
  ): Promise<PatternDto> {
    isMutating.value = true;
    try {
      const updated = await patternsService.updatePattern(patternId, payload);
      invalidateCache();
      await refreshActiveTab(activeTab);
      return updated;
    } finally {
      isMutating.value = false;
    }
  }

  async function deletePattern(
    patternId: string,
    activeTab: "global" | "group",
  ): Promise<void> {
    isMutating.value = true;
    try {
      await patternsService.deletePattern(patternId);
      invalidateCache();
      await refreshActiveTab(activeTab);
    } finally {
      isMutating.value = false;
    }
  }

  return {
    globalPatterns,
    groupGlobalPatterns,
    groupLocalPatterns,
    activeGroupId,
    isLoading,
    isMutating,
    error,
    forbidden,
    invalidateCache,
    fetchGlobalPatterns,
    fetchGroupPatterns,
    createPattern,
    updatePattern,
    deletePattern,
  };
});
