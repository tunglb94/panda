# FAIRRIDE Business Rule Bible — Version 1.0

**Document Classification:** Internal — Confidential  
**Authority:** CPO · CFO · COO · CTO  
**Effective Date:** 2026-07-09  
**Review Cycle:** Quarterly  
**Supersedes:** All prior pricing memos, incentive sheets, and policy drafts

---

## Preamble

This Business Rule Bible (BRB) is the single source of truth for every business decision at FAIRRIDE. It governs how money flows, how drivers earn, how riders pay, how promotions work, how fraud is penalised, and how the platform sustains itself over the long term.

Every product manager, engineer, data analyst, and operations manager must read this document in full before touching any revenue-related system. When a feature conflicts with this document, this document wins. When this document is silent on a topic, a CPO-level decision is required before implementation proceeds.

This document describes **business rules only**. It intentionally contains no code, no database schemas, no API contracts, no architecture diagrams. Those are implementation details — subordinate to the business logic defined here.

---

# PART 1 — BUSINESS PHILOSOPHY

## 1.1 Mission

FAIRRIDE exists to create a transportation platform where every ride is priced fairly, every driver earns a dignified income, and every rider trusts that the price they see is the price they pay.

We are not optimising for market dominance at any cost. We are building a sustainable platform that earns rider trust through price transparency and earns driver loyalty through honest earnings. We believe these two goals are complementary, not competing.

## 1.2 Core Principles

**Principle 1 — Transparency Before Conversion.**  
A rider must always see the complete fare breakdown before confirming a booking. There are no hidden fees. Surge pricing is always labelled. Promotions always show their exact discount. If we cannot show the full price upfront, we do not show the booking button.

**Principle 2 — Rules Are Public, Algorithms Are Honest.**  
Pricing rules are published. Surge thresholds are disclosed. Commission tiers are visible to drivers. We may hold proprietary optimisations inside our algorithms, but the underlying rules that determine how money moves are never hidden from the people who depend on them.

**Principle 3 — A Rule Exists for a Reason, and the Reason Is Documented.**  
Every business rule in this document carries a rationale. If a rule cannot be justified, it should not exist. If a rule's justification becomes obsolete, the rule must be reviewed. Rules without rationale create technical debt and operational confusion.

**Principle 4 — Simplicity Scales, Complexity Breaks.**  
When two approaches deliver similar outcomes, we choose the simpler one. Complex pricing tables, multi-variable commission calculations, and opaque scoring systems all produce the same outcome: confusion on the street, errors in the ledger, and erosion of trust. Simple rules that everyone understands are a competitive advantage.

**Principle 5 — Every Rule Must Be Fraud-Resistant by Design.**  
We do not assume good faith from anonymous actors. Every incentive scheme, every promotion, every referral programme is designed from the start with its abuse case in mind. We are not naïve; we are careful.

## 1.3 Platform Fairness

Fairness at FAIRRIDE has three dimensions:

**Rider Fairness.** Riders pay a price that reflects the real cost of their trip. They are not charged more simply because our algorithm detects they are willing to pay more. Surge pricing reflects genuine supply-demand imbalance, not price discrimination based on device type, app version, or payment history. Two riders requesting the same trip at the same moment in the same conditions receive the same price.

**Driver Fairness.** Drivers earn a percentage of the fare they generate. When the fare rises due to surge, drivers benefit proportionally. When promotions are funded by FAIRRIDE, drivers are made whole — they never absorb a discount that the platform chose to offer. A driver's income is predictable, understandable, and increasing over time as they build seniority on the platform.

**Platform Fairness.** FAIRRIDE takes a transparent commission. We do not extract value through hidden mechanisms — late payments, delayed settlements, or opaque deductions. Our revenue is visible in the settlement breakdown that every driver sees after every completed trip.

## 1.4 Driver-First Philosophy

Drivers are the physical product of FAIRRIDE. Without them, there is no service. Without a good service, there are no riders. Without riders, the platform dies. This causal chain means driver welfare is a business imperative, not a corporate social responsibility exercise.

**Drivers are treated as partners, not contractors in the dismissive sense.** This means:
- Commission reductions are communicated 30 days in advance, never retroactively
- Incentive rules are published at the start of each period, never changed mid-period
- Disputes are resolved within 48 hours with a written outcome
- Driver accounts are never suspended without a documented reason and an appeal path

**Driver income is predictable.** We design incentives so that a driver who follows normal professional behaviour — online during busy hours, accepting most trip requests, maintaining a good rating — earns a reliable income. Drivers should not need to gamble on secret optimisations or exploit loopholes to earn well.

## 1.5 Customer Trust

Riders trust FAIRRIDE when:
- The price they saw is the price they paid
- The driver who arrived matches the profile they saw
- When something goes wrong, it is resolved quickly and fairly
- Their payment data and personal data are handled responsibly

We earn and maintain trust through consistency. A platform that offers surprise discounts one week and unexplained surcharges the next trains riders to be suspicious. FAIRRIDE trains riders to be comfortable. Every interaction should confirm that we operate exactly as we described.

## 1.6 Long-Term Sustainability

FAIRRIDE is not an investor-subsidy operation. We do not plan to burn cash for years acquiring users at a loss, then raise prices once competitors are gone. This approach harms the ecosystem and eventually harms the platform.

Our path to sustainability:
- Platform revenue from the first paying trip
- Promotions funded with a finite budget, measured against actual ROI
- Commission rates designed to sustain driver income and platform operations simultaneously
- No cross-subsidisation that masks the true cost of growth

We will grow more slowly than a subsidy-funded competitor in the short term. We will outlast them.

## 1.7 Platform Revenue Philosophy

FAIRRIDE earns money in four ways, in order of importance:

1. **Commission on completed trips.** A percentage of every fare paid by a rider flows to the platform. This is our primary revenue source and must remain so.
2. **Booking fee.** A flat fee charged per completed trip, separate from the commission. This covers payment processing and platform operating costs.
3. **Promotional placement fees.** Corporate and third-party sponsors pay to fund promotions distributed through our platform. We do not earn a margin on these; we charge a placement and administration fee.
4. **Premium driver services.** Fast withdrawal, verified badge, priority placement — optional services that drivers pay for if they want them.

We do not earn revenue from:
- Surge pricing disproportionately retained by the platform (drivers benefit from surge too)
- Rider data sold to third parties
- Hidden fees disguised as line items

## 1.8 Risk Philosophy

FAIRRIDE manages three categories of financial risk:

**Operational Risk** — Trips where something goes wrong (driver no-show, accident, incorrect route). Managed through the dispute process, insurance partners, and refund policy.

**Financial Risk** — Wallet balances, uncollected payment, chargeback, settlement failure. Managed through pre-authorisation, balance checks, and settlement timing rules.

**Fraud Risk** — GPS spoofing, fake accounts, incentive manipulation, voucher abuse. Managed through the Risk Engine described in Part 12.

Risk management is not a separate department at FAIRRIDE. Every product team owns the fraud and operational risk of the features it builds. Engineers are required to read the Fraud Rules in Part 11 before shipping any incentive, referral, or payment feature.

---

# PART 2 — TRIP PRICING RULES

## 2.1 Philosophy of Pricing

A fare should be easy to explain to a driver and a rider in under sixty seconds. If a fare cannot be explained in sixty seconds, the pricing model is too complex and must be simplified.

FAIRRIDE uses a **component-based fare model**. The total fare is the sum of individual components, each with a clear rule. Riders see each component in the fare breakdown. Drivers see their share of each component.

## 2.2 Fare Components

### 2.2.1 Base Fare

**Definition.** A flat fee charged at the moment a trip begins — when the driver confirms pickup.

**Purpose.** Compensates the driver for the time and cost of travelling to the pickup location, which is not captured by distance or time metering.

**Rule.** Base fare is fixed per city per vehicle class. It does not vary by time of day, weather, or demand level. It is not multiplied by surge.

**Rationale.** The base fare is the predictable anchor of every fare. Subjecting it to surge creates volatility that confuses riders and is difficult to justify. All surge is applied via the Surge Multiplier (see 2.12).

**Default values (primary city, launch):**

| Vehicle Class | Base Fare |
|---|---|
| Standard (4-seat) | 10,000 VND |
| Premium (4-seat, higher quality) | 15,000 VND |
| XL (7-seat) | 18,000 VND |

### 2.2.2 Distance Fare

**Definition.** A per-kilometre charge applied to the metered distance of the trip.

**Metering rule.** Distance is measured as the GPS-tracked distance travelled by the driver from the moment the rider enters the vehicle to the moment the driver confirms trip completion. FAIRRIDE's own Route Engine provides the distance measurement. Google Maps or third-party services may be used for route display, but billing distance is always our own measurement.

**Why our own measurement?** Third-party APIs can fail, change pricing, or have outages. Billing on our own measurement insulates riders and drivers from external dependencies.

**Rule.** Distance fare is fixed per city per vehicle class per kilometre. It is not applied during waiting time (waiting has its own component). It is not applied while the driver is en route to pickup.

**Default values (primary city, launch):**

| Vehicle Class | Distance Fare |
|---|---|
| Standard | 4,000 VND/km |
| Premium | 5,500 VND/km |
| XL | 5,000 VND/km |

### 2.2.3 Time Fare

**Definition.** A per-minute charge applied when the vehicle is moving but below a defined speed threshold.

**Purpose.** In heavy traffic, a ride that covers 2 km may take 20 minutes. Pure distance pricing would produce an absurdly low fare. Time fare compensates for slow-moving conditions.

**Activation threshold.** Time fare activates when GPS-measured speed is below **10 km/h** for more than 15 consecutive seconds. Time fare deactivates when speed exceeds 10 km/h.

**Rule.** Time fare and distance fare are **mutually exclusive within each second**. A given second of the trip is either billed as distance (speed ≥ 10 km/h) or as time (speed < 10 km/h), never both.

**Rationale.** This "greater of time or distance" approach, used by Uber and Lyft, prevents double-counting during slow movement.

**Default values (primary city, launch):**

| Vehicle Class | Time Fare |
|---|---|
| Standard | 400 VND/minute |
| Premium | 550 VND/minute |
| XL | 500 VND/minute |

### 2.2.4 Minimum Fare

**Definition.** The lowest total fare (before promotions) that can be charged for any completed trip.

**Purpose.** Very short trips generate disproportionate overhead — app matching, payment processing, and driver positioning cost — that the metered fare would not cover.

**Rule.** If (Base Fare + Distance Fare + Time Fare) < Minimum Fare, then the rider is charged the Minimum Fare. The driver receives their normal commission percentage of the Minimum Fare.

**Default values (primary city, launch):**

| Vehicle Class | Minimum Fare |
|---|---|
| Standard | 25,000 VND |
| Premium | 35,000 VND |
| XL | 40,000 VND |

### 2.2.5 Booking Fee

**Definition.** A flat fee charged per completed trip, on top of the metered fare.

**Purpose.** Covers platform operating costs: payment processing fees, customer support, insurance, and technology infrastructure.

**Rule.** The booking fee is fixed. It is NOT subject to surge. It is NOT shared with the driver. It flows entirely to FAIRRIDE. It is shown as a separate line item in the fare breakdown.

**Default value:** 2,000 VND per trip (all vehicle classes).

**Rationale.** A flat booking fee is transparent and simple. Riders know exactly what the platform charges for each trip. Because it is fixed and not surged, it does not inflate during peak periods, which would feel punitive.

### 2.2.6 Platform Fee (Commission)

**Definition.** The percentage of the metered fare that FAIRRIDE retains as its commission.

**Rule.** Commission is applied to (Base Fare + Distance Fare + Time Fare + applicable surcharges). It is not applied to the Booking Fee (which is 100% platform revenue already) and not applied to Toll Fees (which are 100% pass-through).

**Default rate:** 20% platform / 80% driver for new and Bronze drivers. Lower rates apply for higher driver tiers (see Part 7).

