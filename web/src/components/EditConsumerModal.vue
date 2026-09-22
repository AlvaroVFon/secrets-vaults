<script setup lang="ts">
import { ref } from 'vue'
import { Eye, EyeOff, Wand2 } from '@lucide/vue'
import { ApiError, updateConsumer } from '@/api/client'
import { logout, storedToken } from '@/stores/auth'
import type { Consumer, Role } from '@/types'
import { generateApikey } from '@/utils/apikey'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

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
  <Dialog :open="true" @update:open="(v) => !v && emit('close')">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>Editar consumer</DialogTitle>
        <DialogDescription>Modifica el nombre, el rol o la apikey.</DialogDescription>
      </DialogHeader>

      <form class="grid gap-4" @submit.prevent="onSave">
        <div class="grid gap-2">
          <Label for="edit-consumer-name">Nombre</Label>
          <Input id="edit-consumer-name" v-model="name" autocomplete="off" />
        </div>

        <div class="grid gap-2">
          <Label>Rol</Label>
          <Select v-model="roleId">
            <SelectTrigger class="w-full">
              <SelectValue placeholder="Seleccionar rol" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="r in roles" :key="r.id" :value="r.id">{{ r.name }}</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div class="grid gap-2">
          <Label for="edit-consumer-apikey">Apikey</Label>
          <div class="flex gap-2">
            <Input
              id="edit-consumer-apikey"
              v-model="newApikey"
              :type="showApikey ? 'text' : 'password'"
              autocomplete="off"
            />
            <Button
              type="button"
              variant="ghost"
              size="icon"
              @click="showApikey = !showApikey"
              :aria-label="showApikey ? 'Ocultar' : 'Mostrar'"
            >
              <EyeOff v-if="showApikey" />
              <Eye v-else />
            </Button>
            <Button type="button" variant="outline" @click="onGenerate">
              <Wand2 />
              Generar
            </Button>
          </div>
        </div>

        <p v-if="error" class="text-sm text-destructive">{{ error }}</p>

        <DialogFooter>
          <Button type="button" variant="outline" @click="emit('close')">Cancelar</Button>
          <Button type="submit" :disabled="loading">
            {{ loading ? 'Guardando…' : 'Guardar' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
