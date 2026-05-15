# Napkin Notes — Implementation Plan Index

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a full-stack notes app where users write on realistic napkin textures in their own handwriting font.

**Architecture:** Monorepo with Go REST API, Vue 3 SPA, Python font worker, PostgreSQL, Redis, Docker + Traefik.

**Tech Stack:** Go (Chi), Vue 3 + TypeScript + Vite + Pinia, Python (fonttools), PostgreSQL 16, Redis 7, Docker, Traefik v3

---

## Phases (each = one PR)

| # | Phase | Plan File | Description |
|---|-------|-----------|-------------|
| 1 | Project scaffold | `phase-01-scaffold.md` | Monorepo structure, tooling, CI basics |
| 2 | Database & migrations | `phase-02-database.md` | PostgreSQL schema, migration tooling |
| 3 | Auth API | `phase-03-auth.md` | Registration, login, JWT, refresh tokens |
| 4 | Notes CRUD API | `phase-04-notes-api.md` | Notes endpoints with tests |
| 5 | Fonts API | `phase-05-fonts-api.md` | Font upload, metadata, file serving |
| 6 | Font generation worker | `phase-06-fontforge-worker.md` | Python worker, Redis queue integration |
| 7 | Vue SPA scaffold | `phase-07-vue-scaffold.md` | Vite project, router, stores, auth views |
| 8 | Gallery & editor views | `phase-08-gallery-editor.md` | Napkin cards, editor, font rendering |
| 9 | Rip-to-delete animation | `phase-09-rip-animation.md` | Drag gesture, tear animation, trash |
| 10 | Image export | `phase-10-image-export.md` | Server-side PNG rendering |
| 11 | Docker & deployment | `phase-11-docker.md` | Compose files, Traefik, production config |
| 12 | E2E tests | `phase-12-e2e.md` | Playwright end-to-end test suite |

## PR Strategy

- Each phase = one feature branch off `main`
- Branch naming: `feat/phase-NN-description`
- PR gets GitHub Copilot review before merge
- Must pass unit tests before PR creation
- E2E tests added in phase 12 cover the full flow

## Test Strategy

- **Unit tests:** Every phase includes unit tests (Go: `testing` + `testify`, Vue: `vitest`, Python: `pytest`)
- **Integration tests:** API phases test against real PostgreSQL (Docker test container)
- **E2E tests:** Phase 12 adds Playwright tests covering golden paths
