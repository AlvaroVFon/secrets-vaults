<script setup lang="ts">
import { ref } from 'vue'
import { ApiError, fetchGroupedSecrets } from '../api/client'
import { login } from '../stores/auth'

const key = ref('')
const error = ref('')
const loading = ref(false)

async function onSubmit(): Promise<void> {
  error.value = ''
  if (key.value.trim() === '') {
    error.value = 'Introduce tu apikey'
    return
  }
  loading.value = true
  try {
    await fetchGroupedSecrets(key.value.trim())
    login(key.value)
  } catch (err) {
    if (err instanceof ApiError) {
      if (err.status === 403) {
        error.value = 'Esta apikey no tiene rol superadmin'
      } else if (err.status === 401 || err.status === 404) {
        error.value = 'Apikey inválida'
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
  <main class="page">
    <div class="card">
      <h1>Secrets Vault</h1>
      <p class="muted">Acceso de gestión. Necesitas una apikey de un consumer con rol superadmin.</p>
      <form @submit.prevent="onSubmit">
        <label class="field">
          <span>Apikey</span>
          <input
            v-model="key"
            type="password"
            placeholder="superadmin-api-key"
            autocomplete="off"
          />
        </label>
        <p v-if="error" class="error">{{ error }}</p>
        <button type="submit" class="btn primary" :disabled="loading">
          {{ loading ? 'Entrando…' : 'Entrar' }}
        </button>
      </form>
    </div>
  </main>
</template>
