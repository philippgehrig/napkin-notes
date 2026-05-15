# Phase 11: Docker & Deployment

> **For agentic workers:** Use superpowers:subagent-driven-development or superpowers:executing-plans.

**Goal:** Production-ready Docker setup with Traefik, TLS, and dev override.

**Branch:** `feat/phase-11-docker`

---

## File Structure

```
docker/
├── docker-compose.yml        (production)
├── docker-compose.dev.yml    (dev override)
├── traefik/
│   └── traefik.yml           (static config)
├── api/
│   └── Dockerfile
├── fontforge/
│   └── Dockerfile
├── web/
│   └── Dockerfile
└── .env.example (updated)
```

---

### Task 1: API Dockerfile

**Files:**
- Create: `docker/api/Dockerfile`

- [ ] **Step 1: Create multi-stage Go Dockerfile**

Create `docker/api/Dockerfile`:
```dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY services/api/go.mod services/api/go.sum ./
RUN go mod download

COPY services/api/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /api .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /api .
COPY services/api/migrations ./migrations

ENV PORT=8080
ENV MIGRATIONS_PATH=/app/migrations
EXPOSE 8080

CMD ["./api"]
```

- [ ] **Step 2: Commit**

```bash
git add docker/api/
git commit -m "feat: add API Dockerfile (multi-stage Go build)"
```

---

### Task 2: Fontforge worker Dockerfile

**Files:**
- Create: `docker/fontforge/Dockerfile`

- [ ] **Step 1: Create Python Dockerfile**

Create `docker/fontforge/Dockerfile`:
```dockerfile
FROM python:3.11-slim

WORKDIR /app
COPY services/fontforge/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY services/fontforge/ .

CMD ["python", "worker.py"]
```

- [ ] **Step 2: Commit**

```bash
git add docker/fontforge/
git commit -m "feat: add fontforge worker Dockerfile"
```

---

### Task 3: Web Dockerfile

**Files:**
- Create: `docker/web/Dockerfile`
- Create: `docker/web/nginx.conf`

- [ ] **Step 1: Create Vue build + nginx Dockerfile**

