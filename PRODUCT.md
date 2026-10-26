# Product

<!-- impeccable:product-schema 1 -->

## Platform

web (responsive design for mobile and desktop browsers)

## Users

**Primary**: Strategy game players and boxing enthusiasts who manage boxer careers through training, fight scheduling, and contract negotiation. They enjoy long-form progression, tactical decision-making, and emergent storytelling from AI-driven world simulation.

**Secondary**: Other human players in the same persistent world who become negotiation partners, rivals, or opponents in a shared PvPVE ecosystem.

## Product Purpose

A managerial PvPVE boxing simulation where players guide boxers from amateur to champion through strategic training, fight booking, and community contract negotiation. The game runs on accelerated time (1 real minute = 1 game hour) with a server-authoritative world clock that progresses continuously, enabling persistent multiplayer interaction without requiring simultaneous play.

Success means building legendary boxer careers while participating in a living economy of fights, rivalries, and rankings—where AI opponents exhibit emergent goal-driven behavior and human players negotiate contracts to create meaningful matchups.

## Positioning

Unlike Football Manager or traditional sports management games, this product combines:

1. **Community-driven fight economy**: Players book fights with each other through contract negotiation rather than facing only AI opponents
2. **Emergent Utility AI**: Boxers develop goal-oriented behavior (retiring after career-defining losses, refusing unfavorable contracts, building rivalries) similar to RimWorld's pawn system
3. **Asynchronous multiplayer persistence**: The accelerated world clock enables continuous progression without real-time PvP pressure
4. **Narrative generation potential**: Fight logs and news systems feed into visual novel-style recaps for marquee matchups

## Operating Context

- Web browser interface (mobile-responsive)
- Accelerated game time: 1 real minute = 1 game hour, configurable speeds for testing
- Server-authoritative simulation: all state processed backend-side
- Event-driven architecture: scheduled events (training completion, fights, recovery) processed by background worker
- Persistent world: game continues progressing while players are offline

## Capabilities and Constraints

**Confirmed Functionality:**
- Boxer creation and stat management (Power, Speed, Chin, Cardio, Defense, Fight IQ)
- Training system with fatigue mechanics and diminishing returns
- Automatic recovery/rest scheduling based on fatigue levels
- Fight simulation with round-by-round deterministic replay
- Contract-based fight booking between players
- Ranking ladders per weight class
- World clock with configurable speed multipliers

**Technical Constraints:**
- Go backend with PostgreSQL persistence and Redis caching
- React frontend with Webpack bundling
- JWT authentication for user sessions
- REST API architecture

**Future Scope (Confirmed but Not Implemented):**
- News system with world events
- Fog-of-war information gating
- Role system for boxer careers
- Prestige/reincarnation mechanics
- Visual novel fight recaps from generated logs

## Brand Commitments

- Game name: TBD (currently "Boxing Simulator")
- Tagline in use: "Train. Schedule. Conquer. Own the Ring."
- No established visual identity or brand assets yet
- Terminology: "boxer" (not "fighter"), "contract" for fight agreements, "training session" for stat improvement activities

## Evidence on Hand

- Existing codebase with authentication, boxer CRUD, training system, fight simulation, and world clock
- Documentation in `docs/` covering database design, game mechanics, configuration, and testing strategy
- Architecture Decision Records in `adr/` documenting technology choices and patterns
- React components for dashboard, boxer profile, fight booking, training page, rankings

**No testimonials, customer data, press coverage, or legal commitments exist.**

## Product Principles

1. **Emergent narrative over scripted content**: The game generates stories through systems (AI goals, contract negotiations, fight outcomes) rather than authored scenarios
2. **Community as competitive landscape**: Human players are the primary obstacle and opportunity, not AI填充
3. **Strategic depth without twitch skill**: All decisions are planning-based; no real-time reflex gameplay
4. **Persistent world respect**: The clock keeps turning—design for players who join mid-career or return after long absences

## Accessibility & Inclusion

No specific accessibility requirements established. Target is responsive web compatible with standard browser assistive technologies.
