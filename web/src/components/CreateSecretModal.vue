<script setup lang="ts">
import { ref } from 'vue'
import { ApiError, createSecret } from '../api/client'
import { logout, storedApikey } from '../stores/auth'
import type { Consumer } from '../types'
import ModalBase from './ModalBase.vue'

const props = defineProps<{ consumers: Consumer[]; preselect?: string }>()
const emit = defineEmits<{ (e: 'close'): void; (e: 'created'): void }>()

const consumerId = ref(props.preselect ?? '')
const key = ref('')
const value = ref('')
const error = ref('')
const loading = ref(false)

function consumerName(id: string): string {
  return props.consumers.find((c) => c.id === id)?.name ?? id
}

async function onSave(): Promise<void> {
  error.value = ''
  if (consumerId.value === '') {
    error.value = 'Selecciona un consumer'
    return
  }
  if (key.value.trim() === '' || value.value.trim() === '') {
    error.value = 'Key y value no pueden estar vacíos'
    return
  }
  loading.value = true
  try {
    await createSecret(storedApikey.value, consumerId.value, key.value.trim(), value.value)
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
      <label class="field">
        <span>Consumer</span>
        <select v-model="consumerId">
          <option value="" disabled>— Seleccionar consumer —</option>
          <option v-for="c in consumers" :key="c.id" :value="c.id">
            {{ c.name }}
          </option>
        </select>
      </label>
      <label class="field">
        <span>Key</span>
        <input v-model="key" placeholder="db.password" autocomplete="off" />
      </label>
      <label class="field">
        <span>Value</span>
        <input v-model="value" type="password" placeholder="s3cret" autocomplete="off" />
      </label>
      <p v-if="error" class="error">{{ error }}</p>
      <div class="row end">
        <button type="button" class="btn" @click="emit('close')">Cancelar</button>
        <button type="submit" class="btn primary" :disabled="loading">
          {{ loading ? 'Guardando…' : `Crear en ${consumerName(consumerId) || '…'}` }}
        </button>
      </div>
    </form>
  </ModalBase>
</template>
