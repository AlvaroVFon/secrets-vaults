<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  ApiError,
  deleteConsumer,
  deleteSecret,
  fetchConsumers,
  fetchRoles,
  fetchSecrets,
  updateConsumer,
} from '../api/client'
import { currentUsername, logout, storedToken } from '../stores/auth'
import type { Consumer, Role, Secret } from '../types'
import ConfirmDeleteModal from '../components/ConfirmDeleteModal.vue'
import CreateConsumerModal from '../components/CreateConsumerModal.vue'
import CreateSecretModal from '../components/CreateSecretModal.vue'
import EditConsumerModal from '../components/EditConsumerModal.vue'
import EditSecretModal from '../components/EditSecretModal.vue'
import GenerateConfigModal from '../components/GenerateConfigModal.vue'

const consumers = ref<Consumer[]>([])
const roles = ref<Role[]>([])
const selected = ref<Consumer | null>(null)
const secrets = ref<Secret[]>([])
const loading = ref(false)
const loadingSecrets = ref(false)
const error = ref('')
const visible = ref<Set<string>>(new Set())

const showCreateConsumer = ref(false)
const editingConsumer = ref<Consumer | null>(null)
const deletingConsumer = ref<Consumer | null>(null)

const showCreateSecret = ref(false)
const editingSecret = ref<Secret | null>(null)
const deletingSecret = ref<Secret | null>(null)
const showConfig = ref(false)

const roleName = computed(() => {
  const map = new Map(roles.value.map((r) => [r.id, r.name]))
  return (id: string): string => map.get(id) ?? ''
})

function handleError(err: unknown): void {
  if (err instanceof ApiError) {
    if (err.status === 401) {
      onLogout()
      return
    }
    error.value = err.message
  } else {
    error.value = 'No se pudo conectar con la API'
  }
}

async function loadConsumers(): Promise<void> {
  loading.value = true
  error.value = ''
  try {
    const [allConsumers, allRoles] = await Promise.all([
      fetchConsumers(storedToken.value),
      fetchRoles(storedToken.value),
    ])
    consumers.value = allConsumers
    roles.value = allRoles
    if (selected.value) {
      const still = allConsumers.find((c) => c.id === selected.value?.id)
      selected.value = still ?? null
    }
  } catch (err) {
    handleError(err)
  } finally {
    loading.value = false
  }
}

async function select(consumer: Consumer): Promise<void> {
  selected.value = consumer
  secrets.value = []
  error.value = ''
  visible.value = new Set()
  showConfig.value = false
  await loadSecrets(consumer.id)
}

async function loadSecrets(consumerId: string): Promise<void> {
  loadingSecrets.value = true
  try {
    secrets.value = await fetchSecrets(storedToken.value, consumerId)
  } catch (err) {
    handleError(err)
  } finally {
    loadingSecrets.value = false
  }
}

function toggleVisible(id: string): void {
  const next = new Set(visible.value)
  if (next.has(id)) {
    next.delete(id)
  } else {
    next.add(id)
  }
  visible.value = next
}

async function copy(value: string): Promise<void> {
  try {
    await navigator.clipboard.writeText(value)
  } catch {
    error.value = 'No se pudo copiar al portapapeles'
  }
}

async function onToggleActive(): Promise<void> {
  if (!selected.value) return
  try {
    await updateConsumer(storedToken.value, selected.value.id, { active: !selected.value.active })
    await loadConsumers()
  } catch (err) {
    handleError(err)
  }
}

async function onDeleteConsumer(): Promise<void> {
  if (!deletingConsumer.value) return
  const id = deletingConsumer.value.id
  try {
    await deleteConsumer(storedToken.value, id)
    deletingConsumer.value = null
    if (selected.value?.id === id) {
      selected.value = null
      secrets.value = []
    }
    await loadConsumers()
  } catch (err) {
    deletingConsumer.value = null
    handleError(err)
  }
}

async function onDeleteSecret(): Promise<void> {
  if (!deletingSecret.value) return
  try {
    await deleteSecret(storedToken.value, deletingSecret.value.id)
    deletingSecret.value = null
    await loadSecrets(selected.value?.id ?? '')
  } catch (err) {
    deletingSecret.value = null
    handleError(err)
  }
}

function onLogout(): void {
  logout()
  consumers.value = []
  roles.value = []
  selected.value = null
  secrets.value = []
}

onMounted(loadConsumers)
</script>

