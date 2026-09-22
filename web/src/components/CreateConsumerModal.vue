<script setup lang="ts">
import { ref } from 'vue'
import { Eye, EyeOff, Wand2 } from '@lucide/vue'
import { ApiError, createConsumer } from '@/api/client'
import { logout, storedToken } from '@/stores/auth'
import type { Role } from '@/types'
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

defineProps<{ roles: Role[] }>()
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
  <Dialog :open="true" @update:open="(v) => !v && emit('close')">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>Nuevo consumer</DialogTitle>
        <DialogDescription>Crea un consumer con su apikey y rol.</DialogDescription>
      </DialogHeader>

      <form class="grid gap-4" @submit.prevent="onSave">
        <div class="grid gap-2">
          <Label for="consumer-name">Nombre</Label>
          <Input id="consumer-name" v-model="name" placeholder="mi-app" autocomplete="off" />
        </div>

        <div class="grid gap-2">
          <Label for="consumer-apikey">Apikey</Label>
          <div class="flex gap-2">
            <Input
              id="consumer-apikey"
              v-model="apikey"
              :type="showApikey ? 'text' : 'password'"
              placeholder="clave-secreta"
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

        <p v-if="error" class="text-sm text-destructive">{{ error }}</p>

        <DialogFooter>
          <Button type="button" variant="outline" @click="emit('close')">Cancelar</Button>
          <Button type="submit" :disabled="loading">
            {{ loading ? 'Guardando…' : 'Crear consumer' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
