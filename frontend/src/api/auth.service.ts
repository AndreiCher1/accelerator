import { api } from "./api";
import type {
  AuthTokensResponse,
  ChangeTempPasswordRequest,
  LoginRequest,
  RefreshTokenRequest,
} from "./auth.types";

const AUTH_PREFIX = "/api/v1/auth";
const USERS_ME_PREFIX = "/api/v1/users/me";

export const authService = {
  login(payload: LoginRequest): Promise<AuthTokensResponse> {
    return api.post<AuthTokensResponse>(`${AUTH_PREFIX}/login`, payload).then((r) => r.data);
  },

  refresh(payload: RefreshTokenRequest): Promise<AuthTokensResponse> {
    return api.post<AuthTokensResponse>(`${AUTH_PREFIX}/refresh`, payload).then((r) => r.data);
  },

  changeTempPassword(payload: ChangeTempPasswordRequest): Promise<void> {
    return api
      .post<void>(`${USERS_ME_PREFIX}/change-temp-password`, payload)
      .then(() => undefined);
  },
};
