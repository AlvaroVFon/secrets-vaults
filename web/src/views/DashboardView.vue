<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Braces, Copy, Eye, EyeOff, LogOut, Pencil, Plus, Search, ShieldCheck, Trash2, User, X } from '@lucide/vue'
import { toast } from 'vue-sonner'
import {
  ApiError,
  deleteConsumer,
  deleteSecret,
  fetchConsumers,
  fetchRoles,
  fetchSecrets,
  updateConsumer,
} from '@/api/client'
import { currentUsername, logout, storedToken } from '@/stores/auth'
import type { Consumer, Role, Secret } from '@/types'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import ConfirmDeleteModal from '@/components/ConfirmDeleteModal.vue'
import CreateConsumerModal from '@/components/CreateConsumerModal.vue'
import CreateSecretModal from '@/components/CreateSecretModal.vue'
import EditConsumerModal from '@/components/EditConsumerModal.vue'
import EditSecretModal from '@/components/EditSecretModal.vue'
import GenerateConfigModal from '@/components/GenerateConfigModal.vue'
import ThemeToggle from '@/components/ThemeToggle.vue'

const consumers = ref<Consumer[]>([])
const roles = ref<Role[]>([])
const selected = ref<Consumer | null>(null)
const secrets = ref<Secret[]>([])
const loading = ref(false)
const loadingSecrets = ref(false)
const error = ref('')
const visible = ref<Set<string>>(new Set())
const search = ref('')

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

