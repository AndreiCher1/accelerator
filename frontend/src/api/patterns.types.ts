/** JSON-поле additional_prompt: объект, массив или null */
export type AdditionalPrompt = Record<string, unknown> | unknown[] | null;

export type PatternScope = "global" | "group";

export interface PatternDto {
  pattern_id: string;
  group_id?: string;
  name: string;
  description: string;
  summary_prompt: string;
  additional_prompt: AdditionalPrompt;
  created_at: string;
  change_flag: boolean;
}

export interface CreatePatternRequest {
  name: string;
  description: string;
  summary_prompt: string;
  additional_prompt?: AdditionalPrompt;
  /** Пусто или не передаётся — глобальный паттерн */
  group_id?: string;
}

export interface EditPatternRequest {
  name?: string;
  description?: string;
  summary_prompt?: string;
  additional_prompt?: AdditionalPrompt;
}

export interface GroupPatternsResponse {
  global_patterns: PatternDto[];
  group_patterns: PatternDto[];
}

export interface GlobalPatternsResponse {
  global_patterns: PatternDto[];
}

export interface GroupWithPatternsDto {
  group_id: string;
  name: string;
  description: string;
  patterns: PatternDto[];
}

export interface AllGroupPatternsResponse {
  groups: GroupWithPatternsDto[];
}

export interface PatternFormValues {
  patternId: string;
  name: string;
  description: string;
  summaryPrompt: string;
  additionalPromptJson: string;
  scope: PatternScope;
  groupId: string;
}
