<script setup lang="ts">
import { Moon, Sun } from '@lucide/vue'
import { currentTheme, toggleTheme } from '../stores/theme'
</script>

<template>
  <button
    class="btn ghost icon theme-toggle"
    type="button"
    @click="toggleTheme"
    :aria-label="currentTheme === 'dark' ? 'Activar tema claro' : 'Activar tema oscuro'"
    :title="currentTheme === 'dark' ? 'Tema claro' : 'Tema oscuro'"
  >
    <Sun v-if="currentTheme === 'dark'" :size="18" />
    <Moon v-else :size="18" />
  </button>
</template>
