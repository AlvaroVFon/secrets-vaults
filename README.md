# Secrets Vault

## Objetivo del proyecto

El objetivo del proyecto es tener un vault de secretos para almacenar la configuración
sensible de los diferentes proyectos alojados en un servidor local.
Algo similar a `AWS Secrets Manager`

## Componentes

- Carga de configuraciones desde `config.go`

- Database para integration test dedicada, preparar setup y teardown. Mantener aislamiento entre test.

- Los secretos de este proyecto serán gestionados mediante variables de entorno
  usando un módulo de configuración

- Será necesarión agregar una Entidad `Roles`, para poder limitar el acceso a
  la api de management

  ```go
  type Role struct {
    ID string
    Role string
    Permissions []string
  }
  ```

- Entidad `Consumer`, es quien, mediante api-key, podrá acceder a los diferentes
  secretos almacenados únicamente para ese consumer.

```go
type Consumer struct {
  Name
  ClientID string
  ClientSecret string
  ApiKeyHash string
  Active bool
  Role Role
}
```

- Entidad `Secrets`. Es la entidad que almacenará los secretos por `Consumer`

```go
type Secret struct {
  Key string // único por consumer
  Value string
  Active bool
  ConsumerID string
}
```

- API básica para gestión global de secretos (admin panel) con CRUD básico.

## Decisiones técnicas

- BDD: PostgreSQL con `pgx v5`
- Usaremos `golang-migrate` para las migraciones
- Para la API, usaremos únicamente `stdlib`
