<script setup lang="ts">
import { Moon, Sun } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { currentTheme, toggleTheme } from '@/stores/theme'
</script>

<template>
  <Button
    variant="ghost"
    size="icon"
    type="button"
    @click="toggleTheme"
    :aria-label="currentTheme === 'dark' ? 'Activar tema claro' : 'Activar tema oscuro'"
    :title="currentTheme === 'dark' ? 'Tema claro' : 'Tema oscuro'"
  >
    <Sun v-if="currentTheme === 'dark'" />
    <Moon v-else />
  </Button>
</template>
