<script setup lang="ts">
defineProps<{ title: string }>()
const emit = defineEmits<{ (e: 'close'): void }>()

function onBackdrop(e: MouseEvent): void {
  if ((e.target as HTMLElement).classList.contains('backdrop')) {
    emit('close')
  }
}

function onKey(e: KeyboardEvent): void {
  if (e.key === 'Escape') {
    emit('close')
  }
}
</script>

<template>
  <div class="backdrop" @click="onBackdrop" @keydown="onKey" tabindex="-1">
    <div class="modal" role="dialog" aria-modal="true">
      <div class="modal-head">
        <h2>{{ title }}</h2>
        <button class="btn ghost" @click="emit('close')" aria-label="Cerrar">✕</button>
      </div>
      <div class="modal-body">
        <slot />
      </div>
    </div>
  </div>
</template>