const filteredSecrets = computed(() => {
  const q = search.value.trim().toLowerCase()
  if (!q) return secrets.value
  return secrets.value.filter(
    (s) => s.key.toLowerCase().includes(q) || s.value.toLowerCase().includes(q),
  )
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
  search.value = ''
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
    toast.success('Copiado al portapapeles')
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
    toast.success('Consumer eliminado')
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
    toast.success('Secret eliminado')
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
  <div class="flex h-svh w-full overflow-hidden bg-background">
    <aside class="flex w-64 min-w-64 flex-col border-r bg-sidebar text-sidebar-foreground">
      <div class="flex h-14 items-center gap-2 border-b px-4 text-base font-semibold">
        <ShieldCheck class="size-5 text-primary" />
        <span>Secrets Vault</span>
      </div>

      <div class="flex-1 overflow-y-auto p-3">
        <p class="px-2 pb-2 text-xs font-semibold tracking-wider text-muted-foreground uppercase">
          Consumers
        </p>
        <ul class="flex flex-col gap-1">
          <li v-for="c in consumers" :key="c.id">
            <button
              type="button"
              class="flex w-full items-center gap-2 rounded-md px-2.5 py-2 text-left text-sm transition-colors hover:bg-sidebar-accent"
              :class="{ 'bg-sidebar-accent': selected?.id === c.id }"
              @click="select(c)"
              :title="c.id"
            >
              <span
                class="size-2 shrink-0 rounded-full"
                :class="c.active ? 'bg-emerald-500' : 'bg-muted-foreground/40'"
              ></span>
              <span class="flex-1 truncate font-medium">{{ c.name }}</span>
              <Badge v-if="roleName(c.roleId)" variant="secondary" class="text-[10px]">
                {{ roleName(c.roleId) }}
              </Badge>
            </button>
          </li>
        </ul>
        <p
          v-if="!loading && consumers.length === 0"
          class="px-2 py-4 text-sm text-muted-foreground"
        >
          Sin consumers todavía.
        </p>
      </div>

      <div class="border-t p-3">
        <Button class="w-full" @click="showCreateConsumer = true">
          <Plus />
          Nuevo consumer
        </Button>
      </div>
    </aside>

    <main class="flex min-w-0 flex-1 flex-col overflow-hidden">
      <header class="flex h-14 shrink-0 items-center justify-between gap-4 border-b px-6">
        <div v-if="selected" class="flex min-w-0 items-center gap-2">
          <h1 class="truncate text-base font-semibold">{{ selected.name }}</h1>
          <Badge :variant="selected.active ? 'default' : 'secondary'">
            {{ selected.active ? 'Activo' : 'Inactivo' }}
          </Badge>
          <Badge v-if="roleName(selected.roleId)" variant="outline">
            {{ roleName(selected.roleId) }}
          </Badge>
        </div>
        <h1 v-else class="text-base font-semibold">Panel de gestión</h1>

        <div class="flex items-center gap-1">
          <DropdownMenu>
            <DropdownMenuTrigger as-child>
              <Button variant="ghost" size="sm">
                <User />
                <span class="hidden sm:inline">{{ currentUsername }}</span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem variant="destructive" @select="onLogout">
                <LogOut />
                Salir
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
          <ThemeToggle />
        </div>
      </header>

      <div class="flex-1 overflow-y-auto p-6">
        <p
          v-if="error"
          class="mb-4 rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive"
        >
          {{ error }}
        </p>

        <template v-if="selected">
          <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
            <p class="text-sm text-muted-foreground">
              {{ filteredSecrets.length }} secret{{ filteredSecrets.length === 1 ? '' : 's' }}
              <span v-if="search.trim()"> de {{ secrets.length }}</span>
            </p>
            <div class="flex flex-wrap items-center gap-2">
              <Button variant="outline" size="sm" @click="onToggleActive">
                {{ selected.active ? 'Desactivar' : 'Activar' }}
              </Button>
              <Button variant="outline" size="sm" @click="editingConsumer = selected">
                <Pencil />
                Editar consumer
              </Button>
              <Button
                variant="outline"
                size="sm"
                class="text-destructive hover:text-destructive"
                @click="deletingConsumer = selected"
              >
                <Trash2 />
                Eliminar
              </Button>
              <Button variant="outline" size="sm" @click="showConfig = true">
                <Braces />
                Generar config
              </Button>
              <Button size="sm" @click="showCreateSecret = true">
                <Plus />
                Nuevo secret
              </Button>
            </div>
          </div>

          <div v-if="secrets.length > 0" class="relative mb-4 max-w-sm">
            <Search class="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              v-model="search"
              type="search"
              class="pl-9"
              placeholder="Buscar secret por key o valor…"
              aria-label="Buscar secrets"
            />
            <Button
              v-if="search"
              type="button"
              variant="ghost"
              size="icon-sm"
              class="absolute top-1/2 right-1 -translate-y-1/2"
              aria-label="Limpiar búsqueda"
              @click="search = ''"
            >
              <X />
            </Button>
          </div>

          <div v-if="loadingSecrets" class="space-y-2 rounded-lg border p-4">
            <Skeleton class="h-9 w-full" />
            <Skeleton class="h-9 w-full" />
            <Skeleton class="h-9 w-full" />
          </div>

          <template v-else>
            <div
              v-if="secrets.length === 0"
              class="rounded-lg border border-dashed px-6 py-12 text-center"
            >
              <ShieldCheck class="mx-auto mb-3 size-8 text-muted-foreground" />
              <p class="text-sm text-muted-foreground">Este consumer no tiene secrets todavía.</p>
            </div>

            <div
              v-else-if="filteredSecrets.length === 0"
              class="rounded-lg border border-dashed px-6 py-12 text-center"
            >
              <p class="text-sm text-muted-foreground">
                No hay secrets que coincidan con «{{ search.trim() }}».
              </p>
            </div>

            <div v-else class="overflow-hidden rounded-lg border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Key</TableHead>
                    <TableHead>Value</TableHead>
                    <TableHead class="text-right">Acciones</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  <TableRow v-for="secret in filteredSecrets" :key="secret.id">
                    <TableCell>
                      <div class="flex items-center gap-2">
                        <code class="text-sm font-medium text-primary">{{ secret.key }}</code>
                        <Badge :variant="secret.isSecret ? 'default' : 'secondary'">
                          {{ secret.isSecret ? 'secret' : 'config' }}
                        </Badge>
                      </div>
                    </TableCell>
                    <TableCell class="whitespace-normal">
                      <code class="text-sm break-all text-muted-foreground">
                        {{ !secret.isSecret || visible.has(secret.id) ? secret.value : '••••••••' }}
                      </code>
                    </TableCell>
                    <TableCell>
                      <div class="flex justify-end gap-1">
                        <Button
                          v-if="secret.isSecret"
                          variant="ghost"
                          size="icon-sm"
                          @click="toggleVisible(secret.id)"
                          :aria-label="visible.has(secret.id) ? 'Ocultar valor' : 'Mostrar valor'"
                          :title="visible.has(secret.id) ? 'Ocultar valor' : 'Mostrar valor'"
                        >
                          <EyeOff v-if="visible.has(secret.id)" />
                          <Eye v-else />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon-sm"
                          aria-label="Copiar"
                          title="Copiar"
                          @click="copy(secret.value)"
                        >
                          <Copy />
                        </Button>
                        <Button variant="ghost" size="sm" @click="editingSecret = secret">
                          Editar
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          class="text-destructive hover:text-destructive"
                          @click="deletingSecret = secret"
                        >
                          Eliminar
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </div>
          </template>
        </template>

        <div
          v-else
          class="flex flex-col items-center justify-center rounded-lg border border-dashed px-6 py-20 text-center"
        >
          <ShieldCheck class="mb-3 size-10 text-primary" />
          <h2 class="text-lg font-semibold">Selecciona un consumer</h2>
          <p class="mt-1 text-sm text-muted-foreground">
            Elige un consumer de la barra lateral para ver y gestionar sus secretos.
          </p>
        </div>
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
