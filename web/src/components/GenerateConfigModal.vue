<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import hljs from 'highlight.js/lib/core'
import go from 'highlight.js/lib/languages/go'
import typescript from 'highlight.js/lib/languages/typescript'
import 'highlight.js/styles/github-dark.css'
import { ApiError, fetchConsumerConfig } from '../api/client'
import { logout, storedToken } from '../stores/auth'
import type { ConfigLanguage, Consumer } from '../types'
import ModalBase from './ModalBase.vue'

hljs.registerLanguage('typescript', typescript)
hljs.registerLanguage('go', go)

const props = defineProps<{ consumer: Consumer }>()
const emit = defineEmits<{ (e: 'close'): void }>()

const language = ref<ConfigLanguage>('ts')
const code = ref('')
const filename = ref('')
const loading = ref(false)
const error = ref('')
const copied = ref(false)

const highlighted = computed(() =>
  code.value
    ? hljs.highlight(code.value, { language: language.value === 'ts' ? 'typescript' : 'go' }).value
    : '',
)

let copiedTimer: ReturnType<typeof setTimeout> | undefined

async function load(): Promise<void> {
  loading.value = true
  error.value = ''
  copied.value = false
  try {
    const result = await fetchConsumerConfig(storedToken.value, props.consumer.id, language.value)
    code.value = result.code
    filename.value = result.filename
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

async function copy(): Promise<void> {
  try {
    await navigator.clipboard.writeText(code.value)
    copied.value = true
    clearTimeout(copiedTimer)
    copiedTimer = setTimeout(() => {
      copied.value = false
    }, 1500)
  } catch {
    error.value = 'No se pudo copiar al portapapeles'
  }
}

watch(language, load)
onMounted(load)
</script>

<template>
  <ModalBase class="config-modal" :title="`Config de ${consumer.name}`" @close="emit('close')">
    <div class="config-toolbar">
      <div class="segmented" role="tablist" aria-label="Lenguaje">
        <button
          type="button"
          class="segmented-item"
          :class="{ active: language === 'ts' }"
          @click="language = 'ts'"
        >
          TypeScript
        </button>
        <button
          type="button"
          class="segmented-item"
          :class="{ active: language === 'go' }"
          @click="language = 'go'"
        >
          Go
        </button>
      </div>

      <div class="row">
        <span v-if="filename" class="muted filename">{{ filename }}</span>
        <button type="button" class="btn" :disabled="!code || loading" @click="copy">
          {{ copied ? 'Copiado ✓' : 'Copiar' }}
        </button>
      </div>
    </div>

    <p v-if="loading" class="muted">Generando…</p>
    <p v-else-if="error" class="error">{{ error }}</p>
    <pre v-else class="config-preview"><code class="hljs" v-html="highlighted"></code></pre>
  </ModalBase>
</template>
