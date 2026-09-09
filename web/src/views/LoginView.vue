<script setup lang="ts">
import { ref } from 'vue'
import { ApiError, login } from '../api/client'
import { login as setSession } from '../stores/auth'

const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function onSubmit(): Promise<void> {
  error.value = ''
  if (username.value.trim() === '') {
    error.value = 'Introduce tu usuario'
    return
  }
  if (password.value === '') {
    error.value = 'Introduce tu contraseña'
    return
  }
  loading.value = true
  try {
    const res = await login(username.value.trim(), password.value)
    setSession(res.token, res.username)
  } catch (err) {
    if (err instanceof ApiError) {
      if (err.status === 401) {
        error.value = 'Usuario o contraseña incorrectos'
      } else {
        error.value = err.message
      }
    } else {
      error.value = 'No se pudo conectar con la API'
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main class="login">
    <form class="login-card" @submit.prevent="onSubmit">
      <div class="login-brand">
        <div class="login-logo">🔐</div>
        <h1>Secrets Vault</h1>
        <p class="muted">Panel de gestión</p>
      </div>

      <label class="field">
        <span>Usuario</span>
        <input v-model="username" autocomplete="username" placeholder="admin" />
      </label>
      <label class="field">
        <span>Contraseña</span>
        <input
          v-model="password"
          type="password"
          autocomplete="current-password"
          placeholder="••••••••"
        />
      </label>

      <p v-if="error" class="error">{{ error }}</p>
      <button type="submit" class="btn primary" :disabled="loading">
        {{ loading ? 'Entrando…' : 'Entrar' }}
      </button>
    </form>
  </main>
</template>