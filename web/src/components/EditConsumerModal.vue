<script setup lang="ts">
import { ref } from 'vue'
import { ApiError, updateConsumer } from '../api/client'
import { logout, storedToken } from '../stores/auth'
import type { Consumer, Role } from '../types'
import { generateApikey } from '../utils/apikey'
import ModalBase from './ModalBase.vue'

const props = defineProps<{ consumer: Consumer; roles: Role[] }>()
const emit = defineEmits<{ (e: 'close'): void; (e: 'updated'): void }>()

const name = ref(props.consumer.name)
const roleId = ref(props.consumer.roleId)
const showApikey = ref(false)
const newApikey = ref(props.consumer.apikey)
const error = ref('')
const loading = ref(false)

function onGenerate(): void {
  newApikey.value = generateApikey()
  showApikey.value = true
}

async function onSave(): Promise<void> {
  error.value = ''
  if (name.value.trim() === '') {
    error.value = 'El nombre no puede estar vacío'
    return
  }
  const patch: { name: string; roleId?: string; apikey?: string } = { name: name.value.trim() }
  if (roleId.value !== props.consumer.roleId) {
    patch.roleId = roleId.value
  }
  if (newApikey.value.trim() !== '' && newApikey.value !== props.consumer.apikey) {
    patch.apikey = newApikey.value.trim()
  }

  loading.value = true
  try {
    await updateConsumer(storedToken.value, props.consumer.id, patch)
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
  <ModalBase title="Editar consumer" @close="emit('close')">
    <form @submit.prevent="onSave">
      <label class="field">
        <span>Nombre</span>
        <input v-model="name" autocomplete="off" />
      </label>
      <label class="field">
        <span>Rol</span>
        <select v-model="roleId">
          <option v-for="r in roles" :key="r.id" :value="r.id">{{ r.name }}</option>
        </select>
      </label>
      <label class="field">
        <span>Apikey</span>
        <div class="input-row">
          <input
            v-model="newApikey"
            :type="showApikey ? 'text' : 'password'"
            autocomplete="off"
          />
          <button
            type="button"
            class="btn ghost"
            @click="showApikey = !showApikey"
            :aria-label="showApikey ? 'Ocultar' : 'Mostrar'"
          >
            {{ showApikey ? '🙈' : '👁️' }}
          </button>
          <button type="button" class="btn" @click="onGenerate">Generar</button>
        </div>
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