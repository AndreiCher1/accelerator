/**
 * Лёгкий «мостик» между axios-интерсептором и auth-store: позволяет
 * обновлять роль пользователя в памяти, не создавая циклической
 * зависимости между api-слоем и Pinia-стором.
 */

type RoleSetter = (role: string | null) => void;
type LogoutHandler = () => void;

let roleSetter: RoleSetter | null = null;
let logoutHandler: LogoutHandler | null = null;

export function registerRoleSetter(fn: RoleSetter): void {
  roleSetter = fn;
}

export function applyRole(role: string | null | undefined): void {
  if (!roleSetter) return;
  roleSetter(role ?? null);
}

export function registerLogoutHandler(fn: LogoutHandler): void {
  logoutHandler = fn;
}

export function notifyLogout(): void {
  logoutHandler?.();
}
