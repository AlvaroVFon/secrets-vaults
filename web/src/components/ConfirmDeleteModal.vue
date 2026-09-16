<script setup lang="ts">
import { X } from '@lucide/vue'

defineProps<{ title: string; message: string }>()
const emit = defineEmits<{ (e: 'close'): void; (e: 'confirm'): void }>()
</script>

<template>
  <div class="backdrop" @click.self="emit('close')">
    <div class="modal" role="dialog" aria-modal="true">
      <div class="modal-head">
        <h2>{{ title }}</h2>
        <button class="btn ghost icon" @click="emit('close')" aria-label="Cerrar">
          <X :size="18" />
        </button>
      </div>
      <div class="modal-body">
        <p>{{ message }}</p>
        <div class="row end">
          <button class="btn" @click="emit('close')">Cancelar</button>
          <button class="btn danger" @click="emit('confirm')">Eliminar</button>
        </div>
      </div>
    </div>
  </div>
</template>