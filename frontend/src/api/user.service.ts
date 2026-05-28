import { api } from "./api";
import { cleanPayload } from "./utils";
import type {
  AddCreatorRequest,
  AddCreatorResponse,
  EditUserRequest,
  GetUsersParams,
  GetUsersResponse,
  RegisterUserRequest,
  RegisterUserResponse,
  ResetPasswordResponse,
  UserDto,
} from "./user.types";

const ADMIN_USERS = "/api/v1/admin/users";
const CREATOR_REGISTRATION = "/api/v1/creator/registration";

export const userService = {
  registerCreator(payload: AddCreatorRequest): Promise<AddCreatorResponse> {
    return api
      .post<AddCreatorResponse>(CREATOR_REGISTRATION, payload)
      .then((r) => r.data);
  },

  getUsers(params: GetUsersParams): Promise<GetUsersResponse> {
    return api
      .get<GetUsersResponse>(`${ADMIN_USERS}/`, {
        params: { page: params.page, limit: params.limit },
      })
      .then((r) => r.data);
  },

  registerUser(payload: RegisterUserRequest): Promise<RegisterUserResponse> {
    return api
      .post<RegisterUserResponse>(`${ADMIN_USERS}/`, payload)
      .then((r) => r.data);
  },

  updateUser(userId: string, payload: EditUserRequest): Promise<UserDto> {
    const body = cleanPayload(payload, { stripEmptyStrings: true });
    return api.put<UserDto>(`${ADMIN_USERS}/${userId}`, body).then((r) => r.data);
  },

  resetPassword(userId: string): Promise<ResetPasswordResponse> {
    return api
      .post<ResetPasswordResponse>(`${ADMIN_USERS}/${userId}/reset-password`)
      .then((r) => r.data);
  },

  deleteUser(userId: string): Promise<void> {
    return api.delete<void>(`${ADMIN_USERS}/${userId}`).then(() => undefined);
  },
};
