# Snippet Service

Serviço backend para **armazenamento e busca de snippets de texto**, com suporte a metadados (nome, linguagem, tags), visibilidade (`public` / `private`) e busca fuzzy.  
Este projeto é desenvolvido em **Go**, utilizando **PostgreSQL**, **Docker** e **migrations versionadas**.

---

## Visão Geral

O objetivo do projeto é fornecer um **endpoint simples e eficiente** para:

- Criar snippets de texto
- Consultar snippets públicos
- Evoluir futuramente para autenticação e snippets privados
- Realizar buscas fuzzy por nome, metadados e conteúdo

O foco inicial é um **MVP funcional**, acessado via **Insomnia** ou `curl`.

---

## Stack Tecnológica

- **Linguagem:** Go (stdlib + `chi`)
- **Roteamento HTTP:** `github.com/go-chi/chi`
- **Banco de dados:** PostgreSQL 16
- **Busca:** PostgreSQL Full-Text Search + `pg_trgm`
- **Migrações:** `golang-migrate`
- **Containerização:** Docker + Docker Compose

---

## Arquitetura

```
API (Go + chi)
    |
    v
PostgreSQL
```

Um container separado é utilizado exclusivamente para executar as **migrations**.

---

## Estrutura do Projeto

```
.
├─ cmd/
│  └─ api/
│     └─ main.go
├─ internal/
│  ├─ db/
│  │  └─ db.go
│  ├─ httpapi/
│  │  ├─ router.go
│  │  ├─ handlers_health.go
│  │  └─ handlers_snippets.go
│  └─ snippets/
│     ├─ model.go
│     └─ repo_pg.go
├─ migrations/
│  ├─ 000001_init.up.sql
│  └─ 000001_init.down.sql
├─ Dockerfile
├─ docker-compose.yml
├─ go.mod
├─ go.sum
└─ README.md
```

---

## Banco de Dados

### Entidades

#### users
- Representa usuários do sistema
- No MVP existe um usuário fixo `usr_demo`

#### snippets
- Conteúdo textual do snippet
- Metadados: linguagem e tags
- Visibilidade: `public` ou `private`
- Relacionado a um usuário

---

## Busca

- **Fuzzy search no nome:** `pg_trgm`
- **Busca textual:** Full-Text Search do PostgreSQL
- Campos indexados: `name`, `content`, `tags`

---

## Migrations

O schema do banco é controlado por **migrations versionadas** usando `golang-migrate`.

- Cada alteração estrutural gera uma nova migration
- A API nunca cria ou altera tabelas diretamente
- O histórico é mantido na tabela `schema_migrations`

---

## Docker

### Serviços

| Serviço | Função |
|-------|-------|
| db | PostgreSQL |
| migrate | Executa migrations |
| api | Servidor HTTP Go |

### Subir o banco

```
docker compose up -d db
```

### Rodar migrations

```bash
docker compose run --rm migrate \
  -source file:///migrations \
  -database=postgres://snippet:snippet@db:5432/snippet?sslmode=disable \
  up
```

### Subir a API

```
docker compose up -d api
```

---

## API

### Health Check

```
GET /health
```

### Criar snippet

```
POST /v1/snippets
```

Body:
```json
{
  "name": "Exemplo",
  "content": "print('hello')",
  "language": "python",
  "tags": ["demo"],
  "visibility": "public"
}
```

---

### Buscar snippet público

```
GET /v1/snippets/{id}
```

Somente snippets públicos são retornados no MVP atual.

### Listar snippets (ListAll)

```
GET /v1/snippets
```

Suporta filtros via query params (implementação no handler):
- `q` — termo de busca (full-text / fuzzy)
- `creator` — `creator_id` para filtrar snippets por criador
- `limit`, `offset` — paginação

Exemplo:
```bash
curl 'http://localhost:8080/v1/snippets?q=hello&creator=usr_demo&limit=20'
```

---

### Atualizar snippet (Update)

```
PUT /v1/snippets/{id}
```

Body (mesmo formato do create):
```json
{
  "name": "Exemplo atualizado",
  "content": "print('hello world')",
  "language": "python",
  "tags": ["demo","edit"],
  "visibility": "public"
}
```

Resposta: `200 OK` com o recurso atualizado (JSON).

Exemplo:
```bash
curl -X PUT http://localhost:8080/v1/snippets/snp_abc123 \
  -H 'Content-Type: application/json' \
  -d '{"name":"Novo","content":"x","language":"txt"}'
```

---

### Excluir snippet (Delete)

```
DELETE /v1/snippets/{id}
```

Resposta: `204 No Content` em caso de sucesso.

Exemplo:
```bash
curl -X DELETE http://localhost:8080/v1/snippets/snp_abc123
```

---

## Autenticação (Planejada)

- JWT Bearer Token
- Snippets privados
- Endpoints `/auth/login` e `/auth/register`
- Endpoint `/v1/me/snippets`

---

## Convenções

- Toda mudança de schema passa por migration
- Migrations não são alteradas após aplicadas
- IDs são strings (`snp_*`, `usr_*`) para facilitar debug

---

## Próximos Passos

- Endpoint de busca `/v1/snippets/search`
- Autenticação JWT
- Snippets privados
- Paginação
- OpenAPI / Swagger
- Testes de integração
- CI/CD

---

## Objetivo do MVP

Fornecer uma base **simples, clara e extensível**, pronta para evolução sem retrabalho arquitetural.
