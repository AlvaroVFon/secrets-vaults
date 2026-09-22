<script setup lang="ts">
import { ref } from 'vue'
import { Eye, EyeOff } from '@lucide/vue'
import { ApiError, updateSecret } from '@/api/client'
import { logout, storedToken } from '@/stores/auth'
import type { Secret } from '@/types'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
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
  <Dialog :open="true" @update:open="(v) => !v && emit('close')">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>Editar secret</DialogTitle>
        <DialogDescription>Actualiza el valor o el tipo del secret.</DialogDescription>
      </DialogHeader>

      <form class="grid gap-4" @submit.prevent="onSave">
        <div class="grid gap-2">
          <Label for="edit-secret-key">Key</Label>
          <Input id="edit-secret-key" v-model="key" autocomplete="off" />
        </div>

        <div class="grid gap-2">
          <Label for="edit-secret-value">Value</Label>
          <div class="relative">
            <Input
              id="edit-secret-value"
              v-model="value"
              :type="isSecret && !showValue ? 'password' : 'text'"
              autocomplete="off"
              :class="isSecret ? 'pr-10' : ''"
            />
            <Button
              v-if="isSecret"
              type="button"
              variant="ghost"
              size="icon-sm"
              class="absolute top-1/2 right-1 -translate-y-1/2"
              @click="showValue = !showValue"
              :aria-label="showValue ? 'Ocultar' : 'Mostrar'"
            >
              <EyeOff v-if="showValue" />
              <Eye v-else />
            </Button>
          </div>
        </div>

        <label class="flex items-center gap-2 text-sm">
          <Checkbox v-model="isSecret" />
          <span class="text-muted-foreground">Es un secret (valor sensible)</span>
        </label>

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
