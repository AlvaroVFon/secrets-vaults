<script setup lang="ts">
import { ref } from 'vue'
import { Eye, EyeOff } from '@lucide/vue'
import { ApiError, createSecret } from '@/api/client'
import { logout, storedToken } from '@/stores/auth'
import type { Consumer } from '@/types'
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

const props = defineProps<{ consumer: Consumer }>()
const emit = defineEmits<{ (e: 'close'): void; (e: 'created'): void }>()

const key = ref('')
const value = ref('')
const isSecret = ref(false)
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
    await createSecret(storedToken.value, props.consumer.id, key.value.trim(), value.value, isSecret.value)
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
        <DialogTitle>Nuevo secret</DialogTitle>
        <DialogDescription>
          Consumer: <span class="font-medium text-foreground">{{ consumer.name }}</span>
        </DialogDescription>
      </DialogHeader>

      <form class="grid gap-4" @submit.prevent="onSave">
        <div class="grid gap-2">
          <Label for="secret-key">Key</Label>
          <Input id="secret-key" v-model="key" placeholder="db.password" autocomplete="off" />
        </div>

        <div class="grid gap-2">
          <Label for="secret-value">Value</Label>
          <div class="relative">
            <Input
              id="secret-value"
              v-model="value"
              :type="isSecret && !showValue ? 'password' : 'text'"
              placeholder="s3cret"
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
            {{ loading ? 'Guardando…' : `Crear en ${consumer.name}` }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
