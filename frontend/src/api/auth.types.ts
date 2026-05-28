import type { UserRole } from "./api.types";

export interface LoginRequest {
  login: string;
  password: string;
}

export interface RefreshTokenRequest {
  refresh_token: string;
}

export interface AuthTokensResponse {
  access_token: string;
  refresh_token?: string;
  /** Время истечения access-токена (ISO-строка). На бэкенде поле — `expires_at`. */
  expires_at?: string;
  token_type: string;
  /** Роль пользователя, возвращаемая бэкендом (поле `user_role`). */
  user_role?: UserRole;
  temporary_password?: boolean;
}

export interface ChangeTempPasswordRequest {
  password: string;
}
