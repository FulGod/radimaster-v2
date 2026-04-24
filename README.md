# 🏥 RadiMaster

> **Real-time Medical Imaging Education Platform** — A platform for radiology students to study real clinical images (CT, MRI, X-Ray) through a live virtual classroom system powered by WebSocket.

**Live Demo**: [https://stage.radimaster.com](https://stage.radimaster.com)

---

## 🎯 Overview

RadiMaster is a medical education product developed for **diagnostic imaging specialists** to share clinical images from hospitals with students via a web-based platform. Key capabilities:

- **View CT/MRI/X-Ray images** slice-by-slice with a professional split-view interface
- **Learn from real diagnoses** annotated by specialist doctors
- **Join live classrooms** — doctor leads, students follow in real-time via WebSocket

## ⚡ Tech Stack

| Layer | Technology | Rationale |
|-------|-----------|-----------|
| **Backend** | Go 1.26 + Gin | High performance (~5ms latency), goroutine concurrency |
| **Frontend** | React 18 + TypeScript + Vite | Type-safe SPA, fast HMR dev experience |
| **Database** | PostgreSQL 16 | ACID compliance for medical data integrity |
| **Cache/PubSub** | Redis 7 | WebSocket horizontal scaling via PubSub |
| **Reverse Proxy** | Nginx | Static file serving, WebSocket upgrade, gzip |
| **Container** | Docker Compose | Reproducible multi-service orchestration |
| **CI/CD** | GitHub Actions | Automated lint → test → deploy → smoke test |

## 🏗️ Architecture

```
                  ┌───────────────────────────────────────────────┐
                  │              Nginx Proxy Manager              │
                  │   stage.radimaster.com → frontend (:80)       │
                  │   + SSL termination (Let's Encrypt)           │
                  └──────────────────┬────────────────────────────┘
                                     │
                  ┌──────────────────▼────────────────────────────┐
                  │         Nginx (frontend container :80)         │
                  │  /       → React SPA (static)                  │
                  │  /api/*  → Go Backend (proxy_pass :8080)       │
                  │  /ws/*   → WebSocket (upgrade + proxy_pass)    │
                  │  /media/*→ Go Backend (static files)           │
                  └──────────────────┬────────────────────────────┘
                                     │
                  ┌──────────────────▼────────────────────────────┐
                  │           Go Backend (:8080)                   │
                  │                                                │
                  │  ┌──────────────┐   ┌───────────────────────┐ │
                  │  │  REST API    │   │  WebSocket Hub        │ │
                  │  │  (Gin)       │   │  (goroutine per conn) │ │
                  │  │              │   │                       │ │
                  │  │  • Auth JWT  │   │  • Room management    │ │
                  │  │  • Cases     │   │  • Client read/write  │ │
                  │  │  • Classroom │   │  • Redis PubSub relay │ │
                  │  │  • Media     │   │  • Presence tracking  │ │
                  │  └──────┬───────┘   └──────────┬────────────┘ │
                  └─────────┼──────────────────────┼──────────────┘
                            │                      │
                  ┌─────────▼────────┐   ┌─────────▼────────────┐
                  │  PostgreSQL 16   │   │     Redis 7          │
                  │                  │   │                      │
                  │  users           │   │  PubSub channel:     │
                  │  medical_cases   │   │  radimaster:ws:      │
                  │  phases          │   │  broadcast           │
                  │  slices          │   │                      │
                  │  classrooms      │   │  (horizontal scale)  │
                  └──────────────────┘   └──────────────────────┘
```

### WebSocket Horizontal Scaling

```
Doctor (Instance A)  ──PUBLISH──▶  Redis Channel  ──SUBSCRIBE──▶  Student (Instance B)
                                   radimaster:ws:broadcast
```

When running multiple Go instances behind a load balancer, all WebSocket messages are relayed through Redis PubSub, ensuring students connected to any instance receive real-time updates.

## 📁 Project Structure

```
radimaster/
├── .github/workflows/
│   └── deploy.yml              # CI/CD: lint → test → deploy → smoke test
├── deploy.sh                   # Shell deploy script (stage / production)
├── .env.example                # Environment template (no secrets)
│
├── docker-compose.yml          # Base orchestration (4 services)
├── docker-compose.stage.yml    # Staging overrides (isolated ports, global_net)
├── docker-compose.production.yml # Production overrides
│
├── backend/
│   ├── Dockerfile              # Multi-stage: Go build → Alpine (15MB)
│   ├── cmd/server/main.go      # Entry point, graceful shutdown
│   └── internal/
│       ├── config/             # Environment config loader
│       ├── database/           # GORM + PostgreSQL (auto-migrate)
│       ├── handlers/           # REST (auth, cases, classroom) + WebSocket
│       ├── middleware/         # JWT auth, CORS, rate-limit, logger
│       ├── models/            # User, MedicalCase, Phase, Slice, Classroom
│       ├── services/          # Data seed (natural sort, directory scan)
│       └── ws/                # WebSocket Hub + Client + Redis PubSub
│
├── frontend/
│   ├── Dockerfile              # Multi-stage: Node build → Nginx (static)
│   ├── nginx.conf              # Reverse proxy (/api, /ws, /media → backend)
│   └── src/
│       ├── App.tsx             # Router + Protected Routes
│       ├── AuthContext.tsx     # JWT auth state (Context API)
│       ├── api.ts              # REST client + WebSocket URL builder
│       ├── useWebSocket.ts    # Custom WS hook (auto-reconnect)
│       ├── types.ts            # TypeScript interfaces
│       └── pages/
│           ├── LoginPage.tsx
│           ├── CaseLibraryPage.tsx     # Search + filter by modality
│           ├── ViewerPage.tsx          # Dual-pane split-view
│           └── ClassroomPage.tsx       # Real-time WS classroom
│
├── README.md
└── TECHNICAL_QA.md             # Architecture decisions & tradeoffs
```

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose v2+

### Run Locally

```bash
# 1. Clone
git clone https://github.com/FulGod/radimaster-v2.git
cd radimaster-v2

# 2. Configure environment
cp .env.example .env
# Edit .env with your values (or use defaults for local dev)

# 3. Start all services
docker compose up -d --build
```

4 services will start:

| Service | Container | Internal Port | Health Check |
|---------|-----------|---------------|--------------|
| PostgreSQL 16 | radimaster_postgres | 5432 | `pg_isready` |
| Redis 7 | radimaster_redis | 6379 | `redis-cli ping` |
| Go Backend | radimaster_backend | 8080 | — |
| React + Nginx | radimaster_frontend | **80** | — |

Access: **http://localhost** (port 80)

### Default Accounts

| Role | Email | Password |
|------|-------|----------|
| 👨‍⚕️ Doctor | `doctor@radimaster.com` | `doctor123` |
| 🎓 Student | `student@radimaster.com` | `student123` |

### Local Development (without Docker)

```bash
# Backend (Go)
cd backend
cp .env.example .env          # Uses SQLite in dev mode
go run cmd/server/main.go     # http://localhost:8080

# Frontend (React) — separate terminal
cd frontend
npm install
npm run dev                   # http://localhost:5173
```

## 🚢 Deployment

### Infrastructure

```
                    ┌─────────────────────────────┐
                    │         VPS (Ubuntu)         │
                    │                              │
                    │  Nginx Proxy Manager (:443)  │
                    │     ├── SSL (Let's Encrypt)  │
                    │     ├── stage.radimaster.com  │
                    │     └── radimaster.com        │
                    │                              │
                    │  ┌────────────────────────┐  │
                    │  │  Stage Environment     │  │
                    │  │  docker-compose.stage  │  │
                    │  │  (isolated network)    │  │
                    │  └────────────────────────┘  │
                    │                              │
                    │  ┌────────────────────────┐  │
                    │  │  Production Environment│  │
                    │  │  docker-compose.prod   │  │
                    │  │  (isolated network)    │  │
                    │  └────────────────────────┘  │
                    └─────────────────────────────┘
```

### CI/CD Pipeline (GitHub Actions)

```
Push to 'stage' → Lint (Go vet + TSC) → Build → Deploy to Staging → Smoke Test ✅
Manual trigger   → Lint (Go vet + TSC) → Build → Deploy to Production → Smoke Test ✅
```

| Trigger | Branch | Environment | Action |
|---------|--------|-------------|--------|
| `git push` | `stage` | Staging | Automatic deploy |
| `workflow_dispatch` | `master` | Production | Manual deploy (review required) |

### Manual Deploy

```bash
# Deploy to staging
./deploy.sh stage

# Deploy to production (manual trigger only)
./deploy.sh production
```

### Environment Configuration

```bash
# .env.example — copy and customize per environment
COMPOSE_PROJECT_NAME=radimaster          # Namespace for containers
DB_PASSWORD=change_me                    # PostgreSQL password
JWT_SECRET=change_me_32_chars           # JWT signing key
```

Each environment uses a separate Docker Compose override:

| File | Purpose |
|------|---------|
| `docker-compose.yml` | Base services (Postgres, Redis, Backend, Frontend) |
| `docker-compose.stage.yml` | Staging: container names, external network, isolated ports |
| `docker-compose.production.yml` | Production: container names, external network, standard ports |

## ✨ Key Features

### 1. Medical Image Viewer — Split View
- **Dual-pane display**: Compare 2 phases side-by-side (e.g., PLAIN vs ARTERIAL)
- **Synchronized scrolling**: Scroll one pane → both sync
- **Natural numeric sort**: `1.jpg, 2.jpg, ... 10.jpg` (not `1, 10, 11, 2`)
- **Image controls**: Brightness/Contrast adjustment
- **Multi-input**: Scroll wheel, slider, keyboard (←→↑↓)

### 2. Virtual Classroom — Real-time Sync
- **Doctor Mode**: Create session, navigate slices/phases → broadcast to all students
- **Student Mode**: Auto-follow doctor's view in real-time
- **Presence**: Online user list updated live via WebSocket
- **Horizontal Scaling**: Redis PubSub enables multi-instance deployment

### 3. Case Library
- 19 real medical cases (CT/MRI/X-Ray) with Vietnamese diagnoses
- Search by title, diagnosis, body part
- Filter by modality (CT, MRI, X-Ray)

## 🔌 WebSocket Protocol

| Type | Direction | Payload | Description |
|------|-----------|---------|-------------|
| `slice:changed` | Doctor → Students | `{slice_index, phase_id}` | Sync image slice |
| `phase:changed` | Doctor → Students | `{phase_id}` | Change phase |
| `cursor:moved` | Doctor → Students | `{x, y}` | Doctor's cursor position |
| `presence_update` | Server → All | `[{id, name}]` | Online user list |

## 🔒 Security

- **JWT Authentication**: Token-based with configurable secret & expiry
- **bcrypt**: Password hashing (cost 12)
- **CORS**: Configurable allowed origins
- **Rate Limiting**: 100 requests/minute per IP
- **Non-root Docker**: Backend runs as `appuser` (UID 1001)
- **Read-only media**: Medical images mounted as `:ro`
- **Environment isolation**: Stage & Production on separate Docker networks
- **No secrets in repo**: `.env.example` template only, `.env.*` in `.gitignore`

## 📊 Performance

| Metric | Value |
|--------|-------|
| API latency (p50) | ~5ms |
| Image serve latency | ~3ms (Nginx cache 7d) |
| WebSocket message relay | < 10ms |
| Backend binary size | ~15MB (Alpine) |
| Frontend bundle | ~180KB gzipped |
| Docker full stack boot | ~5s (after image pull) |

## 🌍 Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `COMPOSE_PROJECT_NAME` | `radimaster` | Docker container namespace |
| `DB_PASSWORD` | `radimaster_secret` | PostgreSQL password |
| `JWT_SECRET` | `radimaster-dev` | JWT signing key |
| `PORT` | `8080` | Backend listen port |
| `DATABASE_URL` | (empty = SQLite) | PostgreSQL connection string |
| `REDIS_URL` | (empty = in-memory) | Redis URL for PubSub |
| `MEDIA_DIR` | `./med-data` | Path to medical image directory |
| `ENVIRONMENT` | `development` | `development` / `staging` / `production` |

## 📝 License

Private — Developed for RadiMaster Medical Education Platform.
