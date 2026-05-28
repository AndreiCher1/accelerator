import type { AxiosError } from "axios";
import { ApiError, type ApiErrorResponse } from "./api.types";

const ACCESS_TOKEN_KEYS = ["access_token"] as const;
const REFRESH_TOKEN_KEYS = ["refresh_token"] as const;
const LEGACY_USER_ROLE_KEY = "user_role";

function readToken(storages: readonly Storage[], keys: readonly string[]): string | null {
  for (const storage of storages) {
    for (const key of keys) {
      const value = storage.getItem(key);
      if (value) return value;
    }
  }
  return null;
}

export function getStoredAccessToken(): string | null {
  return readToken([localStorage, sessionStorage], ACCESS_TOKEN_KEYS);
}

export function getStoredRefreshToken(): string | null {
  return readToken([localStorage, sessionStorage], REFRESH_TOKEN_KEYS);
}

export function rememberMeFromStorage(): boolean {
  return localStorage.getItem("access_token") !== null;
}

export function setStoredTokens(access: string, refresh: string, persistent: boolean): void {
  clearStoredTokens();
  const storage = persistent ? localStorage : sessionStorage;
  storage.setItem("access_token", access);
  storage.setItem("refresh_token", refresh);
}

export function clearStoredTokens(): void {
  for (const key of ACCESS_TOKEN_KEYS) {
    localStorage.removeItem(key);
    sessionStorage.removeItem(key);
  }
  for (const key of REFRESH_TOKEN_KEYS) {
    localStorage.removeItem(key);
    sessionStorage.removeItem(key);
  }
  // Подчищаем легаси-ключ, если он остался от старых версий приложения.
  localStorage.removeItem(LEGACY_USER_ROLE_KEY);
  sessionStorage.removeItem(LEGACY_USER_ROLE_KEY);
}

// Одноразовая миграция: если роль ещё лежит в браузерном хранилище от
// предыдущих версий — удаляем. Теперь роль приходит только из ответа
// /auth/login и /auth/refresh и хранится исключительно в памяти.
if (typeof window !== "undefined") {
  localStorage.removeItem(LEGACY_USER_ROLE_KEY);
  sessionStorage.removeItem(LEGACY_USER_ROLE_KEY);
}

export function extractApiErrorMessage(error: unknown, fallback = "Произошла ошибка"): string {
  if (error instanceof ApiError) {
    return error.message;
  }

  if (isAxiosErrorWithBody(error)) {
    const data = error.response?.data;
    if (data && typeof data.error === "string" && data.error.length > 0) {
      return data.error;
    }
  }

  if (error instanceof Error && error.message) {
    return error.message;
  }

  return fallback;
}

function isAxiosErrorWithBody(error: unknown): error is AxiosError<ApiErrorResponse> {
  return typeof error === "object" && error !== null && "isAxiosError" in error;
}

/**
 * Удаляет `undefined` из объекта перед отправкой.
 * Опционально убирает пустые строки (для PATCH/PUT, где пустая строка — ошибка на бекенде).
 */
export function cleanPayload<T extends object>(
  payload: T,
  options?: { stripEmptyStrings?: boolean },
): Partial<T> {
  const result = {} as Partial<T>;

  for (const key of Object.keys(payload) as Array<keyof T>) {
    const value = payload[key];
    if (value === undefined) continue;
    if (options?.stripEmptyStrings && value === "") continue;
    result[key] = value;
  }

  return result;
}
