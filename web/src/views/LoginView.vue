<script setup lang="ts">
import { ref } from 'vue'
import { ShieldCheck } from '@lucide/vue'
import { ApiError, login } from '@/api/client'
import { login as setSession } from '@/stores/auth'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import ThemeToggle from '@/components/ThemeToggle.vue'

const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function onSubmit(): Promise<void> {
  error.value = ''
  if (username.value.trim() === '') {
    error.value = 'Introduce tu usuario'
    return
  }
  if (password.value === '') {
    error.value = 'Introduce tu contraseña'
    return
  }
  loading.value = true
  try {
    const res = await login(username.value.trim(), password.value)
    setSession(res.token, res.username)
  } catch (err) {
    if (err instanceof ApiError) {
      if (err.status === 401) {
        error.value = 'Usuario o contraseña incorrectos'
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
  <main class="relative flex min-h-svh items-center justify-center bg-background p-4">
    <div
      class="pointer-events-none absolute inset-x-0 top-0 h-[500px] bg-[radial-gradient(1100px_420px_at_50%_-20%,var(--primary),transparent_70%)] opacity-15"
    ></div>
    <div class="absolute top-4 right-4">
      <ThemeToggle />
    </div>

    <Card class="relative w-full max-w-sm">
      <CardHeader class="justify-items-center gap-2 text-center">
        <div class="flex size-12 items-center justify-center rounded-xl bg-primary/10 text-primary">
          <ShieldCheck class="size-6" />
        </div>
        <CardTitle class="text-xl">Secrets Vault</CardTitle>
        <CardDescription>Panel de gestión</CardDescription>
      </CardHeader>
      <CardContent>
        <form class="grid gap-4" @submit.prevent="onSubmit">
          <div class="grid gap-2">
            <Label for="username">Usuario</Label>
            <Input id="username" v-model="username" autocomplete="username" placeholder="admin" />
          </div>
          <div class="grid gap-2">
            <Label for="password">Contraseña</Label>
            <Input
              id="password"
              v-model="password"
              type="password"
              autocomplete="current-password"
              placeholder="••••••••"
            />
          </div>

          <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
          <Button type="submit" class="w-full" :disabled="loading">
            {{ loading ? 'Entrando…' : 'Entrar' }}
          </Button>
        </form>
      </CardContent>
    </Card>
  </main>
</template>
