---
document_id: DOC-0001
title: Project Constitution
version: 0.1.0
status: Draft
classification: Internal — All Engineering
owner: CTO
approved_by: ~
depends_on: []
required_by:
  - DOC-0002  # project-vision
  - DOC-0003  # project-principles
  - DOC-0004  # business-model
  - DOC-0005  # success-metrics
  - DOC-0006  # product-strategy
  - DOC-0007  # roadmap
  - DOC-0008  # competitor-analysis
  - "[ALL subsequent EOS documents — DOC-0009 through DOC-0072+]"
related_adrs:
  - ADR-0001  # backend-technology-stack (pending creation)
  - ADR-0002  # mobile-technology-stack (pending creation)
  - ADR-0003  # admin-web-technology-stack (pending creation)
  - ADR-0004  # documentation-first-engineering-process (pending creation)
  - ADR-0005  # microservices-architecture-approach (pending creation)
  - ADR-0006  # event-driven-communication-pattern (pending creation)
glossary_refs:
  - EOS
  - SSOT
  - ADR
  - RFC
  - Domain
  - Service
  - Platform
  - MVP
  - Driver
  - Rider
  - Trip
  - Fare
  - Dispatch
  - Wallet
created: 2026-06-30
last_updated: 2026-06-30
next_review: 2027-06-30
---

# FAIRRIDE — Project Constitution

> This document is the supreme law of the FAIRRIDE Engineering Organization.
> All engineering decisions, all technical designs, all process choices, and all
> AI-generated artifacts must conform to the principles established herein.
> No document, no individual, and no automated system may override this Constitution
> without completing the formal amendment process defined in Article XI.

---

## TABLE OF CONTENTS

