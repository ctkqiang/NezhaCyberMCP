type Theme = "dark" | "light";

const STORAGE_KEY = "nezha-theme";

const theme = ref<Theme>("dark");

function applyTheme(t: Theme) {
  if (import.meta.client) {
    document.documentElement.setAttribute("data-theme", t);
  }
}

export function useTheme() {
  function init() {
    if (!import.meta.client) return;

    const stored = localStorage.getItem(STORAGE_KEY) as Theme | null;
    if (stored === "dark" || stored === "light") {
      theme.value = stored;
    } else {
      const prefersDark = window.matchMedia("(prefers-color-scheme: dark)").matches;
      theme.value = prefersDark ? "dark" : "light";
    }
    applyTheme(theme.value);

    window
      .matchMedia("(prefers-color-scheme: dark)")
      .addEventListener("change", (e) => {
        if (!localStorage.getItem(STORAGE_KEY)) {
          theme.value = e.matches ? "dark" : "light";
          applyTheme(theme.value);
        }
      });
  }

  function toggle() {
    theme.value = theme.value === "dark" ? "light" : "dark";
    applyTheme(theme.value);
    if (import.meta.client) {
      localStorage.setItem(STORAGE_KEY, theme.value);
    }
  }

  function setTheme(t: Theme) {
    theme.value = t;
    applyTheme(t);
    if (import.meta.client) {
      localStorage.setItem(STORAGE_KEY, t);
    }
  }

  const isDark = computed(() => theme.value === "dark");

  return { theme: readonly(theme), isDark, toggle, setTheme, init };
}