Create `docker/web/Dockerfile`:
```dockerfile
FROM node:20-alpine AS builder

WORKDIR /app
COPY web/package.json web/yarn.lock ./
RUN yarn install --frozen-lockfile

COPY web/ ./
RUN yarn build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY docker/web/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

- [ ] **Step 2: Create nginx.conf**

Create `docker/web/nginx.conf`:
```nginx
server {
    listen 80;
    server_name _;
    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://api:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

- [ ] **Step 3: Commit**

```bash
git add docker/web/
git commit -m "feat: add web Dockerfile with nginx and SPA routing"
```

---

### Task 4: Traefik configuration

**Files:**
- Create: `docker/traefik/traefik.yml`

- [ ] **Step 1: Create Traefik static config**

Create `docker/traefik/traefik.yml`:
```yaml
api:
  dashboard: true

entryPoints:
  web:
    address: ":80"
    http:
      redirections:
        entryPoint:
          to: websecure
          scheme: https
  websecure:
    address: ":443"

providers:
  docker:
    exposedByDefault: false

certificatesResolvers:
  letsencrypt:
    acme:
      email: "${ACME_EMAIL}"
      storage: /letsencrypt/acme.json
      httpChallenge:
        entryPoint: web
```

- [ ] **Step 2: Commit**

```bash
git add docker/traefik/
git commit -m "feat: add Traefik static configuration"
```

---

### Task 5: Production docker-compose

**Files:**
- Modify: `docker/docker-compose.yml`

- [ ] **Step 1: Replace docker-compose.yml with full production config**

Replace `docker/docker-compose.yml`:
```yaml
services:
  traefik:
    image: traefik:v3.0
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - ./traefik/traefik.yml:/etc/traefik/traefik.yml:ro
      - letsencrypt:/letsencrypt
    environment:
      - ACME_EMAIL=${ACME_EMAIL}
    restart: unless-stopped

  api:
    build:
      context: ..
      dockerfile: docker/api/Dockerfile
    environment:
      - DB_HOST=db
      - DB_PORT=5432
      - DB_NAME=${DB_NAME:-napkin_notes}
      - DB_USER=${DB_USER:-postgres}
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_SSLMODE=disable
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=${JWT_SECRET}
      - STORAGE_PATH=/data/storage
      - RUN_MIGRATIONS=true
      - MIGRATIONS_PATH=/app/migrations
    volumes:
      - storage:/data/storage
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.api.rule=Host(`${DOMAIN}`) && PathPrefix(`/api`)"
      - "traefik.http.routers.api.entrypoints=websecure"
      - "traefik.http.routers.api.tls.certresolver=letsencrypt"
      - "traefik.http.services.api.loadbalancer.server.port=8080"
    restart: unless-stopped

  fontforge:
    build:
      context: ..
      dockerfile: docker/fontforge/Dockerfile
    environment:
      - REDIS_URL=redis://redis:6379
      - STORAGE_PATH=/data/storage
      - DB_HOST=db
      - DB_PORT=5432
      - DB_NAME=${DB_NAME:-napkin_notes}
      - DB_USER=${DB_USER:-postgres}
      - DB_PASSWORD=${DB_PASSWORD}
    volumes:
      - storage:/data/storage
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy
    restart: unless-stopped

  web:
    build:
      context: ..
      dockerfile: docker/web/Dockerfile
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.web.rule=Host(`${DOMAIN}`)"
      - "traefik.http.routers.web.entrypoints=websecure"
      - "traefik.http.routers.web.tls.certresolver=letsencrypt"
      - "traefik.http.services.web.loadbalancer.server.port=80"
    restart: unless-stopped

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: ${DB_NAME:-napkin_notes}
      POSTGRES_USER: ${DB_USER:-postgres}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 3s
      retries: 5
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5
    restart: unless-stopped

volumes:
  pgdata:
  storage:
  letsencrypt:
```

- [ ] **Step 2: Commit**

```bash
git add docker/docker-compose.yml
git commit -m "feat: add production docker-compose with Traefik TLS"
```

---

### Task 6: Dev docker-compose override

**Files:**
- Create: `docker/docker-compose.dev.yml`

- [ ] **Step 1: Create dev override**

Create `docker/docker-compose.dev.yml`:
```yaml
services:
  traefik:
    ports:
      - "80:80"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
    command:
      - "--api.insecure=true"
      - "--providers.docker=true"
      - "--providers.docker.exposedbydefault=false"
      - "--entrypoints.web.address=:80"

  api:
    build:
      context: ..
      dockerfile: docker/api/Dockerfile
    environment:
      - DB_HOST=db
      - DB_PORT=5432
      - DB_NAME=napkin_notes
      - DB_USER=postgres
      - DB_PASSWORD=postgres
      - DB_SSLMODE=disable
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=dev-secret-change-me
      - STORAGE_PATH=/data/storage
      - RUN_MIGRATIONS=true
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.api.rule=PathPrefix(`/api`)"
      - "traefik.http.routers.api.entrypoints=web"
      - "traefik.http.services.api.loadbalancer.server.port=8080"

  web:
    build:
      context: ..
      dockerfile: docker/web/Dockerfile
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.web.rule=PathPrefix(`/`)"
      - "traefik.http.routers.web.entrypoints=web"
      - "traefik.http.routers.web.priority=1"
      - "traefik.http.services.web.loadbalancer.server.port=80"

  db:
    environment:
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"

  redis:
    ports:
      - "6379:6379"
```

- [ ] **Step 2: Update .env.example**

Replace `docker/.env.example`:
```env
# Production
DOMAIN=napkin.yourdomain.com
ACME_EMAIL=your@email.com
DB_NAME=napkin_notes
DB_USER=postgres
DB_PASSWORD=CHANGE_ME_IN_PRODUCTION
JWT_SECRET=CHANGE_ME_IN_PRODUCTION
STORAGE_PATH=/data/storage

# Dev (unused in production)
# DB_PASSWORD=postgres
# JWT_SECRET=dev-secret-change-me
```

- [ ] **Step 3: Update Makefile dev target**

Update `Makefile`:
```makefile
dev:
	docker compose -f docker/docker-compose.yml -f docker/docker-compose.dev.yml up --build
```

- [ ] **Step 4: Commit**

```bash
git add docker/ Makefile
git commit -m "feat: add dev docker-compose override and update Makefile"
```

---

### Task 7: Create PR

- [ ] **Step 1: Push and create PR**

```bash
git push -u origin feat/phase-11-docker
gh pr create --title "feat: Docker deployment with Traefik and TLS" --body "## Summary
- Add multi-stage Dockerfiles for API, fontforge, and web
- Add production docker-compose with Traefik v3 and Let's Encrypt TLS
- Add dev docker-compose override (plain HTTP, exposed ports)
- Add nginx config for SPA routing and API proxy
- Add Traefik static configuration

## Test plan
- [ ] \`docker compose -f docker/docker-compose.yml -f docker/docker-compose.dev.yml build\` succeeds
- [ ] \`make dev\` starts all services
- [ ] API accessible at http://localhost/api/health
- [ ] Web SPA loads at http://localhost/

🤖 Generated with [Claude Code](https://claude.com/claude-code)"
```