1. [Purpose](#1-purpose)
2. [Scope](#2-scope)
3. [Dependencies](#3-dependencies)
4. [Related Documents](#4-related-documents)
5. [Definitions](#5-definitions)
6. [Functional Description](#6-functional-description)
   - [Article I — Engineering Charter](#article-i--engineering-charter)
   - [Article II — Core Values and Engineering Principles](#article-ii--core-values-and-engineering-principles)
   - [Article III — The Engineering Operating System](#article-iii--the-engineering-operating-system)
   - [Article IV — Technology Philosophy and Approved Stack](#article-iv--technology-philosophy-and-approved-stack)
   - [Article V — Governance and Decision-Making Authority](#article-v--governance-and-decision-making-authority)
   - [Article VI — Document Hierarchy and Authority](#article-vi--document-hierarchy-and-authority)
   - [Article VII — Engineering Standards Mandate](#article-vii--engineering-standards-mandate)
   - [Article VIII — Platform Scope Definition](#article-viii--platform-scope-definition)
   - [Article IX — Security and Privacy Mandate](#article-ix--security-and-privacy-mandate)
   - [Article X — AI-Assisted Engineering Policy](#article-x--ai-assisted-engineering-policy)
   - [Article XI — Constitutional Amendments](#article-xi--constitutional-amendments)
7. [Non-Functional Requirements](#7-non-functional-requirements)
8. [Constraints](#8-constraints)
9. [Risks](#9-risks)
10. [Future Extension](#10-future-extension)
11. [Open Questions](#11-open-questions)
12. [Decision References (ADR)](#12-decision-references-adr)
13. [Revision History](#13-revision-history)

---

## 1. PURPOSE

This document establishes the foundational charter, values, governance model, and operating principles of the FAIRRIDE Engineering Organization. It is the first and supreme document in the FAIRRIDE Engineering Operating System (EOS).

Every engineering decision made at FAIRRIDE — whether made by a human engineer, a product manager, a security architect, or an AI coding agent — MUST be traceable to a principle or rule established in this Constitution or in a document that is itself authorized by this Constitution.

This document exists to solve three persistent engineering problems that destroy platforms at scale:

**Problem 1 — Inconsistency:** Without a supreme reference, teams make contradictory decisions. Services are built to different standards. APIs have incompatible error formats. Databases have inconsistent naming. The result is a system that cannot be maintained.

**Problem 2 — Undocumented decisions:** When decisions are made verbally or implicitly, they cannot be reviewed, challenged, or traced. Over time, the system accumulates invisible constraints that nobody understands. Onboarding new engineers becomes impossible.

**Problem 3 — Short-term optimism:** Engineering teams under delivery pressure routinely trade long-term maintainability for short-term speed. Without an explicit, approved standard that defines what is non-negotiable, this trade-off happens constantly and silently until the system collapses under its own technical debt.

This Constitution resolves all three problems by making the foundational rules explicit, approved, versioned, and immutable.

---

## 2. SCOPE

### 2.1 Who Is Bound

This Constitution applies to ALL of the following without exception:

- All full-time engineers at FAIRRIDE
- All engineering contractors and freelancers
- All product managers making technical decisions
- All QA engineers and DevOps engineers
- All security architects
- All engineering managers
- All AI coding agents, AI assistants, and automated code generation tools operating on the FAIRRIDE codebase or documentation
- All external technology partners with write access to any FAIRRIDE system

### 2.2 What Is Governed

This Constitution governs:

- All software components comprising the FAIRRIDE Platform
- All engineering documentation in the FAIRRIDE EOS
- All CI/CD pipelines and deployment processes
- All data systems: databases, caches, message queues, data warehouses
- All API contracts: internal and external
- All security practices and controls
- All monitoring, alerting, and operational processes
- All AI assistant interactions with the codebase and documentation

### 2.3 What Is Not Governed

This Constitution does not govern:

- Business strategy and commercial decisions (governed by executive leadership)
- Product pricing strategy (governed by product and commercial teams, implemented per PRICING_ENGINE)
- Legal strategy (governed by legal counsel)
- Organizational structure and HR policies (governed by People Operations)
- Personal tooling choices that do not affect shared systems or codebase

---

## 3. DEPENDENCIES

| Type | Document | Reason |
|------|----------|--------|
| Upstream | None | DOC-0001 is the root document of the EOS. It has no upstream dependencies. |
| Context | PHASE-0-EOS-BLUEPRINT | The EOS folder structure, naming convention, versioning convention, and documentation maintenance policy established in Phase 0 are incorporated by reference into this Constitution. |

---

## 4. RELATED DOCUMENTS

### 4.1 Documents Immediately Required By This Constitution

| Document ID | Title | Relationship |
|-------------|-------|-------------|
| DOC-0002 | Project Vision | Must be consistent with Article I of this Constitution |
| DOC-0003 | Project Principles | Must elaborate on Article II of this Constitution |
| DOC-0004 | Business Model | Must be consistent with Article VIII (Platform Scope) |
| DOC-0005 | Success Metrics | Must be consistent with Article I (Engineering Charter) |
| DOC-0006 | Product Strategy | Must be consistent with Articles II, IV, VIII |
| DOC-0007 | Roadmap | Must be consistent with Article VIII (Platform Scope Definition) |
| DOC-0008 | Competitor Analysis | Must be consistent with Article I (Mission Alignment) |

### 4.2 All Subsequent EOS Documents

Every document in the FAIRRIDE EOS (DOC-0009 through DOC-0072+) MUST:
- Reference this Constitution in their Related Documents section
- Not contradict any Article herein
- Derive their authority from this Constitution

---

## 5. DEFINITIONS

The following terms are defined here and constitute the FAIRRIDE Engineering Vocabulary. Every subsequent EOS document MUST use these terms consistently. If a new term is required, it MUST be introduced in the document where it first appears and included in that document's Definitions section with a proposal to elevate it to the canonical glossary via RFC.

| Term | Definition |
|------|-----------|
| **FAIRRIDE** | The company, product, and ecosystem described by this EOS. Refers to both the organization and the platform it operates. |
| **Platform** | The complete technical ecosystem of FAIRRIDE, including all backend services, data systems, mobile applications, web interfaces, and external integrations. |
| **EOS (Engineering Operating System)** | The complete body of documentation, processes, standards, and tools that govern how the FAIRRIDE Platform is designed, built, operated, and evolved. The EOS is the system within which engineering happens. |
| **SSOT (Single Source of Truth)** | The principle that every piece of information in the FAIRRIDE ecosystem exists in exactly one authoritative location. Duplication of authoritative data is prohibited. |
| **Constitution** | This document (DOC-0001). The supreme governing document of the FAIRRIDE Engineering Organization. |
| **Document ID** | A unique, immutable identifier assigned to every EOS document. Format: `DOC-NNNN` where NNNN is a zero-padded sequential integer. |
| **ADR (Architecture Decision Record)** | An immutable, append-only record of a significant architectural or technical decision, including its context, the alternatives considered, the decision made, and its consequences. |
| **RFC (Request for Comments)** | A proposal document inviting structured engineering community input before a significant decision is formalized as an ADR. |
| **Domain** | A bounded area of business logic within the FAIRRIDE Platform, owned by a specific team or individual engineer. Examples: Dispatch, Pricing, Payments, Geo, Fraud. |
| **Service** | An independently deployable software component that owns a Domain. A Service has its own data store, its own API, and its own SLA. No Service may directly access another Service's data store. |
| **Bounded Context** | The explicit boundary within which a Domain model applies. The same concept may have different meanings in different Bounded Contexts. These differences MUST be documented. |
| **Contract** | A formal, versioned specification of the interface between two systems. The primary form of Contract in FAIRRIDE is the API Contract. No Service may be built without an approved Contract. |
| **Rider** | A human user who requests and receives transportation services through the FAIRRIDE Platform. Also referred to as "Passenger" in vehicle-specific contexts. |
| **Driver** | A human transportation provider registered and verified on the FAIRRIDE Platform who fulfills trip requests. |
| **Fleet Operator** | A legal entity that owns and manages a fleet of vehicles and one or more Drivers registered on the FAIRRIDE Platform. |
| **Merchant** | A business entity participating in FAIRRIDE's commerce ecosystem (e.g., delivery pickup location, food vendor). Applicable from Phase 3 onwards. |
| **Corporate Account** | An enterprise customer who has negotiated a business agreement with FAIRRIDE for employee transportation. Applicable from Phase 2 onwards. |
| **Admin / Operator** | An internal FAIRRIDE employee using the admin portal to manage city operations, drivers, disputes, and configuration. |
| **Trip** | A single completed transportation transaction. A Trip begins when a Driver accepts a Rider's request and ends when the Rider is delivered to their destination and payment is confirmed. |
| **Fare** | The price charged to a Rider for a Trip, calculated by the Pricing Engine. |
| **Dispatch** | The real-time process of receiving a Rider's trip request, identifying eligible Drivers, selecting the optimal Driver, and notifying them. |
| **Matching** | The specific algorithmic step within Dispatch that scores and ranks eligible Drivers for a given Rider request. |
| **Surge** | A dynamic pricing multiplier applied when demand exceeds supply in a given zone and time window. |
| **Wallet** | A digital account within the FAIRRIDE Platform that stores a monetary balance on behalf of a Rider or Driver. Wallets participate in the FAIRRIDE financial ledger. |
| **Ledger** | The double-entry accounting record of all financial transactions within the FAIRRIDE Platform. Every debit has a corresponding credit. |
| **KYC (Know Your Customer)** | The process of verifying a user's identity to meet regulatory requirements. KYC is required for Drivers, Fleet Operators, and any user whose Wallet exceeds defined thresholds. |
| **PII (Personally Identifiable Information)** | Any data that, alone or in combination, can identify a specific natural person. Examples: name, phone number, email address, national ID, GPS location history. All PII is subject to heightened protection standards defined in SECURITY_PRINCIPLES. |
| **SLA (Service Level Agreement)** | A formal commitment to service performance, expressed as measurable targets. Example: the Dispatch Service MUST achieve 99.9% availability. |
| **SLO (Service Level Objective)** | A specific measurable target within an SLA. Example: P99 latency for the booking API endpoint MUST be below 500ms. |
| **Error Budget** | The maximum acceptable downtime or error rate permitted within a given SLA period. |
| **On-Call** | An engineering rotation responsible for monitoring production systems and responding to incidents within defined response time SLAs. |
| **Incident** | Any unplanned interruption to, or degradation of, Platform services that affects real users or violates an SLA. |
| **Postmortem** | A structured, blameless review of an Incident, documenting the timeline, root cause, impact, and actions to prevent recurrence. |
| **MVP (Minimum Viable Product)** | The first production release of the FAIRRIDE Platform, scoped to deliver core ride-hailing functionality to real users in a single launch city. |
| **Phase** | A major release milestone that groups related product features and technical capabilities. The FAIRRIDE Platform is developed in Phases: MVP, Phase 1, Phase 2, Phase 3 (see Article VIII). |
| **Gate** | A mandatory quality checkpoint that a code change, document, or deployment MUST pass before advancing to the next stage. |
| **Definition of Done (DoD)** | A shared, documented checklist that defines when any work artifact — feature, service, API, document — is considered complete and ready for its next lifecycle stage. |
| **Documentation-First** | The principle that a design document MUST exist and be approved before implementation begins. Code is never the first artifact for any non-trivial work. |
| **PCI-DSS** | Payment Card Industry Data Security Standard. A set of security standards applicable to any system that stores, processes, or transmits cardholder data. |
| **GDPR / PDPA** | General Data Protection Regulation (Europe) / Personal Data Protection Act (Southeast Asia). Privacy regulations that govern the collection, storage, and use of personal data. |
| **GMV (Gross Merchandise Value)** | The total value of all Trips completed on the FAIRRIDE Platform in a given period, before FAIRRIDE's commission is deducted. The primary top-line business metric. |
| **Take Rate** | FAIRRIDE's commission percentage on each Trip. The primary revenue metric. |

---

## 6. FUNCTIONAL DESCRIPTION

---

### Article I — Engineering Charter

#### 1.1 Statement of Purpose

The FAIRRIDE Engineering Organization exists to build, operate, and continuously improve the technological foundation of the fairest ride-hailing ecosystem.

Engineering at FAIRRIDE is not a support function. It is a core business function. The platform is the product. The quality of the engineering directly determines the quality of the experience for every Rider, every Driver, and every city where FAIRRIDE operates.

#### 1.2 Mission Alignment

The mission of FAIRRIDE is to build the fairest ride-hailing ecosystem.

"Fairest" is not a marketing word. It is an engineering constraint. It means:

- **Fair to Riders:** Transparent pricing. No hidden fees. Accurate ETAs. Reliable service.
- **Fair to Drivers:** Fair earnings distribution. No arbitrary deactivations. Transparent performance metrics. Sustainable income.
- **Fair to Cities:** Compliance with local regulation. Contribution to urban mobility. No race-to-the-bottom competitive behavior.
- **Fair in Algorithm:** The Dispatch algorithm MUST NOT systematically disadvantage any group of users. Any algorithmic bias MUST be measurable and remediable.

Every engineering decision MUST be evaluated against this definition of fairness. If a technical choice creates unfairness for any participant — even if it improves a business metric — it MUST be escalated to the CTO for review before implementation.

#### 1.3 Engineering Mandate

The FAIRRIDE Engineering Organization is mandated to:

1. **Build a platform that serves tens of millions of users** at a reliability standard no lower than 99.9% availability for critical paths, without degradation of response time as volume grows.
2. **Start small and stay honest:** The MVP is intentionally lightweight. Engineering MUST NOT prematurely optimize for scale that does not yet exist. Premature optimization is a form of engineering waste.
3. **Prepare for scale without building it:** All architectural decisions MUST be made with awareness of the 10x, 100x, and 1,000x traffic scenarios, even if implementation defers until scale demands it.
4. **Earn trust by being reliable:** Reliability is not an operational concern. It is an engineering contract. Every service, every API, every database query is part of this contract.
5. **Protect users:** Every system that touches user data, user money, or user safety carries the full weight of FAIRRIDE's responsibility to those users. This is not negotiable.

#### 1.4 Organizational Responsibility

The Engineering Organization is collectively responsible for:

- Defining and maintaining the architecture of the FAIRRIDE Platform
- Building and releasing software that meets defined quality gates
- Operating the Platform in production with defined reliability standards
- Protecting user data and financial assets from unauthorized access or misuse
- Producing and maintaining engineering documentation as a first-class deliverable
- Ensuring that any AI assistant contributing to the codebase operates within approved boundaries

#### 1.5 The Engineering Promise

FAIRRIDE makes the following engineering promises to its users:

| Promise | Engineering Commitment |
|---------|----------------------|
| Rides are available when I need one | Dispatch Service SLA: 99.9% availability |
| My fare is what was quoted | Pricing Engine: zero variance between estimate and final charge unless explicitly communicated |
| My payment is safe | PCI-DSS compliance; end-to-end encryption of payment data |
| My location data is private | PII policy; location data is never sold or shared without explicit consent |
| If something goes wrong, it is fixed fast | Incident response SLA defined per severity |
| My driver earnings are accurate | Ledger integrity: double-entry accounting; automated reconciliation |

---

### Article II — Core Values and Engineering Principles

The following seven values are non-negotiable. They are the basis for all engineering decisions. When a technical choice violates one of these values, the choice MUST be reconsidered, not the value.

#### 2.1 Fairness First

**Definition:** The platform must be objectively fair to all participants. Fairness is a measurable engineering property, not a feeling.

**In practice:**
- Dispatch algorithms MUST be documented, auditable, and bias-tested before deployment.
- Pricing algorithms MUST produce consistent results for identical inputs. Personalized pricing that disadvantages users based on protected characteristics is prohibited.
- Driver earnings calculations MUST be deterministic and reconcilable. Drivers MUST be able to verify their own earnings independently.
- Any A/B test that affects pricing, earnings, or service availability MUST be reviewed by the ethics review process before launch.

**Engineering test:** "Can we prove this decision is fair to Riders, Drivers, and Cities? Can we measure fairness and detect when it degrades?"

#### 2.2 Reliability as Contract

**Definition:** Reliability is a promise to users, not a best-effort aspiration. Every service has a defined SLA. Meeting that SLA is an engineering obligation, not a target.

**In practice:**
- Every service MUST have a documented SLA before going to production.
- Every SLA breach MUST trigger an incident, regardless of user-visible impact.
- Every incident MUST produce a postmortem.
- Every postmortem MUST produce actionable items with owners and deadlines.
- Error budgets govern release velocity. When an error budget is exhausted, the team stops shipping features and focuses exclusively on reliability.

**Engineering test:** "What is our SLA? Are we currently meeting it? What happens when we do not?"

#### 2.3 Security by Default

**Definition:** Security is not a feature to be added later. Every system is built secure from the first line of the first document. The cost of retrofitting security is 10x the cost of building it in.

**In practice:**
- All threat models MUST be completed before service implementation begins.
- All PII MUST be identified, classified, and protected per the ENCRYPTION standard before the system handles real data.
- No secret, credential, key, or token MUST ever appear in source code, documentation, or logs.
- Every API endpoint MUST require authentication. Unauthenticated endpoints are a deliberate exception requiring explicit ADR approval.
- The principle of least privilege applies to all service-to-service communication, all admin access, and all database permissions.

**Engineering test:** "What is the worst thing an attacker could do with access to this system? Have we mitigated it?"

#### 2.4 Simplicity Over Cleverness

**Definition:** A simple solution that works is always preferred over a complex solution that is elegant. Complexity is a liability, not an asset. Every unnecessary layer of abstraction is future maintenance burden.

**In practice:**
- Before introducing a new abstraction, a new framework, or a new dependency, the engineer MUST articulate the specific problem it solves that cannot be solved more simply.
- Three similar lines of code are better than a premature abstraction.
- A new service MUST be justified by a domain boundary, not by a desire to isolate code.
- Technology choices are made to solve today's proven problems, not tomorrow's hypothetical ones.
- Performance optimizations are applied only where profiling data proves a bottleneck, not preemptively.

**Engineering test:** "Is there a simpler way to achieve the same outcome? Have we tried it?"

#### 2.5 Data Over Opinion

**Definition:** Engineering decisions are made from evidence. When data and opinion conflict, data wins. When no data exists, we build the minimal system needed to generate data, then decide.

**In practice:**
- All product and engineering hypotheses MUST be instrumented with metrics before release.
- Dispatch algorithm changes MUST be validated through simulation and A/B testing with defined success metrics before full rollout.
- Capacity planning is based on measured traffic growth, not intuition.
- Performance budgets are set based on user research and competitive benchmarks, not internal preference.
- Fraud rules are based on measured signal distributions, not assumptions about fraudster behavior.

**Engineering test:** "What data supports this decision? If we have no data, what is the cheapest experiment to get it?"

#### 2.6 Documentation Before Code

**Definition:** Documentation is a first-class engineering deliverable. The specification of a system is written and approved before the system is implemented. This is not bureaucracy — it is the practice that makes complex systems maintainable by teams that change over time.

**In practice:**
- No feature MUST be implemented without an approved design document.
- No API MUST be implemented without an approved API Contract.
- No database schema MUST be implemented without an approved data model.
- No service boundary MUST be crossed without a defined service contract.
- No significant technical decision MUST be made without an ADR.
- The complete specification MUST be understandable by a competent engineer who was not involved in creating it.
- AI coding agents MUST read all relevant domain documents before generating any artifact.

**Engineering test:** "If this engineer left tomorrow, could a new engineer understand and maintain this system from the documentation alone?"

#### 2.7 Long-Term Thinking Over Short-Term Speed

**Definition:** FAIRRIDE is building a platform intended to operate for decades. Every decision that trades long-term maintainability for short-term delivery speed degrades the platform. This trade-off is never acceptable without explicit, documented approval.

**In practice:**
- Technical debt is never "temporary." Every instance of accepted technical debt MUST be tracked in the backlog with a remediation plan and a deadline.
- "We'll fix it later" requires a ticket, an owner, and a delivery commitment. Without these three, the statement is not acceptable.
- Backward-incompatible API changes require a deprecation period and a migration plan. Breaking changes to production APIs without notice are prohibited.
- Database migrations MUST be zero-downtime capable. Any migration that requires downtime MUST be approved by the CTO.
- Vendor lock-in is evaluated explicitly. Any dependency on a proprietary API that has no viable alternative MUST be reviewed and accepted via ADR.

**Engineering test:** "How much will this decision cost to undo or maintain in two years? Five years? Is that acceptable?"

---

### Article III — The Engineering Operating System

#### 3.1 What the EOS Is

The FAIRRIDE Engineering Operating System (EOS) is the complete body of:

- **Documentation:** All design documents, API contracts, data models, runbooks, standards, and policies
- **Processes:** All engineering workflows from design through deployment to operations
- **Governance:** All rules governing how decisions are made, documented, and enforced
- **Tools:** All infrastructure, pipelines, and automation that support the engineering process

The EOS is not a project management system. It is not a ticketing system. It is not a wiki. It is a structured, versioned, dependency-managed repository of engineering truth.

#### 3.2 The Single Source of Truth Principle

The EOS is the SSOT for all engineering decisions. This principle has three implications:

**3.2.1 No duplication of authoritative information.** If a decision, rule, or specification appears in more than one place, exactly one of those places is authoritative and the others are stale by definition. Every piece of information MUST have a single authoritative home.

**3.2.2 The document, not the person, is the authority.** When an engineer says "I know how this works, I built it," and their knowledge contradicts the document, the document is authoritative and MUST be updated or superseded through the proper process — not bypassed. People leave. Documents stay.

**3.2.3 If it is not documented, it does not exist.** An undocumented decision is not a decision — it is an opinion. Undocumented behavior is not a feature — it is an accident waiting to be broken. Any system behavior that is not documented MUST be treated as a defect.

#### 3.3 EOS Document Tiers

EOS documents are organized in ten tiers, each requiring the tier above to be approved before it can be authored. Tier 1 is supreme.

| Tier | Level | Contains |
|------|-------|---------|
| 1 | Constitution | This document. Supreme authority. |
| 2 | Business Foundation | Vision, principles, business model, success metrics, strategy, roadmap |
| 3 | Product Definition | Requirements, personas, journeys, feature catalog, MVP scope |
| 4 | System Architecture | System architecture, service boundaries, event design, data flow |
| 5 | Data Architecture | Database design, data dictionary, entity relationships |
| 6 | API Governance | API guidelines, naming, errors, authentication, versioning |
| 7 | Security Framework | Security principles, RBAC, authentication flows, encryption, audit |
| 8 | Domain Engines | Dispatch, Pricing, Matching, Payments, Wallet, Fraud, Geo, Notifications, Analytics |
| 9 | Engineering Practice | Coding standards, language standards, testing, documentation standards |
| 10 | Operations | CI/CD, deployment, observability, monitoring, logging, backup, DR, scaling |
| 11 | AI & Quality | AI rulebook, AI agent rules, prompt library, review process, quality gates, DoD |

#### 3.4 Documentation-First Enforcement

The Documentation-First principle (Value 2.6) is enforced at the following gates:

| Gate | What Must Exist Before Proceeding |
|------|----------------------------------|
| Architecture Gate | ADR + Architecture document approved by CTO |
| Design Gate | API contract + data model approved by Principal Architect |
| Implementation Gate | Coding standard read; DoD agreed; design document approved by Tech Lead |
| Code Review Gate | Review checklist complete; tests authored; design document referenced in PR |
| Staging Gate | E2E tests passing; security checklist cleared; runbook drafted |
| Production Gate | Pre-launch checklist complete; SLA confirmed; runbook approved; monitoring live |

No gate may be bypassed for any reason, including delivery deadlines, without CTO written approval. Such approvals MUST be documented as technical debt with a remediation plan.

---

### Article IV — Technology Philosophy and Approved Stack

#### 4.1 Platform Principles

The FAIRRIDE Platform is built on the following non-negotiable architectural principles:

**Cloud-Native by Design.** The Platform MUST be designed to run entirely on managed cloud infrastructure. No physical servers. No self-hosted infrastructure in the critical path. Operational burden MUST be minimized through managed services where they do not introduce unacceptable vendor lock-in risk.

**API-First.** Every capability of the Platform MUST be exposed through a well-defined API before it is consumed. Internal and external consumers are treated identically. No service may call another service's database directly.

**Mobile-First User Experience.** The primary user interface of FAIRRIDE is the mobile application. Every product decision MUST consider the mobile experience as primary and the web experience as secondary. Network latency, battery consumption, and offline resilience are first-class engineering concerns.

**Event-Driven Where State Changes Matter.** Significant state changes within the Platform — trip state transitions, payment events, driver status changes — MUST be propagated via a durable event system. This enables auditability, decoupling, and real-time reactivity across services without tight coupling.

**Stateless Services.** Services MUST be horizontally scalable. No critical business state MUST reside in service memory. All state MUST be persisted in a durable store managed independently of the service process.

**Idempotency by Design.** All operations that modify state MUST be idempotent. This is non-negotiable for all payment operations, dispatch operations, and any operation that could be retried by network infrastructure or by client code.

#### 4.2 Architecture Philosophy

**Domain-Driven.** Service boundaries follow business domain boundaries, not technical convenience. A service exists because a domain exists, not because a developer wanted to separate code.

**Shared Nothing.** Services do not share databases. Services do not share message queues (except where explicitly designed as a fan-out mechanism). The only permitted mechanism of service-to-service communication is via API or via durable events.

**Designed for Failure.** Every service is designed with the assumption that its dependencies will fail. Services MUST implement circuit breakers, timeouts, and graceful degradation for all external calls. The system MUST remain partially functional when components fail.

**Observability as Architecture.** Observability is not instrumented after the fact. Metrics, logging, and distributed tracing are designed into every service as core requirements, defined before implementation, not added as afterthoughts.

#### 4.3 Technology Selection Criteria

When evaluating any technology choice, the following criteria MUST be applied in priority order:

1. **Proven at scale:** Has this technology been proven by companies operating at similar or greater scale to our 5-year target? Unproven experimental technology is not permitted in critical paths.
2. **Operational maturity:** Are managed services available? Is the operational burden understood and acceptable?
3. **Team expertise or acquirability:** Does the team have expertise, or can it be acquired within the hiring plan?
4. **Ecosystem and community:** Is the ecosystem active? Are security patches regularly released? Is there a viable long-term support commitment?
5. **Cost at scale:** What does this technology cost at 10x current volume? Is it predictable?
6. **Vendor concentration risk:** Does this choice create unacceptable dependency on a single vendor?
7. **License compatibility:** Is the license compatible with FAIRRIDE's commercial model?

Technology MUST NOT be chosen for novelty, personal preference, or resume enhancement.

#### 4.4 Approved Technology Stack

The following technology choices are the approved FAIRRIDE stack. They are in effect for all new service development. Any deviation MUST be approved via ADR.

| Layer | Technology | Rationale ADR |
|-------|-----------|--------------|
| Backend Services | Go (Golang) | ADR-0001 (pending) |
| Mobile Applications (iOS + Android) | Flutter | ADR-0002 (pending) |
| Web Admin Console | Next.js | ADR-0003 (pending) |

**Important:** The approval of this Constitution constitutes provisional approval of these technology choices. The ADRs listed above MUST be produced within 14 days of this Constitution's approval. Until ADRs are produced, no language-specific style guides or standards documents may be authored.

Language-specific standards (`go-standard`, `flutter-standard`, `nextjs-standard`) are downstream of these ADRs and MUST NOT be authored until the ADRs are approved.

#### 4.5 Infrastructure Philosophy

Infrastructure specifics — cloud provider, region strategy, containerization approach, orchestration platform — are governed by downstream EOS documents and their respective ADRs. This Constitution establishes only the following infrastructure principles:

- No infrastructure choice MUST lock FAIRRIDE into a single cloud provider for the critical path
- Containerized deployment is the baseline for all services
- Infrastructure-as-Code is mandatory; no infrastructure resource MUST be created through a cloud console without a corresponding code definition
- All infrastructure changes MUST pass through CI/CD and require the same review process as application code

---

### Article V — Governance and Decision-Making Authority

#### 5.1 Decision Framework

Decisions in the FAIRRIDE Engineering Organization are classified by scope and reversibility:

| Class | Scope | Reversibility | Authority | Process |
|-------|-------|--------------|-----------|---------|
| Class 1 — Constitutional | Entire EOS | Irreversible | CTO + full review | Amendment process (Article XI) |
| Class 2 — Architectural | Cross-service, long-lived | Expensive | CTO + Principal Architect | ADR required |
| Class 3 — Domain | Single domain, medium-lived | Moderate | Domain Tech Lead + Architecture review | ADR recommended, design doc required |
| Class 4 — Implementation | Single service, short-lived | Easy | Any Engineer | Code review |
| Class 5 — Operational | Production runtime | Immediate | On-call Engineer | Runbook; documented post-resolution |

Class 1 and Class 2 decisions are permanent until explicitly reversed through a new ADR. They MUST NOT be made implicitly, verbally, or through convention. Every Class 2 decision MUST have an ADR.

#### 5.2 Authority Levels

| Role | Decisions They Own |
|------|--------------------|
| **CTO** | Class 1 (Constitutional) and final authority on Class 2. Escalation endpoint. Platform scope changes. Technology stack changes. |
| **Principal Architect** | Class 2 (Architectural) in collaboration with CTO. Owns the architecture review process. Cross-domain design decisions. |
| **Tech Lead** | Class 3 (Domain). Owns design decisions within their domain. Must escalate Class 2 items. |
| **Senior Engineer** | Class 4 (Implementation). Owns implementation approach within defined architecture. |
| **Engineer** | Class 4 within reviewed PRs. Participates in RFC process. |
| **Engineering Manager** | Process and team structure decisions. Not technical architecture decisions. |
| **On-Call Engineer** | Class 5 (Operational) during active incidents only. |

No engineer MUST make a Class 2 decision unilaterally, regardless of seniority. The absence of a Principal Architect does not authorize unilateral Class 2 decisions — it requires escalation to the CTO.

#### 5.3 ADR Process

An Architecture Decision Record (ADR) MUST be created for every Class 2 decision and is strongly recommended for significant Class 3 decisions.

**When an ADR is mandatory:**
- Technology stack selection or change
- Introduction of a new database technology
- Service boundary creation, merger, or split
- API versioning strategy change
- Security protocol change
- Data retention policy change
- Third-party integration selection for critical path
- Any decision that, if wrong, would require more than one sprint to reverse

**ADR Lifecycle:**
```
Proposed → In Review → Approved → Superseded
```

ADRs are immutable after approval. A decision to reverse an ADR requires a new ADR that explicitly supersedes it, explaining the reason for the reversal.

**ADR Template:** Governed by the template in `docs/templates/adr/`.

#### 5.4 RFC Process

A Request for Comments (RFC) is used when a significant decision needs broad engineering community input before being formalized as an ADR. An RFC is a discussion document, not a decision document.

**When an RFC is appropriate:**
- Proposing a new engineering process that affects all teams
- Proposing a cross-domain API standard
- Evaluating multiple viable technology alternatives where team input is valuable
- Proposing a change to the EOS structure

An RFC expires after 30 days of inactivity with no decision. The RFC author or CTO may extend it or close it. A closed RFC with a decision becomes an ADR.

#### 5.5 Escalation Path

When an engineer is uncertain about the appropriate authority for a decision:

```
Engineer → Tech Lead → Principal Architect → CTO
```

No step may be skipped. An engineer who disagrees with a Tech Lead's decision escalates to the Principal Architect. A Tech Lead who disagrees with a Principal Architect's decision escalates to the CTO. The CTO's decision is final unless it requires a Constitutional Amendment.

---

### Article VI — Document Hierarchy and Authority

#### 6.1 Document Supremacy Order

When two EOS documents conflict, the document in the higher tier takes precedence. Specifically:

```
Constitution (DOC-0001)
    └── Business Foundation (DOC-0002 to DOC-0008)
            └── Product Definition (DOC-0009 to DOC-0014)
                    └── System Architecture (DOC-0015 to DOC-0020)
                            └── Data Architecture (DOC-0021 to DOC-0025)
                                    └── API Governance (DOC-0026 to DOC-0032)
                                            └── Security Framework (DOC-0033 to DOC-0040)
                                                    └── Domain Engines (DOC-0041 to DOC-0049)
                                                            └── Engineering Practice (DOC-0050 to DOC-0057)
                                                                    └── Operations (DOC-0058 to DOC-0065)
                                                                            └── AI & Quality (DOC-0066 to DOC-0072)
```

An ADR may override a document in a lower tier. An ADR CANNOT override the Constitution. An ADR that conflicts with the Constitution is null and void until the Constitution is amended.

#### 6.2 Conflict Detection and Resolution

Any engineer who identifies a conflict between two EOS documents MUST:

1. Stop any implementation work that depends on the conflicting information
2. Document the conflict in writing (GitHub issue or RFC)
3. Reference both conflicting documents and the specific sections in conflict
4. Escalate to the owner of the lower-tier document for resolution
5. If unresolved within 48 hours, escalate to the Principal Architect

No engineer MUST resolve a document conflict by choosing the interpretation that is more convenient for their implementation. The process above is mandatory.

#### 6.3 Document Lifecycle

All EOS documents follow this lifecycle:

```
draft (v0.x) → in-review → approved (v1.0) → [minor revision (v1.x)] → deprecated
                                    ↓
                              [major revision requires new draft (v2.0)]
```

An approved document is immutable. Changes require creating a new version through the PR and review process. ADRs are permanently immutable — they are superseded, never edited.

---

### Article VII — Engineering Standards Mandate

This Article establishes mandatory baseline standards. Specific standards are governed by downstream EOS documents. This Article establishes what is non-negotiable at the Constitutional level.

#### 7.1 Code Quality Standards

**All code MUST:**
- Pass automated linting configured to the approved language-specific style guide before merge
- Pass all automated tests before merge
- Meet the minimum test coverage floor defined in the Testing Standard
- Be reviewed by at least one other engineer before merge
- Reference the design document or ADR that authorized the feature being implemented

**All code MUST NOT:**
- Contain secrets, credentials, or environment-specific configuration values
- Contain commented-out code that is not explained with a specific reason and ticket reference
- Introduce a dependency not listed in the approved dependency registry without Tech Lead approval
- Bypass a quality gate through any mechanism (including `--no-verify`, `--force`, or equivalent)

#### 7.2 Testing Mandate

Testing is not optional. The following minimum test requirements apply to all production services:

| Test Type | Requirement |
|-----------|-------------|
| Unit Tests | MUST exist for all business logic. Coverage floor defined in Testing Standard. |
| Integration Tests | MUST exist for all service boundaries and external integrations |
| Contract Tests | MUST exist for all published APIs (consumer-driven contract testing) |
| E2E Tests | MUST exist for all critical user journeys as defined in the Product Requirements |
| Load Tests | MUST be executed before any production release and on any change to a critical path |
| Security Tests | SAST MUST run in CI on every commit. DAST MUST run before every production release. |

The absence of tests for a production service is a critical defect, not a technical debt item. No service goes to production without tests.

#### 7.3 API Standards Mandate

All APIs at FAIRRIDE MUST:
- Have an approved API contract before implementation begins
- Use the canonical error format defined in the Error Standard (DOC-0028)
- Implement authentication via the approved mechanism defined in the Authentication Standard (DOC-0029)
- Be versioned per the Versioning Standard (DOC-0031)
- Respect rate limiting as defined per the API Guidelines (DOC-0026)
- Be idempotent for all state-mutating operations

No API MUST be deployed to production without passing the API Gate in the CI/CD pipeline. The API Gate validates the running service against its approved contract.

#### 7.4 Operational Standards Mandate

No service MUST be deployed to production without:

1. A documented and approved operational runbook covering the following minimum scenarios: normal startup, graceful shutdown, deployment rollback, database failover, dependency degradation
2. Health check endpoints (readiness probe and liveness probe)
3. Structured logging in the approved format
4. Metrics emission in the approved format
5. Distributed tracing instrumentation
6. An alert definition for the service's SLO
7. Assignment to an on-call rotation

---

### Article VIII — Platform Scope Definition

#### 8.1 Long-Term Vision

The FAIRRIDE Platform is designed to eventually support the following product lines. This vision defines the architecture targets and MUST inform every architectural decision, even when a product line is not yet in scope for implementation.

| Product Line | Description | Phase Target |
|-------------|-------------|-------------|
| Ride | Standard passenger vehicle hailing | MVP |
| Taxi | Licensed taxi integration | Phase 1 |
| Bike | Two-wheel vehicle hailing | Phase 1 |
| EV | Electric vehicle specific features | Phase 2 |
| Delivery | Package and document delivery | Phase 2 |
| Food | Restaurant food delivery | Phase 3 |
| Logistics | B2B freight and logistics | Phase 3 |
| Corporate Transport | Enterprise employee transportation | Phase 2 |
| Merchant Platform | Commercial partner ecosystem | Phase 3 |
| Open API Platform | Developer ecosystem and third-party integrations | Phase 4 |

**Critical architectural implication:** The architecture MUST support the multi-product future even though the MVP implements only the Ride product. Service boundaries, data models, and APIs MUST be designed to accommodate additional product lines without requiring rearchitecting of existing services.

#### 8.2 MVP Scope

The MVP is the minimum set of capabilities required to operate a real ride-hailing service in a single city with real users. The MVP scope is:

**Rider Capabilities:**
- Account registration with phone OTP verification
- Booking a standard Ride (Economy tier)
- Real-time driver tracking on map
- In-app payment (card and wallet)
- Trip history and receipt
- Rating and tipping the driver
- Cancellation with policy-driven fee

**Driver Capabilities:**
- Account registration with KYC document submission
- Online/offline status control
- Receiving and accepting trip requests
- Navigation assistance
- Earnings dashboard (daily/weekly)
- Payout to bank account
- Rating visibility

**Admin Capabilities:**
- Driver approval and document management
- Trip management and dispute resolution
- Refund processing
- Pricing configuration per city
- Basic city health dashboard

**Platform Capabilities (not user-visible):**
- Real-time dispatch engine
- Dynamic pricing engine
- Payment processing and ledger
- In-app wallet
- Notifications (push, SMS)
- Fraud detection (baseline rule set)
- Admin portal
- Basic analytics

#### 8.3 Out of MVP Scope

The following capabilities are explicitly out of scope for the MVP. Any request to include them in the MVP MUST be approved by the CTO with a written justification.

- Multiple vehicle categories beyond Economy (Premium, XL, EV)
- Ride pooling / shared rides
- Scheduled rides
- Corporate billing and accounts
- Bike or taxi integration
- Delivery or logistics
- Food delivery
- Open API for third-party developers
- Multi-city operations management
- International payments or multi-currency wallet
- Machine learning-based fraud detection (baseline rules only for MVP)
- Driver incentive programs beyond basic trip completion

#### 8.4 Architecture Must Support What Is Not Yet Built

Even though the above items are out of MVP scope, the following architectural rules apply:

- The fare calculation model MUST support multiple vehicle categories, even if only Economy is configured
- The user account model MUST support multiple actor roles, even if only Rider and Driver are active
- The dispatch engine MUST be designed as a pluggable matching algorithm, not hardcoded logic
- The notification system MUST support multiple channel types, even if only push and SMS are configured for MVP
- The payment system MUST support multiple currencies, even if only one is active for MVP
- The admin portal MUST support multi-city configuration from day one, even if only one city is launched

---

### Article IX — Security and Privacy Mandate

#### 9.1 Security by Design

Security at FAIRRIDE is a design requirement, not a post-launch activity. The following rules are inviolable:

**Threat Modeling is Mandatory.** Every service MUST have a completed STRIDE threat model before implementation begins. Threat models MUST be reviewed by the Security Architect before the service moves to implementation.

**Least Privilege.** Every process, service, user, and AI agent operates with the minimum permissions required to perform its function. Permissions are never granted broadly when they can be granted specifically.

**Defense in Depth.** No single security control is relied upon. Critical protections MUST exist at multiple layers: network, application, data, and operations.

**Security Testing is Non-Optional.** Penetration testing MUST be completed before the first production launch and after any significant architectural change. The security testing scope is defined in SECURITY_PRINCIPLES (DOC-0033).

**Secrets Management.** All secrets (API keys, database credentials, encryption keys, service tokens) MUST be managed through an approved secrets management system. Secrets MUST be rotated on a schedule defined in the Encryption Standard.

#### 9.2 PII Policy

The following rules govern all PII in the FAIRRIDE Platform:

| Rule | Requirement |
|------|------------|
| Classification | All data fields containing PII MUST be explicitly tagged in the Data Dictionary |
| Encryption at Rest | All PII at rest MUST be encrypted using approved algorithms |
| Encryption in Transit | All data in transit MUST be encrypted using TLS 1.2 or higher |
| Access Logging | All access to PII MUST be logged to the audit log with requestor identity and timestamp |
| Minimization | Only PII required for the stated business purpose MUST be collected |
| Retention | PII MUST be deleted or anonymized per the Data Retention Policy (governed by legal) |
| Location Data | Trip location histories are PII. They MUST be retained only as long as legally required and never shared without explicit user consent |

#### 9.3 Financial Data Protection

FAIRRIDE handles real money on behalf of real users. The following rules are absolute:

- All payment data MUST be handled in compliance with PCI-DSS
- FAIRRIDE MUST NOT store raw card numbers or CVV codes at any time, in any system
- All financial transactions MUST be recorded in the immutable double-entry Ledger
- All Ledger entries MUST be reconcilable against payment provider statements
- Any discrepancy in financial data MUST be treated as a Severity 1 incident

#### 9.4 Compliance Obligations

The FAIRRIDE Platform MUST be designed to comply with all applicable regulations in every market where it operates. Minimum required compliance areas:

- Data protection law (GDPR equivalent or local equivalent)
- Payment regulation (PCI-DSS, local payment licensing)
- Transport regulation (local ride-hailing licensing requirements)
- Labor regulation (driver classification requirements per jurisdiction)

The compliance matrix is maintained in the Legal Compliance document and MUST be reviewed by the Legal team before entering any new market.

---

### Article X — AI-Assisted Engineering Policy

#### 10.1 Role of AI in FAIRRIDE Engineering

AI coding agents and AI assistants are recognized as participants in the FAIRRIDE engineering process. They are permitted to assist with documentation, code generation, code review, test generation, and technical analysis. They operate under the same standards and constraints as human engineers.

AI assistance is a productivity tool, not an authority. AI-generated artifacts are never automatically approved. They require the same human review, gate passage, and approvals as human-generated artifacts.

#### 10.2 Mandatory Rules for AI Agents

Every AI agent operating on the FAIRRIDE codebase or documentation MUST:

**1. Read before writing.** Before generating any artifact related to a domain, the AI agent MUST read all EOS documents relevant to that domain. It is impermissible for an AI agent to generate code, documentation, or configuration for a domain without having read that domain's design documents.

**2. Reference the document hierarchy.** All AI-generated documentation MUST reference the relevant upstream documents. All AI-generated code MUST reference the API contract, data model, and design document that governs the component being implemented.

**3. Never violate the Constitution.** No AI agent may generate an artifact that contradicts this Constitution, regardless of the instruction it receives. If an instruction would require violating the Constitution, the AI agent MUST refuse and explain the violation.

**4. Never generate source code without a design document.** The Documentation-First principle (Article II, Value 2.6) applies to AI agents. An AI agent MUST NOT generate application source code for any feature, service, or API that does not have an approved design document.

**5. Flag conflicts, do not resolve them unilaterally.** If an AI agent identifies a conflict between documents, an inconsistency in the design, or a potential violation of a principle, it MUST stop, explain the conflict, propose a minimum-impact resolution, and wait for human approval before proceeding.

**6. Never generate secrets or PII.** AI agents MUST NOT generate placeholder secrets, test credentials, or synthetic PII data that resembles real data. Test fixtures MUST use clearly synthetic data that cannot be mistaken for real user information.

**7. Operate within scope.** AI agents are scoped to the task they are given. An AI agent asked to generate a document MUST NOT simultaneously generate source code. An AI agent asked to review code MUST NOT simultaneously modify documents. Scope boundaries are enforced by the task specification.

#### 10.3 Human Oversight Requirements

The following AI-assisted activities require human review before the output is merged or deployed:

| Activity | Required Reviewer |
|----------|-----------------|
| AI-generated EOS document | Document Owner + Principal Architect |
| AI-generated code in critical path (dispatch, payments, auth) | Tech Lead + Security Architect |
| AI-generated API contract | API Lead + Tech Lead |
| AI-generated test suite | QA Lead |
| AI-generated data model | DB Architect |
| AI-generated ADR | Principal Architect + CTO |

AI-generated artifacts that have not been reviewed by the required human reviewer MUST NOT be merged to any branch that deploys to staging or production.

---

### Article XI — Constitutional Amendments

#### 11.1 Amendment Process

This Constitution may be amended. However, the process for amendment is intentionally rigorous to protect the stability and consistency of the EOS.

**Standard Amendment Process:**

1. Any engineer may propose a Constitutional amendment by opening an RFC
2. The RFC MUST explain: what is being changed, why it is being changed, what the impact is on all downstream EOS documents, and what the migration plan is
3. The RFC MUST be open for review for a minimum of 7 calendar days
4. The RFC requires sign-off from: Principal Architect, all Tech Leads, and CTO
5. Upon approval, the amendment is incorporated into a new version of this document (minor version bump for clarifications; major version bump for principle changes)
6. All downstream documents affected by the amendment MUST be updated within 30 days of the amendment's approval

**No amendment may be made to this Constitution through:**
- A verbal agreement
- An ADR (ADRs operate below the Constitutional tier)
- A code comment or inline documentation
- An emergency process (see 11.2 for exceptions)

#### 11.2 Emergency Amendments

In the event of a constitutional conflict that is blocking critical business operations, the CTO may issue a "Constitutional Directive" — a temporary override of a specific principle for a specific, bounded scope. A Constitutional Directive:

- MUST be documented in writing within 24 hours
- MUST define an expiry date no more than 30 days from issuance
- MUST specify the exact scope of the override
- MUST trigger the standard amendment process within 7 days of issuance
- Is automatically void after its expiry date

No Constitutional Directive may be renewed. If the situation requiring it persists beyond 30 days, the standard amendment process MUST be used.

#### 11.3 Immutable Clauses

The following clauses MUST NOT be amended under any circumstances:

1. **Article I, Section 1.2 — Fairness First:** The principle that the FAIRRIDE Platform must be fair to Riders, Drivers, and Cities is inviolable. It may be elaborated but never removed or weakened.
2. **Article II, Section 2.3 — Security by Default:** The principle that security is never optional is inviolable.
3. **Article III, Section 3.2 — SSOT Principle:** The single source of truth principle is inviolable.
4. **Article IX, Section 9.3 — Financial Data Protection:** The rules governing financial data and PCI-DSS compliance are inviolable.
5. **Article X, Section 10.2, Rule 3 — Never violate the Constitution:** AI agents must always refuse unconstitutional instructions. This rule is inviolable.

These clauses may be amended only by unanimous consent of the founding team (CTO + CEO + CPO) and MUST be documented in a Board resolution.

---

## 7. NON-FUNCTIONAL REQUIREMENTS

These requirements govern the document itself and all EOS documents derived from it.

| Requirement | Standard |
|------------|---------|
| **Longevity** | This document MUST remain substantively valid for a minimum of 5 years with only incremental (minor version) updates. Any major version update implies a significant change in company or product direction. |
| **Clarity** | Every principle MUST be interpretable by a competent software engineer who has never worked at FAIRRIDE, a product manager, and a compliance officer. Where a principle could be interpreted in multiple ways, the intended interpretation MUST be made explicit with examples. |
| **Completeness** | This document MUST answer the question: "When two engineers disagree on a fundamental technical approach and cannot resolve it among themselves, what does this document say?" If an important class of disagreement is not addressed, it MUST be added in the next revision. |
| **Enforceability** | Each principle MUST have a verifiable outcome. Principles that cannot be enforced — because they have no observable measure of compliance or violation — are aspirational slogans, not engineering rules. |
| **Traceability** | Every principle MUST trace to a business objective. Engineering for its own sake is not permitted. If a principle cannot be traced to protecting a user, enabling the business, or ensuring long-term operability, it MUST be removed. |
| **Discoverability** | This document MUST be the first document listed in the EOS index and MUST be prominently linked from the README of every FAIRRIDE repository. |
| **Accessibility to AI agents** | This document MUST be structured so that an AI coding agent can extract the essential rules programmatically. The YAML frontmatter, section numbering, and definition tables serve this purpose. |

---

## 8. CONSTRAINTS

This Constitution deliberately does not govern the following. Each of these is addressed by a downstream EOS document.

| Topic | Governing Document |
|-------|-------------------|
| Specific technology versions | Technology stack ADRs (ADR-0001 through ADR-0003) |
| Database schema design | database-design (DOC-0022) |
| Specific API endpoint definitions | API contracts per service |
| Specific fraud detection rules | fraud-strategy (DOC-0040) |
| Language-specific coding rules | go-standard (DOC-0051), flutter-standard (DOC-0052), nextjs-standard (DOC-0053) |
| Deployment pipeline configuration | ci-cd (DOC-0058) |
| Pricing formula variables | pricing-engine (DOC-0042) |
| Specific SLA numerical targets | SLA documents per service |
| On-call rotation schedules | Operations runbooks |
| Vendor selection | Domain-specific ADRs |
| Compensation and team structure | Human Resources (outside EOS) |

---

## 9. RISKS

| Risk ID | Risk Description | Likelihood | Impact | Mitigation |
|---------|----------------|-----------|--------|-----------|
| R-001 | **Document Drift:** This Constitution becomes outdated as the company grows and principles are violated in practice without being formally updated. | Medium | High | Annual review mandatory. Quarterly staleness check automated. Clear amendment process reduces the inertia to update. |
| R-002 | **Interpretation Variance:** Different teams read the same principle and reach different conclusions, defeating the purpose of having a shared standard. | Medium | High | Principle elaboration with concrete examples. Conflict detection process (Article VI, Section 6.2). Design review culture ensures alignment before implementation. |
| R-003 | **Onboarding Gap:** New engineers do not read this Constitution and operate outside its principles unknowingly. | High | Medium | Constitution reading is mandatory in the onboarding checklist. Code review process includes checks against principles. Constitution is linked from every repository README. |
| R-004 | **Over-Specification:** If the Constitution is too prescriptive, it becomes an obstacle to pragmatic engineering decisions. Developers bypass it rather than challenge it. | Low | High | This Constitution governs only foundational principles. Implementation details are delegated to downstream documents. The amendment process is accessible, not bureaucratic. |
| R-005 | **Scale Mismatch:** Principles written for a startup become inappropriate at 500 engineers without formal review. | Low | High | Annual review cycle specifically evaluates whether the governance model fits current scale. Amendment process creates a formal path for scaling the governance model. |
| R-006 | **AI Agent Non-Compliance:** AI agents generate artifacts that violate this Constitution because the Constitution context was not provided at inference time. | Medium | High | Constitution is loaded as context in all AI assistant configurations. AI rulebook (DOC-0066) operationalizes AI compliance rules into machine-readable format. |
| R-007 | **Immutable Clause Erosion:** Business pressure causes gradual erosion of immutable clauses through informal workarounds without formal amendment. | Low | Critical | Immutable clauses are explicitly named and tracked. Any deviation triggers escalation to CTO. Quarterly compliance audit includes immutable clause verification. |

---

## 10. FUTURE EXTENSION

### 10.1 Scaling the Governance Model

This Constitution is written for a founding engineering team. As FAIRRIDE scales, the governance model will need to evolve. Anticipated evolution points:

**At 20+ engineers:** The RFC process becomes more formal. A standing Architecture Review Board (ARB) replaces ad-hoc Principal Architect reviews for Class 2 decisions. An amendment to Article V will be required.

**At 50+ engineers:** Domain teams become autonomous engineering units with their own governance sub-layers. The Constitution will need a new Article defining the relationship between the central EOS and domain-level engineering standards.

**At 200+ engineers:** The Technology Philosophy may need to evolve to support more specialized sub-stacks per domain. The single approved stack (Article IV) may expand to a curated approved set. A Constitutional amendment will be required.

**At multi-region:** Data sovereignty requirements in multiple jurisdictions will require the Security and Privacy Mandate (Article IX) to be expanded. Compliance will vary by region. A Constitutional amendment or a Regional Supplement model will be required.

### 10.2 Open Platform Evolution

When the Open API Platform (Phase 4) is reached, this Constitution requires a new Article governing the relationship between FAIRRIDE's internal engineering standards and external developer obligations. External developers who integrate with the FAIRRIDE Open API do not operate under this Constitution, but the APIs they consume must be designed to FAIRRIDE's standards.

### 10.3 AI Evolution

Article X is written for the current generation of AI coding assistants. As AI capabilities evolve — particularly toward autonomous AI agents capable of deploying code without direct human instruction — this Article will require expansion. Key open questions for future consideration are captured in Section 11.

---

## 11. OPEN QUESTIONS

The following questions have not yet been resolved. They are documented here to ensure they are formally addressed before the relevant downstream documents are authored.

| OQ ID | Question | Impact | Target Resolution |
|-------|----------|--------|-----------------|
| OQ-001 | What is the primary launch city/market? Regulatory, compliance, and payment requirements are market-specific. | High — affects legal compliance, payment provider selection, language support | Before PRODUCT_STRATEGY is authored |
| OQ-002 | What is the target funding stage and runway at MVP launch? This affects architectural conservatism vs. speed trade-offs. | Medium — affects infrastructure investment decisions | Before SYSTEM_ARCHITECTURE is authored |
| OQ-003 | Is there a Board-level commitment to the multi-product vision, or is it aspirational? | High — affects architectural investment in multi-product support | Before SYSTEM_ARCHITECTURE is authored |
| OQ-004 | What is FAIRRIDE's policy on using open-source AI models for internal tooling vs. commercial AI APIs? | Medium — affects AI cost model and data privacy posture | Before AI_RULEBOOK is authored |
| OQ-005 | Does FAIRRIDE intend to build a proprietary mapping stack or exclusively use third-party providers? | High — affects the Geo Engine design fundamentally | Before GEO_ENGINE is authored |
| OQ-006 | What is the driver classification model in the target market (employee vs. independent contractor)? | High — affects earnings model, wallet design, and tax reporting | Before BUSINESS_MODEL is authored |
| OQ-007 | Is there a requirement for offline-capable dispatch in markets with unreliable connectivity? | High — affects the entire mobile architecture approach | Before SYSTEM_ARCHITECTURE is authored |
| OQ-008 | What is the target for time-to-first-production-deployment? | Medium — affects MVP scope discipline | Before ROADMAP is authored |
| OQ-009 | Will FAIRRIDE operate its own payment infrastructure or exclusively use payment provider rails? | High — affects the Payment Engine design and PCI scope | Before PAYMENT_ENGINE is authored |

---

## 12. DECISION REFERENCES (ADR)

This Constitution generates the following ADR requirements. These ADRs MUST be produced in the order listed before their dependent EOS documents are authored.

| ADR ID | Decision Required | Dependent Documents | Priority |
|--------|-----------------|--------------------|----|
| ADR-0001 | Backend Technology Stack: Rationale for Go as primary backend language | go-standard, all service design docs | P1 — Before any service design |
| ADR-0002 | Mobile Technology Stack: Rationale for Flutter as cross-platform mobile framework | flutter-standard, mobile architecture | P1 — Before mobile architecture |
| ADR-0003 | Web Admin Technology Stack: Rationale for Next.js as admin console framework | nextjs-standard | P1 — Before admin portal design |
| ADR-0004 | Documentation-First Process: Formal adoption of documentation-before-code as engineering process | All EOS documents | P1 — Concurrent with Constitution approval |
| ADR-0005 | Microservices Architecture: Rationale for domain-driven service decomposition vs. modular monolith | system-architecture, all service docs | P1 — Before system architecture |
| ADR-0006 | Event-Driven Communication Pattern: When to use events vs. synchronous API calls | system-architecture, dispatch-engine, payment-engine | P2 — Before domain engine design |

---

## 13. REVISION HISTORY

| Version | Date | Author | Status | Change Summary |
|---------|------|--------|--------|---------------|
| 0.1.0 | 2026-06-30 | Office of the CTO | Draft | Initial creation of FAIRRIDE Project Constitution. All 11 Articles authored. Pending review and CTO approval. |

---

*End of Document — DOC-0001 — Project Constitution — v0.1.0*

*This document MUST be approved by the CTO before any subsequent EOS document is authored.*
*Approval converts this document to status: Approved and triggers a rename to project-constitution-v1.0.md*
