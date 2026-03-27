import { computed, ref } from 'vue';

export type ThemeMode = 'light' | 'dark' | 'system';

const STORAGE_KEY = 'latex-qapp-theme';
const themeMode = ref<ThemeMode>('system');
const activeTheme = ref<'light' | 'dark'>('light');
let mediaListenerBound = false;

function resolveMode(mode: ThemeMode) {
  if (mode !== 'system') {
    return mode;
  }

  if (typeof window === 'undefined') {
    return 'light';
  }

  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

function applyMode(mode: ThemeMode) {
  if (typeof document === 'undefined') {
    return;
  }

  const resolved = resolveMode(mode);
  activeTheme.value = resolved;
  const root = document.documentElement;
  root.classList.toggle('dark', resolved === 'dark');
  root.setAttribute('data-theme-mode', mode);
}

function handleSystemThemeChange() {
  if (themeMode.value === 'system') {
    applyMode('system');
  }
}

function bindSystemListener() {
  if (mediaListenerBound || typeof window === 'undefined') {
    return;
  }

  const media = window.matchMedia('(prefers-color-scheme: dark)');
  media.addEventListener('change', handleSystemThemeChange);

  mediaListenerBound = true;
}

function initTheme() {
  if (typeof window === 'undefined') {
    return;
  }

  const saved = window.localStorage.getItem(STORAGE_KEY) as ThemeMode | null;
  if (saved === 'light' || saved === 'dark' || saved === 'system') {
    themeMode.value = saved;
  }

  applyMode(themeMode.value);
  bindSystemListener();
}

function setTheme(mode: ThemeMode) {
  themeMode.value = mode;
  if (typeof window !== 'undefined') {
    window.localStorage.setItem(STORAGE_KEY, mode);
  }
  applyMode(mode);
}

const resolvedTheme = computed(() => activeTheme.value);

export function useTheme() {
  return {
    themeMode,
    resolvedTheme,
    initTheme,
    setTheme,
  };
}