**Rationale.** 20% is below the industry high of 25–30% used by some competitors. FAIRRIDE's lower commission is a deliberate driver acquisition strategy. It is sustainable because our operational model relies on volume and the flat Booking Fee rather than maximising the commission percentage.

### 2.2.7 Airport Fee

**Definition.** A fixed surcharge applied when the trip origin or destination is an airport zone.

**Purpose.** Airport operations have higher overhead: airport waiting zones, longer driver positioning time, higher congestion. The fee compensates drivers and covers platform airport operations.

**Rule.** Airport fee is charged once per trip (not twice if both origin and destination are airport zones in edge cases — the higher fee applies). The fee is shared with the driver at the same commission split as the metered fare.

**Airport zone definition.** A geographic polygon defined in the admin portal for each airport. Any pickup or dropoff within the polygon triggers the fee.

**Default value:** 10,000 VND per trip.

### 2.2.8 Toll Fee

**Definition.** Road tolls incurred during a trip.

**Rule.** All toll costs are **passed through at cost to the rider**. FAIRRIDE does not take a commission on toll fees. Toll fees are a separate line item.

**Collection method.** Drivers declare tolls encountered during the trip. The declared toll must match known toll amounts on the route. Admin can set a toll rate table for known toll roads. If a driver declares a toll not on a known toll road, the claim is flagged for review.

**Fraud note.** Drivers cannot declare tolls on trips where the route does not pass through a toll plaza. The Route Engine flags this automatically.

### 2.2.9 Waiting Fee

**Definition.** A per-minute charge applied after a grace period once the driver has arrived at the pickup location.

**Grace period.** The first 3 minutes after the driver marks "Arrived" are free. This is the rider's reasonable preparation time.

**Activation.** Starting at minute 4 after "Arrived", waiting fee accumulates at the defined rate.

**Maximum waiting time.** After 10 minutes of total waiting (7 minutes of chargeable time), the driver is permitted to cancel the trip without penalty. If the driver chooses to wait longer, waiting fee continues to accumulate.

**Rule.** Waiting fee is included in the total fare subject to commission split. The driver receives 80% (or their tier rate) of the waiting fee.

**Default value:** 500 VND/minute (all vehicle classes, applied from minute 4 onward).

**Example.** Driver arrives at 14:00. Rider enters vehicle at 14:08. Grace period ends at 14:03. Chargeable waiting = 5 minutes × 500 VND = 2,500 VND.

### 2.2.10 Night Surcharge

**Definition.** An additional percentage applied to the metered fare for trips that begin during night hours.

**Night hours:** 22:00 – 05:00 local time.

**Rule.** Night surcharge is a multiplier applied to (Base Fare + Distance Fare + Time Fare). It is not applied to Booking Fee or Toll Fee. The surcharge is shared with the driver at the standard commission split.

**Trigger.** The trip's start time determines whether the night surcharge applies for the entire trip. A trip that begins at 22:01 is a night trip even if it ends at 23:30.

**Default rate:** +20% (multiplier: 1.20).

**Rationale.** Night driving is harder, involves higher personal risk for drivers, and reduces their ability to go home. The night surcharge compensates for this without requiring a separate rate table for every fare component.

### 2.2.11 Holiday Surcharge

**Definition.** An additional percentage applied to the metered fare on defined public holidays.

**Holiday list.** Maintained in the admin portal per city/country. Examples: National Day, Lunar New Year (all applicable days), Independence Day.

**Rule.** Same structure as Night Surcharge. Holiday surcharge and Night Surcharge can stack. If both apply, the combined multiplier is applied to the base metered fare.

**Default rate:** +15% (multiplier: 1.15).

**Stacking with night:** Trip on a holiday night = (Base Fare + Distance + Time) × 1.20 (night) × 1.15 (holiday).

**Cap.** Combined surcharge multipliers from Night + Holiday cannot exceed ×1.50.

### 2.2.12 Peak Hour Surcharge

**Definition.** An additional percentage applied during defined high-demand time windows.

**Peak windows (defaults, adjustable per city):**
- Morning peak: 07:00 – 09:00, Monday–Friday
- Evening peak: 17:00 – 20:00, Monday–Friday

**Rule.** Peak hour surcharge does not stack with Dynamic Surge (see 2.13). When Dynamic Surge is active, it supersedes the Peak Hour Surcharge. Peak Hour Surcharge is a static rule; Dynamic Surge is a real-time calculation.

**Default rate:** +10% (multiplier: 1.10).

### 2.2.13 Rain Surcharge

**Definition.** An additional surcharge applied during rain conditions in a city zone.

**Trigger.** FAIRRIDE Operations activates Rain Surcharge manually, or it is triggered automatically when a verified weather API reports rainfall above a defined threshold (≥2mm/hour) in the active zone. The trigger and its status are shown to riders in the app.

**Rule.** Rain surcharge is always labelled in the app. Riders see: "Rain demand: +X%". Rain surcharge can stack with Night and Holiday surcharges, subject to the combined cap below.

**Default rate:** +15% (multiplier: 1.15).

**Combined cap.** Night + Holiday + Rain multipliers applied simultaneously cannot exceed ×1.60.

## 2.13 Dynamic Pricing (Surge)

### 2.13.1 Philosophy

Dynamic pricing is a mechanism to balance supply and demand in real time. When there are more riders requesting trips than drivers available to take them, the fare increases. The higher fare accomplishes two things simultaneously: some riders defer their trip (demand reduction), and offline drivers are incentivised to come online (supply increase).

Dynamic pricing is not a tool to extract maximum willingness-to-pay from individual riders. It is a market-clearing mechanism.

### 2.13.2 Surge Multiplier Calculation

**Input signals:**
- Active trip requests in a zone over the last 5 minutes
- Available drivers in or approaching a zone
- Estimated pickup wait time (EWT)

**Demand-supply ratio (DSR):** Active requests ÷ Available drivers in zone.

| DSR Range | Surge Multiplier | Label shown to rider |
|---|---|---|
| < 1.2 | ×1.0 (no surge) | Normal pricing |
| 1.2 – 1.5 | ×1.2 | Busy |
| 1.5 – 2.0 | ×1.4 | High demand |
| 2.0 – 2.5 | ×1.6 | Very high demand |
| 2.5 – 3.0 | ×1.8 | Peak demand |
| > 3.0 | ×2.0 (maximum) | Maximum surge |

### 2.13.3 Maximum Surge

**Rule.** Surge multiplier cannot exceed ×2.0. No exception.

**Rationale.** Unlimited surge destroys rider trust and invites regulatory action. A ×2.0 cap means the maximum fare is double the base rate — painful but not exploitative. Above this level, no amount of additional driver incentive from the surge is worth the permanent rider trust damage.

### 2.13.4 Surge Application

Surge multiplier is applied to: (Base Fare + Distance Fare + Time Fare + Airport Fee if applicable).

Surge is NOT applied to: Booking Fee, Toll Fee, Waiting Fee.

Surge IS shared with the driver at their tier commission split. A ×2.0 surge trip produces twice the metered fare, of which the driver earns their normal percentage.

### 2.13.5 Surge Transparency

- The app shows the surge multiplier explicitly before the rider confirms the booking
- The rider must tap a confirmation acknowledging the surge before booking proceeds
- The driver is shown the surged fare, not the base fare, so they know what they're earning

### 2.13.6 Price Cap (Absolute Maximum Fare)

**Rule.** No trip fare (before promotions/vouchers) can exceed a defined per-city maximum fare, regardless of surge, distance, or time.

**Purpose.** Prevents extreme scenarios — a 5-hour traffic jam producing a 2,000,000 VND fare — that would destroy trust even if technically justified.

**Default cap:** 500,000 VND per trip for Standard class (city trips). Airport long-distance trips have a higher cap negotiated separately.

**When the cap is reached:** The fare stops accumulating. The driver is still paid as if the metered fare had been applied (the platform absorbs the cap if needed in extreme scenarios, or the cap is set high enough that this is essentially never triggered in normal operations).

## 2.14 Minimum Fare Guarantee

**Rule.** A driver who completes a trip will always earn at least 20,000 VND net (after commission), regardless of how short the trip is.

**How it works.** If the driver's earnings from their commission percentage on the metered fare fall below 20,000 VND, the platform tops up the difference. This top-up is a platform cost, not recovered from the rider (the rider pays the Minimum Fare defined in 2.2.4, not necessarily a higher amount).

**Rationale.** No driver should accept a trip that costs them money to complete. The Minimum Fare Guarantee ensures that every completed trip is worth the driver's time.

## 2.15 Price Rounding Rules

**Rule.** All fare calculations are performed at full precision. The final total fare charged to the rider is rounded to the nearest 500 VND (up).

**Driver payout** is calculated from the precise (unrounded) metered fare, then rounded to the nearest 100 VND (up).

**Rationale for upward rounding.** Rounding up ensures the driver never loses money due to rounding. The maximum overcharge from upward rounding is 499 VND — negligible for the rider, meaningful in aggregate for driver earnings.

**Example:**
- Metered fare components sum to: 47,320 VND
- Booking Fee: 2,000 VND
- Total billed to rider: 47,320 → rounded to 47,500 VND + 2,000 VND Booking Fee = 49,500 VND
- Driver earns 80% of 47,320 = 37,856 → rounded to 37,900 VND

## 2.16 Currency Rules

**MVP launch:** Single currency (VND). All fares, wallets, and settlements are in VND.

**Decimal handling.** VND has no decimal subunit. All calculations use integer VND after rounding. Intermediate calculations may use floating-point but must round before storing.

**Future multi-currency rule (see Part 15):** When operating in multiple countries, each country maintains its own currency ledger. Cross-currency conversion is not supported at the trip level. Drivers and riders in each country are always paid/charged in their local currency.

## 2.17 Complete Fare Calculation Example

**Scenario:** Standard car, 8 km trip, 15 minutes total duration, 3 minutes in traffic (below 10 km/h), trip begins at 22:30 on a normal weekday, rain is active, surge is ×1.4.

**Components:**
- Base Fare: 10,000 VND
- Distance: 8 km × 4,000 = 32,000 VND (for the 12 minutes of movement above 10 km/h)
- Time: 3 minutes × 400 VND = 1,200 VND (traffic time)
- Metered sub-total before surcharges: 43,200 VND
- Night surcharge: ×1.20 → metered becomes 43,200 × 1.20 = 51,840 VND
- Rain surcharge: ×1.15 → 51,840 × 1.15 = 59,616 VND
- Surge multiplier (×1.4) supersedes individual surcharges? **No.** Rain is a static surcharge stacked on top of Night. Surge is separately applied to the pre-surcharge metered sub-total and then Night/Rain are applied.

**Revised calculation order:**
1. Metered sub-total: 43,200 VND
2. Apply surge ×1.4: 43,200 × 1.4 = 60,480 VND
3. Apply Night ×1.20: 60,480 × 1.20 = 72,576 VND
4. Apply Rain ×1.15: 72,576 × 1.15 = 83,462 VND
5. Check combined surcharge: Night + Rain = ×1.20 × ×1.15 = ×1.38 (within cap of ×1.60) ✓
6. Add Booking Fee: 83,462 + 2,000 = 85,462 VND
7. Round to nearest 500: 85,500 VND ← **rider pays**
8. Driver earns 80% of 83,462 = 66,770 → rounded to 66,800 VND ← **driver earns**
9. Platform earns: 16,692 VND (commission) + 2,000 VND (booking fee) = **18,692 VND**

## 2.17B Additional Fare Examples

**Example A — Standard short city trip, no surcharge:**
- Vehicle: Standard. Distance: 3 km. Duration: 8 minutes. No traffic. No surcharge. No surge.
- Base fare: 10,000 VND
- Distance: 3 × 4,000 = 12,000 VND
- Time (all above 10 km/h): 0 VND (time fare only activates when speed < 10 km/h)
- Metered sub-total: 22,000 VND
- Minimum fare check: 22,000 < 25,000 → Minimum Fare applies
- Rider pays: 25,000 + 2,000 (booking fee) = **27,000 VND**
- Driver earns (Bronze, 80%): 80% of 25,000 = 20,000 VND ✓ (meets Minimum Earning Guarantee)
- Platform earns: 5,000 (commission) + 2,000 (booking fee) = 7,000 VND

