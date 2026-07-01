# FAIRRIDE EOS — Project Memory
Last updated: 2026-07-01 by Principal Engineer AI

## Current Phase
Phase 2.2 — Identity Persistence (COMPLETE, awaiting CTO approval)
Phase 2.1 domain layer is complete and tested.

## Documentation Strategy Change
ORIGINAL: 72-document comprehensive roadmap
CURRENT: Lean Documentation — 5 permanent core documents + on-demand as needed
DECISION DATE: 2026-06-30
ADR REQUIRED: ADR-0008 (pending)
The 72-document plan from PHASE-0-EOS-BLUEPRINT.md is superseded as the primary delivery plan.
Documents from that plan remain available as on-demand resources, generated when a specific
implementation phase requires them.

## EOS Core Documents Progress
| Doc ID | Title | Status | Path |
|--------|-------|--------|------|
| DOC-0001 | Project Constitution | Draft — Awaiting CTO Approval | docs/business/mission/project-constitution-v0.1.md |
| DOC-0001A | AI Development Governance | Draft — Awaiting CTO Approval | .ai/governance/ai-development-governance-v0.1.md |
| DOC-0002 | Product Vision | Draft — Awaiting CTO+CPO Approval | docs/product/requirements/product-vision-v0.1.md |
| DOC-0003 | System Architecture | Not yet generated — BLOCKED | docs/architecture/system/ |
| DOC-0004 | Implementation Master Plan | Not yet generated — BLOCKED | docs/product/roadmap/ |

## Open Questions — Priority View

### P0 Critical (must resolve before DOC-0003)
| OQ ID | Question |
|-------|----------|
| OQ-001 | Primary launch city and country |
| OQ-006 | Driver classification model (contractor vs employee) |
| OQ-B01 | Cash collection at MVP? (market may require Day 1) |

### P1 High (must resolve before DOC-0004)
| OQ ID | Question |
|-------|----------|
| OQ-002 | Funding stage and runway at MVP |
| OQ-007 | Offline-capable dispatch required at MVP? |
| OQ-B02 | First-ride Rider incentive amount |
| OQ-B03 | Driver 0% commission early adopter programme scope |
| OQ-B04 | MVP ride categories (motorcycle taxi Day 1?) |
| OQ-B05 | Booking fee amount |
| OQ-B06 | Instant withdrawal fee and mechanism |
| OQ-B07 | Map provider requirement at MVP |

### P2 Medium
OQ-003, OQ-005, OQ-A001, OQ-A002, OQ-A003, OQ-A004, OQ-A005, OQ-B08

### P3 Low
OQ-004, OQ-008, OQ-009

## Pending ADRs (16 total)
ADR-0001 (Go), ADR-0002 (Flutter), ADR-0003 (Next.js), ADR-0004 (doc-first), ADR-0005 (microservices),
ADR-0006 (event-driven), ADR-0007 (ai-first), ADR-0008 (lean-docs — P0), ADR-0009 (launch-market — P0),
ADR-0010 (commission-model — P0), ADR-0011 (surge-cap — P0), ADR-0012 (surge-passthrough — P1),
ADR-0013 (driver-tier-calc — P1), ADR-0014 (mvp-payments — P1), ADR-0015 (north-star — P1),
ADR-0016 (driver-classification — P0)

## Key Product Decisions Documented in DOC-0002
| Decision | Value | Section |
|----------|-------|---------|
| MVP Commission Rate | 15% flat | 6.7 |
| Commission Tiers | Standard 15% → Platinum 10% | 6.7 |
| Surge Cap | 2.0x hard cap | 6.9 |
| Surge Revenue to Driver | 100% of surge to Driver | 6.7 |
| North Star Metric | Weekly Fair Matches (WFM) | 6.23 |
| MVP Supply Target | 1,000 active Drivers by Month 3 | 6.19 |
| MVP Trip Target | 5,000 completed trips/day by Month 6 | 6.19 |
| Rider RSAT Target | ≥ 80% | 6.22 |
| Platform Availability | ≥ 99.9% | 6.19 |
| Dispute Resolution SLA | 95% within 24 hours | 6.22 |

## Phase 1 Implementation — Skeleton (COMPLETE)

