import { ref, watch } from "vue";

const isDark = ref(false);

export function useTheme() {
  const initTheme = () => {
    if (localStorage.getItem("theme") === "dark" || (!("theme" in localStorage) && window.matchMedia("(prefers-color-scheme: dark)").matches)) {
      isDark.value = true;
    } else {
      isDark.value = false;
    }
    applyTheme(isDark.value);
  };

  const toggleTheme = () => {
    isDark.value = !isDark.value;
  };

  const applyTheme = (dark: boolean) => {
    if (dark) {
      document.documentElement.classList.add("dark");
      localStorage.setItem("theme", "dark");
    } else {
      document.documentElement.classList.remove("dark");
      localStorage.setItem("theme", "light");
    }
  };

  watch(isDark, (newValue) => {
    applyTheme(newValue);
  });

  return {
    isDark,
    toggleTheme,
    initTheme,
  };
}
