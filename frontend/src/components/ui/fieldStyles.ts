export const fieldWrapperClass = "flex w-full flex-col gap-1.5";

export const fieldLabelClass =
  "text-sm font-medium leading-none text-gray-700 dark:text-gray-300";

type ControlState = {
  error?: boolean;
  disabled?: boolean;
};

export function fieldControlClass(state: ControlState = {}): string {
  const { error = false, disabled = false } = state;

  const base = [
    "w-full rounded-lg border text-sm leading-normal transition-colors",
    "placeholder:text-gray-400 dark:placeholder:text-gray-500",
    "focus:outline-none focus:ring-2 focus:ring-offset-0",
  ];

  if (disabled) {
    base.push(
      "cursor-not-allowed border-gray-200 bg-gray-50 text-gray-500",
      "dark:border-dark-border dark:bg-dark-card/50 dark:text-gray-500",
    );
    return base.join(" ");
  }

  if (error) {
    base.push(
      "border-red-300 bg-red-50 text-red-900",
      "focus:border-red-500 focus:ring-red-500/20",
      "dark:border-red-600 dark:bg-red-900/20 dark:text-red-100",
    );
    return base.join(" ");
  }

  base.push(
    "border-gray-200 bg-white text-gray-900",
    "hover:border-gray-300",
    "focus:border-blue-500 focus:ring-blue-500/20",
    "dark:border-dark-border dark:bg-dark-card dark:text-white",
    "dark:hover:border-gray-500",
    "dark:focus:border-gray-400 dark:focus:ring-gray-400/20",
  );

  return base.join(" ");
}

/** Однострочные поля: input, trigger select */
export const fieldControlSizeClass = "h-10 px-3.5";

/** Многострочные поля */
export const fieldTextareaSizeClass = "min-h-[5.5rem] px-3.5 py-2.5 resize-y";

export const fieldSelectMenuClass =
  "absolute z-50 mt-2 w-full overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xl ring-1 ring-black/5 dark:border-white/10 dark:bg-dark-elevated dark:shadow-black/40 dark:ring-white/5";

export const fieldSelectMenuScrollClass = "max-h-60 overflow-y-auto";

export const fieldSelectOptionClass =
  "flex w-full items-center justify-between gap-2 px-3.5 py-2.5 text-left text-sm transition-colors duration-150";
