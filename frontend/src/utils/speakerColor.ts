/**
 * Палитра пилюль для спикеров стенограммы. Цвет стабилен для одного и того же
 * имени, повторяется циклически.
 */

const SPEAKER_COLORS = [
  "bg-blue-100 text-blue-700 dark:bg-blue-500/20 dark:text-blue-300",
  "bg-green-100 text-green-700 dark:bg-green-500/20 dark:text-green-300",
  "bg-pink-100 text-pink-700 dark:bg-pink-500/20 dark:text-pink-300",
  "bg-purple-100 text-purple-700 dark:bg-purple-500/20 dark:text-purple-300",
  "bg-orange-100 text-orange-700 dark:bg-orange-500/20 dark:text-orange-300",
  "bg-teal-100 text-teal-700 dark:bg-teal-500/20 dark:text-teal-300",
  "bg-yellow-100 text-yellow-700 dark:bg-yellow-500/20 dark:text-yellow-300",
];

function hashString(value: string): number {
  let hash = 0;
  for (let i = 0; i < value.length; i++) {
    hash = (hash << 5) - hash + value.charCodeAt(i);
    hash |= 0;
  }
  return Math.abs(hash);
}

export function speakerPillClass(speaker: string): string {
  return SPEAKER_COLORS[hashString(speaker) % SPEAKER_COLORS.length];
}
