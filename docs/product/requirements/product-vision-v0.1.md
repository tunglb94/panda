---
document_id: DOC-0002
title: Product Vision
version: 0.1.0
status: Draft
classification: Internal — All Teams
owner: CPO / CTO
approved_by: ~
depends_on:
  - DOC-0001   # project-constitution
  - DOC-0001A  # ai-development-governance
required_by:
  - DOC-0003   # system-architecture
  - DOC-0004   # implementation-master-plan
related_adrs:
  - ADR-0001   # backend-technology-stack (pending)
  - ADR-0002   # mobile-technology-stack (pending)
  - ADR-0003   # admin-web-technology-stack (pending)
  - ADR-0005   # microservices-architecture (pending)
  - ADR-0008   # lean-documentation-strategy (pending)
  - ADR-0009   # primary-launch-market (pending)
  - ADR-0010   # commission-model (pending)
  - ADR-0011   # surge-cap-policy (pending)
absorbs:
  - "[planned] project-vision"
  - "[planned] business-model"
  - "[planned] product-strategy"
  - "[planned] roadmap"
  - "[planned] competitor-analysis"
  - "[planned] success-metrics"
  - "[planned] mvp-definition"
  - "[planned] user-personas"
  - "[planned] user-journeys"
  - "[planned] feature-catalog"
  - "[planned] out-of-scope"
glossary_refs:
  - Rider
  - Driver
  - Trip
  - Fare
  - Dispatch
  - Surge
  - Wallet
  - GMV
  - Take Rate
  - KYC
  - MVP
  - Phase
created: 2026-06-30
last_updated: 2026-06-30
next_review: 2026-12-30
---

# FAIRRIDE — Product Vision

> This document is the Single Source of Truth for the FAIRRIDE product.
> Every engineering decision, every architectural choice, every feature prioritization,
> and every design constraint must be traceable to a principle or specification
> established in this document.
>
> This document absorbs eleven previously planned EOS documents into one permanent
> reference under the FAIRRIDE Lean Documentation Strategy.

---

## TABLE OF CONTENTS

