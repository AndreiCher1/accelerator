/** Стандартный формат ошибки бекенда: `{ "error": "..." }` */
export interface ApiErrorResponse {
  error: string;
}

export type HttpStatusCode = 400 | 401 | 403 | 404 | 409 | 500;

/** Роли в системе (из domains/user и JWT) */
export type UserRole = "creator" | "admin" | "user";

/** Роли, доступные при регистрации админом */
export type AssignableUserRole = "admin" | "user";

export interface Pagination {
  page: number;
  limit: number;
  total: number;
}

export class ApiError extends Error {
  readonly status: number;
  readonly isApiError = true;
  readonly original?: unknown;

  constructor(message: string, status: number, original?: unknown) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.original = original;
  }

  get isUnauthorized(): boolean {
    return this.status === 401;
  }

  get isForbidden(): boolean {
    return this.status === 403;
  }

  get isNotFound(): boolean {
    return this.status === 404;
  }

  get isConflict(): boolean {
    return this.status === 409;
  }
}
