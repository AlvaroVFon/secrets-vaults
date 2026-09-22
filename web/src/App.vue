<script setup lang="ts">
import { isLoggedIn } from './stores/auth'
import LoginView from './views/LoginView.vue'
import DashboardView from './views/DashboardView.vue'
import { Toaster } from '@/components/ui/sonner'
</script>

<template>
  <DashboardView v-if="isLoggedIn" />
  <LoginView v-else />
  <Toaster position="top-right" rich-colors />
</template>
