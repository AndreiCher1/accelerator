import type { AdditionalPrompt } from "./patterns.types";

export function formatAdditionalPrompt(value: AdditionalPrompt | undefined): string {
  if (value === undefined || value === null) return "";
  if (typeof value === "string") return value;
  return JSON.stringify(value, null, 2);
}

export function parseAdditionalPrompt(raw: string): AdditionalPrompt | undefined {
  const trimmed = raw.trim();
  if (!trimmed) return undefined;

  let parsed: unknown;
  try {
    parsed = JSON.parse(trimmed);
  } catch {
    throw new Error("Дополнительный промпт должен быть валидным JSON (объект {}, массив [] или null)");
  }

  if (parsed === null) return null;
  if (Array.isArray(parsed) || (typeof parsed === "object" && parsed !== null)) {
    return parsed as AdditionalPrompt;
  }

  throw new Error("additional_prompt должен быть объектом {}, массивом [] или null");
}

export function isGlobalPattern(pattern: { group_id?: string }): boolean {
  return !pattern.group_id || pattern.group_id.trim() === "";
}
