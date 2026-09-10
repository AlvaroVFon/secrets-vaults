<script setup lang="ts">
import { ref } from 'vue'
import { ApiError, updateSecret } from '../api/client'
import { logout, storedToken } from '../stores/auth'
import type { Secret } from '../types'
import ModalBase from './ModalBase.vue'

const props = defineProps<{ secret: Secret }>()
const emit = defineEmits<{ (e: 'close'): void; (e: 'updated'): void }>()

const key = ref(props.secret.key)
const value = ref(props.secret.value)
const isSecret = ref(props.secret.isSecret)
const showValue = ref(false)
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
    await updateSecret(storedToken.value, props.secret.id, {
      key: key.value.trim(),
      value: value.value,
      isSecret: isSecret.value,
    })
    emit('updated')
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
  <ModalBase title="Editar secret" @close="emit('close')">
    <form @submit.prevent="onSave">
      <label class="field">
        <span>Key</span>
        <input v-model="key" autocomplete="off" />
      </label>
      <label class="field">
        <span>Value</span>
        <div class="input-row">
          <input v-model="value" :type="isSecret && !showValue ? 'password' : 'text'" autocomplete="off" />
          <button
            v-if="isSecret"
            type="button"
            class="btn ghost"
            @click="showValue = !showValue"
            :aria-label="showValue ? 'Ocultar' : 'Mostrar'"
          >
            {{ showValue ? '🙈' : '👁️' }}
          </button>
        </div>
      </label>
      <label class="checkbox-field">
        <input v-model="isSecret" type="checkbox" />
        <span>Es un secret (valor sensible)</span>
      </label>
      <p v-if="error" class="error">{{ error }}</p>
      <div class="row end">
        <button type="button" class="btn" @click="emit('close')">Cancelar</button>
        <button type="submit" class="btn primary" :disabled="loading">
          {{ loading ? 'Guardando…' : 'Guardar' }}
        </button>
      </div>
    </form>
  </ModalBase>
</template>