<template>
  <div class="shell">
    <aside class="sidebar">
      <div class="brand">
        <span class="brand-logo">🔐</span>
        <span>Secrets Vault</span>
      </div>

      <div class="sidebar-scroll">
        <p class="nav-label">Consumers</p>
        <ul class="nav-list">
          <li v-for="c in consumers" :key="c.id">
            <button
              class="nav-item"
              :class="{ active: selected?.id === c.id }"
              @click="select(c)"
              :title="c.id"
            >
              <span class="dot" :class="c.active ? 'on' : 'off'"></span>
              <span class="nav-name">{{ c.name }}</span>
              <span v-if="roleName(c.roleId)" class="nav-role">{{ roleName(c.roleId) }}</span>
            </button>
          </li>
        </ul>
        <p v-if="!loading && consumers.length === 0" class="muted empty-sidebar">
          Sin consumers todavía.
        </p>
      </div>

      <button class="btn primary sidebar-add" @click="showCreateConsumer = true">
        ＋ Nuevo consumer
      </button>
    </aside>

    <main class="content">
      <header class="topbar">
        <div v-if="selected" class="topbar-info">
          <h1>{{ selected.name }}</h1>
          <span class="badge" :class="selected.active ? 'on' : 'off'">
            {{ selected.active ? 'Activo' : 'Inactivo' }}
          </span>
          <span class="badge">{{ roleName(selected.roleId) }}</span>
        </div>
        <h1 v-else>Panel de gestión</h1>

        <div class="topbar-actions">
          <span class="muted user-label">{{ currentUsername }}</span>
          <button class="btn ghost" @click="onLogout">Salir</button>
        </div>
      </header>

      <p v-if="error" class="error">{{ error }}</p>

      <template v-if="selected">
        <div class="panel-head">
          <p class="muted">
            {{ secrets.length }} secret{{ secrets.length === 1 ? '' : 's' }}
          </p>
          <div class="row">
            <button class="btn" @click="onToggleActive">
              {{ selected.active ? 'Desactivar' : 'Activar' }}
            </button>
            <button class="btn" @click="editingConsumer = selected">Editar consumer</button>
            <button class="btn danger" @click="deletingConsumer = selected">Eliminar</button>
            <button class="btn" @click="showConfig = true">Generar config</button>
            <button class="btn primary" @click="showCreateSecret = true">＋ Nuevo secret</button>
          </div>
        </div>

        <p v-if="loadingSecrets" class="muted">Cargando secrets…</p>

        <div v-else-if="secrets.length === 0" class="empty-state">
          <p class="muted">Este consumer no tiene secrets todavía.</p>
        </div>

        <div v-else class="secrets-table">
          <div class="table-row table-head">
            <span>Key</span>
            <span>Value</span>
            <span class="table-actions">Acciones</span>
          </div>
          <div v-for="secret in secrets" :key="secret.id" class="table-row">
            <div class="key-cell">
              <code class="key">{{ secret.key }}</code>
              <span class="badge" :class="secret.isSecret ? 'secret' : 'config'">
                {{ secret.isSecret ? 'secret' : 'config' }}
              </span>
            </div>
            <code class="value">{{ !secret.isSecret || visible.has(secret.id) ? secret.value : '••••••••' }}</code>
            <div class="table-actions row">
              <button
                v-if="secret.isSecret"
                class="btn ghost icon"
                @click="toggleVisible(secret.id)"
                :aria-label="visible.has(secret.id) ? 'Ocultar valor' : 'Mostrar valor'"
                :title="visible.has(secret.id) ? 'Ocultar valor' : 'Mostrar valor'"
              >
                {{ visible.has(secret.id) ? '🙈' : '👁️' }}
              </button>
              <button
                class="btn ghost icon"
                @click="copy(secret.value)"
                aria-label="Copiar"
                title="Copiar"
              >
                📋
              </button>
              <button class="btn ghost" @click="editingSecret = secret">Editar</button>
              <button class="btn ghost danger-text" @click="deletingSecret = secret">Eliminar</button>
            </div>
          </div>
        </div>
      </template>

      <div v-else class="empty-state welcome">
        <div class="welcome-icon">🔐</div>
        <h2>Selecciona un consumer</h2>
        <p class="muted">
          Elige un consumer de la barra lateral para ver y gestionar sus secretos.
        </p>
      </div>
    </main>
  </div>

  <CreateConsumerModal
    v-if="showCreateConsumer"
    :roles="roles"
    @close="showCreateConsumer = false"
    @created="showCreateConsumer = false; loadConsumers()"
  />
  <EditConsumerModal
    v-if="editingConsumer"
    :consumer="editingConsumer!"
    :roles="roles"
    @close="editingConsumer = null"
    @updated="editingConsumer = null; loadConsumers()"
  />
  <ConfirmDeleteModal
    v-if="deletingConsumer"
    title="Eliminar consumer"
    :message="`¿Eliminar el consumer «${deletingConsumer.name}»? Esta acción no se puede deshacer.`"
    @close="deletingConsumer = null"
    @confirm="onDeleteConsumer"
  />

  <CreateSecretModal
    v-if="showCreateSecret && selected"
    :consumer="selected"
    @close="showCreateSecret = false"
    @created="showCreateSecret = false; loadSecrets(selected.id)"
  />
  <EditSecretModal
    v-if="editingSecret"
    :secret="editingSecret"
    @close="editingSecret = null"
    @updated="editingSecret = null; loadSecrets(selected!.id)"
  />
  <ConfirmDeleteModal
    v-if="deletingSecret"
    title="Eliminar secret"
    :message="`¿Eliminar el secret «${deletingSecret.key}»?`"
    @close="deletingSecret = null"
    @confirm="onDeleteSecret"
  />

  <GenerateConfigModal
    v-if="showConfig && selected"
    :consumer="selected"
    @close="showConfig = false"
  />
</template>
