import axios, {
  type AxiosError,
  type AxiosResponse,
  type InternalAxiosRequestConfig,
} from "axios";
import { ApiError, type ApiErrorResponse } from "./api.types";
import {
  clearStoredTokens,
  getStoredAccessToken,
  getStoredRefreshToken,
  rememberMeFromStorage,
  setStoredTokens,
} from "./utils";
import { applyRole, notifyLogout } from "./auth-bridge";
import type { AuthTokensResponse } from "./auth.types";

const API_URL = import.meta.env.VITE_API_URL ?? "";

export const api = axios.create({
  baseURL: API_URL,
  headers: {
    "Content-Type": "application/json",
  },
});

type RetryableRequestConfig = InternalAxiosRequestConfig & { _retry?: boolean };

let isRefreshing = false;
let failedQueue: Array<{
  resolve: (token: string) => void;
  reject: (error: unknown) => void;
}> = [];

function processQueue(error: unknown, token: string | null = null): void {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error);
    } else if (token) {
      prom.resolve(token);
    }
  });
  failedQueue = [];
}

function toApiError(error: AxiosError<ApiErrorResponse>): ApiError {
  const status = error.response?.status ?? 0;
  const message =
    error.response?.data?.error ??
    error.message ??
    "Произошла ошибка при выполнении запроса";

  return new ApiError(message, status, error);
}

function isAuthEndpoint(url: string | undefined): boolean {
  if (!url) return false;
  return url.includes("/auth/login") || url.includes("/auth/refresh");
}

api.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = getStoredAccessToken();
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error),
);

api.interceptors.response.use(
  (response: AxiosResponse) => response,
  async (error: AxiosError<ApiErrorResponse>) => {
    const originalRequest = error.config as RetryableRequestConfig | undefined;
    const refreshToken = getStoredRefreshToken();

    if (
      error.response?.status === 401 &&
      originalRequest &&
      !originalRequest._retry &&
      !isAuthEndpoint(originalRequest.url) &&
      refreshToken
    ) {
      if (isRefreshing) {
        return new Promise<string>((resolve, reject) => {
          failedQueue.push({ resolve, reject });
        }).then((token) => {
          originalRequest.headers.Authorization = `Bearer ${token}`;
          return api(originalRequest);
        });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        const { data } = await axios.post<AuthTokensResponse>(
          `${API_URL}/api/v1/auth/refresh`,
          { refresh_token: refreshToken },
        );

        const newRefresh = data.refresh_token ?? refreshToken;
        setStoredTokens(data.access_token, newRefresh, rememberMeFromStorage());
        // Роль приходит с каждым refresh — обновляем её в памяти стора.
        if (data.user_role) {
          applyRole(data.user_role);
        }

        processQueue(null, data.access_token);

        originalRequest.headers.Authorization = `Bearer ${data.access_token}`;
        return api(originalRequest);
      } catch (refreshError) {
        processQueue(refreshError, null);
        clearStoredTokens();
        notifyLogout();
        window.location.href = "/login";
        return Promise.reject(refreshError);
      } finally {
        isRefreshing = false;
      }
    }

    if (error.response?.status === 401 && !isAuthEndpoint(originalRequest?.url)) {
      clearStoredTokens();
      notifyLogout();
      window.location.href = "/login";
      return Promise.reject(toApiError(error));
    }

    return Promise.reject(toApiError(error));
  },
);

export { API_URL };
