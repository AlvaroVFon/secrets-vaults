import { computed, ref } from 'vue';

export type Theme = 'light' | 'dark';

const STORAGE_KEY = 'secrets-vault-theme';

function storedTheme(): Theme | null {
  const value = localStorage.getItem(STORAGE_KEY);
  return value === 'light' || value === 'dark' ? value : null;
}

function systemTheme(): Theme {
  return window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark';
}

const theme = ref<Theme>(storedTheme() ?? systemTheme());

export const currentTheme = computed(() => theme.value);

export function applyTheme(value: Theme): void {
  theme.value = value;
  document.documentElement.dataset.theme = value;
  document.documentElement.classList.toggle('dark', value === 'dark');
  localStorage.setItem(STORAGE_KEY, value);
}

export function toggleTheme(): void {
  applyTheme(theme.value === 'dark' ? 'light' : 'dark');
}
