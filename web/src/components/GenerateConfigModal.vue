<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { Check, Copy } from '@lucide/vue'
import hljs from 'highlight.js/lib/core'
import go from 'highlight.js/lib/languages/go'
import typescript from 'highlight.js/lib/languages/typescript'
import { ApiError, fetchConsumerConfig } from '@/api/client'
import { logout, storedToken } from '@/stores/auth'
import type { ConfigLanguage, Consumer } from '@/types'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'

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
  <Dialog :open="true" @update:open="(v) => !v && emit('close')">
    <DialogContent class="sm:max-w-3xl">
      <DialogHeader>
        <DialogTitle>Config de {{ consumer.name }}</DialogTitle>
        <DialogDescription>Genera el fichero de configuración desde sus secrets.</DialogDescription>
      </DialogHeader>

      <div class="flex flex-wrap items-center justify-between gap-3">
        <Tabs v-model="language">
          <TabsList>
            <TabsTrigger value="ts">TypeScript</TabsTrigger>
            <TabsTrigger value="go">Go</TabsTrigger>
          </TabsList>
        </Tabs>

        <div class="flex items-center gap-3">
          <span v-if="filename" class="font-mono text-xs text-muted-foreground">{{ filename }}</span>
          <Button variant="outline" :disabled="!code || loading" @click="copy">
            <Check v-if="copied" />
            <Copy v-else />
            {{ copied ? 'Copiado' : 'Copiar' }}
          </Button>
        </div>
      </div>

      <p v-if="loading" class="py-6 text-center text-sm text-muted-foreground">Generando…</p>
      <p v-else-if="error" class="text-sm text-destructive">{{ error }}</p>
      <pre
        v-else
        class="config-preview max-h-[60vh] overflow-auto rounded-md border bg-muted/30 p-4"
      ><code class="hljs text-xs" v-html="highlighted"></code></pre>
    </DialogContent>
  </Dialog>
</template>