**Example B — Premium car, airport trip, peak hour:**
- Vehicle: Premium. Distance: 18 km. Duration: 35 minutes. 10 minutes in traffic. Morning peak (07:30). No surge active.
- Base fare: 15,000 VND
- Distance: 18 × 5,500 = 99,000 VND
- Time: 10 × 550 = 5,500 VND
- Metered sub-total: 119,500 VND
- Peak surcharge: ×1.10 → 119,500 × 1.10 = 131,450 VND
- Airport fee: 10,000 VND × 1.10 = 11,000 VND (airport fee also multiplied by peak? **No.** Airport fee is fixed, not a metered component. Airport fee = 10,000 VND flat)
- Total metered + airport: 131,450 + 10,000 = 141,450 VND
- Booking fee: 2,000 VND
- Rider pays: 141,450 → rounded to 141,500 + 2,000 = **143,500 VND**
- Driver earns (Silver, 82%): 82% of 141,450 = 116,000 VND (rounded up)
- Platform earns: 25,461 + 2,000 = **27,461 VND**

**Example C — XL, New Year's Eve, night, heavy traffic, surge ×1.6:**
- Vehicle: XL. Distance: 12 km. Duration: 45 minutes. 25 minutes in heavy traffic. Holiday (New Year's Eve). Night (23:00). Rain active. Surge ×1.6.
- Base fare: 18,000 VND
- Distance: 12 × 5,000 = 60,000 VND (20 minutes of movement above 10 km/h estimated)
- Time: 25 × 500 = 12,500 VND
- Metered sub-total before surcharges: 90,500 VND
- Apply surge ×1.6: 90,500 × 1.6 = 144,800 VND
- Apply Night ×1.20: 144,800 × 1.20 = 173,760 VND
- Apply Holiday ×1.15: 173,760 × 1.15 = 199,824 VND
- Apply Rain ×1.15: 199,824 × 1.15 = 229,798 VND
- Combined static multiplier (Night × Holiday × Rain): 1.20 × 1.15 × 1.15 = 1.587 → within cap of ×1.60 ✓
- Add booking fee: 229,798 + 2,000 = 231,798 VND
- Round: 232,000 VND ← **rider pays**
- Driver earns (Gold, 84%): 84% of 229,798 = 193,030 VND (rounded up)
- Platform earns: 36,768 + 2,000 = **38,768 VND**
- Rider sees: "Holiday Night Rain Surge — your trip is priced with active multipliers. Total: 232,000 VND" with full breakdown visible before confirmation.

## 2.18 Multi-City and Multi-Country Fare Rules

**Multi-city.** Each city has its own rate table for base fare, distance fare, time fare, minimum fare, and airport fee. Surge parameters (DSR thresholds) may also vary by city. The booking fee is the same nationwide.

**Multi-country.** Each country has its own rate table, its own currency, and its own regulatory compliance requirements. Country-level fare rules are set by the local operations team and approved by the CPO. They cannot contradict the core principles in this document.

**Roaming trips** (origin in City A, destination in City B): Rate applied is City A's rate table. The driver completes the trip under City A rules. This is the simplest rule and avoids disputes about which city's rates apply mid-trip.

---

# PART 3 — PROMOTION ENGINE

## 3.1 Philosophy of Promotions

Promotions are investments. Every promotion has a budget, a target outcome, and a measurement methodology. Promotions without measurable ROI are not approved.

FAIRRIDE distinguishes between:
- **Rider acquisition promotions:** Attract new riders. Typically first-ride discounts or referral bonuses. High cost per acquisition (CPA) accepted because the lifetime value (LTV) of a rider justifies it.
- **Rider retention promotions:** Keep existing riders active. Loyalty rewards, birthday bonuses, streak discounts.
- **Demand generation promotions:** Stimulate rides during slow periods. Golden Hour, weekend campaigns, rain campaigns.
- **Sponsored promotions:** Funded by corporate partners or merchants. FAIRRIDE distributes them, but the cost is borne by the sponsor.

## 3.2 Promotion Types

### 3.2.1 First Ride Promotion

**Target:** New riders who have never completed a trip on FAIRRIDE.

**Eligibility:** Account created within the last 30 days AND zero completed trips. Phone number verification is required. One promotion per phone number, per device ID.

**Discount:** 50% off the total fare (metered + surcharges + booking fee), maximum discount of 30,000 VND.

**Driver impact:** Driver is paid their full commission on the unDiscounted fare. The discount is funded entirely by FAIRRIDE.

**Fraud protection:** If a first-ride promotion is used on a trip that is later flagged as fraudulent (same driver, GPS spoofed, very short trip), the promotion funds are clawed back and the account is suspended.

### 3.2.2 Birthday Promotion

**Target:** Existing rider on their birthday.

**Eligibility:** Rider has completed at least 3 trips in the past 90 days. Birthday must be verified from account profile (not just self-declared on the day).

**Discount:** 40% off, maximum 20,000 VND discount. Valid on birthday +/- 1 day (3 days total).

**Limit:** One use per birthday period.

### 3.2.3 Golden Hour Promotion

**Target:** All riders. Designed to stimulate demand during slow hours.

**Active window:** Defined per campaign. Example: 10:00–12:00 on weekdays.

**Discount:** 20% off, maximum 15,000 VND. No minimum fare required.

**Limit:** Each rider can use once per Golden Hour window per day.

**Budget:** Campaign budget defined upfront. When budget is exhausted, the promotion ends automatically.

### 3.2.4 Weekend Promotion

**Active window:** Saturday and Sunday.

**Discount:** 15% off metered fare, maximum 20,000 VND.

**Limit:** Two rides per rider per weekend.

### 3.2.5 Rain Campaign

**Trigger:** Automatically activated when Rain Surcharge is active in a zone (see 2.2.13). The Rain Campaign partially offsets the Rain Surcharge for riders.

**Discount:** 10% off, up to 10,000 VND. Applied to riders who have been active on the platform for at least 7 days.

**Rationale.** The Rain Surcharge compensates drivers. The Rain Campaign softens the impact on loyal riders. New riders (< 7 days) do not receive it because they may be using rain as an exploit to test promotions.

### 3.2.6 Festival Promotion

**Trigger:** Manually activated by Operations for major festivals (Tết, National Day, Christmas, etc.).

**Discount:** Custom per festival. Example: Tết → 25% off for 3 days, maximum 30,000 VND.

**Budget:** Pre-approved by CFO before activation. Budget exhaustion ends the campaign automatically.

### 3.2.7 Referral Programme

**Mechanics:**
- Rider A invites Rider B with a unique referral code
- Rider B completes their first ride using the code
- Rider A receives 20,000 VND wallet credit
- Rider B receives 30,000 VND off their first ride (separate from or stacking with First Ride Promotion, resolved by Campaign Priority rules in 3.4)

**Fraud rules:**
- Rider A and Rider B must have different phone numbers, different device IDs, and different payment methods
- Rider A cannot refer a rider who has the same home WiFi IP as them consistently (soft flag)
- Rider A's referral earnings are held for 7 days after Rider B's first trip. If Rider B's account is suspended within 7 days, the referral reward is not paid
- Maximum 50 referral rewards per rider account, per year

### 3.2.8 Cashback Promotion

**Mechanics:** Rider receives a percentage of the trip fare back as wallet credit after trip completion.

**Common use:** Post-trip retention. Example: "Earn 10% cashback on your next 5 rides this week."

**Rule:** Cashback is credited 24 hours after trip completion (not instantly). This delay allows for trip disputes and cancellations to be resolved before credit is issued.

**Maximum cashback:** 15,000 VND per trip.

### 3.2.9 Coupon Campaign

**Definition.** A specific-value or percentage-off coupon issued to a defined segment of riders.

**Difference from voucher:** A coupon is campaign-based (many riders receive it simultaneously via campaign targeting). A voucher is individual (issued to a specific rider). The rules around usage, transfer, and fraud are different (see Part 4 for Vouchers).

**Coupon rules:**
- Cannot be transferred between accounts
- Expires on the campaign end date
- One coupon use per trip

## 3.3 Campaign Budget Rules

**Rule 1.** Every promotion campaign must have an approved budget before activation. No budget = no campaign.

**Rule 2.** Budget is the total platform cost of the promotion. For rider discounts funded by FAIRRIDE, this is the sum of all discounts paid out.

**Rule 3.** When campaign budget is 90% consumed, an automatic alert goes to the CPO and the relevant Operations manager. At 100%, the campaign pauses automatically.

**Rule 4.** Mid-campaign budget increases require CPO + CFO approval.

**Rule 5.** Campaign budget is tracked in real time. No campaign can produce spend beyond its approved budget.

## 3.4 Campaign Priority Rules

When a rider is eligible for multiple promotions on the same trip, Campaign Priority determines which promotion (or combination) is applied.

**Priority order (highest to lowest):**
1. Vouchers (individual, highest priority — see Part 4)
2. Referral Programme (if applicable — the rider's first trip)
3. First Ride Promotion (if applicable)
4. Festival Promotion (if active)
5. Golden Hour / Rain / Weekend Campaign (highest discount value wins if multiple active)
6. Birthday Promotion
7. Cashback (applied post-trip, does not conflict with upfront discounts)

**Stacking rule:** By default, promotions do **not** stack. The highest-priority eligible promotion is applied. The sole exception is Cashback, which always applies post-trip regardless of what upfront discount was used.

**Override:** The CPO can explicitly configure stacking for specific campaign pairs. This must be documented at campaign creation.

## 3.5 Campaign Conflict Resolution

**Rule.** If two promotions of equal priority are active simultaneously, the one with the **higher monetary value** to the rider is applied. If equal value, the promotion with the earlier expiry date is applied (encouraging the rider to use their "scarcer" credit).

## 3.6 Campaign Expiration

- All campaigns have an explicit end date. No campaign is "open-ended."
- Campaigns created without an end date cannot be activated.
- Expired campaigns auto-archive. Archived campaigns can be cloned but not reactivated.

## 3.7 Campaign Quota

**Rider-level quota:** Maximum uses per rider per campaign (e.g., 2 rides per week in a Weekly Campaign).

**Total quota:** Maximum total rides subsidised in a campaign (linked to budget).

When a rider's quota is exhausted, they see "Offer used up" in the app. When total quota is exhausted, the campaign ends.

## 3.8 Campaign ROI Measurement

**ROI metrics tracked for every campaign:**
- Total rides attributed to the campaign (incremental vs baseline)
- Total spend (discounts paid out)
- Cost per incremental ride (CPIR)
- Rider retention rate 30 days post-campaign vs control group
- Average ride frequency change for participating riders

**ROI threshold for renewal:** A campaign must demonstrate at least one of:
- CPIR below the approved acquisition cost target, OR
- Measurable uplift in 30-day rider retention ≥ 10%, OR
- Demonstrated brand/marketing value approved by CPO

---

# PART 4 — VOUCHER ENGINE

## 4.1 Voucher Definition

A voucher is a unique, individually-issued discount code or credit assigned to a specific rider account. Unlike campaign promotions (which are rule-based and segment-wide), a voucher is a discrete asset with its own lifecycle.

## 4.2 Voucher Lifecycle

**States:** Created → Issued → Active → Redeemed → Expired / Cancelled

- **Created:** Voucher is generated in the system. Not yet assigned to a rider.
- **Issued:** Voucher is assigned to a specific rider account. Rider can see it.
- **Active:** Within validity period, not yet used.
- **Redeemed:** Used on a completed trip.
- **Expired:** Validity period passed without redemption.
- **Cancelled:** Admin has manually cancelled the voucher (e.g., fraud detected).

## 4.3 Voucher Ownership

A voucher is owned by the rider account it was issued to. Ownership transfer is prohibited unless the voucher type explicitly allows it (see Transferability below).

## 4.4 Transferability

**Default:** Vouchers are NOT transferable.

**Exception:** Corporate vouchers may be issued to a corporate account and distributed by the company admin to employee accounts. This internal distribution is not considered transfer — it is bulk issuance to defined recipients.

**Why non-transferable by default?** Transferability enables voucher marketplaces and grey-market fraud. A voucher sold between strangers cannot be verified as legitimately acquired.

## 4.5 Expiration

All vouchers have an explicit expiry date. No voucher is permanent.

**Rule.** A voucher cannot be redeemed after 23:59:59 on its expiry date. If a trip is in progress at expiry, and the voucher was applied before the trip started, the discount is still applied. Expiry prevents new application, not mid-trip revocation.

**Default expiry:** 30 days from issuance.

**Compensation vouchers** (issued after a bad experience) expire in 90 days. They are given more time because the rider may be less active during the period following a negative experience.

## 4.6 Quota

**Per-rider usage limit:** Most vouchers have a usage limit of 1 (single-use). Multi-use vouchers must explicitly define their quota (e.g., "Use up to 5 times").

**Total issuance quota:** A voucher campaign has a maximum total number of vouchers that can be issued. When the quota is met, no further issuance occurs.

## 4.7 Stacking

**Rule:** A maximum of ONE voucher can be applied per trip.

**Exception:** A voucher and a post-trip cashback promotion may coexist (cashback is not a voucher).

**Rationale.** Allowing voucher stacking creates uncontrolled discount depth and is the primary vector for organised voucher fraud.

## 4.8 Minimum Fare

**Rule:** Vouchers can define a minimum trip fare below which they cannot be applied.

**Example.** A 30,000 VND voucher with a 40,000 VND minimum fare cannot be applied to a 35,000 VND trip.

**Purpose.** Prevents incentivised very-short trips designed purely to extract voucher value.

## 4.9 Maximum Discount

**Rule.** A voucher discount can never reduce the amount paid by the rider below 0. If a voucher value exceeds the trip fare, the excess is forfeited — not credited to wallet.

**Exception.** Vouchers explicitly marked "Wallet Cashback" return the excess to the rider's wallet (corporate vouchers and compensation vouchers only).

## 4.10 City Limitation

**Rule.** Vouchers can be restricted to a specific city or a set of cities. A rider cannot use a "Ho Chi Minh City" voucher while requesting a ride in Hanoi.

**Default:** If no city restriction is set at issuance, the voucher is valid nationwide.

## 4.11 Driver Limitation

**Rule.** Vouchers may be restricted to specific vehicle classes (Standard only, Premium only, XL only). They are never restricted to specific drivers.

**Rationale.** Restricting to specific drivers would create unfair competitive dynamics and enable collusion (issuing vouchers that only work with a specific driver friend).

## 4.12 Corporate Vouchers

Corporate vouchers are issued to corporate accounts for distribution to employees. Rules:

- Issued in bulk to a corporate account at a defined total value
- Corporate admin distributes to employees via internal portal
- Employee accounts receive individual vouchers with standard lifecycle
- Corporate is billed for total voucher value upfront (pre-paid) or on settlement (post-paid, credit-approved corporates only)
- FAIRRIDE earns no margin on corporate voucher face value; revenue comes from the placement/management fee charged to the corporate account

## 4.13 Refund Behaviour

**Rule.** If a trip is refunded:
- The voucher used on that trip is **reinstated** to the rider's account (returned to Active state) with its original expiry date, unless the refund was triggered by rider fault
- If the refund was due to rider cancellation after driver arrival, the voucher is forfeited

**Rationale.** If the rider had a bad experience that wasn't their fault, they lost their ride AND their voucher. Reinstating the voucher is fair and builds trust.

## 4.14 Cancellation Behaviour

**Rule.** If a rider cancels before the driver arrives, no fare is charged, and the voucher (if selected) is not consumed. The voucher returns to Active state.

**If the rider cancels after the driver arrives** (cancellation fee applies), the voucher is applied to the cancellation fee if the fee is lower than the voucher value. The remainder is not returned to the voucher.

## 4.15 Fraud Protection

**Rule 1.** A voucher account can only be applied by the account it was issued to. System-level verification confirms rider account ID matches voucher owner ID.

**Rule 2.** If more than 3 vouchers are redeemed on the same account within 7 days and each trip is under 2 km, the account is flagged for review.

**Rule 3.** Voucher codes are cryptographically random (minimum 12 characters, alphanumeric). Sequential or predictable voucher codes are not permitted.

**Rule 4.** The same voucher code cannot be issued to multiple accounts. Each issuance creates a unique voucher instance.

---

# PART 5 — WALLET SYSTEM

## 5.1 Wallet Philosophy

The FAIRRIDE wallet is a pre-funded payment method. Riders add money to their wallet, and that balance is used for trip payments. The wallet is not a bank account. It is not interest-bearing. It does not have an IBAN. It is a pre-payment mechanism.

**Trust rule:** Every unit of currency in the rider wallet represents real money that the rider deposited. FAIRRIDE is the custodian of those funds. We must be able to return every wallet balance to its owner at any time upon request. Wallet funds are never used to finance platform operations.

## 5.2 Wallet Ledger

The wallet is a **ledger**, not a running balance. Every change to the wallet is recorded as an immutable transaction. The current balance is the sum of all transactions.

**Why a ledger?** A running-balance approach requires updating a single number and is prone to race conditions and audit failures. A ledger is auditable, immutable, and reconstructable. The balance is always derivable from the ledger.

## 5.3 Transaction Types

| Transaction Type | Direction | Description |
|---|---|---|
| TOP_UP | Credit (+) | Rider adds money via payment method |
| TRIP_PAYMENT | Debit (−) | Trip fare deducted on completion |
| REFUND | Credit (+) | Refund from cancelled or disputed trip |
| PROMOTION_DISCOUNT | Credit (+) | Platform-funded promotion applied |
| VOUCHER_DISCOUNT | Credit (+) | Voucher applied to trip fare |
| CASHBACK | Credit (+) | Post-trip cashback reward |
| REFERRAL_REWARD | Credit (+) | Referral programme reward |
| ADJUSTMENT_CREDIT | Credit (+) | Manual admin adjustment (positive) |
| ADJUSTMENT_DEBIT | Debit (−) | Manual admin adjustment (negative) |
| WITHDRAWAL | Debit (−) | Rider withdraws balance to bank |
| EXPIRED_CREDIT | Debit (−) | Promotion/bonus credit expiry |

## 5.4 Top-Up

**Minimum top-up:** 20,000 VND.

**Maximum top-up:** 2,000,000 VND per transaction.

**Daily top-up limit:** 5,000,000 VND per rider account per day.

**Monthly top-up limit:** 20,000,000 VND per rider account per month.

**Limits rationale.** These limits comply with payment regulations in Vietnam and prevent the wallet from being used as a money-transfer vehicle.

**Top-up methods:** Credit/debit card, domestic bank transfer, cash at partner agents (future). Each method may have its own processing time and fee (passed through to rider at cost).

## 5.5 Withdrawal

**Rule.** Riders can withdraw their wallet balance to their verified bank account.

**Minimum withdrawal:** 50,000 VND.

**Processing time:** T+1 business day.

**Verification:** Bank account must be verified (name on account matches FAIRRIDE account name) before withdrawal is permitted.

**Withdrawal fee:** None. FAIRRIDE absorbs the bank transfer cost.

**Limit:** Maximum 2 withdrawals per rider per month (prevents wallet from being used as a payment router).

## 5.6 Refund

**Rule.** Refunds are credited to the rider's wallet within 24 hours of refund approval.

**Exception.** Refunds originating from a credit/debit card top-up that occurred less than 7 days ago may be returned to the original payment method instead of the wallet, at the rider's choice.

**Refund types:**
- Trip refund (full or partial)
- Cancellation fee reversal (driver fault)
- Overcharge correction

## 5.7 Promotion and Voucher Credits

**Rule.** Promotion and voucher credits are applied directly at the point of trip payment — they reduce the fare charged to the wallet. They are NOT credited to the wallet as a balance and then deducted separately. The ledger records a PROMOTION_DISCOUNT transaction showing the platform funding the discount.

**Why?** Pre-crediting promotions to wallet balance obscures the true source of funds and creates reconciliation complexity. Applying at payment point is cleaner and auditable.

## 5.8 Negative Balance Policy

**Rule.** The rider wallet balance cannot go below zero through normal transactions.

**Exception.** If a trip fare exceeds the wallet balance (due to a race condition or an approved credit arrangement), the deficit is allowed up to a maximum of 10,000 VND. The rider must top up before their next trip is permitted.

**Recovery.** If a rider's wallet is negative, the next top-up is first applied to clear the negative balance. Riders with a negative balance cannot apply vouchers or promotions until the balance is restored.

## 5.9 Frozen Balance

**Definition.** A portion of the wallet that cannot be used for transactions, typically because of a dispute or fraud investigation.

**Rule.** Only Admin can freeze a wallet balance. Frozen amounts are shown to the rider as "Frozen" in the wallet screen. A reason is provided. An estimated resolution timeline is provided.

**Maximum freeze period.** 30 days without a formal investigation outcome. After 30 days, the freeze must either be extended with documented justification or released.

## 5.10 Locked Balance

**Definition.** Balance that is earmarked for a specific in-progress trip. When a rider confirms a booking, the trip fare is locked (pre-authorised) in their wallet.

**Rule.** Locked balance cannot be used for other transactions. Locking happens at booking confirmation. Lock is released either on trip completion (and the amount is debited as TRIP_PAYMENT) or on trip cancellation (and the lock is released, amount returns to available).

## 5.11 Expired Balance

**Definition.** Credits of promotional origin (cashback, referral reward, promotional top-up) may have an expiry date. On expiry, the balance is removed from the wallet.

**Rule.** Promotional credits expire within 12 months of issuance. Rider-deposited funds (TOP_UP) never expire.

**Pre-expiry notification.** Rider receives a push notification 7 days and 1 day before promotional credit expires.

## 5.12 Ledger Immutability

**Rule.** Once a wallet transaction is recorded, it cannot be deleted or modified. Corrections are made through offsetting transactions (e.g., an ADJUSTMENT_CREDIT to reverse an incorrect ADJUSTMENT_DEBIT).

**Every ledger entry records:** transaction ID, rider account ID, type, amount, timestamp (UTC), initiating system or admin ID, reference trip ID (where applicable), pre-transaction balance, post-transaction balance.

## 5.13 Reconciliation

**Daily reconciliation:** Total of all wallet balances must equal the platform's custodial cash holding. Any discrepancy triggers an alert to the CFO within 1 hour.

**Monthly reconciliation:** Full ledger reconciliation against payment processor reports. Any unmatched transaction is investigated within 5 business days.

## 5.14 Accounting Principles

Wallet balances are a **liability** on FAIRRIDE's balance sheet. They represent money owed to riders. Promotional credits that are funded by FAIRRIDE are expensed at the time of issuance. Those funded by sponsors are offset by sponsor receivable.

---

# PART 6 — SETTLEMENT ENGINE

## 6.1 Money Flow

When a rider completes a trip, money flows as follows:

```
Rider pays → Platform collects → Driver earns → FAIRRIDE retains
```

In detail:

1. Rider's wallet (or payment method) is debited for the total fare
2. Platform receives the full amount
3. Platform calculates driver earnings (commission-based)
4. Platform calculates platform revenue (commission retained + booking fee)
5. Driver earnings are credited to the driver's earning account (settled on payout schedule)
6. Platform revenue is recognised in the income ledger

## 6.2 Money Ownership

- **Rider payment:** Owned by the platform from the moment of debit, as a trust liability to drivers and operators
- **Driver earnings:** Owned by the driver from the moment of trip completion. The platform holds it in trust until payout
- **Platform revenue:** Owned by FAIRRIDE from the moment of trip completion
- **Promotion discount funded by FAIRRIDE:** FAIRRIDE expense at promotion application
- **Promotion discount funded by sponsor:** Sponsor liability until settlement

## 6.3 Platform Revenue Composition

For each completed trip:

| Revenue Item | Formula |
|---|---|
| Commission | (Metered fare) × Commission Rate |
| Booking fee | Fixed (e.g., 2,000 VND) |
| Airport fee share | Airport fee × Commission Rate |
| Total platform revenue | Commission + Booking fee + Airport fee share |

## 6.4 Driver Earnings Composition

| Earnings Item | Formula |
|---|---|
| Metered fare share | (Metered fare) × (1 − Commission Rate) |
| Airport fee share | Airport fee × (1 − Commission Rate) |
| Toll fee | 100% of toll declared |
| Waiting fee share | (Waiting fee) × (1 − Commission Rate) |
| Driver incentive bonuses | Added on top (see Part 8) |

## 6.5 Voucher Settlement

When a rider uses a voucher:

**Case A — FAIRRIDE-funded voucher:**
- Rider pays: (Fare − Voucher discount)
- Driver earns: Commission on full (unDiscounted) fare
- FAIRRIDE absorbs the voucher cost
- Net platform revenue is reduced by the voucher amount

**Case B — Sponsor-funded voucher:**
- Rider pays: (Fare − Voucher discount)
- Driver earns: Commission on full fare
- FAIRRIDE receives sponsor payment for the discount amount (settled monthly with sponsor)
- Net platform revenue is not reduced; revenue includes sponsor reimbursement

**Rule.** In both cases, the driver always earns based on the unDiscounted fare. Drivers are never penalised for platform-chosen promotions.

## 6.6 Promotion Settlement

Same logic as vouchers. Driver always earns on full unDiscounted fare. Platform absorbs or recovers from sponsor.

## 6.7 Refund Settlement

**Driver-fault refund:** Driver earnings are clawed back. If the earnings were already paid to the driver, the amount is deducted from future earnings. Platform revenue from the trip is reversed.

**Platform-fault refund:** Driver keeps earnings. Platform absorbs the refund cost (this is an operational loss, tracked separately).

**Rider-fault refund:** Generally not applicable — if the fault is the rider's, a refund is not granted.

## 6.8 Chargeback

A chargeback occurs when a rider disputes a payment with their bank and the bank reverses the charge.

**Rule:**
1. FAIRRIDE initiates an internal investigation
2. If the chargeback is fraudulent (rider did receive the service), FAIRRIDE provides evidence to the bank and fights the chargeback
3. If the chargeback is valid (service not rendered), the driver's earnings for the trip are forfeited and the rider's account is flagged
4. Repeated chargebacks on a single rider account lead to account suspension and payment method block

**Driver protection.** If a chargeback is the result of the rider's bad faith (fraudulent dispute of a legitimate trip), the driver is made whole from the platform. FAIRRIDE does not pass chargeback risk to drivers.

## 6.9 Driver Payout Schedule

**Standard payout:** Weekly. Every Monday, drivers are paid their earnings from the previous Monday–Sunday.

**Fast payout (Premium feature):** Daily, available to Gold tier and above drivers for a fee (detailed in Part 7). Processed by 10:00 AM the following day.

**Minimum payout:** 50,000 VND. Earnings below this threshold are carried to the next payout cycle.

## 6.10 Tax Handling

**Driver taxes:** Drivers are classified as independent contractors. FAIRRIDE provides earnings reports but does not withhold income tax. Each driver is responsible for their own tax compliance.

**Platform taxes:** FAIRRIDE is responsible for corporate tax, VAT on platform services, and any applicable digital service taxes. These are calculated and remitted by the Finance team independently of this document.

**Invoice generation:** Riders receive an e-invoice for each completed trip (per regulations). Drivers receive monthly earnings summaries.

## 6.11 Settlement Example

**Trip:** Standard car, 10 km, 20 minutes, no surcharge. Rider used a 20,000 VND FAIRRIDE-funded voucher. Driver is Gold tier (16% commission).

- Metered fare: 10,000 (base) + 40,000 (distance) + 8,000 (time) = 58,000 VND
- Booking fee: 2,000 VND
- Rider pays: 58,000 − 20,000 (voucher) + 2,000 = **40,000 VND**
- Driver earns: 58,000 × (1 − 0.16) = 48,720 VND
- Platform earns: 58,000 × 0.16 = 9,280 VND (commission) + 2,000 (booking fee) = **11,280 VND**
- Platform voucher cost: **20,000 VND**
- Platform net: 11,280 − 20,000 = **−8,720 VND** (net loss on this trip — acceptable if the rider is a retained user with high LTV)

---

# PART 7 — DRIVER ECONOMY

## 7.1 Commission Structure

FAIRRIDE's driver tier system rewards seniority and performance with lower commission rates. A driver who demonstrates consistent quality and volume earns a progressively better deal.

| Tier | Commission Rate | Driver Earns |
|---|---|---|
| New / Bronze | 20% | 80% |
| Silver | 18% | 82% |
| Gold | 16% | 84% |
| Platinum | 14% | 86% |
| Diamond | 12% | 88% |

Commission rate is applied to the metered fare (Base + Distance + Time + surcharges, excluding Booking Fee and Toll).

## 7.2 Driver Tier Requirements

### Bronze (Entry Level)
- New drivers start here automatically
- No minimum requirements
- Maintained as long as the driver is active

### Silver
Requirements (evaluated over the trailing 30 days):
- Completed trips: ≥ 100
- Acceptance Rate: ≥ 80%
- Completion Rate: ≥ 95%
- Average Rating: ≥ 4.5
- No active violations

### Gold
Requirements (trailing 90 days):
- Completed trips: ≥ 500
- Acceptance Rate: ≥ 85%
- Completion Rate: ≥ 97%
- Average Rating: ≥ 4.7
- No active violations
- Online time per week: ≥ 30 hours average

### Platinum
Requirements (trailing 90 days):
- Completed trips: ≥ 1,500
- Acceptance Rate: ≥ 88%
- Completion Rate: ≥ 98%
- Average Rating: ≥ 4.8
- Zero serious violations in 180 days
- Online time per week: ≥ 35 hours average

### Diamond
Requirements (trailing 180 days):
- Completed trips: ≥ 4,000 lifetime AND ≥ 1,800 in the last 180 days
- Acceptance Rate: ≥ 90%
- Completion Rate: ≥ 99%
- Average Rating: ≥ 4.85
- Zero violations in 360 days
- Sustained top 5% driver in their city

## 7.3 Tier Benefits Summary

| Benefit | Bronze | Silver | Gold | Platinum | Diamond |
|---|---|---|---|---|---|
| Commission rate | 20% | 18% | 16% | 14% | 12% |
| Fast withdrawal | No | No | Yes (fee applies) | Yes (reduced fee) | Yes (free) |
| Priority dispatch | No | No | Yes | Yes | Yes (highest) |
| Airport queue priority | No | No | Yes | Yes | Yes |
| Dedicated support | No | No | Yes | Yes | Yes (24/7) |
| Badge in app | Bronze | Silver | Gold | Platinum | Diamond |
| Monthly earnings bonus | — | — | +2% on monthly | +3% | +5% |

## 7.4 Tier Evaluation

**Frequency:** Tiers are re-evaluated on the 1st of every month.

**Upgrade:** If a driver meets the requirements for a higher tier, they are automatically upgraded.

**Downgrade:** If a driver falls below their current tier's requirements, they receive a warning month before downgrade. If they do not recover in the following month, they are downgraded one tier only (not straight to Bronze from Platinum).

**Grace period.** A driver who becomes inactive due to verified illness or emergency can request a 60-day freeze on tier evaluation. This requires documentation reviewed by Ops.

## 7.5 Priority Dispatch

Gold tier and above drivers receive trip offers before Bronze and Silver drivers for the same pickup location, under identical conditions.

**Rule.** Priority dispatch bias is a maximum of 15 seconds. If no Gold+ driver accepts within 15 seconds, the offer is extended to all drivers. This prevents rider wait times from becoming unreasonably long due to tier ordering.

**Airport queue.** At airport waiting zones, Gold+ drivers are positioned higher in the queue than lower-tier drivers who arrived at the same time.

## 7.6 Badge System

Each tier displays a visible badge on the driver's public profile in the rider app. Badges show:
- Current tier
- Lifetime completed trips
- Rating

Badges are a trust signal to riders and a recognition mechanism for drivers.

---

# PART 8 — DRIVER INCENTIVE ENGINE

## 8.1 Incentive Philosophy

Incentives are designed to achieve one or more of:
1. Increase driver supply during high-demand periods
2. Improve driver quality metrics (acceptance rate, completion rate, rating)
3. Retain high-performing drivers from churning to competitors
4. Create predictable additional income that drivers can plan around

Incentives are never designed to:
- Extract maximum working hours from drivers regardless of their wellbeing
- Create income that is only achievable through physically unsustainable effort
- Incentivise behaviours that harm rider experience (e.g., accepting trips just to cancel them)

## 8.2 Daily Quest

**Mechanics.** A set of 2–3 goals a driver must complete within a calendar day (00:00–23:59) to earn a bonus.

**Example Quest:** "Complete 8 trips today → earn 30,000 VND bonus."

**Rule.** Daily Quests are pushed to drivers by 07:00 each day. A driver must opt in (tap "Accept Quest") by 10:00 or the quest is considered skipped. Quests cannot be retroactively accepted.

**Anti-abuse.** Only completed, non-cancelled trips by real riders (verified by anti-fraud checks) count toward quests.

## 8.3 Weekly Mission

**Mechanics.** A set of goals across the full week (Monday–Sunday) for a larger bonus.

**Example Mission:** "Complete 50 trips this week → earn 150,000 VND bonus."

**Tier multiplier.** Diamond drivers receive a 1.3x multiplier on Weekly Mission bonuses.

## 8.4 Monthly Incentive

**Mechanics.** Targets for the full calendar month, with tiered rewards.

**Example:**
- 150 trips → 200,000 VND
- 200 trips → 350,000 VND
- 250 trips → 500,000 VND + Tier bonus (if Silver+)

## 8.5 Streak Bonus

**Mechanics.** Completing trips on consecutive days earns increasing bonuses.

| Consecutive days | Daily bonus |
|---|---|
| 3 days | +5,000 VND/day |
| 5 days | +10,000 VND/day |
| 7 days | +20,000 VND/day |
| 10 days | +30,000 VND/day |

**Streak reset rule.** A streak resets if the driver completes fewer than 3 trips in any calendar day and is marked "online" for less than 2 hours that day. A driver who was "offline" that day (not online at all) does not lose their streak — rest days are allowed.

## 8.6 Peak Hour Bonus

**Mechanics.** Completing trips during designated peak hours earns a per-trip cash bonus on top of the normal fare commission.

**Peak hours:** 07:00–09:00 and 17:00–20:00 on weekdays.

**Bonus:** 5,000 VND per trip completed (start and end both within peak hours window).

**Cap:** Maximum 4 peak bonuses per peak window (8 per day).

**Rationale cap.** Without a cap, drivers are incentivised to take very short trips during peak hours to maximise per-trip bonuses rather than providing rides to riders who need longer trips.

## 8.7 Airport Bonus

**Mechanics.** Completing a trip to or from the airport earns an additional bonus.

**Bonus:** 15,000 VND per airport trip (pickup or dropoff inside the airport zone).

**Cap:** Maximum 3 airport bonuses per day.

## 8.8 Rain Bonus

**Mechanics.** When Rain Surcharge is active, drivers earn an additional per-trip cash bonus for completing trips during the rain period.

**Bonus:** 8,000 VND per trip.

**Why in addition to surge?** Surge increases the fare, which the driver earns a percentage of. The rain bonus is an additional fixed recognition of driving in difficult conditions. It is funded by FAIRRIDE, not recovered from the rider.

## 8.9 Guaranteed Income Programme

**Target.** New drivers in their first 90 days.

**Mechanics.** If a new driver completes at least 20 trips in a week and earns less than 2,000,000 VND gross that week, FAIRRIDE tops up the difference.

**Rule.** The top-up is a one-time payment per qualifying week. It is capped at 500,000 VND top-up (i.e., if the driver earned 1,200,000 VND, the gap is 800,000 VND but only 500,000 VND is paid — the guaranteed income floor is effectively 1,500,000 VND for this cap scenario).

**Purpose.** Removes the early-period income uncertainty that causes new drivers to quit before they learn the platform.

**Fraud protection.** Trips used for the guarantee count must pass all fraud checks. Collusion trips, fake trips, and GPS-spoofed trips are excluded.

## 8.10 Acceptance Bonus

**Mechanics.** Maintaining an Acceptance Rate ≥ 90% for a rolling 7-day period earns a weekly bonus.

**Bonus:** 50,000 VND per week.

**Why?** High acceptance rates reduce rider wait times. Rewarding drivers for accepting trips more consistently improves rider experience.

## 8.11 Completion Bonus

**Mechanics.** Completing ≥ 98% of accepted trips (not cancelling after accepting) earns a weekly bonus.

**Bonus:** 30,000 VND per week.

## 8.12 Rating Bonus

**Mechanics.** Maintaining an average rating ≥ 4.9 for a trailing 30-day period earns a monthly recognition bonus.

**Bonus:** 100,000 VND per month.

**Minimum base.** Driver must have completed at least 50 rated trips in the 30-day period for the bonus to apply.

## 8.13 Referral Bonus (Driver Referral)

**Mechanics.** A driver who refers another driver earns a bonus when the referred driver completes their first 50 trips.

**Bonus:** 100,000 VND paid to the referring driver when the referred driver's 50th trip is completed.

**Fraud rules:** Referring driver and referred driver cannot share a device ID, residential address, or consistent GPS pattern. Drivers cannot refer themselves through alternate accounts.

## 8.13B Incentive Payout Timing

**Rule.** All driver incentive bonuses (Daily Quest, Weekly Mission, Streak, Peak, Airport, Rain, Rating, Referral) are calculated and added to the driver's earnings balance by 06:00 the following day. They appear as a distinct line item on the earnings breakdown labelled with the incentive type and the qualifying period.

**Dispute window.** A driver who believes an incentive was miscalculated has 7 days to raise a dispute through the in-app earnings support channel. Operations must respond within 48 hours.

**Correction rule.** If an incentive was underpaid due to a system error, the difference is credited in the next settlement cycle. If an incentive was overpaid, the overpayment is deducted from future earnings in the next cycle (not clawed back from an already-processed payout). Underpayment corrections are never waived; overpayment recovery is subject to a maximum monthly deduction of 10% of that cycle's earnings to prevent hardship.

## 8.14 Anti-Abuse Rules for Incentives

**Rule 1 — Minimum trip quality.** A trip counts toward any incentive only if: the trip distance is ≥ 1 km, the trip duration is ≥ 3 minutes, and the trip passes fraud detection (no GPS spoofing flag, real rider, real movement).

**Rule 2 — Account integrity.** Incentives are credited only to accounts in Good Standing (no active suspension, no pending fraud investigation).

**Rule 3 — No incentive stacking on minimum fare trips.** Trips that trigger the Minimum Fare Guarantee do not also count toward quest/mission completion unless the trip distance is ≥ 2 km.

**Rule 4 — Incentive clawback.** If a trip that was used to qualify for an incentive is later found to be fraudulent, the incentive payment is clawed back from future earnings. If the driver has already withdrawn the amount, the clawback is treated as a debt recorded against the account.

---

# PART 9 — DRIVER PERFORMANCE

## 9.1 Acceptance Rate

**Definition.** The percentage of trip requests shown to the driver that the driver accepts.

**Formula.** (Trips accepted) ÷ (Trip requests received) × 100, trailing 30 days.

**What counts as "received":** Any trip offer that was displayed to the driver's app for at least 5 seconds.

**What does NOT reduce the rate:**
- Trip requests received when the driver was in a tunnel or low-signal area (automatically excluded by GPS signal quality flag)
- Trip requests received within 60 seconds of the driver completing a previous trip (driver needs time to prepare)

**Update frequency:** Recalculated in real time after each accepted/declined event. The displayed metric is updated hourly for display.

**Impact:** Acceptance Rate is the primary input for Priority Dispatch eligibility and tier retention.

## 9.2 Completion Rate

**Definition.** The percentage of accepted trips that are completed (not cancelled after acceptance).

**Formula.** (Trips completed) ÷ (Trips accepted) × 100, trailing 30 days.

**What counts as a cancellation:** Driver cancels after accepting. Driver does not arrive within the maximum arrival time. Driver's app loses connectivity and the trip auto-cancels while the driver was en route (disputed — reviewed individually).

**Driver-initiated cancellation without penalty:** A driver may cancel without affecting Completion Rate if:
- The rider requests a destination change to a city outside the driver's available area
- The pickup location is physically inaccessible (documented with evidence)
- The driver feels unsafe (safety cancellation — never penalised)

## 9.3 Cancellation Rate

**Definition.** The percentage of all trips accepted by the driver that the driver cancelled.

**Formula.** (Driver-initiated cancellations) ÷ (Trips accepted) × 100, trailing 30 days.

**Note.** Cancellation Rate is the complement of Completion Rate but measured separately to distinguish driver-initiated from other cancellation causes.

**Thresholds:**

| Cancellation Rate | Consequence |
|---|---|
| < 3% | No action |
| 3–5% | Warning notification |
| 5–8% | Temporary dispatch priority reduction |
| > 8% | Account review, potential suspension |

## 9.4 Late Arrival

**Definition.** The frequency with which the driver arrives at the pickup location more than 5 minutes beyond the estimated arrival time (ETA) shown to the rider at booking.

**Update frequency:** Tracked per trip; average computed over trailing 30 days.

**Impact:** High late arrival rates reduce the driver's Trust Score (see 9.9).

## 9.5 Online Time

**Definition.** Total hours per week the driver's app is in "Online" mode (available for trips).

**Measurement.** GPS-verified. Online time does not count during periods where GPS signal is absent for more than 5 consecutive minutes.

**Impact:** Input to tier requirements and some incentive programmes.

## 9.6 Trip Count

**Definition.** Total completed trips, tracked as both lifetime and trailing-period metrics (30-day, 90-day, 180-day).

**Impact:** Primary input for tier upgrades.

## 9.7 Rating

**Definition.** Average rider rating of the driver, on a 1–5 star scale.

**Calculation.** Simple average of all ratings received in the trailing 90 days. Minimum 10 ratings required for the metric to be displayed.

**Rating expiry.** Ratings older than 90 days do not count in the average. This allows drivers to recover from a bad period.

**Rating fraud.** If a rider rates a driver 1 star on more than 3 trips in 30 days without providing a written reason, those ratings are excluded from the average and the rider account is flagged for review.

**Impact:** Ratings are a tier requirement and affect Trust Score.

## 9.8 Complaint Score

**Definition.** A weighted count of rider complaints about the driver in the trailing 90 days.

**Complaint weights:**

| Complaint Type | Weight |
|---|---|
| Rude behaviour | 3 |
| Unsafe driving | 5 |
| Wrong route | 2 |
| Dirty vehicle | 1 |
| Excessive phone use | 2 |
| Overcharge attempt | 4 |

**Formula.** Sum of (complaint count × weight) / 100 trips completed = Complaint Score per 100 trips.

**Thresholds:**

| Score | Action |
|---|---|
| 0–0.5 | No action |
| 0.5–1.0 | Warning |
| 1.0–2.0 | Mandatory training |
| > 2.0 | Account review |

## 9.9 Driver Trust Score

**Definition.** A composite score from 0–1000 that summarises the driver's overall platform standing.

**Components:**

| Component | Maximum points |
|---|---|
| Acceptance Rate (≥90% = 200 pts, scaled) | 200 |
| Completion Rate (≥99% = 200 pts) | 200 |
| Average Rating (5.0 = 200 pts) | 200 |
| Trip Volume (scaled vs peer group) | 150 |
| Complaint Score (zero complaints = 150) | 150 |
| Violations (zero = 100 pts) | 100 |
| **Total** | **1,000** |

**Update frequency:** Recalculated weekly.

**Use:** Trust Score determines dispatch priority in ambiguous situations, eligibility for special programmes (corporate trips, premium events), and is visible to the driver in their dashboard.

---

# PART 10 — RIDER RULES

## 10.1 Cancellation Policy

**Free cancellation window:** Riders may cancel any trip for free within 2 minutes of booking confirmation.

**After free window:** If the rider cancels after 2 minutes from booking AND before the driver has arrived:
- Cancellation fee: 10,000 VND (charged to rider wallet or payment method)
- Driver receives 80% of the cancellation fee (8,000 VND)
- Platform retains 20% (2,000 VND)

**After driver arrival:** If the rider cancels after the driver marks "Arrived":
- Cancellation fee: 20,000 VND
- Driver receives 80% (16,000 VND)
- Platform retains 20% (4,000 VND)

**Rationale.** Drivers incur real costs (fuel, time) when riders cancel after the driver has committed to the trip. The cancellation fee compensates them and disincentivises casual cancellations.

**Abuse.** Riders who cancel more than 30% of their trips in a 30-day period receive a warning. Riders who maintain a >40% cancellation rate after warning face account restrictions.

## 10.2 Refund Policy

**Automatic refund triggers:**
- Driver did not show up within 10 minutes of "Arrived" mark
- Trip was cancelled by the driver after accepting
- Rider was overcharged due to a verified metering error

**Refund processing:** Credited to wallet within 24 hours.

**Dispute-triggered refund:** Rider submits a dispute within 24 hours of trip completion. Reviewed within 48 hours. Partial or full refund at Operations discretion.

**What is NOT refunded:**
- Rider remorse (trip completed, rider unsatisfied with service for reasons not related to a platform failure)
- Cancellation fee for rider-initiated cancellation outside the free window

## 10.3 Rating System

**Rider rating of driver:** 1–5 stars, submitted within 24 hours of trip completion. After 24 hours, the rating window closes.

**Driver rating of rider:** Drivers may rate riders 1–5 stars. Rider ratings are NOT displayed publicly but are used internally.

**Impact of rider rating on driver:** Low ratings reduce driver's average. See 9.7.

**Rider accountability:** If a rider consistently receives low ratings from drivers (< 3.0 average from driver ratings over 10+ trips), Operations may investigate and restrict the account.

## 10.4 Referral Programme (Rider)

Covered in Part 3 (Promotion Engine, Section 3.2.7).

## 10.5 Rider Loyalty

**Loyalty tiers are not implemented at launch.** Future consideration (see Part 15).

**Planned behaviour.** Riders who complete more trips earn better access to promotions, faster support, and loyalty rewards. Rules to be defined in a future BRB amendment.

## 10.6 Account Suspension Rules

**Rider account is suspended for:**
- Fraudulent payment (chargeback after legitimate trip)
- Voucher fraud
- Abusive or threatening behaviour toward drivers (complaint received and verified)
- GPS manipulation (rare on rider side, but possible via fake location apps)
- Operating multiple accounts for the same person
- Failure to pay cancellation fees (balance negative > 30 days)

**Suspension process:**
1. System flags the account
2. Rider receives notification with reason
3. Rider has 7 days to appeal via in-app form
4. Operations reviews appeal within 3 business days
5. Outcome: reinstate, permanent ban, or restricted access

## 10.7 Abuse Detection

Patterns that trigger automatic review:
- Requesting trips from the same origin/destination repeatedly and cancelling
- Rating every driver 1 star without written reason
- Using multiple devices under one account
- Top-up from payment methods that are flagged by payment processor

---

# PART 11 — FRAUD RULES

## 11.1 GPS Spoofing

**Definition.** A driver or rider uses software to simulate a GPS location that does not match their physical location.

**Detection signals:**
- GPS altitude mismatch (mock location apps often produce sea-level altitude in elevated areas)
- Speed physically impossible for the claimed route (0 to 60 km/h instantaneously)
- Polyline path does not follow road network (straight-line movement through buildings)
- Device sensor mismatch (accelerometer suggests stationary device while GPS shows movement)
- Multiple drivers showing identical GPS coordinates simultaneously

**Consequence (driver):**
- First detected: 3-day suspension + investigation
- Confirmed: Permanent ban. All earnings in the disputed period held pending review.

**Consequence (rider):**
- Permanent ban.

## 11.2 Fake Trips

**Definition.** A driver and a "rider" (who may be the driver's associate) complete a trip that did not actually occur, to harvest incentive bonuses or manipulate metrics.

**Detection signals:**
- Trip origin and destination are home addresses of people known to be related
- Trip GPS trace shows minimal actual movement despite recorded distance
- Payment was made from a wallet topped up by the same bank account as the driver uses
- Driver and rider share device IDs or login from the same IP over time
- No rider rating submitted after trip

**Consequence:** Permanent ban for both driver and rider accounts. Earnings clawed back.

## 11.3 Voucher Abuse

**Definition.** Creating or using multiple accounts to exploit first-use vouchers, referral bonuses, or other account-level promotions.

**Detection signals:**
- Multiple accounts from same device ID
- Multiple accounts with same phone number root (different number but same SIM carrier IMSI)
- Multiple accounts created from the same IP within 24 hours
- Referral chain where all participants share physical addresses

**Consequence:**
- All accounts involved suspended
- Promotion credits clawed back
- Primary abuse account permanently banned

## 11.4 Referral Abuse

**Definition.** Creating fake accounts or recruiting strangers without intent to use the platform, purely to harvest referral bonuses.

**Detection signals:**
- Referred riders complete exactly 1 trip (the minimum to trigger referral bonus) and never ride again
- Referred riders' single trip is to/from the referring rider's address
- Unusually high referral counts from a single account (> 10 in one week)

**Consequence:** Referral rewards withheld, investigation opened, potential account restriction.

## 11.5 Collusion

**Definition.** Driver and rider work together to extract platform value — through incentive gaming, inflated fares, or voucher exploitation.

**Detection.** Combination of GPS spoofing signals, payment relationship patterns, and repeated pairings (same driver-rider pair on more than 30% of a driver's trips).

**Consequence:** Both accounts permanently banned.

## 11.6 Multi-Account Operation

**Rule.** One person may only operate one rider account and one driver account. Operating multiple accounts of the same type is prohibited.

**Detection.** Device ID matching, phone number matching, biometric matching (future), payment method matching.

**Consequence.** All duplicate accounts suspended. Primary account reviewed.

## 11.7 Device Farming

**Definition.** Using multiple devices, emulators, or virtual machines to simulate multiple user accounts — typically for voucher or referral exploitation.

**Detection.** Multiple accounts linked to the same device hardware ID (IMEI/IMSI/IDFA) within a short window, or accounts tied to emulators (emulator-specific device flags).

**Consequence.** All accounts on the device network permanently banned.

## 11.8 Repeated Cancellations

**Definition.** A driver or rider who repeatedly accepts and then cancels trips, either to manipulate dispatch or to avoid unfavourable routes.

**Covered in:**
- Driver Completion Rate (Part 9.3)
- Rider Cancellation Policy (Part 10.1)

## 11.9 Payment Fraud

**Definition.** Using stolen, cloned, or fraudulent payment methods to top up the rider wallet.

**Detection.** Payment processor fraud signals, chargeback patterns, card BIN analysis.

**Consequence.** Wallet frozen immediately. Account suspended. Funds held pending investigation. Law enforcement notified if evidence warrants.

## 11.10 Automatic Actions

The fraud system takes the following actions automatically without human review:

| Event | Automatic Action |
|---|---|
| GPS spoofing flag (high confidence) | Trip invalidated, account flagged |
| 3+ chargebacks in 30 days | Payment method blocked |
| 5+ voucher uses in 7 days, all < 2 km | Account flagged, vouchers frozen |
| Multi-account device detected | All accounts suspended pending review |
| Referral count > 10 in 7 days | Referral rewards frozen |

## 11.10B Fraud Response Escalation Matrix

| Severity | Trigger | Automatic Action | Human Review SLA |
|---|---|---|---|
| Low | Single anomaly, no pattern | Log and monitor | 72 hours |
| Medium | Pattern detected (2+ signals) | Account flag, enhanced monitoring | 48 hours |
| High | Confirmed fraud signal or multiple medium flags | Trip invalidation, payment hold | 24 hours |
| Critical | Confirmed GPS spoof, fake trips, payment fraud | Immediate suspension | 4 hours |

**Escalation rule.** A Medium-severity case that is not resolved within 48 hours automatically escalates to High. A High-severity case that is not resolved within 24 hours automatically escalates to Critical. This ensures no fraud case falls through the cracks due to workload.

**Documentation requirement.** Every fraud case, regardless of outcome, must be documented with: trigger event, evidence collected, action taken, resolution, and any pattern noted for future detection improvement. This documentation feeds into the monthly Fraud Intelligence Report reviewed by the COO and CTO.

## 11.11 Manual Review

Cases flagged by the automatic system enter a manual review queue (see Part 12 and 13). A human Operations agent completes the review within 24 hours for high-severity flags and 72 hours for medium severity.

---

# PART 12 — RISK ENGINE

## 12.1 Risk Score

Every driver and rider account has a Risk Score from 0–100. A higher score means higher risk.

**Score composition:**

| Risk Factor | Max Score |
|---|---|
| Fraud flags (confirmed) | 40 |
| Fraud flags (unresolved) | 20 |
| Cancellation pattern | 10 |
| Payment incidents | 15 |
| Complaint pattern | 10 |
| Account age and tenure | −5 (negative = reduces risk) |
| **Total** | **up to 100** |

## 12.2 Risk Thresholds and Actions

| Risk Score | Status | Action |
|---|---|---|
| 0–20 | Low | No action |
| 20–40 | Medium | Enhanced monitoring |
| 40–60 | High | Soft warning issued, some features restricted |
| 60–80 | Very High | Temporary suspension (7 days) |
| 80–100 | Critical | Immediate suspension, investigation required |

## 12.3 Soft Warning

A Soft Warning is a notification to the driver or rider that their account behaviour has been flagged. It explains what was detected and what the consequence will be if the behaviour continues. No restriction is applied at the Soft Warning stage.

**Purpose.** Many flagged patterns are accidental or have innocent explanations. A warning gives the account holder an opportunity to correct the behaviour before penalties apply.

## 12.4 Temporary Suspension

**Duration:** 7 days (standard). 30 days (serious violations).

**What is preserved:** Account, earnings balance, wallet balance.

**What is restricted:** No trips accepted or requested. No voucher use.

## 12.5 Permanent Ban

Triggered by:
- Confirmed GPS spoofing
- Confirmed fake trips
- Confirmed payment fraud
- Failure to resolve after repeated temporary suspensions
- Physical threats or violence toward other platform participants

**Effect:** Account permanently deactivated. Wallet balance returned to the rider (after 30-day dispute period). Earnings balance paid out to the driver (after deducting any fraud-related clawbacks).

## 12.6 Appeal Process

All suspension decisions (temporary and permanent) can be appealed within 30 days.

**Appeal process:**
1. Driver or rider submits appeal via in-app form
2. Operations receives the appeal within 1 hour
3. A senior Operations agent (not the original reviewer) reviews the case
4. Decision delivered within 5 business days for temporary suspensions, 10 business days for permanent bans
5. Decision is final (no further appeals within the platform — external legal recourse available)

---

# PART 13 — ADMIN OPERATIONS

## 13.1 Refund Permissions

| Refund Type | Who Can Approve |
|---|---|
| Automatic refund (platform-fault, < 50,000 VND) | System (automatic) |
| Dispute refund (< 100,000 VND) | Operations Agent |
| Dispute refund (100,000–500,000 VND) | Operations Lead |
| Dispute refund (> 500,000 VND) | COO or CFO |
| Full trip reversal (all amounts) | COO |

All refunds are logged with the approving agent ID, reason, and reference trip ID. No refund can be processed without a reason code.

## 13.2 Voucher Creation Permissions

| Voucher Type | Who Can Create |
|---|---|
| Standard rider voucher (< 50,000 VND) | Marketing Manager |
| Standard rider voucher (50,000–200,000 VND) | Marketing Director |
| Corporate voucher (any value) | Sales Director + Finance |
| Compensation voucher (any value) | Operations Lead |
| High-value voucher (> 200,000 VND) | CPO approval required |

## 13.3 Campaign Approval

| Campaign Budget | Approval Required |
|---|---|
| < 10,000,000 VND | Marketing Manager |
| 10–50M VND | Marketing Director |
| 50–200M VND | CPO + CFO |
| > 200M VND | CEO approval |

## 13.4 Driver Verification

Driver onboarding requires verification of:
- National ID / Passport
- Driver's licence (valid, correct class for vehicle type)
- Vehicle registration
- Vehicle insurance
- Vehicle inspection certificate (annual)
- Face-matching selfie vs. ID photo (automated)

**Re-verification:** Annual for licence and insurance. Immediate if a complaint involves identity verification.

## 13.5 Manual Payout

If a driver's payout fails (bank account error, bank system outage), Operations can initiate a manual payout:
- Requires Finance Manager approval
- Method: direct bank transfer or cash at partner agent
- Documented within 24 hours

## 13.6 Manual Adjustment

Operations may make manual wallet adjustments for:
- Customer goodwill compensation
- Error correction
- Fraud recovery

All manual adjustments are logged with justification and approver ID. Adjustments > 100,000 VND require Lead approval. Adjustments > 500,000 VND require COO approval.

## 13.7 Audit Trail

Every administrative action — refund, suspension, adjustment, voucher creation, campaign activation — is recorded in an immutable audit log that captures:
- Admin user ID
- Timestamp (UTC)
- Action type
- Target account ID
- Amount (if applicable)
- Justification
- Before/after state

The audit log cannot be edited or deleted. It is retained for 7 years.

## 13.8 Customer Support SLA

| Issue Type | Target Resolution |
|---|---|
| Overcharge dispute | 24 hours |
| Refund request | 48 hours |
| Driver complaint | 72 hours |
| Account suspension appeal | 5 business days |
| Fraud investigation outcome | 10 business days |

---

# PART 14 — FINANCIAL REPORTS

## 14.1 Key Metrics Definitions

**Gross Merchandise Value (GMV).** The total value of trip fares billed to riders, before any deductions. GMV includes booking fees, surcharges, and toll fees. GMV is the top-line measurement of platform transaction volume.

**Net Revenue.** GMV minus driver earnings minus toll pass-through. Net Revenue = Platform Commission + Booking Fees. This is FAIRRIDE's economic output from trip operations.

**Take Rate.** Net Revenue ÷ GMV × 100. Expressed as a percentage. The platform's effective cut of every transaction dollar. Take rate includes the effect of promotions (which reduce it) and booking fees (which increase it). Target take rate at launch: 12–15%.

**Driver Earnings.** Total amount paid or owed to drivers for completed trips, including commissions, incentives, and bonuses.

**Promotion Cost.** Total value of FAIRRIDE-funded discounts given to riders or bonuses given to drivers. Measured against campaign budgets and GMV.

**Voucher Cost.** Total value of FAIRRIDE-funded vouchers redeemed. Tracked separately from Promotion Cost for accounting clarity.

**Wallet Liability.** The sum of all rider wallet balances at a point in time. Represents the platform's obligation to return funds to riders.

**Outstanding Settlement.** Driver earnings that have been earned but not yet paid out (between trip completion and payout date).

## 14.2 Daily Report

Produced automatically by 08:00 the following day. Contents:

- Total trips completed (by city, by vehicle class)
- GMV
- Net Revenue
- Average fare per trip
- Promotion Cost
- Voucher Cost
- Active drivers (online at least 1 hour)
- Average driver earnings
- Surge activity (zones, duration, peak multiplier)
- Fraud flags raised and resolved

## 14.3 Weekly Report

Produced by Monday 09:00 for the prior week. Contents:
- All daily metrics aggregated
- Driver tier distribution (how many at each tier)
- New driver activations and churn
- Rider cohort retention (% of riders from prior week who rode again this week)
- Incentive spend vs. target
- Campaign performance summary
- Risk Engine summary (suspensions, bans, appeals)

## 14.4 Monthly Report

Produced by the 5th of the following month. Contents:
- All weekly metrics aggregated
- GMV and Net Revenue trend vs. prior month and prior year
- Take Rate trend
- Wallet Liability snapshot (1st vs. last day of month)
- Outstanding Settlement amount at month end
- Tax liability estimate
- Campaign ROI summary
- Fraud incident summary
- Driver income distribution (P10, P25, P50, P75, P90)

## 14.5 Financial Controls

**Rule.** No single person can both approve a refund and process it in the ledger. The approver and the processor must be different individuals.

**Rule.** Driver payout files are generated automatically by the system; no manual modification is permitted. Manual payout requests (13.5) are a separate exception process.

**Rule.** Wallet liability is reconciled against custodial cash daily. A variance > 0.1% triggers an immediate freeze on new top-ups pending investigation.

---

# PART 15 — FUTURE EXPANSION

This part documents planned expansion directions. These are not current business rules; they are design constraints that current rules must not contradict.

## 15.1 Food Delivery

**Philosophy.** FAIRRIDE delivery uses the same driver network as ride-hailing. A driver can toggle between "Ride" and "Delivery" mode.

**Pricing.** Delivery pricing will use its own fare component table (base fee per order, distance-based, weight-based), separate from ride-hailing. The commission structure will be separate.

**Constraint for current rules:** The wallet system must be designed to handle multiple service categories from day 1. A rider who is also a food customer uses one wallet for both.

## 15.2 Parcel Delivery

**Philosophy.** Peer-to-peer parcel delivery using idle driver capacity.

**Key business rule to define later:** Liability for lost or damaged parcels. This requires an insurance product before launch.

## 15.3 Electric Vehicle (EV) Programme

**Planned rules:**
- EV drivers receive a commission tier bonus of −1% (i.e., the same as the next tier up in commission rate) as an environmental incentive
- EV trips may be labelled for riders who prefer green transport
- EV charging time (at approved stations) is excluded from "Online Time" metrics

**Constraint for current rules:** Commission structure must allow for additive adjustments without restructuring the whole tier table.

## 15.4 Corporate Programme

**Planned rules:**
- Corporate accounts have a credit facility (post-pay) instead of pre-pay wallet
- Trip approval workflows for corporate riders
- Expense reporting integration
- Dedicated corporate support line

**Constraint for current rules:** The voucher and settlement systems must support bulk issuance and corporate billing from day 1 (or near day 1). See Part 4.12.

## 15.5 Subscription Programme

**Planned feature.** Riders pay a monthly subscription for benefits (reduced booking fee, guaranteed surge cap, priority matching).

**Constraint.** Current pricing rules must not hard-code the booking fee as non-negotiable — it must be overridable per account type.

## 15.6 Merchant Platform

**Planned feature.** Merchants (restaurants, retailers) can offer FAIRRIDE credits to their customers as a loyalty benefit.

**Constraint for current rules.** The voucher system must support a "Merchant Sponsor" voucher type where the merchant pre-funds a voucher pool.

## 15.7 Open API

**Planned.** Third-party developers can integrate FAIRRIDE ride requests into their own apps.

**Business rule.** Third-party apps pay a higher booking fee. Riders booked through third-party apps receive the same fare and driver as direct app riders.

## 15.8 Insurance Products

**Planned.** In-trip insurance products sold at the time of booking.

**Constraint.** Pricing engine must support optional add-on line items in the fare breakdown.

## 15.9 International Expansion

**Rules that must hold in every country:**
- Local currency only (no cross-border wallet transfers)
- Local regulatory compliance reviewed before launch
- Core commission structure (tier-based) maintained; specific rates may vary by country economics
- This BRB applies to all countries; country-specific amendments are documented as appendices to this document, never as replacements

## 14.5B Revenue Forecasting Model

FAIRRIDE uses the following simplified forecasting formula to project monthly Net Revenue:

**Net Revenue = (Average trips/day × Active days) × (Average Fare × Take Rate) + (Average trips/day × Active days × Booking Fee)**

**Example forecast for Year 1, Month 6:**
- Active days: 30
- Average trips/day: 3,000
- Average fare: 60,000 VND
- Take rate: 13%
- Booking fee: 2,000 VND

Net Revenue = (3,000 × 30) × (60,000 × 0.13) + (3,000 × 30 × 2,000)
= 90,000 × 7,800 + 90,000 × 2,000
= 702,000,000 + 180,000,000
= **882,000,000 VND (~37,000 USD/month)**

This model is a management tool for direction, not a contractual commitment. Actual results will vary with surge activity, cancellation rates, and promotion costs.

## 14.6 KPI Targets and Review Cadence

| KPI | Monthly Target | Review |
|---|---|---|
| Take Rate | 12–16% | Monthly by CFO |
| Promotion Cost / GMV | < 8% | Monthly by CPO + CFO |
| Driver Churn Rate | < 5%/month | Monthly by COO |
| Rider Retention (D30) | ≥ 40% | Monthly by CPO |
| Fraud Loss / GMV | < 0.5% | Monthly by COO + CTO |
| Customer Support SLA adherence | ≥ 90% | Weekly by COO |
| Wallet Liability / Cash Ratio | 1:1 exact | Daily by CFO |

**Breach response.** Any KPI breaching its target for 2 consecutive months triggers a formal root-cause analysis and recovery plan, presented to the CEO within 15 business days.

---

# APPENDIX A — INTERNAL CONSISTENCY REVIEW

The following contradictions and edge cases were identified and resolved during drafting.

**Issue 1.** Parts 2.12 (Night Surcharge) and 2.13 (Dynamic Pricing) both act as multipliers. **Resolution:** Applied sequentially. Dynamic Surge is applied first to the base metered fare; static surcharges (Night, Holiday, Rain) are applied next. Combined cap of ×1.60 on static surcharges prevents compounding into extreme territory.

**Issue 2.** Part 6 (Settlement) states the driver earns on the full unDiscounted fare; Part 2.14 (Minimum Fare Guarantee) states drivers earn at least 20,000 VND. When a promotion reduces the rider's payment to nearly zero but the driver would normally earn 18,000 VND (below the minimum), does the platform top up? **Resolution:** Yes. The Minimum Fare Guarantee is a driver floor, not a rider floor. The platform tops up the driver regardless of how much the rider paid.

**Issue 3.** Part 4.9 (Voucher Maximum Discount) states excess is forfeited unless the voucher is marked "Wallet Cashback." Part 5.9 (Locked Balance) states that when a booking is confirmed, fare is locked. If a voucher makes the net fare zero, is any balance locked? **Resolution:** The locked amount is the net fare (after voucher). If the voucher covers the full fare, the wallet lock is zero. No wallet funds are at risk during the trip.

**Issue 4.** Part 8.2 (Daily Quest) requires driver to opt in by 10:00. Part 8.5 (Streak Bonus) requires the driver to be online for a minimum time. A driver who starts their day at 11:00 cannot opt in to the Quest — does this break their streak? **Resolution:** Streak is based on trips completed, not Quest participation. A driver who starts late and completes 3+ trips in a day preserves their streak even if they miss the Quest opt-in window.

**Issue 5.** Part 12 (Risk Engine) and Part 11 (Fraud Rules) overlap on automatic account suspension. **Resolution:** Part 11 defines the triggering conditions for fraud actions. Part 12 defines the scoring system that governs escalation. Fraud-triggered actions in Part 11 bypass the score-based escalation and apply directly.

---

# APPENDIX B — UNRESOLVED BUSINESS DECISIONS

The following items require a formal CPO-level decision before implementation can proceed. They are listed here to ensure they are not silently resolved by engineers.

**UBD-001 — Cash Payment.**
Does FAIRRIDE accept cash payment at launch? Cash introduces settlement complexity (driver holds cash, platform cannot credit wallet automatically), fraud risk (driver keeps cash and claims rider didn't pay), and regulatory considerations. Many Southeast Asian markets require cash acceptance. **Status: Unresolved. Requires COO + CFO decision.**

**UBD-002 — Rider Trust Score.**
Should riders have a Trust Score (analogous to the Driver Trust Score in Part 9.9)? A rider Trust Score could gate access to premium features, promotions, and discounts. However, it introduces complexity and potential bias. **Status: Deferred to Phase 2 product planning.**

**UBD-003 — Dispute Arbitration.**
When a driver and rider have conflicting accounts of a trip (e.g., driver claims rider was aggressive; rider claims driver took a wrong route), who bears the burden of proof? The current document gives Operations discretion. A more formal arbitration rule is needed. **Status: Unresolved. Requires COO decision.**

**UBD-004 — Insurance Partnership.**
FAIRRIDE's risk philosophy (Part 1.8) references insurance partners for operational risk. No insurance product has been procured. Until it is, what is the platform's liability for passenger injury during a trip? **Status: Requires Legal + CFO decision before commercial launch.**

**UBD-005 — Tipping.**
Should riders be able to tip drivers? Tipping increases driver income but can also create pressure dynamics and uneven earnings distribution. **Status: Deferred. No rule exists yet. Until a decision is made, tipping is not supported.**

**UBD-006 — Scheduled Rides.**
Riders scheduling trips in advance (e.g., 6:00 AM airport pickup scheduled the night before) require a pricing lock commitment. If surge is active at the scheduled trip time, does the rider pay the surge they could not have anticipated? **Status: Unresolved. Requires CPO decision before scheduling feature is built.**

**UBD-007 — Driver Vehicle Ownership.**
The current rules assume drivers own their vehicles. If FAIRRIDE partners with a fleet operator who provides vehicles to drivers (rental model), the commission and settlement structure must account for a three-party split (FAIRRIDE, fleet operator, driver). **Status: Deferred to fleet partnership phase.**

**UBD-008 — Negative Driver Earning Edge Case.**
If a driver incurs a fraud clawback that exceeds their current outstanding earnings, and they have already been paid their weekly settlement, the platform has an unrecovered debt. How is this debt enforced? Deducting from future earnings is only possible if the driver continues driving. If they quit, the debt is unrecoverable through platform means. **Status: Requires Legal + CFO decision on enforcement mechanism.**

**UBD-009 — Multi-Stop Trips.**
A rider who requests to stop at a waypoint during a trip (e.g., stop at ATM before continuing to destination). How is waiting time at the waypoint charged? Is a second Minimum Fare triggered? **Status: Unresolved. Multi-stop trip feature requires its own pricing addendum.**

**UBD-010 — Driver Income Ceiling.**
Should FAIRRIDE impose a maximum earning ceiling on any single driver per day, to prevent one driver from using automation or GPS tricks to simulate inhuman trip volumes? **Status: Under review by Risk team. Preliminary rule suggested: flag any driver completing > 20 trips in a single 24-hour period for review, without automatic penalty.**

---

*End of FAIRRIDE Business Rule Bible — Version 1.0*

*This document was authored under the joint authority of the Chief Product Officer, Chief Financial Officer, Chief Operations Officer, and Chief Technology Officer.*

*Next scheduled review: Q4 2026*

*All amendments must be versioned (v1.1, v1.2, etc.) and distributed to all stakeholders before taking effect. No amendment takes effect retroactively.*