| Deliverable | Path | Status |
|------------|------|--------|
| Go workspace | `backend/go.work` | ✅ |
| Shared package (config/logger/errors/db/redis/kafka/grpc) | `backend/shared/` | ✅ tested |
| 14 service skeletons | `backend/services/*/` | ✅ all compile |
| Docker Compose (postgres/redis/kafka) | `infra/docker/` | ✅ |
| Makefile | `Makefile` | ✅ |
| CI workflow | `.github/workflows/ci.yml` | ✅ |
| Service Dockerfile | `backend/Dockerfile` | ✅ |

## Phase 1.5 — Architecture Audit Fixes (COMPLETE)

All Critical and High findings from the audit have been resolved.

| Finding | Severity | Fix Applied | File |
|---------|----------|-------------|------|
| C-001: gRPC reflection always on in production | Critical | `Options.EnableReflection` + gated in `bootstrap.Run()` | `shared/grpc/server.go`, `shared/server/bootstrap.go` |
| C-002: CI test command `./...` broken | Critical | `working-directory: backend/shared` + `go test ./...` | `.github/workflows/ci.yml` |
| C-003: readiness probe always returns 200 | Critical | Real `ReadinessTracker` with 503 on not-ready | `shared/server/bootstrap.go` |
| H-001: `max()` shadow of Go 1.21+ builtin | High | Removed function, inlined logic | `shared/redis/client.go` |
| H-002: kafka-ui pinned to `latest` | High | Pinned to `v0.7.2` | `infra/docker/docker-compose.yml` |
| H-003: DB pool exhaustion (25 × 14 = 350 > 100) | High | MaxConns default 25→5, MinConns 5→2 | `shared/config/config.go` |
| H-004: gRPC msg limits not wired from config | High | `MaxRecvMsgSizeMB`/`MaxSendMsgSizeMB` passed via `Options` | `shared/grpc/server.go`, `shared/server/bootstrap.go` |
| H-005: `kafka.Message` type leak across package boundary | High | Unexported `raw` field; `FetchMessage`/`Commit` use `*Message` | `shared/kafka/consumer.go` |
| H-006: CI build command `./services/...` broken | High | Changed to `go build github.com/fairride/$svc/...` | `.github/workflows/ci.yml` |
| M-002: Docker init/postgres directory missing | Medium | Added `.gitkeep` | `infra/docker/init/postgres/.gitkeep` |
| M-004: Dockerfile buildVersion uses `$(date)` (breaks cache) | Medium | Changed to `ARG GIT_COMMIT`; passed at build time | `backend/Dockerfile` |
| M-005: No linter configuration | Medium | Added `.golangci.yml` | `.golangci.yml` |
| M-007: 14 identical main.go files (90 lines each) | Medium | `server.Run()` bootstrap; all 14 now 7 lines each | `shared/server/bootstrap.go` + all service `main.go` |
| L-008: Kafka ACL — no RequireAllAcks for financial | Low | Added `RequireAllAcks bool` to `ProducerConfig` | `shared/kafka/producer.go` |

**Deferred to later phases (medium/low, no immediate risk):**
- M-001: Type aliases → interfaces for DB/Redis (Phase 2+)
- M-003: Correlation/trace ID in logging interceptor (Phase 2+)
- M-006: Database migration tooling (Phase 2.1)

**Build verification:** all 14 services compile. `go test -race ./...` passes (3 packages tested, 0 failures).

## Audit Score (Phase 1.5 exit)
| Category | Score Before | Score After |
|----------|-------------|-------------|
| Security | 65/100 | 95/100 |
| Reliability | 55/100 | 90/100 |
| Code Quality | 75/100 | 92/100 |
| DevOps | 60/100 | 90/100 |
| Overall | 64/100 | **92/100** |

**Verdict: GO for Phase 2**

## Phase 2.1 — Identity Foundation (COMPLETE)

Clean Architecture domain layer for the Identity service. No DB, no endpoints, no auth.