1. [Purpose](#1-purpose)
2. [Scope](#2-scope)
3. [Dependencies](#3-dependencies)
4. [Related Documents](#4-related-documents)
5. [Definitions](#5-definitions)
6. [Functional Description](#6-functional-description)
   - [6.1 Mission](#61-mission)
   - [6.2 Vision](#62-vision)
   - [6.3 Core Values](#63-core-values)
   - [6.4 Product Philosophy](#64-product-philosophy)
   - [6.5 Business Goals](#65-business-goals)
   - [6.6 Revenue Model](#66-revenue-model)
   - [6.7 Commission Strategy](#67-commission-strategy)
   - [6.8 Driver Fairness Strategy](#68-driver-fairness-strategy)
   - [6.9 Customer Protection Strategy](#69-customer-protection-strategy)
   - [6.10 Competitive Advantages](#610-competitive-advantages)
   - [6.11 Market Position](#611-market-position)
   - [6.12 Target Users](#612-target-users)
   - [6.13 User Personas](#613-user-personas)
   - [6.14 Customer Journey](#614-customer-journey)
   - [6.15 Driver Journey](#615-driver-journey)
   - [6.16 Product Scope](#616-product-scope)
   - [6.17 Out of Scope](#617-out-of-scope)
   - [6.18 Feature Roadmap](#618-feature-roadmap)
   - [6.19 MVP Scope](#619-mvp-scope)
   - [6.20 Version 2 Vision](#620-version-2-vision)
   - [6.21 Version 3 Vision](#621-version-3-vision)
   - [6.22 Success Metrics](#622-success-metrics)
   - [6.23 North Star Metric](#623-north-star-metric)
   - [6.24 Product KPIs](#624-product-kpis)
   - [6.25 Definition of Success](#625-definition-of-success)
7. [Non-Functional Requirements](#7-non-functional-requirements)
8. [Constraints](#8-constraints)
9. [Risks](#9-risks)
10. [Future Extension](#10-future-extension)
11. [Open Questions](#11-open-questions)
12. [Decision References (ADR)](#12-decision-references-adr)
13. [Revision History](#13-revision-history)

---

## 1. PURPOSE

This document establishes the complete product identity, business strategy, user understanding, and feature roadmap of FAIRRIDE. It is the canonical reference that answers every business and product question before implementation begins.

Engineers reading this document MUST understand not just what to build, but why. Every technical trade-off made in DOC-0003 (System Architecture) and DOC-0004 (Implementation Master Plan) MUST be traceable to a product principle or user need expressed here.

This document deliberately consolidates content that would traditionally span eleven separate documents. This consolidation is intentional and is governed by the FAIRRIDE Lean Documentation Strategy (ADR-0008, pending). Consolidation does not mean superficiality — every section is written to the depth required for implementation decisions.

---

## 2. SCOPE

This document governs product decisions for the entire FAIRRIDE Platform, from MVP through Version 3. It covers:

- The full business model of FAIRRIDE
- All user types, their needs, and their journeys
- The complete feature surface across all phases
- The metrics framework for measuring success
- The competitive context and differentiation strategy
- All product decisions that affect architectural design

This document does NOT govern:
- Technical implementation details → DOC-0003
- Sprint plans, implementation timelines → DOC-0004
- API specifications, data models, service design → DOC-0003 and on-demand documents
- Security implementation → on-demand security documents
- Operational procedures → on-demand operational documents

---

## 3. DEPENDENCIES

| Type | Document ID | Title | Reason |
|------|------------|-------|--------|
| Required upstream | DOC-0001 | Project Constitution | Supreme authority. Mission (Section 6.1) must align with DOC-0001 Article I. Core Values (Section 6.3) extend DOC-0001 Article II into product-specific form. |
| Required upstream | DOC-0001A | AI Development Governance | All AI agents contributing to or reading this document are bound by DOC-0001A. |

---

## 4. RELATED DOCUMENTS

| Document ID | Title | Relationship |
|-------------|-------|-------------|
| DOC-0003 | System Architecture | Must be consistent with and derived from the product scope defined in Sections 6.16–6.21 of this document |
| DOC-0004 | Implementation Master Plan | Must be consistent with the feature roadmap in Section 6.18 and MVP scope in Section 6.19 |
| ADR-0009 | Primary Launch Market | Resolves Open Question OQ-001. Required before city-specific configuration in DOC-0004 |
| ADR-0010 | Commission Model | Formalizes the commission structure proposed in Section 6.7 |
| ADR-0011 | Surge Cap Policy | Formalizes the surge cap proposed in Section 6.9 |

---

## 5. DEFINITIONS

The following product-specific terms supplement the canonical glossary established in DOC-0001 Section 5 and DOC-0001A Section 5. All previous definitions remain in force.

| Term | Definition |
|------|-----------|
| **Gross Trip Value (GTV)** | The total fare paid by a Rider for a single Trip, including base fare, surge component, and any booking fee, before any commission or fee deduction. Used interchangeably with trip-level GMV. |
| **Net Trip Revenue** | The GTV minus payment processing fees and minus the driver's earnings payout. FAIRRIDE's financial contribution from a single Trip. |
| **Take Rate** | FAIRRIDE's commission percentage applied to each Trip's GTV. The primary revenue lever. |
| **Driver Tier** | A performance classification assigned to Drivers based on trip volume and quality rating. Determines the commission rate the Driver pays. Defined in Section 6.7. |
| **Surge Multiplier** | The dynamic pricing factor applied to the base fare during high-demand periods. Always displayed to the Rider before booking. |
| **Surge Cap** | The maximum Surge Multiplier FAIRRIDE will apply, regardless of supply-demand ratio. A driver fairness and rider trust mechanism. |
| **Surge Revenue Share** | The portion of surge pricing revenue passed to the Driver. FAIRRIDE policy: 100% of surge is passed to the Driver. |
| **Marketplace Match** | A Trip that has been dispatched to a Driver, accepted, and completed with no post-trip dispute. The quality unit of FAIRRIDE's marketplace. |
| **Fair Match** | A Marketplace Match where both Rider and Driver independently rated the experience at ≥ 4 out of 5 stars. The basis of the North Star Metric. |
| **Driver Utilization Rate** | The percentage of time an online Driver spends actively on a Trip (pickup travel + passenger transport) versus waiting for a request. A key supply efficiency metric. |
| **Acceptance Rate** | The percentage of Dispatch requests accepted by a Driver in a given period. |
| **Trip Completion Rate** | The percentage of accepted Trip requests that result in a completed Trip without cancellation. |
| **Cancellation Rate** | The percentage of Trip requests that result in cancellation by either Rider or Driver before Trip completion. |
| **Booking Fee** | A fixed platform fee charged per Trip, separate from the distance/time-based Fare. Disclosed to Rider before booking. |
| **Driver Activation** | The event at which a Driver completes all onboarding and KYC requirements and places their first active trip offer. |
| **Driver Deactivation** | The suspension or termination of a Driver's right to operate on the FAIRRIDE Platform. Subject to formal documented process. |
| **Supply Health** | The availability of active Drivers in a given zone and time window. Low supply health triggers Surge. |
| **Demand Heat** | The concentration of pending Rider requests in a given zone and time window. High demand heat combined with low supply health produces Surge. |
| **Upfront Pricing** | The FAIRRIDE policy that Riders are shown the complete Trip Fare before confirming a booking. The fare shown is the fare charged, subject only to route deviation by Rider request. |
| **ETA** | Estimated Time of Arrival. The predicted time for the Driver to reach the Rider's pickup location. Displayed before Rider confirms booking. |
| **Geofence** | A virtual geographic boundary that defines operational zones, pricing zones, restricted areas, or surge zones within the FAIRRIDE Platform. |
| **City Launch** | The formal activation of FAIRRIDE operations in a new urban market, including onboarding a minimum viable Driver supply base and opening the app to Riders. |
| **Supply Seeding** | The pre-launch phase of a City Launch in which FAIRRIDE recruits, onboards, and activates a threshold number of Drivers before accepting Rider bookings. |
| **ARPU** | Average Revenue Per User. Computed separately for Riders (revenue per monthly active Rider) and Drivers (earnings per monthly active Driver). |
| **LTV** | Lifetime Value. The projected total revenue attributable to a Rider or Driver over their entire relationship with FAIRRIDE. |
| **CAC** | Customer Acquisition Cost. The total marketing and incentive spend required to activate one new Rider or Driver. |
| **Unit Economics** | The profitability calculation at the level of a single Trip. Positive unit economics means each Trip contributes net positive revenue after all variable costs. |
| **Driver DSAT** | Driver Dissatisfaction Rate. The percentage of Driver sessions in a period that receive a post-session rating below 3 stars or a submitted complaint. Inverse of satisfaction. |
| **Rider RSAT** | Rider Satisfaction Rate. The percentage of completed Trips in which the Rider rated the experience ≥ 4 stars. |

---

## 6. FUNCTIONAL DESCRIPTION

---

### 6.1 Mission

> **Build the fairest ride-hailing ecosystem.**

FAIRRIDE exists to solve a specific, persistent injustice in the ride-hailing industry: the platforms that connect Riders with Drivers extract disproportionate value from both sides.

Drivers on major platforms surrender 25–30% of every fare. They receive opaque earnings reports they cannot independently verify. Their dispatch access decreases as they age on the platform. They face arbitrary deactivations with no right of appeal. They are the essential supply of the marketplace, yet they are treated as interchangeable commodity.

Riders pay prices set by algorithms they cannot understand. They face unlimited surge multipliers with no ceiling. They have no guarantee that the fare shown before booking equals the fare charged after arrival. They pay fees that are not clearly itemized.

FAIRRIDE reverses this. Fairness is not a brand claim. It is an engineering constraint, a business rule, and a product policy simultaneously. Every feature built on this platform MUST be evaluated against the question: **Is this fair to the Driver? Is this fair to the Rider? Is this fair to the City?**

---

### 6.2 Vision

**Short-term Vision (MVP — Year 1):**
FAIRRIDE launches in a single metropolitan market as the most driver-friendly, rider-transparent ride-hailing app in that city. Drivers earn demonstrably more per hour than on competitor platforms. Riders experience fully upfront pricing with no surprises. Within 12 months, FAIRRIDE achieves 10% of daily urban trip volume in the launch city.

**Medium-term Vision (Year 2–3):**
FAIRRIDE expands to five cities in the primary market and begins cross-border expansion into a second country. The platform broadens to include additional transport categories (motorcycle taxi, premium sedan, multi-stop) and introduces a lightweight corporate account tier. FAIRRIDE is recognized by Drivers as the platform of choice for professional drivers who care about earnings transparency.

**Long-term Vision (Year 4–7):**
FAIRRIDE operates as a comprehensive urban mobility and commerce platform across Southeast Asia. The platform supports ride, bike, EV, delivery, food, logistics, and corporate transport. The FAIRRIDE Merchant Platform enables local businesses to participate in the ecosystem. The Open API Platform allows third-party developers to build on FAIRRIDE infrastructure. FAIRRIDE is the standard bearer for ethical technology in the mobility sector.

**10-Year Aspiration:**
FAIRRIDE defines the global standard for ethical marketplace design. The FAIRRIDE model — transparent algorithms, graduated commission, surge caps, 100% surge pass-through — becomes the industry norm because it outcompetes the extractive model at scale.

---

### 6.3 Core Values

The following six product values govern every product decision at FAIRRIDE. They extend and make product-specific the engineering values established in DOC-0001 Article II.

**1. Radical Transparency**
Every number shown to a user — fare, earnings, commission, surge — is the real number. No hidden fees. No retroactive adjustments. No opaque algorithms. If FAIRRIDE cannot explain a calculation to a Driver or Rider in plain language, the calculation MUST be redesigned.

*In practice:* Fare breakdowns show every component (base, distance, time, surge multiplier, booking fee). Driver earnings statements show gross fare, FAIRRIDE commission percentage and amount, payment processing deduction, and net earnings. Every line is itemized.

**2. Driver Prosperity First**
Drivers are not gig workers to be optimized away. They are the supply that makes the marketplace function. Without financially healthy, motivated Drivers, there is no FAIRRIDE. Every product decision that affects Driver economics MUST pass a driver prosperity test: does this increase or decrease Driver earnings per hour on FAIRRIDE?

*In practice:* Commission rates decrease as Driver tenure and quality increase. 100% of surge revenue is passed to Drivers. Dispatch algorithm explicitly avoids penalizing Drivers for short trips or low-demand pickup areas.

**3. Rider Trust**
Riders trust FAIRRIDE because FAIRRIDE never surprises them. The price shown is the price paid. The ETA shown is within ±2 minutes of actual arrival 90% of the time. If something goes wrong, it is fixed within 4 hours.

*In practice:* Upfront pricing is mandatory — no trip begins without Rider confirmation of the exact fare. Surge is always disclosed with the multiplier visible, not hidden in the fare. Refunds are processed automatically where fault is clear.

**4. City Partnership**
FAIRRIDE operates as a partner to cities, not an interloper to be regulated away. FAIRRIDE proactively shares anonymized mobility data with city governments. FAIRRIDE complies with local regulations before being required to. FAIRRIDE does not participate in predatory pricing that harms local transportation ecosystems.

*In practice:* FAIRRIDE conducts regulatory review before entering any market. FAIRRIDE maintains a City Partnership team that proactively engages with transport authorities.

**5. Safety Without Compromise**
The physical safety of Riders and Drivers is a product requirement, not a feature. Every version of the FAIRRIDE app MUST include real-time trip sharing, an emergency contact feature, and an in-trip emergency button. These features are never removed, never deprioritized, and never gated behind a subscription tier.

*In practice:* Safety features are in the MVP. They are not V2 or V3. They are V0.

**6. Proportional Technology**
FAIRRIDE does not build technology for its own sake. The technology serves the mission of fairness. FAIRRIDE uses the simplest technology that can deliver the required experience at the required scale. When a simple rule works as well as a machine learning model, the simple rule is chosen.

*In practice:* MVP fraud detection uses rule-based heuristics. ML models are introduced only when rules demonstrably fail and data volume justifies model training.

---

### 6.4 Product Philosophy

**The marketplace is bilateral — serve both sides equally.**
Most ride-hailing platforms are built rider-first: they optimize the Rider experience and treat Drivers as a variable supply input. FAIRRIDE treats the marketplace as bilateral and designs for both sides simultaneously. Driver experience is a first-class product concern. Driver app quality, earnings visibility, and dispatch fairness receive the same product investment as Rider booking experience.

**Default to restraint in feature scope.**
FAIRRIDE will not build features because competitors have them. Features are built because they solve a documented user problem. The MVP feature set is small by design. Every feature added to scope must answer: What specific user problem does this solve? What would a user lose if we didn't build this?

**The algorithm must be auditable.**
FAIRRIDE will never deploy an algorithm — dispatch, pricing, fraud, rating — that cannot be explained in plain language to the user it affects. Algorithmic decisions that affect a Driver's access to trips or a Rider's experience of pricing MUST have a documented, human-readable rationale. Black-box systems that affect user livelihoods are incompatible with FAIRRIDE's mission.

**Build for the emerging market user.**
FAIRRIDE's primary market is an emerging metropolitan economy. Users have mid-range Android smartphones. Network connectivity is variable. Cash handling is still common. Bank account penetration is incomplete. The app must be light (<30MB initial install), fast on 3G connections, functional offline for key flows, and support non-card payment methods from MVP.

**Earn trust through consistency, not campaigns.**
FAIRRIDE does not launch with a "we are different" marketing campaign. FAIRRIDE earns trust by being demonstrably different every day: lower commission reported by Drivers, accurate ETAs reported by Riders, disputes resolved within the promised SLA. Trust is a product metric, not a marketing metric.

---

### 6.5 Business Goals

#### Short-Term Goals (MVP — Months 1–12)

| Goal | Target | Metric |
|------|--------|--------|
| Achieve positive unit economics | By Month 6 | Net Trip Revenue > variable cost per trip |
| Establish supply base | 1,000 active Drivers by Month 3 | Weekly Active Drivers |
| Achieve demand traction | 5,000 completed trips/day by Month 6 | Daily Completed Trips |
| Prove driver fairness | FAIRRIDE Driver hourly earnings ≥ 20% above primary competitor | Driver Earnings Survey (monthly) |
| Prove rider trust | Rider RSAT ≥ 80% | In-app post-trip rating |
| Zero critical safety incidents | 0 verified safety incidents per 10,000 trips | Safety Incident Rate |

#### Medium-Term Goals (Year 2–3)

| Goal | Target | Metric |
|------|--------|--------|
| Multi-city operation | 5 cities in primary market by Month 24 | Active Cities |
| Market share in launch city | 20% of daily urban trip volume by Month 24 | Third-party market share estimate |
| Revenue milestone | $10M Annual GMV by Month 18 | Monthly GMV |
| Driver retention | 70% of Drivers active in Month 1 still active in Month 12 | Driver Cohort Retention |
| Corporate product launch | First corporate account active by Month 18 | Corporate Accounts |
| Funding milestone | Series A close by Month 18 | — |

#### Long-Term Goals (Year 4–7)

| Goal | Target |
|------|--------|
| Regional presence | Operations in 3 countries |
| Platform diversity | At least 3 product lines generating revenue |
| Driver earnings leadership | FAIRRIDE Drivers earn the highest effective hourly rate of any platform in operating markets |
| Brand recognition | "FAIRRIDE" is synonymous with fair treatment in the ride-hailing industry in operating markets |

---

### 6.6 Revenue Model

FAIRRIDE's revenue model is commission-based at its core, with ancillary revenue streams introduced at later phases.

#### Primary Revenue: Trip Commission

FAIRRIDE deducts a commission from every completed Trip fare. This commission is the primary and MVP-phase revenue stream.

```
Driver Gross Earnings = Trip Fare × (1 − Commission Rate)
FAIRRIDE Commission = Trip Fare × Commission Rate
Net FAIRRIDE Revenue = Trip Fare × Commission Rate − Payment Processing Fee
```

**MVP Commission Rate:** 15% flat for all Drivers at launch
**Payment Processing Cost:** ~2.5% of Trip Fare (charged by payment provider)
**FAIRRIDE Net Take after processing:** ~12.5% per trip

#### Secondary Revenue: Booking Fee (MVP Phase)

A fixed booking fee is charged to the Rider per Trip, separate from the distance/time fare. This fee is:
- Disclosed to the Rider before booking confirmation
- Not shared with the Driver
- Set at [TBD by pricing team — estimated $0.30–$0.50 USD equivalent per trip]

#### Tertiary Revenue: Premium Driver Subscription (Phase 1)

An optional monthly subscription for Drivers that provides:
- Guaranteed minimum trip assignment hours per week
- Priority dispatch queue during high-competition periods
- Enhanced driver support response SLA
- Monthly earnings report with tax-ready breakdown

**Target price:** [TBD — estimated equivalent of 2–3 hours earnings per month]

#### Phase 2+ Revenue Streams

| Stream | Phase | Description |
|--------|-------|-------------|
| Corporate Billing | Phase 2 | Monthly invoicing for corporate accounts with volume discounts |
| Delivery Commission | Phase 2 | Commission on package delivery trips |
| In-App Advertising | Phase 2 | Non-intrusive destination and area-relevant promotions during trip wait |
| Open API Licensing | Phase 3 | Developer access fees for the FAIRRIDE Open API platform |
| Merchant Platform Fees | Phase 3 | Listing and transaction fees for Merchant partners |
| Data Insights (anonymized) | Phase 3 | Aggregated urban mobility data sold to city planners and research institutions |

#### Unit Economics Model (MVP Target City)

| Metric | MVP Month 6 Target | MVP Month 12 Target |
|--------|-------------------|---------------------|
| Daily Completed Trips | 5,000 | 10,000 |
| Average Trip Fare | $4.50 USD equivalent | $5.00 USD equivalent |
| Daily GTV | $22,500 | $50,000 |
| Commission Rate | 15% | 15% |
| Daily Gross Commission | $3,375 | $7,500 |
| Payment Processing (2.5%) | $563 | $1,250 |
| Daily Net Revenue | $2,812 | $6,250 |
| Annual Revenue Run Rate | $1.03M | $2.28M |

*All figures are estimates and assume single city operation. Subject to revision upon OQ-001 resolution (market selection).*

---

### 6.7 Commission Strategy

The commission structure is FAIRRIDE's most powerful tool for driver acquisition and retention. It is both a business strategy and a fairness mechanism.

#### Graduated Commission Tiers

FAIRRIDE uses a graduated commission model where Driver performance — measured by trip volume and quality rating — is rewarded with lower commission rates. This creates a positive feedback loop: better Drivers earn more, which attracts better Drivers.

| Tier | Name | Commission Rate | Qualification Criteria |
|------|------|----------------|----------------------|
| Tier 0 | **Standard** | 15% | All Drivers at activation (no qualification required) |
| Tier 1 | **Silver** | 14% | 500+ trips in trailing 30 days AND rating ≥ 4.6 |
| Tier 2 | **Gold** | 12% | 1,000+ trips in trailing 30 days AND rating ≥ 4.7 |
| Tier 3 | **Platinum** | 10% | 2,000+ trips in trailing 30 days AND rating ≥ 4.8 |

**Tier mechanics:**
- Tiers are calculated and updated every 30 days on the Driver's platform anniversary date
- Tier downgrade is gradual: a Driver who falls below threshold enters a 30-day grace period before downgrading
- Tier status is visible to the Driver in real-time on their dashboard
- Tier calculation methodology is published in the app's "How tiers work" section

**Competitive context:**
At Tier 3 (Platinum), FAIRRIDE Drivers keep 90% of every fare. This compares to:
- Grab standard: ~75% retained by driver
- Uber standard: ~72–75% retained by driver
- inDrive: ~85–90% but with highly variable fares from negotiation

FAIRRIDE Platinum drivers earn the best commission economics of any major platform.

#### Commission Transparency Commitment

FAIRRIDE MUST display on every Trip completion notification:
- Gross fare: $X.XX
- FAIRRIDE commission (X%): −$X.XX
- Booking fee: −$X.XX (if applicable)
- Payment processing: −$X.XX
- **Your earnings: $X.XX**

This is non-negotiable. Any product design that hides the commission deduction violates FAIRRIDE's radical transparency value.

#### Surge Revenue: 100% to Drivers

When Surge pricing is active, 100% of the surge component of the fare is earned by the Driver. FAIRRIDE's commission applies only to the base fare.

```
Example:
  Base Fare:         $4.00
  Surge (1.5x):      +$2.00 surge component
  Total Fare:        $6.00

  FAIRRIDE commission (15% of BASE fare only): $0.60
  Driver earnings: $4.00 − $0.60 + $2.00 surge = $5.40
  Driver effective take: 90% of total fare in this example
```

This policy directly incentivizes Drivers to go online during surge periods, improving supply health exactly when demand is highest. It also eliminates the "platform double-dips on surge" complaint that damages competitor trust with drivers.

---

### 6.8 Driver Fairness Strategy

FAIRRIDE's driver fairness strategy addresses the seven most common driver grievances against ride-hailing platforms.

#### Grievance 1: Opaque earnings

**FAIRRIDE Response:** Full earnings itemization per trip. Monthly earnings report in tax-ready format. Earnings history exportable as CSV. No retroactive deductions without Driver notification and 7-day appeal window.

#### Grievance 2: Biased dispatch

**FAIRRIDE Response:** The Dispatch Algorithm (governed by DOC-0003 and Dispatch Engine spec) MUST be designed with the following fairness constraints:
- No systematic disadvantage for Drivers based on vehicle age, driver age, or historical trip volume that is not directly correlated with quality rating
- Short-trip Drivers (multiple short fares in same area) MUST receive equivalent dispatch weighting to long-trip Drivers after adjusting for earnings per hour
- Pickup area bias detection: the algorithm MUST be monitored for systematic dispatch inequality across city zones. Any zone where identified Drivers receive disproportionately fewer requests triggers an algorithmic audit.
- Dispatch fairness report: published quarterly as an internal audit. Available to Drivers on request.

#### Grievance 3: Arbitrary deactivation

**FAIRRIDE Response:** The Driver Deactivation Policy governs:
- **Warning system:** All policy violations receive a written warning with specific violation cited before any deactivation
- **Escalation tiers:** Minor violations (1 warning) → Temporary suspension (3 days) → Deactivation
- **Automatic deactivation triggers** (no warning, immediate): Verified fraud, verified sexual harassment complaint, verified physical assault, driving under influence
- **Appeal process:** All deactivations (including automatic) are appealable within 14 days. Appeals reviewed by a human reviewer (not algorithm) within 5 business days
- **Documentation requirement:** Every deactivation must have a documented reason code, evidence reference, and reviewer ID

#### Grievance 4: Unfair ratings

**FAIRRIDE Response:**
- Ratings below 3 stars REQUIRE a written explanation from the Rider (mandatory field, cannot be skipped)
- Ratings given during demonstrably abnormal events (detected vehicle incident, extreme surge period, platform disruptions) are flagged for review and may be excluded from the Driver's rating average
- Drivers may flag up to 3 ratings per month as "disputed." Disputed ratings trigger a human review within 48 hours
- Rating average is calculated using a trailing 200-trip window (not all-time), giving Drivers the ability to recover from a bad period

#### Grievance 5: Slow/complex payouts

**FAIRRIDE Response:**
- Driver Wallet is credited in real-time on Trip completion (not end-of-day batch)
- Withdrawal to bank account: available once per day, processed next business day
- In-app instant withdrawal (for fee): available 24/7 with immediate transfer for a small fee [TBD by Payments team]
- No minimum withdrawal amount for daily transfers
- No holding period for new Drivers beyond the first Trip verification period (48 hours)

#### Grievance 6: No Driver voice

**FAIRRIDE Response:**
- Driver Advisory Panel: A group of 20 high-tenure Drivers selected quarterly, consulted before any policy change that affects earnings or dispatch
- In-app feedback mechanism: Drivers can submit platform feedback on any policy or feature, with a committed response SLA of 7 business days
- Policy Change Notice: Any change to commission rates, payout structure, or dispatch rules requires 30 days advance notice to active Drivers

#### Grievance 7: Driver safety

**FAIRRIDE Response:**
- In-app SOS button linked to local emergency services
- Trip recording feature: optional, locally stored, privacy-compliant
- Rider identity verification before Trip begins (name + photo visible to Driver before acceptance)
- Incident report submission: available within 24 hours of any Trip
- Driver insurance partnership: FAIRRIDE partners with a local insurer to offer optional in-trip insurance for Drivers [Phase 1]

---

### 6.9 Customer Protection Strategy

Rider protection builds the trust that drives demand. FAIRRIDE's commitments to Riders:

#### Price Protection

**Upfront Pricing Guarantee:** The fare displayed at booking confirmation is the fare charged at Trip completion, subject only to:
- Rider requests a route deviation that adds distance
- Road closures or exceptional traffic events requiring significant re-routing (disclosed in-app immediately)
- Rider adds a stop that was not in the original booking

**Surge Cap:** FAIRRIDE applies a maximum Surge Multiplier of **2.0x** regardless of demand-supply ratio. No Trip fare will exceed 2× the base fare for the same route and vehicle type. This cap is permanent platform policy and requires a CTO-level decision to change.

*Rationale:* Surge caps reduce FAIRRIDE's potential revenue during peak periods. This is an intentional trade-off: rider trust and predictability are worth more long-term than uncapped surge revenue.

**Surge Notification:** When Surge is active, the Rider sees:
- The surge multiplier displayed prominently (e.g., "1.4× surge active in your area")
- The surge-adjusted total fare
- An estimated time until surge ends (where computable)
- An option to schedule the trip for after the surge period

#### Service Quality Protection

**ETA Accuracy Commitment:** FAIRRIDE targets P90 ETA accuracy of ±3 minutes. When a Driver's arrival is more than 5 minutes later than the displayed ETA, the Rider receives an automatic partial credit [TBD amount].

**Driver Verification:** Every Driver on the platform has completed:
- Government ID verification
- Driving license validity check
- Vehicle registration verification
- Criminal background check (scope per market regulatory requirement)
- Profile photo verification

**Trip Sharing:** Riders can share real-time trip status with up to 5 trusted contacts from within the app.

**In-Trip Emergency:** A single-press SOS button is accessible during every active Trip, linking to local emergency services and simultaneously notifying FAIRRIDE's safety team.

#### Dispute Resolution

**Automatic Resolution:** The following disputes are resolved automatically within 2 hours:
- Driver cancelled after acceptance: Rider receives full refund + $0.50 platform credit
- Driver marked trip as complete but trip did not happen (geofence verification fails): Full refund
- Rider charged for route longer than actual route by more than 15%: Automatic refund of excess

**Human Resolution:** Complex disputes are reviewed by a human agent within 4 hours (guaranteed SLA) with a decision communicated within 24 hours.

**Refund Processing:** Refunds are processed to the original payment method within 3 business days or to the Rider Wallet within 1 hour.

---

### 6.10 Competitive Advantages

#### Primary Differentiators

| Advantage | FAIRRIDE | Grab | Uber | inDrive |
|-----------|----------|------|------|---------|
| Standard commission rate | 15% | 25–28% | 25–30% | 10–15% |
| Surge cap | 2.0x | No cap | No cap | N/A (negotiation) |
| Surge revenue to driver | 100% | ~70–75% | ~70–75% | ~85–90% |
| Commission transparency | Full itemization | Summary only | Summary only | Variable |
| Driver tier system | Yes (10–15%) | No | No | No |
| Driver deactivation appeal | Formal process, 5-day SLA | Limited | Limited | Limited |
| Upfront pricing guarantee | Guaranteed | Best effort | Best effort | Negotiated |
| Free daily bank withdrawal | Yes | Limited | Limited | Limited |
| Fare dispute auto-resolution | Yes | Manual | Manual | No |

#### Secondary Differentiators

**Technology fairness:** FAIRRIDE publishes high-level documentation of how its dispatch algorithm works. Competitors treat their algorithms as trade secrets. FAIRRIDE's transparency on algorithm mechanics is a trust signal to both Drivers and Drivers.

**Community advocacy:** FAIRRIDE operates a Driver Community Programme providing financial literacy resources, vehicle maintenance tips, and legal guidance specific to gig-economy workers. This is funded by a portion of platform revenue and builds deep driver loyalty.

**Speed of support:** FAIRRIDE commits to 4-hour maximum response time for all in-app support requests. Competitors typically respond within 24–72 hours. Speed of resolution is the single highest-impact trust factor after price transparency.

**Lighter app:** FAIRRIDE's rider app targets <30MB install size and functional use on 3G connections with <1MB/hour data usage during active trips. Competitor apps are typically 80–200MB and data-heavy. In emerging markets, this is a genuine differentiator.

---

### 6.11 Market Position

#### Positioning Statement

FAIRRIDE is the **professional driver's preferred platform** and the **conscious rider's transparent choice**.

FAIRRIDE does not compete to be the cheapest ride-hailing service. It competes to be the most trusted. Trust is earned by transparency, fairness, and reliability — not by being the lowest-cost option at launch.

#### Target Market Characteristics

FAIRRIDE targets urban markets that meet the following criteria:

| Criterion | Requirement | Rationale |
|-----------|------------|-----------|
| Urban population | ≥ 1 million in metro area | Sufficient trip density for marketplace liquidity |
| Smartphone penetration | ≥ 60% of urban adults | App-based platform requires smartphone base |
| Existing ride-hailing competition | At least one established player | Proves market demand; FAIRRIDE competes on fairness, not market creation |
| Driver dissatisfaction | Evidence of Driver complaints against incumbent(s) | FAIRRIDE's supply acquisition advantage requires driver pain points to exploit |
| Payment infrastructure | Mobile money OR card penetration ≥ 30% | Payment processing viability |
| Regulatory environment | Ride-hailing is legal or in the process of legalization | Regulatory risk must be manageable |
| Language | English OR market covered by localization plan | MVP localization capacity constraint |

#### Competitive Entry Strategy

FAIRRIDE enters each new market as a **driver-first launch**:
1. Recruit and onboard minimum 500 active Drivers before accepting any Rider bookings
2. Launch aggressive Driver referral programme during supply seeding phase
3. Open to Riders only when Driver density ensures sub-5 minute ETAs in target zones
4. First 90 days: Zero commission for the first 100 Drivers in each city (early adopter programme)
5. Rider launch with referral programme and first-ride incentives

This strategy ensures Riders experience a high-quality first impression (reliable ETAs) because supply is established before demand is opened.

---

### 6.12 Target Users

#### User Type 1: Riders

**Primary Rider Segment — Urban Professional**
- Demographics: 24–42 years old, employed full-time, lives and works in urban center
- Device: Mid-range to high-end Android (primary); iPhone (secondary)
- Usage frequency: 3–7 trips per week (commute + social)
- Key motivation: Reliability and predictability. They want to know the car is coming, the price is what was shown, and the driver is safe.
- Pain points with competitors: Surge surprises, ETA inaccuracy, support response time
- FAIRRIDE value for this segment: Upfront pricing guarantee, ETA accuracy commitment, 4-hour support SLA

**Secondary Rider Segment — Budget-Conscious Traveler**
- Demographics: 18–28, student or early career, price-sensitive
- Device: Low-to-mid-range Android
- Usage frequency: 1–3 trips per week (non-commute occasions)
- Key motivation: Affordable price, app reliability on slow connection
- Pain points: Surge pricing, card requirements (limited banking access), heavy apps
- FAIRRIDE value: Wallet-based payments (no card required), surge cap at 2x, lightweight app

**Tertiary Rider Segment — Corporate Employee (Phase 2)**
- Demographics: 30–55, corporate professional, company-provided transport benefit
- Usage frequency: Daily (work commute)
- Key motivation: Reliability, receipt generation, expense reporting integration
- FAIRRIDE value: Corporate account billing, monthly invoice, receipt in expense-ready format

#### User Type 2: Drivers

**Primary Driver Segment — Full-Time Independent Driver**
- Demographics: 28–55, driving is primary income
- Vehicle: Own car, financed or owned
- Hours: 8–12 hours per day, 5–6 days per week
- Key motivation: Maximize earnings per hour, predictable income, fair treatment
- Pain points: High commission on other platforms, slow payouts, opaque dispatch
- FAIRRIDE value: Lowest commission (graduating to 10%), daily withdrawal, transparent dispatch, formal appeal process

**Secondary Driver Segment — Part-Time Driver**
- Demographics: 25–45, driving supplements another income
- Hours: 3–5 hours per day, flexible schedule
- Key motivation: Flexible income with minimal platform friction
- Pain points: Minimum hours requirements, complex onboarding, earnings delays
- FAIRRIDE value: No minimum hours, streamlined onboarding, real-time wallet credit

**Tertiary Driver Segment — Fleet Operator (Phase 1)**
- Demographics: Business owner managing 5–50 vehicles and drivers
- Key motivation: Fleet management tools, consolidated reporting, volume commission deals
- Pain points: No fleet management on current platforms, per-driver account complexity
- FAIRRIDE value: Fleet dashboard, bulk driver management, volume commission negotiation

#### User Type 3: Administrators

**City Operations Manager**
- Internal FAIRRIDE employee managing driver supply and operational performance in a city
- Key needs: Real-time supply visibility, driver approval workflow, dispute management
- FAIRRIDE tools: Admin portal with city health dashboard, driver management, dispute resolution

**Support Agent**
- Internal FAIRRIDE employee handling Rider and Driver support requests
- Key needs: Trip history access, refund processing, driver account management
- FAIRRIDE tools: Support portal with full trip detail, refund tools, driver communication

---

### 6.13 User Personas

#### Persona R1: "Alex" — The Urban Professional Rider

| Attribute | Detail |
|-----------|--------|
| Age | 32 |
| Occupation | Marketing manager at a technology company |
| Location | Urban center of [LAUNCH_CITY] |
| Device | Samsung mid-range Android, good connectivity |
| Monthly trips | ~20 (daily commute + weekend social) |
| Monthly spend | ~$70–$90 USD equivalent |
| Motivation | Reliability. Alex has been late to meetings because of competitors' ETAs being wrong. |
| Frustration | "I accepted a 1.8x surge because I needed to get there. Then it charged me 2.3x." |
| Dream | A ride app where the price shown is exactly the price paid, always. |
| FAIRRIDE hook | Upfront pricing guarantee + ETA accuracy commitment |
| Conversion trigger | A friend (referral) who says "it's always exactly what they show" |
| Retention driver | Never being surprised by a fare after 3 months on platform |

#### Persona R2: "Sam" — The Budget-Conscious Rider

| Attribute | Detail |
|-----------|--------|
| Age | 21 |
| Occupation | University student, part-time retail |
| Location | University district, needs rides to campus and social venues |
| Device | Budget Android, variable 3G/4G connectivity |
| Monthly trips | ~8 (non-commute occasions) |
| Monthly spend | ~$20–$30 USD equivalent |
| Motivation | Affordability and wallet convenience (limited banking access) |
| Frustration | "I only use it when I have to because the app is huge and eats data." |
| FAIRRIDE hook | Lightweight app (<30MB), in-app wallet (cash top-up at partner locations in Phase 1) |
| Retention driver | Reliable 2x surge cap — Sam knows FAIRRIDE won't suddenly be 4x during rain |

#### Persona D1: "Ahmad" — The Full-Time Driver

| Attribute | Detail |
|-----------|--------|
| Age | 38 |
| Occupation | Full-time ride-hailing driver (7 years experience) |
| Current platform | Primarily Grab, dissatisfied with commission |
| Hours per day | 10 hours, 6 days per week |
| Monthly earnings target | $800–$1,200 USD equivalent |
| Vehicle | 2020 sedan, financed |
| Motivation | "I want to keep more of what I earn. I'm doing all the work." |
| Frustration | 28% commission on Grab. Doesn't understand why his dispatch seems worse after holidays. |
| FAIRRIDE hook | 15% commission (save ~$100/month immediately). Transparent dispatch explanation. |
| Conversion trigger | Driver community meeting where FAIRRIDE explains the tier system on paper |
| Tier progression path | Standard (15%) → Silver in Month 2 (14%) → Gold by Month 5 (12%) at current volume |
| Dream | Platinum tier at 10% commission within 8 months |

#### Persona D2: "Priya" — The Part-Time Driver

| Attribute | Detail |
|-----------|--------|
| Age | 29 |
| Occupation | Office administrator (9–5), drives evenings and weekends |
| Hours per day | 3–4 hours (evenings and weekends) |
| Monthly earnings target | $200–$350 USD equivalent supplemental income |
| Motivation | Flexible extra income, no commitment pressure |
| Frustration | "Other apps want you to be online 6 hours a day. I can't do that." |
| FAIRRIDE hook | No minimum hours. Drive when you want. Earnings in wallet immediately. |
| Retention driver | Simple, no-pressure app that doesn't penalize low activity |

#### Persona A1: "Sarah" — The City Operations Manager

| Attribute | Detail |
|-----------|--------|
| Age | 34 |
| Occupation | FAIRRIDE City Manager (internal) |
| Key responsibility | Driver supply health, city operational performance, dispute escalations |
| Tools needed | Real-time dashboard, driver approval, bulk communication, manual dispatch override |
| FAIRRIDE goal | Keep driver utilization > 60% and daily trips growing by 15% MoM |

---

### 6.14 Customer Journey

The complete Rider journey from first awareness to loyal user.

#### Stage 1: Discovery

| Touchpoint | Description | FAIRRIDE Responsibility |
|-----------|-------------|----------------------|
| Referral | Friend shares FAIRRIDE referral link with $X first-ride credit | Referral system must be seamless; credit must apply automatically |
| Social media | FAIRRIDE driver earnings testimonial content | Marketing — outside EOS scope |
| App store search | User searches "ride app" or "FAIRRIDE" | App Store Optimization (ASO); rating must be maintained ≥ 4.5 |

#### Stage 2: Onboarding

| Step | User Action | Platform Requirement |
|------|-------------|---------------------|
| Download | Install app | App < 30MB; load in < 3s on 3G |
| Registration | Enter phone number | OTP delivered in < 10 seconds P90 |
| OTP verification | Enter 6-digit code | Code valid for 5 minutes; 3 attempts before resend |
| Profile setup | Name, optional email, profile photo | Photo optional at registration; required after 5 trips |
| Payment setup | Add card OR skip to use Wallet | Card optional; Wallet accessible without card |
| First booking | Enter destination | Map loads in < 2s; autocomplete results in < 500ms |

#### Stage 3: First Trip

| Step | User Experience | Quality Standard |
|------|-----------------|-----------------|
| Fare display | See upfront fare + surge multiplier if applicable | Fare calculation: < 1s P99 |
| Booking confirmation | Confirm trip | Driver matched in < 30s P90 |
| Driver tracking | See Driver on map, ETA countdown | Location update: ≤ 3s refresh |
| Driver arrival | Driver arrives, Rider notified | Push notification: < 5s from geofence trigger |
| In-trip navigation | Ride to destination | Driver sees nav; Rider sees estimated arrival |
| Trip completion | Auto-charged, receipt shown | Payment: < 5s P99; receipt: immediate |
| Rating prompt | Rate Driver (optional) | One-tap 5-star prompt; additional comment optional |

#### Stage 4: Retention and Loyalty

| Mechanism | Description |
|-----------|-------------|
| Fare history | Complete fare breakdown visible in trip history |
| Wallet credits | Referral credits, resolution credits accumulate in Wallet |
| Streak rewards | [Phase 1] Rewards for consecutive weeks of trips |
| Corporate account | [Phase 2] Expense-ready receipts, department billing |
| Notifications | Weekly personalized usage summary (opt-in) |

#### Stage 5: Issue Resolution (When Things Go Wrong)

| Issue | Resolution Pathway | SLA |
|-------|-------------------|-----|
| Driver didn't arrive | In-app contact driver → support escalation | Response: 15 min |
| Fare different from estimate | Automatic audit + refund if applicable | Auto: 2 hours; Manual: 24 hours |
| Driver behaviour complaint | In-app report → Safety team review | First response: 4 hours |
| Payment charged twice | Automatic detection + refund | Auto-refund: 2 hours |
| App crashed mid-trip | Support contact with trip ID | Response: 30 min; credit if needed |

---

### 6.15 Driver Journey

The complete Driver journey from recruitment through becoming a platform-loyal professional driver.

#### Stage 1: Recruitment

| Channel | Description |
|---------|-------------|
| Driver community events | City-level events where FAIRRIDE presents the commission model |
| Digital ads | Targeted to ride-hailing drivers; lead with "keep 85–90% of fares" |
| Driver referral | Existing Drivers earn bonus for each referred Driver who completes 50 trips |
| Partner garages | Partnerships with vehicle service centers to reach active drivers |

#### Stage 2: Application and KYC

| Step | Required Documents | FAIRRIDE Commitment |
|------|-------------------|---------------------|
| Identity verification | Government-issued ID (front + back photo) | Result in < 24 hours |
| License verification | Driving license (front + back) | Automated check where possible |
| Vehicle registration | Vehicle registration document | Result in < 24 hours |
| Vehicle inspection | In-person or photo-based depending on market | Appointment within 3 business days |
| Criminal background check | Consent form; FAIRRIDE runs check via partner | Result in < 5 business days |
| Profile photo | Selfie during onboarding | Verified against ID photo |

**Onboarding SLA:** Total time from application submission to activation approval: ≤ 7 business days

#### Stage 3: Activation and First Trip

| Step | Experience | Platform Requirement |
|------|------------|---------------------|
| App download | Driver app install | Driver app < 30MB |
| Training | 10-minute in-app training module | Covers: how dispatch works, rating system, earnings, emergency procedures |
| Policy acknowledgment | Digital sign of Driver Agreement | Legally valid; market-specific |
| Go online | First online session | Notification of first trip request within [target: 15 min in urban zone] |
| First trip | Earn first fare | Earnings credited to Wallet immediately on completion |
| Earnings view | See breakdown in real-time | Commission itemization visible immediately |

#### Stage 4: Active Driver Experience

| Daily Workflow | FAIRRIDE Design |
|----------------|-----------------|
| Start of day | Open app, go online, GPS permission granted |
| Trip request | Push notification with pickup location, distance, estimated fare |
| Acceptance | One tap to accept; destination revealed after acceptance |
| Navigation | Integrated map navigation (Google Maps or local equivalent) |
| Trip completion | One tap to complete; fare appears in Wallet instantly |
| End of day | Go offline; see daily summary (trips, earnings, rating) |
| Withdrawal | Tap "Withdraw" to transfer daily earnings to bank |

#### Stage 5: Tier Progression

| Milestone | Trigger | Reward |
|-----------|---------|--------|
| 100 trips completed | Automatic | "Established Driver" badge; in-app recognition |
| Silver Tier achieved | 500 trips + 4.6 rating in trailing 30 days | Commission drops to 14%; notification with savings estimate |
| Gold Tier achieved | 1,000 trips + 4.7 rating | Commission drops to 12%; Driver community highlight |
| Platinum Tier achieved | 2,000 trips + 4.8 rating | Commission drops to 10%; priority support; Driver advisory panel invitation |

#### Stage 6: Long-Term Retention

| Mechanism | Description |
|-----------|-------------|
| Earnings transparency | Monthly PDF earnings report in tax-ready format |
| Driver Advisory Panel | Quarterly consultation on platform policy |
| Driver Community App | Optional community forum for FAIRRIDE Drivers |
| Vehicle service partner discounts | [Phase 1] Partner discounts for FAIRRIDE Drivers |
| Insurance partnership | [Phase 1] Optional in-trip insurance at group rates |

---

### 6.16 Product Scope

The FAIRRIDE Platform encompasses the following product areas across all phases:

| Product Area | MVP | Phase 1 | Phase 2 | Phase 3 |
|-------------|-----|---------|---------|---------|
| Standard Ride (Economy) | ✓ | ✓ | ✓ | ✓ |
| Premium Ride | — | ✓ | ✓ | ✓ |
| XL Ride | — | ✓ | ✓ | ✓ |
| Motorcycle Taxi (Bike) | — | ✓ (markets) | ✓ | ✓ |
| Electric Vehicle (EV) | — | — | ✓ | ✓ |
| Licensed Taxi Integration | — | ✓ (markets) | ✓ | ✓ |
| Ride Pooling | — | — | ✓ | ✓ |
| Scheduled Rides | — | ✓ | ✓ | ✓ |
| Package Delivery | — | — | ✓ | ✓ |
| Food Delivery | — | — | — | ✓ |
| B2B Logistics | — | — | — | ✓ |
| Corporate Accounts | — | — | ✓ | ✓ |
| Driver Subscription | — | ✓ | ✓ | ✓ |
| Fleet Management | — | ✓ | ✓ | ✓ |
| Merchant Platform | — | — | — | ✓ |
| Open API | — | — | — | ✓ |
| In-App Wallet | ✓ | ✓ | ✓ | ✓ |
| Card Payments | ✓ | ✓ | ✓ | ✓ |
| Cash Collection (Driver) | — | ✓ (markets) | ✓ | ✓ |
| Corporate Billing | — | — | ✓ | ✓ |
| In-App Advertising | — | — | ✓ | ✓ |
| Multi-Currency | — | — | ✓ | ✓ |
| Multi-Language | ✓ (2 languages) | ✓ (5) | ✓ (10) | ✓ (15+) |
| Multi-City | — | ✓ (5 cities) | ✓ (15 cities) | ✓ (50+ cities) |
| Multi-Country | — | — | ✓ (2 countries) | ✓ (5+ countries) |

---

### 6.17 Out of Scope

The following are explicitly and permanently outside the FAIRRIDE product scope. Any proposal to include these requires CTO + CPO written approval and an ADR.

#### Permanently Out of Scope

| Item | Reason |
|------|--------|
| Self-driving / autonomous vehicles | Not a near-term market reality in target markets; requires separate product strategy |
| Vehicle purchase or leasing financing | Outside ride-hailing; different regulatory environment |
| Driver payroll or employment services | FAIRRIDE is not an employer; legal risk in gig-economy classification |
| Personal banking services | Requires banking license; materially different regulatory burden |
| Ride-sharing between strangers in same private vehicle (carpooling) | Regulatory risk in most target markets; safety complexity |
| Social networking features | Outside core mission; complexity with no business model fit |
| Vehicle tracking for fleet owners without driver consent | Privacy violation; against FAIRRIDE values |

#### Out of Scope for MVP

| Item | Planned Phase | Reason for MVP Exclusion |
|------|-------------|--------------------------|
| Multiple ride categories (Premium, XL) | Phase 1 | MVP validates core ride mechanics; categories add supply complexity |
| Ride pooling | Phase 2 | Requires significantly more sophisticated dispatch; validates post-MVP |
| Scheduled rides | Phase 1 | Requires scheduling infrastructure not critical for MVP market validation |
| Corporate accounts | Phase 2 | Requires separate billing infrastructure; business development pipeline needed |
| Cash collection by Driver | Phase 1 | Requires float management, reconciliation; payment-first MVP is cleaner |
| Motorcycle taxi | Phase 1 (market-dependent) | Vehicle category requires separate supply onboarding |
| ML-based fraud detection | Phase 1 | Insufficient data volume at MVP; rule-based sufficient for MVP scale |
| Multi-city operation | Phase 1 | MVP validates single-city unit economics before scaling |
| Driver vehicle insurance integration | Phase 1 | Partnership negotiation required; not on critical path for launch |

---

### 6.18 Feature Roadmap

#### MVP (Target: First City Launch)

**Core Rider App:**
- Phone number registration + OTP
- Home/destination search with autocomplete
- Upfront fare display with surge indicator
- Driver dispatch + real-time tracking
- In-trip navigation progress bar
- In-app payment (card + wallet)
- Post-trip rating (optional)
- Trip history + receipt
- In-trip SOS button
- Trip sharing with contacts
- In-app support chat

**Core Driver App:**
- Registration + KYC document upload
- Online/offline toggle
- Trip request notification + map view
- Navigation integration
- Trip completion
- Real-time wallet earnings display
- Daily earnings summary
- Bank withdrawal request
- Rating visibility
- In-app support
- In-trip incident reporting

**Admin Portal (MVP):**
- Driver application queue + approval workflow
- Driver account management (deactivate, warn, reinstate)
- Trip search and detail view
- Dispute creation + resolution workflow
- Manual refund tool
- City supply health dashboard
- Basic analytics (daily trips, GMV, active drivers)
- Pricing configuration (base fare, per-km, per-min rates per city)
- Surge zone configuration

**Platform Capabilities (MVP):**
- Dispatch engine (proximity + quality scoring)
- Basic surge pricing (demand/supply ratio)
- Upfront fare calculation
- In-app wallet (Rider + Driver)
- Card payment processing
- OTP service
- Push notifications
- SMS notifications (fallback)
- Fraud baseline rule engine
- Driver KYC pipeline
- Audit logging

#### Phase 1 Feature Additions (Target: Month 13–24)

**Product:**
- Premium and XL ride categories
- Scheduled rides
- Driver subscription tier (Premium Driver)
- Fleet management portal
- Motorcycle taxi (market-dependent)
- Cash collection mode (market-dependent)
- Wallet top-up at partner locations
- Driver community in-app hub
- Driver referral programme (improved)
- Rider streak rewards

**Platform:**
- Enhanced fraud detection (velocity rules, device fingerprinting)
- Multi-city configuration
- Driver tier automation (automatic commission adjustment)
- A/B testing framework
- Enhanced analytics dashboard
- Driver support SLA automation
- Data warehouse v1

#### Phase 2 Feature Additions (Target: Year 3)

**Product:**
- Corporate accounts + billing
- Ride pooling
- EV category
- Package delivery
- In-app advertising
- Multi-country expansion tooling
- Expanded wallet: multi-currency
- Corporate spend management portal

**Platform:**
- ML-based fraud detection v1
- ML-based ETA prediction
- Dynamic surge zones (H3 grid based)
- Multi-country payment rails
- Data warehouse v2 with BI tooling
- Experimentation platform

#### Phase 3 Feature Additions (Target: Year 4+)

- Food delivery
- B2B logistics
- Merchant Platform
- Open API + developer portal
- Advanced ML dispatch optimization
- Regional supply forecasting
- Cross-product wallet ecosystem

---

### 6.19 MVP Scope

The MVP is scoped to prove one thing: **FAIRRIDE can operate a fair, reliable, profitable ride-hailing marketplace in a single city with real users.**

#### MVP Success Criteria (must achieve by Day 180)

| Criteria | Target |
|----------|--------|
| Active Drivers | ≥ 1,000 weekly active Drivers |
| Daily Completed Trips | ≥ 5,000 |
| Rider RSAT | ≥ 80% |
| Driver DSAT | ≤ 15% |
| Average ETA | ≤ 6 minutes |
| Surge frequency | ≤ 20% of trips affected by surge |
| Platform availability | ≥ 99.9% uptime on critical path |
| Dispute resolution SLA | 95% of disputes resolved within 24 hours |
| Unit economics | Net Trip Revenue > variable cost per trip |

#### MVP Non-Negotiables

The following MUST be in the MVP regardless of scope pressure:
1. Upfront pricing display before every booking
2. Surge cap at 2.0x — enforced in the pricing engine, not just displayed
3. In-trip SOS button for both Rider and Driver
4. Commission itemization on Driver earnings notification
5. Formal Driver deactivation process with appeal
6. Dispute resolution within 24-hour SLA
7. Daily bank withdrawal for Drivers

These seven items are the product expression of FAIRRIDE's mission. If scope must be cut, cut elsewhere first.

#### MVP Technical Minimums

The MVP must support:
- 500 concurrent Riders requesting trips
- 2,000 concurrent Drivers online
- 200 concurrent trips in progress
- 99.9% availability for booking API and dispatch service
- Sub-3s app load on 3G connection
- Sub-30 second driver match time at P90

These are the minimum viable technical targets. DOC-0003 (System Architecture) must be designed to meet these targets.

---

### 6.20 Version 2 Vision

Version 2 transforms FAIRRIDE from a single-product ride-hailing app into a multi-category urban mobility platform.

**The V2 Product Promise to Drivers:**
- Choice of vehicle categories means more trip volume options
- Corporate accounts provide stable, premium-rated trips during business hours
- Fleet management tools allow professional operators to scale their business on FAIRRIDE

**The V2 Product Promise to Riders:**
- Premium and XL categories for different occasion types
- Scheduled rides for predictable commutes
- Ride pooling for cost-conscious trips with others going the same way

**V2 Architecture Implication:**
Version 2 requires the platform to support multiple vehicle categories in the dispatch algorithm, multiple pricing models, a corporate billing service, and a significantly more sophisticated Driver management system. These architectural requirements MUST be anticipated in DOC-0003 even if V2 implementation is deferred.

**V2 Key Metrics Targets:**
- 5 cities operational
- 30,000 daily completed trips across all cities
- $50M annual GMV run rate
- Driver retention (12-month cohort): ≥ 60%
- Corporate accounts: ≥ 50 active

---

### 6.21 Version 3 Vision

Version 3 makes FAIRRIDE a comprehensive urban commerce and logistics platform.

**The V3 Transformation:**
FAIRRIDE's driver network and logistics infrastructure is repurposed for delivery of packages, food, and commercial goods. The same Driver who transports a Rider in the morning can deliver packages in the afternoon and pick up food orders in the evening. This multi-modal efficiency maximizes Driver earnings per hour and FAIRRIDE's asset utilization.

**The Merchant Platform:**
Local businesses connect to FAIRRIDE as verified Merchants, using the FAIRRIDE network for delivery and potentially as advertising access to FAIRRIDE's user base.

**The Open API:**
Third-party developers build transportation features into their apps using FAIRRIDE's network via a public API. This creates a developer ecosystem and diversifies revenue.

**V3 Key Metrics Targets:**
- 3 countries operational
- 100,000+ daily transactions across all product lines
- $200M annual GMV run rate
- Delivery as ≥ 25% of GMV
- Open API: ≥ 100 active developers

**V3 Architecture Implication:**
The Open API Platform requires DOC-0003 to design the API layer with external developer access in mind from day one. API versioning, rate limiting, sandbox environments, and developer authentication must be designed as first-class architectural concerns, not retrofit after V2.

---

### 6.22 Success Metrics

#### Tier 1: Mission Metrics (quarterly executive review)

These metrics measure whether FAIRRIDE is actually fair, not just commercially successful.

| Metric | Definition | Target |
|--------|-----------|--------|
| Driver Earnings Premium | FAIRRIDE Driver average hourly earnings ÷ primary competitor Driver average hourly earnings | ≥ 1.20 (FAIRRIDE Drivers earn ≥ 20% more) |
| Surge Cap Compliance | % of trips where Surge Multiplier stayed ≤ 2.0x | 100% (system-enforced, not a target) |
| Upfront Pricing Accuracy | % of trips where final charge = displayed upfront fare ± 5% | ≥ 99.5% |
| Driver Deactivation Appeal Rate | % of deactivations that result in an appeal | Tracked; drives policy review if > 15% |
| Rider Fare Surprise Rate | % of trips where Rider contacts support about an unexpected charge | < 0.5% of trips |

#### Tier 2: Business Health Metrics (monthly leadership review)

| Metric | Definition | MVP Month 6 Target |
|--------|-----------|-------------------|
| GMV | Total fare value of completed trips | $22,500/day |
| Net Revenue | GMV × take rate − payment processing | $2,800/day |
| Daily Completed Trips | Count of trips completed | 5,000/day |
| Weekly Active Drivers (WAD) | Drivers completing ≥ 1 trip in trailing 7 days | 1,000 |
| Weekly Active Riders (WAR) | Riders completing ≥ 1 trip in trailing 7 days | 3,000 |
| Driver Utilization Rate | On-trip time ÷ total online time | ≥ 55% |
| Take Rate | Revenue ÷ GMV | ~15% |
| CAC (Rider) | Total rider acquisition spend ÷ new activated Riders | < $3.00 USD equivalent |
| CAC (Driver) | Total driver acquisition spend ÷ newly activated Drivers | < $15.00 USD equivalent |

#### Tier 3: Product Quality Metrics (weekly operations review)

| Metric | Definition | Target |
|--------|-----------|--------|
| Average ETA | Time from booking to Driver arrival | ≤ 6 minutes (P50); ≤ 10 minutes (P90) |
| Dispatch Success Rate | % of trip requests matched to a Driver within 60 seconds | ≥ 92% |
| Trip Completion Rate | % of accepted trips completed without cancellation | ≥ 88% |
| Rider RSAT | % of post-trip ratings ≥ 4 stars | ≥ 80% |
| Driver DSAT | % of Driver sessions with a complaint or rating < 3 | ≤ 15% |
| Dispute Rate | Disputes ÷ completed trips | < 2% |
| Dispute Resolution SLA Compliance | % of disputes resolved within 24 hours | ≥ 95% |
| App Stability (Rider) | Crash-free session rate | ≥ 99.5% |
| App Stability (Driver) | Crash-free session rate | ≥ 99.5% |
| API Availability | Booking + dispatch API uptime | ≥ 99.9% |

---

### 6.23 North Star Metric

**FAIRRIDE's North Star Metric is: Weekly Fair Matches (WFM)**

**Definition:**
A Fair Match is a completed Trip where:
1. The Trip was fully completed (not cancelled by either party after pickup)
2. The Driver independently rated the trip experience ≥ 4 out of 5 stars (or did not submit a negative rating)
3. The Rider independently rated the trip experience ≥ 4 out of 5 stars (or did not submit a negative rating)
4. No dispute was raised by either party within 24 hours of trip completion

Weekly Fair Matches = the count of Fair Matches in a trailing 7-day period.

**Why this metric and not GMV or Trip Count:**

GMV measures financial volume but not quality. A platform maximizing GMV might inflate fares via surge, reducing fairness. Trip count measures activity but not satisfaction. A platform maximizing trip count might accept low-quality trips with dissatisfied participants.

Weekly Fair Matches measures the thing FAIRRIDE exists to create: a completed trip where both participants were satisfied, the platform worked fairly, and no harm was done. A rising WFM means FAIRRIDE is delivering on its mission. A falling WFM is an immediate signal of a systemic problem, regardless of whether GMV is growing.

**WFM Target:**
- MVP Day 90: 10,000 WFM
- MVP Day 180: 25,000 WFM
- End of Year 1: 50,000 WFM
- End of Year 2: 150,000 WFM (multi-city)

**WFM Limitations and Mitigations:**
WFM undercounts satisfaction because most satisfied Riders and Drivers do not leave explicit positive ratings. FAIRRIDE mitigates this by treating "no negative rating" as a neutral-positive signal rather than excluding the trip. The metric is internally consistent even if the absolute number understates total satisfaction.

---

### 6.24 Product KPIs

#### Business KPIs (directly tied to revenue)

| KPI | Formula | Reporting Frequency |
|-----|---------|-------------------|
| GMV | Sum(Trip Fares) for period | Daily |
| Net Revenue | GMV × Take Rate − Payment Processing Costs | Daily |
| Effective Take Rate | Net Revenue ÷ GMV | Monthly |
| Average Revenue Per Active Rider (ARPR) | Net Revenue from Riders ÷ Monthly Active Riders | Monthly |
| LTV:CAC Ratio | LTV ÷ CAC per cohort | Quarterly |
| Unit Economics | Net Trip Revenue − Variable Cost per Trip | Weekly |

#### Supply KPIs (driver health)

| KPI | Formula | Reporting Frequency |
|-----|---------|-------------------|
| Weekly Active Drivers (WAD) | Unique Drivers completing ≥ 1 trip in 7 days | Weekly |
| Driver Utilization Rate | On-trip minutes ÷ Total online minutes | Daily |
| Average Driver Daily Earnings | Total driver payouts ÷ Active drivers | Daily |
| Driver Retention (30-day) | % of Drivers active in Month 1 who are active in Month 2 | Monthly |
| Driver Tier Distribution | % of fleet in Standard / Silver / Gold / Platinum | Monthly |
| Driver Acceptance Rate | Accepted trips ÷ Dispatched trips | Daily |

#### Demand KPIs (rider health)

| KPI | Formula | Reporting Frequency |
|-----|---------|-------------------|
| Weekly Active Riders (WAR) | Unique Riders completing ≥ 1 trip in 7 days | Weekly |
| Daily Completed Trips | Count of completed trips per day | Daily |
| Average Trips Per Rider Per Week | Completed trips ÷ WAR | Weekly |
| Rider Retention (30-day) | % of Riders active in Month 1 who are active in Month 2 | Monthly |
| Rider RSAT | % of rated trips ≥ 4 stars | Daily |
| Booking Conversion Rate | Confirmed bookings ÷ Fare estimate requests | Daily |

#### Operational KPIs (platform health)

| KPI | Formula | Reporting Frequency |
|-----|---------|-------------------|
| Average ETA (P50, P90) | Time from booking confirmation to Driver arrival | Real-time |
| Dispatch Success Rate | Matched bookings ÷ Total booking attempts | Real-time |
| Trip Completion Rate | Completed trips ÷ Accepted trips | Daily |
| Surge Frequency | Trips with surge ÷ Total trips | Daily |
| Cancellation Rate | Cancelled trips ÷ Accepted trips | Daily |
| Dispute Rate | Disputes ÷ Completed trips | Daily |
| Dispute Resolution SLA | % disputes resolved within 24h | Daily |
| Platform Availability | Booking API uptime | Real-time |

#### Fairness KPIs (mission alignment)

| KPI | Formula | Reporting Frequency |
|-----|---------|-------------------|
| Weekly Fair Matches | Count of Fair Match trips in 7 days | Weekly (North Star) |
| Driver Earnings Premium | FAIRRIDE Driver hourly earnings ÷ Competitor estimate | Monthly |
| Upfront Pricing Accuracy | Trips where final = upfront ± 5% ÷ Total | Daily |
| Deactivation Appeal Rate | Driver deactivation appeals ÷ Total deactivations | Monthly |
| Rider Fare Surprise Rate | Support contacts about unexpected charge ÷ Trips | Daily |
| Surge Cap Compliance | Trips where surge ≤ 2.0x ÷ Total surge trips | Daily (must be 100%) |

---

### 6.25 Definition of Success

FAIRRIDE has succeeded at MVP when **all** of the following are true simultaneously:

#### Business Success

- [ ] Unit economics are positive: each completed Trip contributes net positive revenue after all variable costs
- [ ] Driver supply is self-sustaining: new Driver registrations through word-of-mouth and referrals exceed paid acquisition within the cohort
- [ ] Rider demand is self-sustaining: monthly new Rider activations through referral ≥ 30% of total new activations

#### Mission Success

- [ ] FAIRRIDE Drivers in the launch city earn ≥ 20% more per hour than Drivers on the primary competitor platform (independently measured or driver-reported)
- [ ] Rider RSAT ≥ 80% sustained for 30+ consecutive days
- [ ] Zero verified cases of a Rider being charged more than the upfront price displayed (without route deviation)
- [ ] Zero cases of Surge exceeding 2.0x

#### Operational Success

- [ ] Platform availability ≥ 99.9% averaged over the MVP period
- [ ] ETA accuracy: P90 ≤ 10 minutes for 30+ consecutive days
- [ ] Dispute resolution SLA ≥ 95% within 24 hours for 30+ consecutive days
- [ ] All Driver deactivations have a documented reason code and appeal notification sent

#### Strategic Success

- [ ] FAIRRIDE has a documented, validated Unit Economics model that a Series A investor can evaluate
- [ ] FAIRRIDE has evidence of brand differentiation: Drivers explicitly choose FAIRRIDE over competitors at equal volume
- [ ] FAIRRIDE has a city expansion playbook: a documented, repeatable process for launching a new city within 90 days

When all 13 of these conditions are met, FAIRRIDE has succeeded at MVP. At that point, the platform is ready to begin Phase 1 scaling.

---

## 7. NON-FUNCTIONAL REQUIREMENTS

| Requirement | Standard | Governing Document |
|------------|---------|-------------------|
| **Document longevity** | This document must remain substantively valid for 3 years with only minor revisions. Major product changes require a new version (v2.0). | Phase 0 versioning convention |
| **Mission alignment** | Every section of this document must be traceable to the FAIRRIDE mission (Section 6.1) and the Core Values (Section 6.3). If a product decision cannot be traced to the mission, it is not in scope. | DOC-0001, Article I |
| **Architecture derivability** | DOC-0003 (System Architecture) MUST be fully derivable from the product requirements stated in Sections 6.16–6.21 of this document, without requiring additional clarification from business or product stakeholders. | DOC-0001, Article II, Value 2.6 |
| **Fairness testability** | Every KPI in Section 6.24, particularly the Fairness KPIs, must be measurable with data that the platform itself produces. No KPI may require external data sources at MVP. | DOC-0001, Article II, Value 2.5 |
| **Accessibility for AI agents** | Sections 6.3, 6.9, 6.17, 6.19, and 6.25 must be structured so that an AI agent working on a feature or service can determine: (a) Is this feature in scope for MVP? (b) Does this feature comply with FAIRRIDE's fairness values? | DOC-0001A, Section 6.4 |

---

## 8. CONSTRAINTS

| Constraint | Description |
|-----------|-------------|
| **Single launch city (MVP)** | All MVP commitments (ETAs, supply targets, trip volumes) are sized for one city. Multi-city architecture must be designed but multi-city operations are not in MVP scope. |
| **Two mobile platforms only** | iOS and Android. No web-based Rider or Driver app in MVP. Progressive web app is not a planned alternative. |
| **English + 1 local language (MVP)** | App must be localized in English and one local language determined by OQ-001 (launch market). Additional languages are Phase 1. |
| **Card + wallet payments only (MVP)** | Cash collection by Drivers is out of MVP scope. QR code payments are Phase 1. Only card and in-app wallet at MVP. |
| **Rule-based fraud detection (MVP)** | ML-based fraud detection is not in MVP scope. Baseline rule engine must be sufficient for MVP transaction volume. |
| **Commission transparency is non-negotiable** | No MVP scope cut may remove commission itemization from Driver earnings notifications. This is immutable per Section 6.19 MVP Non-Negotiables. |
| **Surge cap is non-negotiable** | 2.0x surge cap is a system constraint, not a configuration option. It must be enforced at the Pricing Engine level, not the UI level. |

---

## 9. RISKS

| Risk ID | Description | Likelihood | Impact | Mitigation |
|---------|-------------|-----------|--------|-----------|
| R-001 | **Supply acquisition failure:** FAIRRIDE cannot recruit sufficient Drivers before rider launch, resulting in poor ETAs and rider churn. | Medium | Critical | Supply seeding strategy (Section 6.11); 500 Drivers minimum before rider launch; 0% commission for first 100 Drivers |
| R-002 | **Competitor retaliation:** Primary competitor (Grab/local equivalent) responds to FAIRRIDE entry with Driver incentive campaign or Rider pricing war. | High | High | FAIRRIDE competes on structural fairness, not promotional spending. Driver tier system creates long-term loyalty that promo campaigns cannot match. |
| R-003 | **Regulatory risk:** Ride-hailing regulations in launch market change unfavorably post-launch. | Medium | High | Legal compliance review before entry (OQ-001 resolution must include regulatory assessment). FAIRRIDE's transparent operating model reduces regulatory adversarial risk. |
| R-004 | **Payment fraud at scale:** Fraudulent trips, account takeover, or payment fraud erodes unit economics. | Medium | High | Baseline fraud rule engine at MVP. GPS spoofing protection in dispatch. KYC for Drivers. Security AI review of payment flows before launch. |
| R-005 | **ETA inaccuracy at launch:** Insufficient Driver density in certain zones leads to poor ETAs, violating the Rider Trust value and damaging early reviews. | High | Medium | Zone-based Driver seeding strategy. Supply health monitoring. Temporary zone restriction during low-supply periods rather than showing poor ETAs. |
| R-006 | **Earnings promise not delivered:** FAIRRIDE claims Drivers earn more, but real-world earnings (accounting for idle time, costs) do not exceed competitor earnings. | Medium | Critical | Earnings Premium KPI tracked from month 1. Driver earnings survey run monthly. If earning premium < 20% for 2 consecutive months, commission model must be reviewed. |
| R-007 | **Safety incident at MVP:** A significant safety incident (assault, accident) during MVP period damages brand and attracts regulatory attention. | Low | Critical | MVP Non-Negotiable: SOS button, trip sharing, Driver identity verification. Incident response plan documented before launch. Insurance partnership explored pre-launch. |
| R-008 | **Scope creep:** Pressure to add features (pooling, multi-category) to MVP erodes focus and delays launch. | High | Medium | Section 6.17 Out of Scope list is formally approved by CTO + CPO. Any addition requires documented override with business justification. |
| R-009 | **Open Question delay:** Unresolved OQs (particularly OQ-001 — launch market, OQ-006 — driver classification) delay DOC-0003 and DOC-0004 authoring. | Medium | High | OQ-001 and OQ-006 are P0 priority for resolution. Target: resolved before DOC-0003 generation begins. |
| R-010 | **Product-architecture mismatch:** DOC-0003 makes architecture decisions that later conflict with product requirements in this document. | Low | High | DOC-0003 explicitly depends on DOC-0002. All system architecture decisions trace to a product requirement documented here. Reviewer AI checks alignment. |

---

## 10. FUTURE EXTENSION

### 10.1 Platform Expansion Paths

The FAIRRIDE Platform has three natural expansion paths beyond Version 3:

**Expansion Path A — Vertical Depth:** Deepen each product line to enterprise level. Corporate transport becomes full corporate travel management. Delivery becomes end-to-end logistics with warehouse integration. Food delivery becomes dark kitchen partnerships.

**Expansion Path B — Geographic Breadth:** Expand to 20+ countries. Requires the Open API Platform and Merchant Platform to be operational so that local partners can accelerate market entry without FAIRRIDE building everything from scratch.

**Expansion Path C — Platform Openness:** FAIRRIDE becomes infrastructure. Third-party apps run on FAIRRIDE's driver network, dispatch system, and payment rails. FAIRRIDE operates as the backend mobility platform for other businesses. This is the highest-value long-term outcome.

### 10.2 Driver Ecosystem Evolution

As FAIRRIDE matures, the Driver relationship evolves from "gig worker on platform" to "business partner in ecosystem":
- Driver-owned mini-businesses using FAIRRIDE as their primary channel
- Driver cooperative structures enabled by FAIRRIDE's transparent earnings model
- Driver financial products: savings accounts, vehicle loans, insurance — all accessible through FAIRRIDE's trusted relationship

### 10.3 Product Vision Updates

This document MUST be reviewed and updated when:
- A new Phase is initiated (Phase 1, 2, or 3 launch)
- A new country is entered (regulatory and cultural context changes product requirements)
- Significant competitive market changes occur (new competitor entry, competitor collapse)
- The North Star Metric requires recalibration based on observed behaviour

---

## 11. OPEN QUESTIONS

The following questions must be resolved before DOC-0003 or DOC-0004 can be finalized. They inherit from DOC-0001 (OQ-001 through OQ-009) and add product-specific open questions.

| OQ ID | Question | Impact | Priority | Owner |
|-------|----------|--------|---------|-------|
| OQ-001 | What is the primary launch city and country? | Critical — affects regulatory, language, payment rails, map provider, currency | P0 | CEO / CPO |
| OQ-006 | Are Drivers classified as independent contractors or employees in the launch market? | High — affects earnings model, payroll tax, benefits, Worker protection disclosure | P0 | Legal / CPO |
| OQ-B01 | What is the cash collection policy for MVP? The market may require cash acceptance from Day 1. | High — if yes, significantly changes Driver earnings flow and reconciliation design | P0 | CPO + Payments Lead |
| OQ-B02 | What is the first-ride Rider incentive amount? | Medium — affects CAC model and launch budget | P1 | CPO + Finance |
| OQ-B03 | What is the Driver 0% commission early adopter programme duration and cap? | Medium — affects supply cost at launch | P1 | CPO + Finance |
| OQ-B04 | What ride categories are in MVP? Is motorcycle taxi required from Day 1 in the launch market? | High — motorcycle taxi requires different dispatch rules and KYC | P1 | CPO |
| OQ-B05 | What is the booking fee amount? | Medium — affects unit economics and rider price perception | P1 | CPO + Finance |
| OQ-B06 | What is the instant withdrawal fee and mechanism? | Medium — affects Driver Wallet design and payment provider selection | P1 | Payments Lead |
| OQ-B07 | What level of map provider service is required at MVP? Does FAIRRIDE require its own geocoding, or is a third-party API sufficient? | Medium — affects Geo Engine architecture and monthly operating cost | P1 | CTO + Geo Lead |
| OQ-007 | Is offline-capable dispatch required at MVP due to connectivity in the launch market? | High — materially changes mobile architecture approach | P1 | CTO + Mobile Lead |
| OQ-B08 | What is the Driver Advisory Panel selection process and schedule for the first quarter? | Low | P2 | CPO |

---

## 12. DECISION REFERENCES (ADR)

The following ADRs are triggered by or directly related to this document. They MUST be resolved before DOC-0003 is authored.

| ADR ID | Decision Required | Section Reference | Priority |
|--------|-----------------|------------------|---------|
| ADR-0008 | Lean Documentation Strategy: formal adoption of 5-document core EOS | Architecture Change decision | P0 — before any new document is generated |
| ADR-0009 | Primary Launch Market selection | OQ-001, Section 6.11 | P0 — before DOC-0003 |
| ADR-0010 | Commission Model: formal adoption of graduated tier structure (15%→10%) | Section 6.7 | P0 — before DOC-0003 (affects Pricing Engine) |
| ADR-0011 | Surge Cap Policy: formal adoption of 2.0x hard cap | Section 6.9 | P0 — before DOC-0003 (affects Pricing Engine) |
| ADR-0012 | Surge Revenue Pass-Through: 100% of surge to Driver | Section 6.7 | P1 — before Pricing Engine spec |
| ADR-0013 | Driver Tier Calculation: trailing 30-day window and graduation rules | Section 6.7 | P1 — before Driver Earnings service design |
| ADR-0014 | MVP Payment Methods: Card + Wallet only at MVP (cash deferred) | Section 6.17 / OQ-B01 | P1 — before Payment Engine design |
| ADR-0015 | North Star Metric definition: Weekly Fair Matches | Section 6.23 | P1 — before analytics design |
| ADR-0016 | Driver Classification Model in launch market | OQ-006 | P0 — legal determination required |

---

## 13. REVISION HISTORY

| Version | Date | Author | Status | Change Summary |
|---------|------|--------|--------|---------------|
| 0.1.0 | 2026-06-30 | Office of the CTO / CPO | Draft | Initial creation of FAIRRIDE Product Vision. Absorbs 11 previously planned EOS documents under Lean Documentation Strategy. All 30 required sections authored. 5 personas defined. Full feature roadmap across MVP, V2, V3. North Star Metric defined. Pending review and CTO + CPO approval. |

---

*End of Document — DOC-0002 — Product Vision — v0.1.0*

*This document requires CTO + CPO approval before DOC-0003 (System Architecture) is generated.*
*Open Questions OQ-001, OQ-006, and OQ-B01 are P0 priority and SHOULD be resolved before or concurrent with DOC-0002 approval.*
