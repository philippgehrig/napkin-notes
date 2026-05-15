# Deployment Guide

Host Napkin Notes on a VPS with Docker and a custom domain.

## Prerequisites

- A VPS with Docker installed (e.g. Hetzner Cloud CX22 — €4.51/mo, Ubuntu 24.04)
- A domain name (e.g. Cloudflare Registrar — ~€9/yr for .com)

## 1. Create the VPS

1. Sign up at [Hetzner Cloud](https://www.hetzner.com/cloud/)
2. Create a server: CX22, Ubuntu 24.04, your preferred location
3. Note the **static IP** shown in the dashboard (e.g. `78.46.123.45`)
4. SSH access is provided via your SSH key or a root password emailed to you

## 2. Register a Domain

1. Register a domain at your preferred registrar (Cloudflare, Namecheap, INWX, etc.)
2. In the DNS settings, add an A record pointing to your server:

| Type | Name | Content | Proxy |
|------|------|---------|-------|
| A | `@` | `78.46.123.45` | DNS only |

If using Cloudflare, set the proxy to "DNS only" (grey cloud) so Traefik can handle TLS directly.

## 3. Set Up the Server

SSH into your server:

```bash
ssh root@78.46.123.45
```

Install Docker:

```bash
curl -fsSL https://get.docker.com | sh
```

Clone the repository:

```bash
git clone https://github.com/philippgehrig/napkin-notes.git
cd napkin-notes/docker
```

Create the environment file:

```bash
cat > .env << 'EOF'
DB_NAME=napkin_notes
DB_USER=napkin
DB_PASSWORD=<generate-a-strong-password>
JWT_SECRET=<generate-a-64-char-random-string>
DOMAIN=yourdomain.com
ACME_EMAIL=your@email.com
EOF
```

Generate secure values for the secrets:

```bash
# Generate DB password
openssl rand -base64 32

# Generate JWT secret
openssl rand -base64 64
```

## 4. Start the Application

```bash
docker compose up -d --build
```

First build takes 5-10 minutes. After that, the app is live at `https://yourdomain.com`.

Traefik automatically provisions a Let's Encrypt TLS certificate on first request.

## 5. Verify

```bash
# Check all services are running
docker compose ps

# Check logs if something is wrong
docker compose logs -f
```

## Maintenance

```bash
# Pull latest changes and rebuild
cd napkin-notes
git pull
cd docker
docker compose up -d --build

# View logs
docker compose logs -f api
docker compose logs -f web

# Restart a service
docker compose restart api

# Stop everything
docker compose down

# Stop and remove all data (destructive)
docker compose down -v
```

## Troubleshooting

| Problem | Fix |
|---------|-----|
| TLS cert not working | Ensure DNS A record points to server IP, port 80 is open, and `DOMAIN` in `.env` is correct |
| 502 Bad Gateway | API may still be starting — check `docker compose logs api` |
| Database connection refused | Check `docker compose ps` — db may not be healthy yet |
| Can't SSH in | Check Hetzner firewall rules allow port 22 |

## Firewall

If using Hetzner's firewall, allow these inbound ports:

- 22 (SSH)
- 80 (HTTP — needed for Let's Encrypt challenge)
- 443 (HTTPS)
