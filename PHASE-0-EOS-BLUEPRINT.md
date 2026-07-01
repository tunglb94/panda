# FAIRRIDE — Engineering Operating System (EOS)
# Phase 0: Blueprint & Documentation Architecture

**Status:** DRAFT — Awaiting CTO Approval
**Version:** v0.1.0
**Date:** 2026-06-30
**Author:** Office of the CTO
**Classification:** Internal — Engineering Leadership

---

> "Architecture is the decisions that are hard to change."
> Every document in this EOS is a decision made before code makes it permanent.

---

## TABLE OF CONTENTS

1. [Complete Folder Tree](#1-complete-folder-tree)
2. [Explanation of Every Folder](#2-explanation-of-every-folder)
3. [Dependency Map Between Folders](#3-dependency-map-between-folders)
4. [Documentation Creation Order](#4-documentation-creation-order)
5. [Engineering Workflow](#5-engineering-workflow)
6. [Documentation Roadmap](#6-documentation-roadmap)
7. [Estimated Number of Documents](#7-estimated-number-of-documents)
8. [Estimated Pages Per Document](#8-estimated-pages-per-document)
9. [Naming Convention](#9-naming-convention)
10. [Versioning Convention](#10-versioning-convention)
11. [Documentation Maintenance Policy](#11-documentation-maintenance-policy)

---

## 1. COMPLETE FOLDER TREE

```
fairride/                                    ← Project root
│
├── PHASE-0-EOS-BLUEPRINT.md                 ← This document
├── README.md                                ← Project entry point (to be written)
│
├── docs/
│   │
│   ├── business/                            ← Layer 1: Why we exist
│   │   ├── mission/
│   │   ├── vision/
│   │   ├── market/
│   │   ├── competitive/
│   │   └── kpis/
│   │
│   ├── product/                             ← Layer 2: What we build
│   │   ├── requirements/
│   │   ├── user-stories/
│   │   ├── roadmap/
│   │   ├── personas/
│   │   └── features/
│   │       ├── ride/
│   │       ├── driver/
│   │       ├── admin/
│   │       ├── delivery/
│   │       └── logistics/
│   │
│   ├── architecture/                        ← Layer 3: How the system works
│   │   ├── overview/
│   │   ├── system/
│   │   ├── domain/
│   │   ├── services/
│   │   ├── data-flow/
│   │   ├── diagrams/
│   │   └── boundaries/
│   │
│   ├── database/                            ← Layer 4: How data is stored
│   │   ├── models/
│   │   ├── schema/
│   │   ├── indexes/
│   │   ├── migrations/
│   │   ├── policies/
│   │   └── replication/
│   │
│   ├── api/                                 ← Layer 4: How systems communicate
│   │   ├── contracts/
│   │   ├── versioning/
│   │   ├── authentication/
│   │   ├── rate-limiting/
│   │   ├── errors/
│   │   ├── webhooks/
│   │   └── open-api/
│   │
│   ├── security/                            ← Layer 4: How we stay safe
│   │   ├── threat-model/
│   │   ├── authentication/
│   │   ├── authorization/
│   │   ├── encryption/
│   │   ├── compliance/
│   │   ├── incident-response/
│   │   └── penetration/
│   │
│   ├── deployment/                          ← Layer 5: How we ship
│   │   ├── environments/
│   │   ├── pipelines/
│   │   ├── infrastructure/
│   │   ├── containers/
│   │   ├── runbooks/
│   │   └── rollback/
│   │
│   ├── coding/                              ← Layer 5: How we write code
│   │   ├── standards/
│   │   ├── style-guides/
│   │   ├── patterns/
│   │   ├── reviews/
│   │   ├── onboarding/
│   │   └── dependencies/
│   │
│   ├── testing/                             ← Layer 5: How we verify quality
│   │   ├── strategy/
│   │   ├── unit/
│   │   ├── integration/
│   │   ├── e2e/
│   │   ├── load/
│   │   ├── security/
│   │   └── chaos/
│   │
│   ├── operations/                          ← Layer 6: How we run in production
│   │   ├── runbooks/
│   │   ├── on-call/
│   │   ├── incident/
│   │   ├── sla/
│   │   ├── capacity/
│   │   └── postmortem/
│   │
│   ├── monitoring/                          ← Layer 6: How we observe the system
│   │   ├── metrics/
│   │   ├── alerting/
│   │   ├── dashboards/
│   │   ├── logging/
│   │   ├── tracing/
│   │   └── health-checks/
│   │
│   ├── performance/                         ← Layer 6: How we stay fast
│   │   ├── benchmarks/
│   │   ├── optimization/
│   │   ├── capacity/
│   │   └── profiling/
│   │
│   ├── fraud/                               ← Domain: Trust & Safety
│   │   ├── detection/
│   │   ├── prevention/
│   │   ├── rules/
│   │   ├── policies/
│   │   └── appeals/
│   │
│   ├── payments/                            ← Domain: Money movement
│   │   ├── flow/
│   │   ├── providers/
│   │   ├── reconciliation/
│   │   ├── refunds/
│   │   ├── compliance/
│   │   └── disputes/
│   │
│   ├── wallet/                              ← Domain: Digital wallet
│   │   ├── design/
│   │   ├── transactions/
│   │   ├── limits/
│   │   ├── policies/
│   │   └── topup/
│   │
│   ├── dispatch/                            ← Domain: Driver-rider matching
│   │   ├── algorithm/
│   │   ├── matching/
│   │   ├── routing/
│   │   ├── pooling/
│   │   ├── fallback/
│   │   └── simulation/
│   │
│   ├── geo/                                 ← Domain: Geospatial intelligence
│   │   ├── mapping/
│   │   ├── geofencing/
│   │   ├── tracking/
│   │   ├── zones/
│   │   ├── providers/
│   │   └── h3/
│   │
│   ├── pricing/                             ← Domain: Fare & revenue
│   │   ├── model/
│   │   ├── surge/
│   │   ├── promotions/
│   │   ├── corporate/
│   │   ├── policies/
│   │   └── estimation/
│   │
│   ├── notifications/                       ← Domain: User communication
│   │   ├── channels/
│   │   ├── templates/
│   │   ├── policies/
│   │   ├── providers/
│   │   └── preferences/
│   │
│   ├── analytics/                           ← Domain: Data intelligence
│   │   ├── data-warehouse/
│   │   ├── pipelines/
│   │   ├── metrics/
│   │   ├── reports/
│   │   └── experimentation/
│   │
│   ├── mobile/                              ← Domain: Client applications
│   │   ├── architecture/
│   │   ├── guidelines/
│   │   ├── platforms/
│   │   ├── ux/
│   │   ├── deep-links/
│   │   └── offline/
│   │
│   ├── admin/                               ← Domain: Back-office
│   │   ├── portal/
│   │   ├── roles/
│   │   ├── tools/
│   │   ├── reports/
│   │   └── audit-logs/
│   │
│   ├── legal/                               ← Domain: Legal & compliance
│   │   ├── terms/
│   │   ├── privacy/
│   │   ├── compliance/
│   │   ├── regulatory/
│   │   └── data-retention/
│   │
│   ├── adr/                                 ← Architecture Decision Records
│   │
│   ├── prompts/                             ← AI assistant instructions
│   │   ├── engineering/
│   │   ├── review/
│   │   ├── generation/
│   │   └── agents/
│   │
│   ├── tasks/                               ← Engineering task specs
│   │   ├── epics/
│   │   ├── stories/
│   │   ├── spikes/
│   │   └── research/
│   │
│   ├── quality/                             ← Engineering quality standards
│   │   ├── gates/
│   │   ├── metrics/
│   │   ├── audits/
│   │   ├── definitions-of-done/
│   │   └── rfcs/
│   │
│   ├── release/                             ← Release management
│   │   ├── process/
│   │   ├── versioning/
│   │   ├── changelog/
│   │   ├── hotfix/
│   │   └── communications/
│   │
│   ├── checklists/                          ← Operational checklists
│   │   ├── pre-launch/
│   │   ├── deployment/
│   │   ├── security/
│   │   ├── code-review/
│   │   └── go-live/
│   │
│   └── templates/                           ← Reusable document templates
│       ├── adr/
│       ├── prd/
│       ├── runbook/
│       ├── incident/
│       ├── rfc/
│       ├── api-contract/
│       └── test-plan/
│
├── .ai/                                     ← AI assistant configuration
│   ├── context/
│   ├── rules/
│   └── prompts/
│
├── .github/                                 ← GitHub automation
│   ├── ISSUE_TEMPLATE/
│   ├── PULL_REQUEST_TEMPLATE/
│   └── workflows/
│
└── scripts/                                 ← Documentation automation scripts
```

**Total folders:** 135 folders
**Total depth:** Maximum 4 levels deep

---

## 2. EXPLANATION OF EVERY FOLDER

### TIER 0 — Root

| Folder | Purpose | Primary Owner |
|--------|---------|---------------|
| `fairride/` | Project root. Entry point for all contributors and AI assistants | CTO |
| `docs/` | All living documentation. No source code lives here | Entire Engineering Team |
| `.ai/` | System prompts, project context, and rules for AI assistants working on this codebase | CTO + Tech Leads |
| `.github/` | GitHub automation: PR templates, issue templates, CI/CD workflow definitions | DevOps Lead |
| `scripts/` | Shell scripts for documentation linting, link checking, and index generation | DevOps Lead |

---

### TIER 1 — Business Layer

#### `docs/business/`
Foundation of the entire EOS. All engineering decisions must trace back to a business reason.

| Sub-folder | Contains | Owner |
|------------|---------|-------|
| `mission/` | The FAIRRIDE mission statement, company purpose, and "why we exist" documents | CEO / CTO |
| `vision/` | 3-year and 10-year product vision, future platform roadmap narrative | CEO / CPO |
| `market/` | Target markets, city launch strategy, TAM/SAM/SOM analysis, user research | CPO / Business |
| `competitive/` | Competitive landscape analysis (Uber, Grab, Gojek, Lyft, inDrive), differentiation strategy | CPO / Strategy |
| `kpis/` | North Star metric, business KPIs, engineering KPIs, OKR framework, success criteria | CTO / CPO |

---

### TIER 2 — Product Layer

#### `docs/product/`
Translates business intent into product specifications. Everything the engineering team builds must trace to a document in this folder.

| Sub-folder | Contains | Owner |
|------------|---------|-------|
| `requirements/` | Master Product Requirements Document (PRD), functional vs. non-functional requirements, constraints | CPO / PM |
| `user-stories/` | Epics and user stories broken down by actor: Rider, Driver, Admin, Corporate, Merchant | PM |
| `roadmap/` | MVP scope definition, Phase 1, Phase 2, Phase 3 feature roadmap with dependencies | CPO / PM |
| `personas/` | Detailed user personas: who our riders are, who our drivers are, city archetypes | UX Lead / PM |
| `features/ride/` | Ride booking, trip flow, ride types (economy, premium, XL) feature specifications | PM |
| `features/driver/` | Driver onboarding, driver app experience, earnings, ratings, cancellation policies | PM |
| `features/admin/` | Admin portal capabilities, city management, driver management, analytics view | PM |
| `features/delivery/` | Package delivery feature specs (future phase, scoped now) | PM |
| `features/logistics/` | B2B logistics feature specs (future phase, scoped now) | PM |

---

### TIER 3 — Architecture Layer

#### `docs/architecture/`
The engineering blueprint. This is where the system design lives. Every service, every boundary, every integration is documented here before it is built.

| Sub-folder | Contains | Owner |
|------------|---------|-------|
| `overview/` | System architecture executive summary, key architectural principles, technology philosophy | Principal Architect |
| `system/` | Full system architecture document: services, databases, queues, CDN, external integrations | Principal Architect |
| `domain/` | Domain model: bounded contexts, aggregate roots, domain events, ubiquitous language glossary | Principal Architect |
| `services/` | Per-service design documents: responsibilities, APIs, data ownership, SLA targets | Tech Leads |
| `data-flow/` | Data flow diagrams for critical flows: booking flow, payment flow, dispatch flow, cancellation flow | Principal Architect |
| `diagrams/` | Source-of-truth architecture diagrams (C4 model: Context, Container, Component, Code) | Principal Architect |
| `boundaries/` | Service boundary decisions, team topology alignment, what is in-scope vs. out-of-scope for MVP | CTO |

---

### TIER 4 — Infrastructure Layers (Database, API, Security)

#### `docs/database/`
Defines the data persistence strategy. The single source of truth for data models before any migration is written.

| Sub-folder | Contains | Owner |
|------------|---------|-------|
| `models/` | Entity relationship diagrams, domain entity definitions, data type specifications | DB Architect |
| `schema/` | Logical schema design per service (not SQL DDL — conceptual schema documents) | DB Architect |
| `indexes/` | Index strategy per query pattern, composite index decisions, read optimization policies | DB Architect |
| `migrations/` | Migration strategy document, zero-downtime migration playbook, rollback procedures | DB Architect |
| `policies/` | Data partitioning policy, sharding strategy, archival policy, TTL rules | DB Architect |
| `replication/` | Replication topology, read replica strategy, cross-region data design | DB Architect + DevOps |

#### `docs/api/`
Defines all contracts between services and external clients. No API is built without a contract document approved here first.

| Sub-folder | Contains | Owner |
|------------|---------|-------|
| `contracts/` | API contract specifications per service: endpoints, request/response shapes, status codes | API Lead |
| `versioning/` | API versioning policy, deprecation lifecycle, backward-compatibility rules | API Lead |
| `authentication/` | Auth mechanisms: JWT, OAuth2, API keys. Token lifecycle, refresh strategy | Security Architect |
| `rate-limiting/` | Rate limiting tiers per client type, quota policies, throttling behavior specs | API Lead |
| `errors/` | Global error schema, error code registry, error message standards | API Lead |
| `webhooks/` | Webhook design: event catalog, delivery guarantees, retry policy, signature verification | API Lead |
| `open-api/` | Open API platform design for third-party developers (future phase, scoped now) | API Lead |

#### `docs/security/`
The security design. Written before any service is built. Every threat is documented. Every control is specified.

| Sub-folder | Contains | Owner |
|------------|---------|-------|
| `threat-model/` | STRIDE threat model for each service, attack surface analysis, trust boundary definitions | Security Architect |
| `authentication/` | Authentication design: multi-factor auth, phone OTP, biometric, session management | Security Architect |
| `authorization/` | Authorization model: RBAC design, permission matrix, role definitions | Security Architect |
| `encryption/` | Encryption standards: data at rest, data in transit, key management, PII handling | Security Architect |
| `compliance/` | PCI-DSS compliance scope, GDPR/PDPA compliance requirements, SOC2 readiness | Security + Legal |
| `incident-response/` | Security incident playbooks, escalation paths, breach response procedures | Security Architect |
| `penetration/` | Penetration testing scope, methodology, schedule, remediation tracking | Security Architect |

---

### TIER 5 — Engineering Practice Layers

#### `docs/deployment/`
How the system moves from code to production. No deployment happens without a documented process.

| Sub-folder | Contains | Owner |
|------------|---------|-------|
| `environments/` | Environment definitions: local, dev, staging, production, canary. Parity requirements | DevOps Lead |
| `pipelines/` | CI/CD pipeline design, build stages, test gates, promotion criteria | DevOps Lead |
| `infrastructure/` | Infrastructure architecture: cloud provider choice, region strategy, networking topology | DevOps Lead |
| `containers/` | Container strategy, image standards, registry policy, image scanning requirements | DevOps Lead |
| `runbooks/` | Deployment runbooks: blue-green deploy, canary release, rollback procedures | DevOps Lead |
| `rollback/` | Rollback decision tree, rollback procedures per service, data rollback considerations | DevOps Lead |

#### `docs/coding/`
The engineering craft standards. How every engineer writes code on this platform.

| Sub-folder | Contains | Owner |
|------------|---------|-------|
| `standards/` | Universal coding standards: naming, error handling, logging, comment policy | Principal Engineer |
| `style-guides/` | Language-specific style guides per technology choice | Tech Leads |
| `patterns/` | Approved design patterns: repository, saga, CQRS, event sourcing (where applicable) | Principal Architect |
| `reviews/` | Code review policy, review checklist, definition of "ready to merge" | Principal Engineer |
| `onboarding/` | New engineer onboarding guide: first day, first week, first month setup | Engineering Manager |
| `dependencies/` | Dependency management policy, approved library registry, license compliance | Principal Engineer |

#### `docs/testing/`
The complete testing strategy. Quality is built in, not bolted on.

| Sub-folder | Contains | Owner |
|------------|---------|-------|
| `strategy/` | Master test strategy: testing pyramid, coverage requirements, test ownership | QA Lead |
| `unit/` | Unit testing standards, mocking policy, test data management | QA Lead |
| `integration/` | Integration test design, test environment requirements, service contract testing | QA Lead |
| `e2e/` | End-to-end test scenarios by critical user journey, device matrix, test data | QA Lead |
| `load/` | Load testing strategy, target throughput, SLA validation, chaos test plans | QA Lead + DevOps |
| `security/` | Security testing policy, DAST/SAST requirements, dependency scanning | Security + QA |
| `chaos/` | Chaos engineering philosophy, failure injection scenarios, game day procedures | SRE |

---

### TIER 6 — Operations Layer

#### `docs/operations/`
How we run the platform in production. Written before go-live.

| Sub-folder | Contains | Owner |
|------------|---------|-------|
| `runbooks/` | Operational runbooks for every critical scenario: scaling, DB failover, cache flush | SRE |
| `on-call/` | On-call rotation policy, escalation matrix, severity levels, response time SLAs | SRE |
| `incident/` | Incident management process, war-room procedures, communication templates | SRE + Engineering Manager |
| `sla/` | Service Level Agreements per service, SLO targets, error budget policy | CTO + SRE |
| `capacity/` | Capacity planning model, traffic projection, scaling trigger thresholds | SRE |
| `postmortem/` | Postmortem template and policy, blameless culture rules, follow-up tracking | SRE |

#### `docs/monitoring/`
Observability design. We cannot operate what we cannot see.

| Sub-folder | Contains | Owner |
|------------|---------|-------|
| `metrics/` | Metrics taxonomy, RED/USE method definitions, business metrics definitions | SRE + Engineering |
| `alerting/` | Alert policy, alert routing, on-call notification design, fatigue prevention | SRE |
| `dashboards/` | Dashboard design per service and per domain, executive dashboard specs | SRE + Analytics |
| `logging/` | Log format standards, log levels, structured logging schema, log retention policy | SRE |
| `tracing/` | Distributed tracing strategy, span design, sampling policy, performance budget | SRE |
| `health-checks/` | Health check endpoints specification, readiness vs. liveness probe design | DevOps |

#### `docs/performance/`
Performance contracts. Written before code. Validated before release.

| Sub-folder | Contains | Owner |
|------------|---------|-------|
| `benchmarks/` | Performance benchmark targets per API endpoint, P50/P95/P99 latency budgets | Principal Engineer |
| `optimization/` | Performance optimization playbook: database query, caching strategy, CDN | Principal Engineer |
| `capacity/` | Infrastructure sizing model, auto-scaling rules, peak traffic projections | SRE |
| `profiling/` | Profiling methodology, profiling toolchain, hotspot identification process | Principal Engineer |

---

### DOMAIN DOCUMENTATION

#### `docs/fraud/` — Trust & Safety Domain
Fraud prevention is non-negotiable for a money-moving platform.

| Sub-folder | Contains | Owner |
|------------|---------|-------|
| `detection/` | Fraud detection model design: rule-based engine, ML signal design, risk scoring | Trust & Safety Lead |
| `prevention/` | Fraud prevention mechanisms: velocity checks, device fingerprinting, location anomaly | Trust & Safety Lead |
| `rules/` | Fraud rule catalog: fake GPS, account takeover, payment fraud, rating manipulation | Trust & Safety Lead |
| `policies/` | Fraud policy: penalty tiers, ban criteria, appeal process | Trust & Safety + Legal |
| `appeals/` | Appeal process design, evidence collection, review workflow | Trust & Safety |

#### `docs/payments/` — Payments Domain
How money flows through the FAIRRIDE ecosystem.

| Sub-folder | Contains | Owner |
|------------|---------|-------|
| `flow/` | End-to-end payment flow diagrams: authorization, capture, settlement, split payment | Payments Lead |
| `providers/` | Payment provider integration specs: Stripe, local rails, bank transfers | Payments Lead |
| `reconciliation/` | Reconciliation design: daily reconciliation process, discrepancy resolution | Payments + Finance |
| `refunds/` | Refund policy, refund flow design, partial refund logic, timeline SLAs | Payments Lead |
| `compliance/` | PCI-DSS scope, PSD2 considerations, local financial regulation requirements | Payments + Legal |
| `disputes/` | Dispute resolution flow, chargeback handling, driver vs. rider dispute process | Payments + Ops |

#### `docs/wallet/` — Digital Wallet Domain
The FAIRRIDE in-app wallet for riders and driver earnings.

| Sub-folder | Contains | Owner |
|------------|---------|-------|
| `design/` | Wallet architecture: ledger design, double-entry accounting model, balance computation | Payments Lead |
| `transactions/` | Transaction type catalog, transaction state machine, idempotency design | Payments Lead |
| `limits/` | Wallet limits: top-up limits, spend limits, withdrawal limits, KYC tiers | Payments + Legal |
| `policies/` | Wallet terms, inactivity policy, expiry rules for promotions/credits | Legal + Payments |
| `topup/` | Top-up flow design: payment method linkage, auto-topup, topup sources | Payments Lead |

#### `docs/dispatch/` — Dispatch Domain
The heart of the platform. Driver-rider matching is our core differentiator.

| Sub-folder | Contains | Owner |
|------------|---------|-------|
| `algorithm/` | Dispatch algorithm design: matching objective function, optimization approach, fairness model | Principal Engineer |
| `matching/` | Matching logic: proximity scoring, ETA calculation, acceptance rate weighting, driver preference | Principal Engineer |
| `routing/` | Route computation: shortest path, traffic-aware routing, route optimization for pooling | Principal Engineer |
| `pooling/` | Ride pooling design: pool eligibility, detour threshold, passenger merging logic | Principal Engineer |
| `fallback/` | Fallback dispatch: surge expansion radius, driver incentive triggers, manual override | Principal Engineer |
| `simulation/` | Dispatch simulation framework design: test market simulation, algorithm A/B testing methodology | Principal Engineer |

#### `docs/geo/` — Geospatial Domain
The location layer that everything depends on.

| Sub-folder | Contains | Owner |
|------------|---------|-------|
| `mapping/` | Map provider strategy, map rendering approach, offline map design | Geo Lead |
| `geofencing/` | Geofence design: city boundaries, airport zones, surge zones, restricted areas | Geo Lead |
| `tracking/` | Real-time tracking design: driver location update frequency, client interpolation, trail retention | Geo Lead |
| `zones/` | Zone system design: pricing zones, supply zones, demand heatmap zones | Geo Lead + Pricing |
| `providers/` | Map provider evaluation: Google Maps, Mapbox, HERE, OpenStreetMap — cost vs. capability | Geo Lead |
| `h3/` | H3 hexagonal grid design for demand forecasting, supply aggregation, surge computation | Geo Lead |

#### `docs/pricing/` — Pricing Domain
How fares are calculated fairly and transparently.

| Sub-folder | Contains | Owner |
|------------|---------|-------|
| `model/` | Base fare model: base charge, per-km, per-minute, minimum fare, vehicle type multipliers | Pricing Lead |
| `surge/` | Dynamic pricing design: demand/supply ratio algorithm, surge cap policy, transparency rules | Pricing Lead |
| `promotions/` | Promotion engine design: promo code system, referral mechanics, discount stacking rules | Pricing Lead |
| `corporate/` | Corporate pricing: negotiated rates, invoice billing, spend controls, department allocation | Pricing + B2B |
| `policies/` | Pricing governance: fare review process, city-specific pricing approval, price display rules | CPO + Legal |
| `estimation/` | Upfront fare estimation design: accuracy standards, re-pricing rules, partial trip pricing | Pricing Lead |

#### `docs/notifications/` — Notifications Domain
How we communicate with users at scale.

| Sub-folder | Contains | Owner |
|------------|---------|-------|
| `channels/` | Channel design: push (FCM/APNs), SMS, email, in-app. Channel fallback hierarchy | Engineering Lead |
| `templates/` | Notification template catalog by event type: trip events, payment events, promo events | Product + Engineering |
| `policies/` | Notification frequency caps, quiet hours, opt-in/opt-out policies, DND rules | PM + Legal |
| `providers/` | Provider evaluation: Firebase, Twilio, SendGrid, local SMS gateways | Engineering Lead |
| `preferences/` | User notification preference design: granular control, preference sync across devices | Engineering Lead |

#### `docs/analytics/` — Analytics Domain
Data is a product. We measure everything.

| Sub-folder | Contains | Owner |
|------------|---------|-------|
| `data-warehouse/` | Data warehouse design: schema strategy (dimensional/OBT), partitioning, retention | Data Lead |
| `pipelines/` | ETL/ELT pipeline design: CDC strategy, stream processing design, batch schedule | Data Lead |
| `metrics/` | Business metrics definitions: GMV, take rate, driver utilization, trip completion rate | Data Lead + CPO |
| `reports/` | Standard report catalog: daily ops report, finance report, driver earnings report | Data Lead |
| `experimentation/` | A/B testing framework design: experiment registry, statistical significance policy | Data Lead |

#### `docs/mobile/` — Mobile Platform Domain
The client applications used by 100% of our end users.

| Sub-folder | Contains | Owner |
|------------|---------|-------|
| `architecture/` | Mobile architecture: state management, offline-first design, API communication layer | Mobile Lead |
| `guidelines/` | Mobile engineering guidelines: performance budgets, battery optimization, data efficiency | Mobile Lead |
| `platforms/` | Platform-specific requirements: iOS and Android version support matrix, device requirements | Mobile Lead |
| `ux/` | UX design principles for mobile, accessibility standards, localization requirements | UX Lead |
| `deep-links/` | Deep link schema design: deferred deep links, universal links, marketing attribution | Mobile Lead |
| `offline/` | Offline capability design: what works offline, sync strategy, conflict resolution | Mobile Lead |

#### `docs/admin/` — Admin & Back-Office Domain
The internal tooling that operations teams use to run the city.

| Sub-folder | Contains | Owner |
|------------|---------|-------|
| `portal/` | Admin portal feature specifications: city management, driver management, trip management | PM + Ops |
| `roles/` | Admin role definitions: City Manager, Support Agent, Finance Ops, Engineering, Super Admin | PM + Security |
| `tools/` | Internal tooling specifications: driver ban tool, refund tool, fare adjustment tool | PM + Engineering |
| `reports/` | Admin report specifications: city health dashboard, weekly ops report, finance summary | PM + Data |
| `audit-logs/` | Audit log design: what is logged, log format, retention, access to audit logs | Security + Legal |

#### `docs/legal/` — Legal & Compliance Domain
Non-negotiable. Documented before launch.

| Sub-folder | Contains | Owner |
|------------|---------|-------|
| `terms/` | Terms of Service requirements for riders and drivers | Legal |
| `privacy/` | Privacy policy requirements, data subject rights, cookie policy | Legal |
| `compliance/` | Regulatory compliance matrix: financial licenses, transport licenses, data protection | Legal |
| `regulatory/` | Country-specific regulatory requirements per launch market | Legal |
| `data-retention/` | Data retention schedule: PII retention, trip data retention, financial data retention | Legal + Engineering |

---

### PROCESS & TOOLING DOCUMENTATION

#### `docs/adr/`
Architecture Decision Records. Every significant technical decision is recorded here with context, options considered, decision made, and consequences. ADRs are immutable after approval — they are never edited, only superseded.

#### `docs/prompts/`
The AI assistant library. Every AI task used in this project is documented here so it can be replicated consistently.

| Sub-folder | Contains |
|------------|---------|
| `engineering/` | Prompts for engineering tasks: code review, refactoring, debugging |
| `review/` | Prompts for document review, ADR drafting, RFC evaluation |
| `generation/` | Prompts for generating boilerplate, test cases, API contracts |
| `agents/` | Multi-step agent definitions: full service scaffolding, migration generation |

#### `docs/tasks/`
Engineering task specifications before they become tickets.

| Sub-folder | Contains |
|------------|---------|
| `epics/` | Large feature epic specifications with acceptance criteria |
| `stories/` | Individual user story specifications |
| `spikes/` | Time-boxed research task specifications |
| `research/` | Technical research reports from spikes |

#### `docs/quality/`
The engineering quality framework.

| Sub-folder | Contains |
|------------|---------|
| `gates/` | Quality gate definitions: what must pass before merge, before deploy, before release |
| `metrics/` | Quality metrics: test coverage floor, code complexity ceiling, duplication threshold |
| `audits/` | Scheduled technical audit specifications: quarterly security audit, annual architecture review |
| `definitions-of-done/` | Definitions of Done per artifact type: feature, service, migration, API |
| `rfcs/` | Request for Comments process and RFC index |

#### `docs/release/`
How software ships to users.

| Sub-folder | Contains |
|------------|---------|
| `process/` | Release process specification: release train, feature flags, release checklist |
| `versioning/` | Software versioning policy: semantic versioning rules, mobile app versioning |
| `changelog/` | Changelog format standard, CHANGELOG maintenance policy |
| `hotfix/` | Hotfix process: severity criteria, hotfix procedure, post-hotfix review |
| `communications/` | Release communication templates: internal engineering, customer-facing, partner communications |

#### `docs/checklists/`
Executable checklists that prevent mistakes.

| Sub-folder | Contains |
|------------|---------|
| `pre-launch/` | Pre-launch checklist: everything that must be verified before opening to users |
| `deployment/` | Deployment checklist: steps for every production deployment |
| `security/` | Security checklist: OWASP top 10, secrets, dependency audit |
| `code-review/` | Code review checklist for reviewers |
| `go-live/` | City go-live checklist: operational readiness for launching in a new market |

#### `docs/templates/`
Reusable document templates that enforce consistency.

| Sub-folder | Contains |
|------------|---------|
| `adr/` | ADR template |
| `prd/` | Product Requirements Document template |
| `runbook/` | Operational runbook template |
| `incident/` | Incident report and postmortem template |
| `rfc/` | Request for Comments template |
| `api-contract/` | API contract specification template |
| `test-plan/` | Test plan template |

#### `.ai/`
Machine-readable project context for AI assistants (Claude, Cursor, GitHub Copilot).

| Sub-folder | Contains |
|------------|---------|
| `context/` | Project context summary: what FAIRRIDE is, tech stack, domain glossary |
| `rules/` | AI assistant rules: what to avoid, conventions to follow, security constraints |
| `prompts/` | Ready-to-run prompts for common AI-assisted engineering tasks |

#### `.github/`
GitHub repository configuration.

| Sub-folder | Contains |
|------------|---------|
| `ISSUE_TEMPLATE/` | Standardized GitHub issue templates: bug, feature, ADR, security |
| `PULL_REQUEST_TEMPLATE/` | PR templates enforcing checklist completion |
| `workflows/` | CI/CD workflow specifications (documentation validation, link checking) |

---

## 3. DEPENDENCY MAP BETWEEN FOLDERS

Documentation is produced in dependency order. No folder is written before its dependencies are approved.

```
TIER 0 — FOUNDATION (no dependencies)
┌─────────────────────────────────────────────────────┐
│  business/mission                                   │
│  business/vision                                    │
│  legal/regulatory  (parallel track)                 │
│  .ai/context       (initial bootstrapping)          │
│  docs/templates/*  (parallel track)                 │
└─────────────────────────────────────────────────────┘
                         │
                         ▼
TIER 1 — DISCOVERY (depends on: mission, vision)
┌─────────────────────────────────────────────────────┐
│  business/market                                    │
│  business/competitive                               │
│  product/personas                                   │
│  legal/compliance                                   │
└─────────────────────────────────────────────────────┘
                         │
                         ▼
TIER 2 — REQUIREMENTS (depends on: discovery)
┌─────────────────────────────────────────────────────┐
│  product/requirements                               │
│  product/user-stories                               │
│  product/features/*                                 │
│  product/roadmap                                    │
│  business/kpis                                      │
└─────────────────────────────────────────────────────┘
                         │
                         ▼
TIER 3 — ARCHITECTURE (depends on: requirements)
┌─────────────────────────────────────────────────────┐
│  architecture/overview                              │
│  architecture/domain          (ubiquitous language) │
│  architecture/boundaries      (service boundaries)  │
│  architecture/system                                │
│  adr/* (ongoing — first ADRs are created here)     │
└─────────────────────────────────────────────────────┘
                         │
              ┌──────────┼──────────┐
              ▼          ▼          ▼
TIER 4 — DESIGN DOMAINS (depends on: architecture)
┌────────────┐  ┌────────────┐  ┌────────────────────┐
│ database/  │  │   api/     │  │   security/        │
│  models    │  │ contracts  │  │  threat-model      │
│  schema    │  │versioning  │  │  authentication    │
└────────────┘  └────────────┘  └────────────────────┘
              │          │          │
              └──────────┼──────────┘
                         │
              ┌──────────┼──────────────────┐
              ▼          ▼                  ▼
TIER 5 — DOMAIN SPECIFICATIONS (depends on: design)
┌────────────┐  ┌────────────┐  ┌────────────────────┐
│ dispatch/  │  │   geo/     │  │   pricing/         │
│ algorithm  │  │  mapping   │  │    model           │
│ matching   │  │   zones    │  │    surge           │
└────────────┘  └────────────┘  └────────────────────┘
┌────────────┐  ┌────────────┐  ┌────────────────────┐
│ payments/  │  │  wallet/   │  │   fraud/           │
│   flow     │  │  design    │  │  detection         │
│ providers  │  │   ledger   │  │  prevention        │
└────────────┘  └────────────┘  └────────────────────┘
┌────────────┐  ┌────────────┐  ┌────────────────────┐
│  mobile/   │  │ analytics/ │  │  notifications/    │
│architecture│  │ warehouse  │  │   channels         │
└────────────┘  └────────────┘  └────────────────────┘
                         │
                         ▼
TIER 6 — ENGINEERING PRACTICE (depends on: domain specs)
┌─────────────────────────────────────────────────────┐
│  coding/standards                                   │
│  coding/patterns                                    │
│  testing/strategy                                   │
│  deployment/environments                            │
│  deployment/infrastructure                          │
│  performance/benchmarks                             │
└─────────────────────────────────────────────────────┘
                         │
                         ▼
TIER 7 — OPERATIONS (depends on: engineering practice)
┌─────────────────────────────────────────────────────┐
│  monitoring/metrics                                 │
│  monitoring/alerting                                │
│  operations/runbooks                                │
│  operations/sla                                     │
│  operations/incident                                │
│  deployment/pipelines                               │
└─────────────────────────────────────────────────────┘
                         │
                         ▼
TIER 8 — QUALITY & RELEASE (depends on: operations)
┌─────────────────────────────────────────────────────┐
│  quality/gates                                      │
│  quality/definitions-of-done                        │
│  checklists/*                                       │
│  release/process                                    │
│  admin/portal                                       │
│  legal/terms + legal/privacy                        │
└─────────────────────────────────────────────────────┘
```

**Critical dependency rules:**
- `dispatch/` cannot be designed without `geo/zones` and `pricing/model`
- `payments/` cannot be designed without `security/encryption` and `legal/compliance`
- `wallet/` cannot be designed without `payments/flow` and `fraud/detection`
- `mobile/` cannot be designed without `api/contracts` and `architecture/system`
- `monitoring/` cannot be designed without `architecture/services` and `operations/sla`
- All `checklists/` depend on their corresponding domain specs

---

## 4. DOCUMENTATION CREATION ORDER

### Wave 1 — Foundation (Week 1)
1. `docs/templates/*` — templates first, so all subsequent docs follow format
2. `docs/business/mission`
3. `docs/business/vision`
4. `docs/legal/regulatory` (parallel, legal team)
5. `.ai/context` (bootstrap AI assistant context)
6. `.github/ISSUE_TEMPLATE/` + `PULL_REQUEST_TEMPLATE/`

### Wave 2 — Discovery (Week 1–2)
7. `docs/business/market`
8. `docs/business/competitive`
9. `docs/product/personas`
10. `docs/legal/compliance`
11. `docs/business/kpis`

### Wave 3 — Requirements (Week 2)
12. `docs/product/requirements`
13. `docs/product/user-stories`
14. `docs/product/features/ride`
15. `docs/product/features/driver`
16. `docs/product/features/admin`
17. `docs/product/roadmap`

### Wave 4 — Architecture (Week 3)
18. `docs/architecture/domain` (ubiquitous language first)
19. `docs/architecture/boundaries`
20. `docs/architecture/overview`
21. `docs/architecture/system`
22. `docs/architecture/data-flow` (booking, payment, dispatch flows)
23. `docs/adr/ADR-0001-*` through `ADR-0005-*` (first 5 ADRs)

### Wave 5 — Design: Data, API, Security (Week 3–4)
24. `docs/database/models`
25. `docs/database/schema`
26. `docs/security/threat-model`
27. `docs/security/authentication`
28. `docs/security/authorization`
29. `docs/security/encryption`
30. `docs/api/contracts` (core booking + driver APIs)
31. `docs/api/authentication`
32. `docs/api/errors`
33. `docs/api/versioning`

### Wave 6 — Domain Specifications (Week 4–5)
34. `docs/geo/mapping`
35. `docs/geo/zones`
36. `docs/geo/tracking`
37. `docs/dispatch/algorithm`
38. `docs/dispatch/matching`
39. `docs/dispatch/routing`
40. `docs/pricing/model`
41. `docs/pricing/surge`
42. `docs/payments/flow`
43. `docs/payments/providers`
44. `docs/payments/compliance`
45. `docs/wallet/design`
46. `docs/wallet/transactions`
47. `docs/fraud/detection`
48. `docs/fraud/rules`
49. `docs/notifications/channels`
50. `docs/mobile/architecture`
51. `docs/analytics/data-warehouse`

### Wave 7 — Engineering Practice (Week 5–6)
52. `docs/coding/standards`
53. `docs/coding/patterns`
54. `docs/coding/reviews`
55. `docs/testing/strategy`
56. `docs/testing/e2e` (critical user journeys)
57. `docs/deployment/environments`
58. `docs/deployment/infrastructure`
59. `docs/deployment/pipelines`
60. `docs/performance/benchmarks`

### Wave 8 — Operations (Week 6)
61. `docs/operations/sla`
62. `docs/monitoring/metrics`
63. `docs/monitoring/alerting`
64. `docs/monitoring/logging`
65. `docs/operations/runbooks`
66. `docs/operations/on-call`
67. `docs/operations/incident`

### Wave 9 — Quality & Launch Readiness (Week 7)
68. `docs/quality/gates`
69. `docs/quality/definitions-of-done`
70. `docs/checklists/pre-launch`
71. `docs/checklists/deployment`
72. `docs/checklists/security`
73. `docs/release/process`
74. `docs/admin/portal`
75. `docs/admin/roles`
76. `docs/legal/terms`
77. `docs/legal/privacy`

### Wave 10 — Remaining & Hardening (Week 7–8)
78. All remaining sub-documents in every domain
79. `docs/prompts/*`
80. `docs/tasks/epics` (all MVP epics)
81. `docs/checklists/go-live`

---

## 5. ENGINEERING WORKFLOW

### The Documentation-First Development Cycle

```
DISCOVER → SPECIFY → REVIEW → APPROVE → IMPLEMENT → VERIFY → RELEASE
    ↓           ↓         ↓        ↓           ↓          ↓        ↓
  docs/      docs/      ADR/     CTO       source     docs/    docs/
 business   product    RFC      sign-     code       testing  release
 product    arch/      review   off      written    testing  changelog
            api/
            domain
```

### Rule: No code without a document

Every engineering task must reference a document that describes what is being built. The traceability chain is:

```
Business Goal → KPI → Product Requirement → User Story → Architecture Decision
     → API Contract → Service Design → Implementation → Test → Release
```

### Gate Model

| Gate | What must exist | Who approves |
|------|----------------|--------------|
| Gate 0: Architecture | ADR + Architecture doc approved | CTO |
| Gate 1: Spec Complete | API contract + DB schema approved | Principal Architect |
| Gate 2: Implementation Start | Coding standards read, Definition of Done agreed | Tech Lead |
| Gate 3: Code Review | Review checklist complete, tests written | 2 Engineers |
| Gate 4: Staging | E2E tests passing, Security checklist cleared | QA Lead |
| Gate 5: Production | Pre-launch checklist complete, SLA confirmed, runbooks ready | CTO + SRE |

### How AI Assistants use this EOS

An AI assistant working on FAIRRIDE must:
1. Read `.ai/context/` to understand the project
2. Read `.ai/rules/` to understand constraints
3. Find the relevant domain document before writing any code
4. Reference the API contract before writing any API
5. Reference the security doc before handling any PII
6. Open an ADR if making a decision not already documented

---

## 6. DOCUMENTATION ROADMAP

| Week | Milestone | Output |
|------|-----------|--------|
| Week 1 | EOS Foundation approved | Templates, mission, vision, regulatory context |
| Week 2 | Product Requirements signed off | Full PRD, user stories, feature specs, personas |
| Week 3 | Architecture approved | System design, domain model, first 5 ADRs |
| Week 4 | Data + Security + API design approved | DB models, threat model, API contracts |
| Week 5 | All domain specifications approved | Dispatch, geo, pricing, payments, wallet, fraud |
| Week 6 | Engineering practice docs approved | Coding standards, test strategy, CI/CD design |
| Week 7 | Operations + Quality docs approved | Runbooks, SLAs, monitoring, release process |
| Week 8 | Launch-readiness docs approved | All checklists, legal docs, go-live plan |

**Documentation complete → Engineering implementation begins → Week 9**

---

## 7. ESTIMATED NUMBER OF DOCUMENTS

| Folder | Est. Documents | Notes |
|--------|---------------|-------|
| `business/` | 8 | Mission, vision, market, competitive, KPIs |
| `product/` | 18 | PRD, personas, 5 feature areas × 2–3 docs each |
| `architecture/` | 14 | Overview, system, domain, services (1 per service ~8), data-flow diagrams |
| `database/` | 12 | Models, schema per service, index strategies, migration policy |
| `api/` | 20 | Contracts per service (~12), versioning, auth, errors, webhooks |
| `security/` | 12 | Threat model per service boundary, auth, encryption, incident playbooks |
| `deployment/` | 10 | Environment specs, pipeline design, infra, runbooks |
| `coding/` | 8 | Standards, style guides, patterns, review policy |
| `testing/` | 12 | Strategy + plans per layer |
| `operations/` | 15 | Runbooks (~6), on-call, incident, SLA, postmortem template |
| `monitoring/` | 12 | Metrics taxonomy, alert catalog, dashboards, logging schema |
| `performance/` | 6 | Benchmark targets, optimization playbook, capacity model |
| `fraud/` | 8 | Detection model, rule catalog, policy, appeals |
| `payments/` | 12 | Flow diagrams, provider specs, reconciliation, refund, compliance |
| `wallet/` | 8 | Ledger design, transaction catalog, limits, topup |
| `dispatch/` | 10 | Algorithm, matching, routing, pooling, fallback, simulation plan |
| `geo/` | 10 | Mapping, geofencing, tracking, zones, H3 grid design |
| `pricing/` | 10 | Model, surge, promotions, corporate, estimation |
| `notifications/` | 8 | Channel design, template catalog, provider comparison |
| `analytics/` | 10 | Warehouse design, pipelines, metrics definitions, reports |
| `mobile/` | 10 | Architecture, guidelines, platform matrix, UX principles, deep links |
| `admin/` | 8 | Portal spec, role matrix, tool specs, reports |
| `legal/` | 8 | Terms requirements, privacy requirements, compliance matrix |
| `adr/` | 25 | Estimated 25 ADRs for MVP architecture decisions |
| `prompts/` | 12 | Engineering, review, generation, and agent prompts |
| `tasks/` | 30 | Epics (~8), stories (~15), spikes (~7) |
| `quality/` | 8 | Gates, metrics, DoD per artifact type |
| `release/` | 6 | Process, versioning policy, changelog format, hotfix process |
| `checklists/` | 10 | Pre-launch, deployment, security, code review, go-live |
| `templates/` | 8 | ADR, PRD, runbook, incident, RFC, API contract, test plan |
| `.ai/` | 5 | Project context, rules, ready-to-run prompts |
| `.github/` | 5 | Issue templates, PR template, workflow specs |
| **TOTAL** | **~340 documents** | Phase 0 to MVP launch |

---

## 8. ESTIMATED PAGES PER DOCUMENT

| Document Category | Pages | Reasoning |
|------------------|-------|-----------|
| Mission / Vision | 2–3 | Concise, directional |
| Market / Competitive Analysis | 8–12 | Substantive research required |
| Business KPIs | 4–6 | Metrics table + rationale |
| Product Requirements (PRD) | 20–30 | Comprehensive feature specification |
| User Personas | 4–6 each | Research + profile |
| User Stories per feature | 6–10 | Multiple stories with acceptance criteria |
| Product Roadmap | 5–8 | Phase breakdown with milestones |
| Architecture Overview | 10–15 | Diagrams + narrative |
| System Architecture | 20–25 | Full service topology |
| Domain Model | 10–15 | Entity catalog + bounded contexts |
| Per-Service Design | 5–8 each | API, data, SLA, dependencies |
| Data Flow Diagrams | 3–5 each | Sequence + state machine |
| Database Schema | 8–12 | Entity list + field catalog |
| Index Strategy | 3–5 | Query patterns + index decisions |
| API Contract per service | 6–10 | Endpoints, payloads, error codes |
| Threat Model | 12–18 | STRIDE per boundary |
| Authentication Design | 8–12 | Flows, token design, edge cases |
| Dispatch Algorithm | 15–20 | Math-heavy specification |
| Geo / H3 Design | 8–12 | Grid design, zone catalog |
| Pricing Model | 10–15 | Formula, variable catalog, edge cases |
| Payment Flow | 12–18 | State machine, error recovery |
| Wallet / Ledger Design | 10–14 | Double-entry model, transaction catalog |
| Fraud Detection Design | 12–16 | Signal catalog, rule engine, ML signals |
| Notification Design | 6–10 | Channel matrix, template catalog |
| Analytics / Data Warehouse | 12–18 | Schema design, pipeline design |
| Mobile Architecture | 10–14 | State, networking, offline |
| Coding Standards | 8–12 | Language-specific rules |
| Test Strategy | 8–12 | Pyramid, coverage policy |
| CI/CD Pipeline Design | 8–12 | Stages, gates, environment flow |
| SLA Document | 6–10 | Service-by-service SLO/SLA |
| Monitoring / Metrics Design | 10–14 | Metric catalog, alert policy |
| Operational Runbook (each) | 4–8 | Step-by-step procedure |
| ADR (each) | 2–4 | Context, options, decision, consequences |
| Quality Gates | 4–6 | Gate criteria per stage |
| Release Process | 6–8 | Release train, approval flow |
| Pre-Launch Checklist | 5–8 | Executable checklist |
| Go-Live Checklist | 6–10 | City launch readiness |
| Legal Terms Requirements | 5–8 | Coverage requirements for legal |
| RFC Template | 2–3 | Reusable |
| AI Context Document | 4–6 | Project summary for AI |

**Average document length:** ~9 pages
**Total estimated pages:** ~3,060 pages across 340 documents

---

## 9. NAMING CONVENTION

### Rule: All filenames are lowercase, hyphen-separated, and version-suffixed.

#### Pattern
```
{category}-{topic}-{subtopic}-v{major}.{minor}.md
```

#### Examples

| Document | Filename |
|----------|---------|
| System Architecture | `arch-system-overview-v1.0.md` |
| Dispatch Algorithm | `dispatch-algorithm-design-v1.0.md` |
| Payment Flow | `payments-flow-rider-trip-v1.2.md` |
| ADR #001 | `ADR-0001-database-selection-v1.0.md` |
| Pricing Model | `pricing-model-base-fare-v1.0.md` |
| Threat Model | `sec-threat-model-booking-service-v1.1.md` |
| API Contract | `api-contract-booking-v2-v1.0.md` |
| Pre-Launch Checklist | `checklist-pre-launch-mvp-v1.0.md` |
| Test Plan | `test-plan-dispatch-service-v1.0.md` |
| Runbook | `runbook-database-failover-v1.0.md` |
| Driver Feature Spec | `feat-driver-onboarding-v1.0.md` |
| OKR Document | `kpi-okr-q3-2026-v1.0.md` |

#### ADR Naming
```
ADR-{NNNN}-{decision-summary}-v{major}.{minor}.md
```
Example: `ADR-0001-technology-stack-backend-v1.0.md`

#### RFC Naming
```
RFC-{NNNN}-{proposal-summary}-v{major}.{minor}.md
```
Example: `RFC-0001-dispatch-algorithm-approach-v1.0.md`

#### Diagram Naming
```
diagram-{type}-{subject}-v{major}.{minor}.{ext}
```
Example: `diagram-c4-context-fairride-v1.0.png`

#### Runbook Naming
```
runbook-{scenario}-{trigger}-v{major}.{minor}.md
```
Example: `runbook-db-failover-primary-down-v1.0.md`

#### Rules
- NO spaces in filenames
- NO camelCase or PascalCase
- NO underscores (hyphens only)
- NO uppercase except for: `ADR-`, `RFC-`, `README.md`, `CHANGELOG.md`
- Version suffix is mandatory on all non-template files
- `.md` extension for all documentation

---

## 10. VERSIONING CONVENTION

### Document Versioning follows `MAJOR.MINOR` format

#### Version Schema

| Version | Meaning | Example trigger |
|---------|---------|-----------------|
| `v0.x` | Draft — not yet approved | Initial authoring, peer review |
| `v1.0` | Approved — baseline | CTO/lead sign-off |
| `v1.x` | Minor revision | Corrections, clarifications, additions that do not change decisions |
| `v2.0` | Major revision | A fundamental change in approach, breaking change to design |
| `deprecated` | Superseded | Replaced by a newer version or ADR |

#### ADR Versioning Exception
ADRs are append-only. An ADR is **never modified** after approval.
- Corrections → new ADR that supersedes the old one
- The old ADR gains status: `Superseded by ADR-NNNN`

#### Version Header (required in every document)

```markdown
---
document-id: arch-system-overview
version: v1.2
status: approved          # draft | in-review | approved | deprecated | superseded
owner: @principal-architect
approved-by: @cto
last-reviewed: 2026-07-15
next-review: 2026-10-15
supersedes: arch-system-overview-v1.1.md
related-adrs: ADR-0001, ADR-0003
---
```

#### Lifecycle States

```
draft → in-review → approved → deprecated
                       ↓
                    (major revision)
                       ↓
                  new version (v2.0 draft)
```

#### Version History Table (in every document)

```markdown
| Version | Date | Author | Change Summary |
|---------|------|--------|---------------|
| v0.1 | 2026-07-01 | @architect | Initial draft |
| v0.2 | 2026-07-03 | @architect | Added threat model section |
| v1.0 | 2026-07-05 | @cto | Approved |
| v1.1 | 2026-08-01 | @architect | Clarified zone boundary behavior |
```

---

## 11. DOCUMENTATION MAINTENANCE POLICY

### Ownership Model

Every document has exactly one **Owner** and one **Approver**.

| Role | Responsibility |
|------|---------------|
| Owner | Maintains the document, initiates reviews, merges changes |
| Approver | Signs off on new versions. No document is approved without sign-off |
| Reviewer | Any engineer may review. Required for `in-review` → `approved` transition |

### Review Cadence

| Document Category | Review Frequency | Trigger |
|------------------|-----------------|---------|
| Mission / Vision | Annually | Strategic planning cycle |
| Product Requirements | Per release cycle | Any feature change |
| Architecture docs | Quarterly | Any system change |
| API contracts | Per API change | Breaking change mandatory |
| Security docs | Quarterly + incident-driven | Any security event |
| ADRs | Never revised — superseded only | New decision required |
| Runbooks | After every incident | Postmortem action item |
| Checklists | Per launch | New launch type |
| Coding standards | Bi-annually | Technology change |
| SLA documents | Quarterly | Performance trend |
| Legal docs | Annually + regulatory change | Regulatory update |

### Change Process

```
1. Author creates PR with document change
2. PR references the document-id and version bump in title
3. At least 1 peer review required for Minor bump
4. At least 2 peer reviews + owner approval for Major bump
5. Approver signs off (comment + PR approval)
6. PR merged → version-stamped file replaces previous
7. Previous version is archived in git history (NOT deleted)
8. If change impacts ADR: new ADR must be opened simultaneously
```

### Staleness Detection

- All documents with `next-review` date past the current date are **flagged as stale**
- A weekly automated check reports stale documents in the engineering channel
- Stale `approved` documents are automatically moved to `needs-review` status after 30 days past review date
- Stale `draft` documents with no activity for 14 days are flagged to the owner

### Deprecation Policy

Documents are deprecated when:
1. The feature or service they describe no longer exists
2. A newer document fully supersedes them
3. The architectural decision has been reversed via ADR

A deprecated document:
- Has its header updated: `status: deprecated`
- Has a banner at the top: `> DEPRECATED: Replaced by [new-doc-name]. Do not use for implementation.`
- Is **never deleted** from the repository (git history preserves all)
- Is moved to a `_deprecated/` sub-folder within its parent

### Documentation Quality Standards

Every approved document must have:
- [ ] Valid frontmatter header with all required fields
- [ ] Version history table
- [ ] Table of contents (for docs > 5 pages)
- [ ] Clear ownership attribution
- [ ] No TODOs or placeholders in approved documents
- [ ] All referenced documents and ADRs must exist and be linked
- [ ] Diagrams must have alt-text and be stored in `docs/architecture/diagrams/`
- [ ] No hardcoded environment values (no prod URLs, no credentials, no IPs)

### Enforcement

- **Pull Request gate:** Documentation PRs are validated by automated linter for:
  - Valid frontmatter
  - Correct filename convention
  - No broken internal links
  - Version bump present
- **Merge protection:** No document can be merged `approved` without the Approver's GitHub review approval
- **Monthly audit:** Documentation audit run monthly by the Principal Architect to catch stale and orphaned docs

---

## SUMMARY STATISTICS

| Metric | Value |
|--------|-------|
| Total folders | 135 |
| Maximum folder depth | 4 levels |
| Estimated total documents | ~340 |
| Estimated total pages | ~3,060 |
| Documentation waves | 10 |
| Estimated calendar time to complete | 8 weeks |
| Team members required | 6–8 (CTO, Architect, PM, Security, DevOps, Data, Legal) |
| Documents requiring CTO approval | ~35 |
| ADRs planned for MVP | ~25 |
| Templates to produce first | 7 |

---

## APPROVAL

| Role | Name | Signature | Date |
|------|------|-----------|------|
| CTO | | _______________ | |
| Principal Architect | | _______________ | |
| CPO | | _______________ | |

---

**END OF PHASE 0 — ENGINEERING OPERATING SYSTEM BLUEPRINT**

**STOP. This document is awaiting CTO approval before Phase 1 begins.**
**No documentation content is written. No source code is written.**
**All decisions after this point require an approved Phase 0.**
