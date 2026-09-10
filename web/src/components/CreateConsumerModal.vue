<script setup lang="ts">
import { ref } from 'vue'
import { ApiError, createConsumer } from '../api/client'
import { logout, storedToken } from '../stores/auth'
import type { Role } from '../types'
import { generateApikey } from '../utils/apikey'
import ModalBase from './ModalBase.vue'

const props = defineProps<{ roles: Role[] }>()
const emit = defineEmits<{ (e: 'close'): void; (e: 'created'): void }>()

const name = ref('')
const apikey = ref('')
const showApikey = ref(false)
const roleId = ref('')
const error = ref('')
const loading = ref(false)

function onGenerate(): void {
  apikey.value = generateApikey()
  showApikey.value = true
}

async function onSave(): Promise<void> {
  error.value = ''
  if (name.value.trim() === '') {
    error.value = 'El nombre no puede estar vacío'
    return
  }
  if (apikey.value.trim() === '') {
    error.value = 'La apikey no puede estar vacía'
    return
  }
  if (roleId.value === '') {
    error.value = 'Selecciona un rol'
    return
  }
  loading.value = true
  try {
    await createConsumer(storedToken.value, {
      name: name.value.trim(),
      apikey: apikey.value.trim(),
      roleId: roleId.value,
    })
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
  <ModalBase title="Nuevo consumer" @close="emit('close')">
    <form @submit.prevent="onSave">
      <label class="field">
        <span>Nombre</span>
        <input v-model="name" placeholder="mi-app" autocomplete="off" />
      </label>
      <label class="field">
        <span>Apikey</span>
        <div class="input-row">
          <input
            v-model="apikey"
            :type="showApikey ? 'text' : 'password'"
            placeholder="clave-secreta"
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
      <label class="field">
        <span>Rol</span>
        <select v-model="roleId">
          <option value="" disabled>— Seleccionar rol —</option>
          <option v-for="r in roles" :key="r.id" :value="r.id">{{ r.name }}</option>
        </select>
      </label>
      <p v-if="error" class="error">{{ error }}</p>
      <div class="row end">
        <button type="button" class="btn" @click="emit('close')">Cancelar</button>
        <button type="submit" class="btn primary" :disabled="loading">
          {{ loading ? 'Guardando…' : 'Crear consumer' }}
        </button>
      </div>
    </form>
  </ModalBase>
</template>