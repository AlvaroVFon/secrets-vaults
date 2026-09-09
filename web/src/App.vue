<script setup lang="ts">
import { isLoggedIn } from './stores/auth'
import LoginView from './views/LoginView.vue'
import DashboardView from './views/DashboardView.vue'
</script>

<template>
  <DashboardView v-if="isLoggedIn" />
  <LoginView v-else />
</template>
