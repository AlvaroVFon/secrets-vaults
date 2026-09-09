import { computed, ref } from 'vue';

const STORAGE_KEY = 'secrets-vault-apikey';

const apikey = ref<string>(localStorage.getItem(STORAGE_KEY) ?? '');

export const isLoggedIn = computed(() => apikey.value !== '');
export const storedApikey = computed(() => apikey.value);

export function login(key: string): void {
  apikey.value = key.trim();
  localStorage.setItem(STORAGE_KEY, apikey.value);
}

export function logout(): void {
  apikey.value = '';
  localStorage.removeItem(STORAGE_KEY);
}
