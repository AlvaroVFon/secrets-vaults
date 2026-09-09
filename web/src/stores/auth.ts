import { computed, ref } from 'vue';

const TOKEN_KEY = 'secrets-vault-token';
const USERNAME_KEY = 'secrets-vault-username';

const token = ref<string>(localStorage.getItem(TOKEN_KEY) ?? '');
const username = ref<string>(localStorage.getItem(USERNAME_KEY) ?? '');

export const isLoggedIn = computed(() => token.value !== '');
export const storedToken = computed(() => token.value);
export const currentUsername = computed(() => username.value);

export function login(newToken: string, user: string): void {
  token.value = newToken;
  username.value = user;
  localStorage.setItem(TOKEN_KEY, newToken);
  localStorage.setItem(USERNAME_KEY, user);
}

export function logout(): void {
  token.value = '';
  username.value = '';
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(USERNAME_KEY);
}