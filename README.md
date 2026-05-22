# LogGraph

A minimal team workflow tool combining a timeline log stream, knowledge graph, and AI-powered report generation.

- **Timeline** — post log entries with `#project` / `@person` / `^reference` tags
- **Graph** — auto-built knowledge graph from parsed tags and relations
- **AI Reports** — LLM-generated project summaries with date-range filtering
- **Webhook** — receive logs from external systems (Lark, custom)

## Architecture

```
loggraph-web (Vue 3 + Vite)   ←→   loggraph-server (Go + Chi)   ←→   PostgreSQL
       ↓                                ↓
  /api/v1/ai proxy              loggraph-ai (Python + FastAPI)
       ↓                                ↓
  port 8081                              LLM (DeepSeek / OpenAI)
```

| Service | Port | Stack |
|---------|------|-------|
| Frontend | 5173 | Vue 3, TypeScript, Tailwind CSS 4, Vue Flow |
| API Server | 8080 | Go, Chi, pgx, PostgreSQL |
| AI Service | 8081 | Python, FastAPI, OpenAI SDK |

## Quick Start

### 1. Database

```bash
createdb loggraph
# Schema auto-migrates on server startup
```

### 2. Go API Server

```bash
cd loggraph-server
cp .env.example .env   # edit DATABASE_URL
go run cmd/server/main.go
```

The server loads `.env`, connects to PostgreSQL, runs migration, then starts on `:8080`.

### 3. AI Service

```bash
cd loggraph-ai
python -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
cp .env.example .env   # add your API key
python main.py
```

Starts on `:8081`.

### 4. Frontend

```bash
cd loggraph-web
npm install
npm run dev
```

Starts on `:5173` with LAN access enabled. Proxies `/api` → `:8080` and `/api/v1/ai` → `:8081`.

Open `http://localhost:5173`.

## Environment Variables

### Go Server (`loggraph-server/.env`)

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | — | PostgreSQL connection string |
| `SERVER_PORT` | `8080` | HTTP listen port |
| `S3_ENDPOINT` | — | S3-compatible endpoint |
| `S3_BUCKET` | `loggraph` | S3 bucket name |
| `S3_ACCESS_KEY` | — | S3 access key |
| `S3_SECRET_KEY` | — | S3 secret key |
| `S3_REGION` | `us-east-1` | S3 region |

### AI Service (`loggraph-ai/.env`)

| Variable | Default | Description |
|----------|---------|-------------|
| `OPENAI_BASE_URL` | `https://api.deepseek.com/v1` | LLM API endpoint |
| `OPENAI_API_KEY` | — | API key |
| `OPENAI_MODEL` | `deepseek-chat` | Model name |
| `LOGGRAPH_API_URL` | `http://localhost:8080/api/v1` | Go API URL |
| `AI_SERVER_PORT` | `8081` | Listen port |

## API Endpoints

### Blocks (`/api/v1/blocks`)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/blocks` | List blocks (cursor pagination, filters) |
| POST | `/blocks` | Create block |
| GET | `/blocks/{id}` | Get block by ID |
| PATCH | `/blocks/{id}` | Update content or status |
| DELETE | `/blocks/{id}` | Delete block |
| GET | `/blocks/{id}/graph` | Block-centric knowledge graph |

### Nodes (`/api/v1/nodes`)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/nodes` | List nodes by type |
| GET | `/nodes/suggest` | Autocomplete node names |
| GET | `/nodes/{id}/graph` | Node-centric graph |

### AI (`/api/v1/ai`)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/ai/report` | Generate project report |
| POST | `/ai/reminders` | Parse reminder from text |
| GET | `/ai/settings` | Get AI settings |
| PUT | `/ai/settings` | Update AI settings |

### Webhook (`/api/v1/webhook`)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/webhook/logs` | Receive external log |
| POST | `/webhook/lark` | Lark bot webhook |
| GET | `/webhook/tokens` | List webhook tokens |
| POST | `/webhook/tokens` | Generate token |
| DELETE | `/webhook/tokens/{id}` | Revoke token |

## Usage

- Type `#project @person ^reference` in the input — autocomplete triggers after each symbol
- Click **< >** toggle (compact mode) or expand to full-editor for markdown formatting
- Double-click a block to edit
- Click the status badge (Active/Done/Blocked) to cycle states
- Select a block to see its graph in the right panel

## Database Schema

- **blocks** — core log entries (id, user_id, content, status, metadata, timestamps)
- **nodes** — knowledge graph entities extracted from `#` `@` tags (name, type)
- **relations** — directed edges: `mentions`, `blocks`, `references`
- **webhook_tokens** — hashed bearer tokens for webhook auth

## License

MIT
