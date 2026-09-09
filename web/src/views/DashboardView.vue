<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ApiError, deleteSecret, fetchConsumers, fetchGroupedSecrets } from '../api/client'
import { logout, storedApikey } from '../stores/auth'
import type { Consumer, ConsumerSecrets, Secret } from '../types'
import ConfirmDeleteModal from '../components/ConfirmDeleteModal.vue'
import CreateSecretModal from '../components/CreateSecretModal.vue'
import EditSecretModal from '../components/EditSecretModal.vue'

const groups = ref<ConsumerSecrets[]>([])
const consumers = ref<Consumer[]>([])
const loading = ref(false)
const error = ref('')
const visible = ref<Set<string>>(new Set())
const showCreate = ref(false)
const preselect = ref('')
const editing = ref<Secret | null>(null)
const deleting = ref<Secret | null>(null)

interface ConsumerSection {
  consumer: Consumer
  secrets: Secret[]
}

const sections = computed<ConsumerSection[]>(() => {
  const byId = new Map<string, Secret[]>()
  for (const g of groups.value) {
    byId.set(g.consumerId, g.secrets)
  }
  const known = new Set(consumers.value.map((c) => c.id))
  const list: ConsumerSection[] = consumers.value.map((c) => ({
    consumer: c,
    secrets: byId.get(c.id) ?? [],
  }))
  for (const g of groups.value) {
    if (!known.has(g.consumerId)) {
      list.push({
        consumer: { id: g.consumerId, name: g.consumerId },
        secrets: g.secrets,
      })
    }
  }
  return list
})

async function load(): Promise<void> {
  loading.value = true
  error.value = ''
  try {
    const [allConsumers, grouped] = await Promise.all([
      fetchConsumers(storedApikey.value),
      fetchGroupedSecrets(storedApikey.value),
    ])
    consumers.value = allConsumers
    groups.value = grouped
  } catch (err) {
    if (err instanceof ApiError) {
      if (err.status === 401) {
        logout()
        return
      }
      error.value = err.message
    } else {
      error.value = 'No se pudo conectar con la API'
    }
  } finally {
    loading.value = false
  }
}

function toggle(id: string): void {
  if (visible.value.has(id)) {
    visible.value.delete(id)
  } else {
    visible.value.add(id)
  }
}

function shortId(id: string): string {
  return id.length > 13 ? `${id.slice(0, 8)}…${id.slice(-4)}` : id
}

function openCreate(consumerId = ''): void {
  preselect.value = consumerId
  showCreate.value = true
}

function closeCreate(): void {
  showCreate.value = false
  preselect.value = ''
}

async function copy(value: string): Promise<void> {
  try {
    await navigator.clipboard.writeText(value)
  } catch {
    error.value = 'No se pudo copiar al portapapeles'
  }
}

async function onDelete(): Promise<void> {
  if (!deleting.value) return
  try {
    await deleteSecret(storedApikey.value, deleting.value.id)
    deleting.value = null
    await load()
  } catch (err) {
    deleting.value = null
    if (err instanceof ApiError) {
      if (err.status === 401) {
        logout()
        return
      }
      error.value = err.message
    } else {
      error.value = 'No se pudo conectar con la API'
    }
  }
}

function onLogout(): void {
  logout()
  groups.value = []
  consumers.value = []
}

onMounted(load)
</script>

<template>
  <main class="page wide">
    <header class="topbar">
      <h1>Secrets Vault</h1>
      <div class="row">
        <button class="btn primary" @click="openCreate()">＋ Nuevo secret</button>
        <button class="btn" @click="onLogout">Salir</button>
      </div>
    </header>

    <p v-if="loading" class="muted">Cargando…</p>
    <p v-if="error" class="error">{{ error }}</p>

    <section v-for="section in sections" :key="section.consumer.id" class="card">
      <div class="section-head">
        <h2 class="consumer" :title="section.consumer.id">
          {{ section.consumer.name }}
          <code>{{ shortId(section.consumer.id) }}</code>
          <span class="count">{{ section.secrets.length }}</span>
        </h2>
        <button class="btn ghost" @click="openCreate(section.consumer.id)" title="Nuevo secret en este consumer">＋</button>
      </div>
      <ul v-if="section.secrets.length > 0" class="secrets">
        <li v-for="secret in section.secrets" :key="secret.id" class="secret-row">
          <div class="secret-main">
            <code class="key">{{ secret.key }}</code>
            <code class="value">{{ visible.has(secret.id) ? secret.value : '••••••••' }}</code>
          </div>
          <div class="row">
            <button class="btn ghost" @click="toggle(secret.id)">
              {{ visible.has(secret.id) ? '🙈' : '👁️' }}
            </button>
            <button class="btn ghost" @click="copy(secret.value)">📋</button>
            <button class="btn" @click="editing = secret">Editar</button>
            <button class="btn danger" @click="deleting = secret">Eliminar</button>
          </div>
        </li>
      </ul>
      <p v-else class="muted">Sin secrets.</p>
    </section>

    <p v-if="!loading && sections.length === 0 && !error" class="muted">Sin consumers todavía.</p>

    <CreateSecretModal
      v-if="showCreate"
      :consumers="consumers"
      :preselect="preselect"
      @close="closeCreate"
      @created="closeCreate(); load()"
    />
    <EditSecretModal
      v-if="editing"
      :secret="editing"
      @close="editing = null"
      @updated="editing = null; load()"
    />
    <ConfirmDeleteModal
      v-if="deleting"
      :secret-key="deleting.key"
      @close="deleting = null"
      @confirm="onDelete"
    />
  </main>
</template>
