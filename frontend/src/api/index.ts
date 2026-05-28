export { api, API_URL } from "./api";
export { ApiError, type ApiErrorResponse, type Pagination, type UserRole, type AssignableUserRole } from "./api.types";
export {
  cleanPayload,
  extractApiErrorMessage,
  getStoredAccessToken,
  getStoredRefreshToken,
  setStoredTokens,
  clearStoredTokens,
} from "./utils";

export { authService } from "./auth.service";
export type * from "./auth.types";

export { userService } from "./user.service";
export type * from "./user.types";

export { groupService } from "./group.service";
export type * from "./group.types";

export { patternsService } from "./patterns.service";
export type * from "./patterns.types";
export { formatAdditionalPrompt, parseAdditionalPrompt, isGlobalPattern } from "./patterns.utils";

export { tasksService } from "./tasks.service";
export type * from "./tasks.types";

/** @deprecated Используйте `authService` */
export { authService as AuthService } from "./auth.service";
/** @deprecated Используйте `tasksService` */
export { tasksService as TasksService } from "./tasks.service";