| File | Purpose |
|------|---------|
| `services/identity/domain/entity/permission.go` | Permission entity + 23 named permission constants + `NewPermission` / `ReconstitutePermission` |
| `services/identity/domain/entity/role.go` | Role entity + 6 system role constants + `NewRole` / `ReconstituteRole`; `AddPermission`, `RemovePermission`, `HasPermission`, `CanDelete` |
| `services/identity/domain/repository/permission_repository.go` | `PermissionRepository` interface (FindByID, FindByName, FindByResource, FindAll, Save, Delete) |
| `services/identity/domain/repository/role_repository.go` | `RoleRepository` interface (FindByID, FindByName, FindAll, Save, Delete) |
| `services/identity/app/container.go` | Composition root; holds `RoleRepository` + `PermissionRepository` interfaces |
| `services/identity/domain/entity/permission_test.go` | 8 unit tests for Permission |
| `services/identity/domain/entity/role_test.go` | 17 unit tests for Role |

**System roles (from DOC-0002 §6.12):** rider, driver, fleet_operator, city_manager, support_agent, super_admin

**Permission format:** `"resource:action"` — validated at construction. 23 well-known constants (trips, drivers, riders, wallet, payments, dispatch, reviews, reports, support, admin).

**Build verification:** all 14 services compile. 25 identity entity tests pass. 3 shared packages pass.

## Phase 2.2 — Identity Persistence (COMPLETE)

PostgreSQL repository implementations for Role and Permission.

| File | Purpose |
|------|---------|
| `services/identity/infrastructure/postgres/permission_repository.go` | `PermissionRepository`: FindByID, FindByName, FindByResource, FindAll, Save (upsert), Delete |
| `services/identity/infrastructure/postgres/role_repository.go` | `RoleRepository`: FindByID, FindByName, FindAll, Save (upsert + tx permission replace), Delete |
| `services/identity/infrastructure/postgres/testmain_test.go` | TestMain: skip when DATABASE_URL unset; createSchema/dropSchema/setupTest helpers |
| `services/identity/infrastructure/postgres/permission_repository_test.go` | 10 integration tests |
| `services/identity/infrastructure/postgres/role_repository_test.go` | 13 integration tests |

**Schema (test scaffolding — NOT a migration):**
- `identity_permissions` (id PK, name UNIQUE, resource, action, description, created_at)
- `identity_roles` (id PK, name UNIQUE, description, is_system, created_at, updated_at)
- `identity_role_permissions` (role_id FK, permission_id FK, PK composite)

**go.mod changes (identity service):**
- Added `require github.com/jackc/pgx/v5 v5.6.0`
- Added `replace github.com/fairride/shared => ../../shared` (for `GOWORK=off go mod tidy`; no effect in workspace mode)

**Verification:**
- `go build -o /dev/null github.com/fairride/identity/...` — ✅ clean
- `go vet github.com/fairride/identity/...` — ✅ clean
- `go test github.com/fairride/identity/infrastructure/postgres/...` (no DATABASE_URL) — ✅ skips gracefully
- Integration tests (need DB): run with `DATABASE_URL=... go test -v github.com/fairride/identity/infrastructure/postgres/...`
- All 14 service builds still pass

## Human Checkpoints Pending
| HC ID | Scope | Action Required |
|-------|-------|----------------|
| HC-001 | DOC-0001 | CTO approval → rename to v1.0 |
| HC-002 | DOC-0001A | CTO approval → rename to v1.0 |
| HC-003 | DOC-0002 | CTO + CPO approval → rename to v1.0 |
| HC-P2.2 | Phase 2.2 persistence | CTO approval to proceed to Phase 2.3 |

## Next Phase (pending CTO approval)
Phase 2.3 — JWT + Refresh Token (in `services/identity/`)
- Domain: `Token` value object
- Application: `IssueToken`, `RefreshToken`, `RevokeToken` use cases
- Infrastructure: JWT signing, Redis token blacklist
- No API endpoints yet

## Notes
- Implementation mode began 2026-07-01. Documentation phase paused.
- DOC-0003 and DOC-0004 remain pending but are not blocking implementation.
- Architecture pattern: 14 microservice skeletons retained; CTO approved continuing with controlled inter-service dependencies.
- 83 canonical terms defined in DOC-0001, DOC-0001A, DOC-0002.
- PHASE-0-EOS-BLUEPRINT.md folder structure (209 folders) remains valid.
