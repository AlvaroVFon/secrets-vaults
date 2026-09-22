# Secrets Vault — web

Panel de gestión de secrets. Vue 3 + TypeScript + Vite, con **Tailwind CSS v4**
y **shadcn-vue** (Reka UI) para los componentes.

## Desarrollo

```bash
npm install
npm run dev    # proxy a la API en :8080, abre http://localhost:5173
```

La API debe estar levantada en `http://localhost:8080`.

## Login

El login pide **usuario y contraseña** del panel (`ADMIN_USERNAME` /
`ADMIN_PASSWORD`, ver README raíz). El usuario se crea en el primer arranque de
la API si la tabla de usuarios está vacía.

## Componentes UI

Los componentes de shadcn-vue se copian al repo en `src/components/ui/` y se
importan con el alias `@/`:

```vue
<script setup lang="ts">
import { Button } from '@/components/ui/button'
</script>
```

Para añadir más componentes:

```bash
npx shadcn-vue@latest add <componente>
```

La configuración (aliases, color base, ruta del CSS) está en `components.json`.

## Tema

Modo claro/oscuro mediante la clase `.dark` en `<html>`, gestionada en
`src/stores/theme.ts` y persistida en `localStorage`. Los tokens de color
(OKLCH) viven en `src/style.css`.

## Producción

```bash
npm run build  # genera web/dist
```

En el stack conjunto (`docker compose up`), el servicio `web` sirve `dist/`
con nginx y proxya `/secrets*` y `/management*` a `api:8080` (mismo origen,
sin CORS).
