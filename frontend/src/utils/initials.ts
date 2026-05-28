/**
 * Возвращает инициалы из ФИО.
 * Берётся первая буква фамилии и первая буква имени.
 */
export function getInitials(fullName: string | null | undefined): string {
  if (!fullName) return "?";

  const parts = fullName
    .trim()
    .split(/\s+/)
    .filter((part) => part.length > 0);

  if (parts.length === 0) return "?";

  const first = parts[0]?.charAt(0) ?? "";
  const second = parts[1]?.charAt(0) ?? "";

  const initials = `${first}${second}`.toUpperCase();
  return initials || "?";
}
