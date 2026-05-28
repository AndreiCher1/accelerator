import type { AssignableUserRole, Pagination, UserRole } from "./api.types";

export interface UserDto {
  user_id: string;
  login: string;
  full_name: string;
  position: string;
  role: UserRole;
  created_at: string;
}

export interface GetUsersResponse {
  users: UserDto[];
  pagination: Pagination;
}

export interface RegisterUserRequest {
  login: string;
  full_name: string;
  position: string;
  role: AssignableUserRole;
}

export interface RegisterUserResponse extends UserDto {
  password: string;
}

export interface EditUserRequest {
  login?: string;
  full_name?: string;
  position?: string;
  role?: AssignableUserRole;
}

export interface ResetPasswordResponse {
  password: string;
}

export interface AddCreatorRequest {
  login: string;
  full_name: string;
  position: string;
  password: string;
}

export interface AddCreatorResponse {
  user_id: string;
  login: string;
  full_name: string;
  position: string;
  role: UserRole;
}

export interface GetUsersParams {
  page: number;
  limit: number;
}
