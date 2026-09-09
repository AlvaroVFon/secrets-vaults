# Secrets Vault — web

Panel de gestión de secrets. Vue 3 + TypeScript + Vite.

## Desarrollo

```bash
npm install
npm run dev    # proxy a la API en :8080, abre http://localhost:5173
```

La API debe estar levantada en `http://localhost:8080`.

## Login

El login pide la **apikey** de un consumer con rol `superadmin`. El seed
(`migrations/000005_seed_roles_and_superadmin.up.sql`) crea uno de prueba:

- apikey: `superadmin-api-key`

## Producción

```bash
npm run build  # genera web/dist
```

En el stack conjunto (`docker compose up`), el servicio `web` sirve `dist/`
con nginx y proxya `/secrets*` y `/management*` a `api:8080` (mismo origen,
sin CORS).
