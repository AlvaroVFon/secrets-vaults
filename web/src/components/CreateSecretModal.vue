<script setup lang="ts">
import { ref } from 'vue'
import { ApiError, createSecret } from '../api/client'
import { logout, storedToken } from '../stores/auth'
import type { Consumer } from '../types'
import ModalBase from './ModalBase.vue'

const props = defineProps<{ consumer: Consumer }>()
const emit = defineEmits<{ (e: 'close'): void; (e: 'created'): void }>()

const key = ref('')
const value = ref('')
const isSecret = ref(true)
const error = ref('')
const loading = ref(false)

async function onSave(): Promise<void> {
  error.value = ''
  if (key.value.trim() === '' || value.value.trim() === '') {
    error.value = 'Key y value no pueden estar vacíos'
    return
  }
  loading.value = true
  try {
    await createSecret(storedToken.value, props.consumer.id, key.value.trim(), value.value, isSecret.value)
    emit('created')
  } catch (err) {
    if (err instanceof ApiError) {
      if (err.status === 401) {
        logout()
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
  <ModalBase title="Nuevo secret" @close="emit('close')">
    <form @submit.prevent="onSave">
      <p class="muted field-hint">
        Consumer: <strong>{{ consumer.name }}</strong>
      </p>
      <label class="field">
        <span>Key</span>
        <input v-model="key" placeholder="db.password" autocomplete="off" />
      </label>
      <label class="field">
        <span>Value</span>
        <input v-model="value" :type="isSecret ? 'password' : 'text'" placeholder="s3cret" autocomplete="off" />
      </label>
      <label class="checkbox-field">
        <input v-model="isSecret" type="checkbox" />
        <span>Es un secret (valor sensible)</span>
      </label>
      <p v-if="error" class="error">{{ error }}</p>
      <div class="row end">
        <button type="button" class="btn" @click="emit('close')">Cancelar</button>
        <button type="submit" class="btn primary" :disabled="loading">
          {{ loading ? 'Guardando…' : `Crear en ${consumer.name}` }}
        </button>
      </div>
    </form>
  </ModalBase>
</template>