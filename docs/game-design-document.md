# Serious Game Design Document: Boxing Simulator Web App

This document maps out the backend architecture and game mechanics for an event-driven, time-accelerated Boxing Simulator. The backend operates on an accelerated virtual clock where player actions are processed at discrete in-game ticks.

---

## 1. Title Page
*   **Game Name**: *To Be Determined* (Boxing Management Simulator)
*   **Tagline**: "Train. Schedule. Conquer. Own the Ring."
*   **Team**: Player & AI Collaborator
*   **Date of Last Update**: August 2026

## 2. Revision History
*   **v1.0.0**: Initial baseline capturing the core entities, accelerated time loop architecture, and microservice boundaries.

## 3. Game Overview
*   **Purpose**: To provide an immersive sports management simulation that models the career, strategy, and economics of professional boxing. It solves the stagnation of real-time wait games by utilizing an accelerated virtual calendar.
*   **Intended Use**: Standalone web application with persistent, asynchronous, multiplayer worlds.
*   **Justification**: Time-mapped strategy loops succeed in titles like *Football Manager*. An accelerated tick-based architecture ensures continuous world progression even when players are offline.
*   **Target Audience**: Strategy, management, and boxing enthusiasts who prefer tactical planning over real-time twitch controls.
*   **Genre(s)**: Sports Management / Strategy / Simulation.

## 4. Gameplay
*   **Objectives**: Manage a boxer's career from a local gym amateur to a unified world champion, maximizing stats, wealth, and legacy.
*   **Flow/Progression**: Continuous cycle: Player plans actions $\rightarrow$ Submits to queue $\rightarrow$ Time ticks advance $\rightarrow$ Actions process $\rightarrow$ Standings update.
*   **Structure**: Weekly in-game schedules consisting of training regimens, promotional events, contract negotiations, and scheduled fight nights.

## 5. Mechanics
*   **Rules & Model**: Explicit rule constraints dictate weight classes, mandatory title defenses, rankings, and medical suspensions.
*   **Accelerated World Clock**: Virtual time progresses continuously ($1\text{ real minute} = 1\text{ game hour}$; $24\text{ real minutes} = 1\text{ game day}$). 
*   **Action Queuing System**: Player actions are saved as future events with an `end_game_time`. A central tick worker processes events once `current_game_time \geq end_game_time`.
*   **Character Actions**: 
    *   *Training*: Strategic allocation of stats (Strength, Speed, Stamina, Agility) against diminishing returns and fatigue accumulation.
    *   *Matchmaking*: Contract bidding, choosing venue locations, and accepting or declining fight proposals.
    *   *Combat*: Asynchronous execution of a round-by-round engine utilizing fighter attributes, strategy sliders, and procedural RNG.

## 6. Story and Narrative
*   **Background**: Every generated boxer (user or AI) features a procedural background story, hometown, and distinct fighting style personality (e.g., Infighter, Out-boxer, Brawler).
*   **Progression**: Career arcs unfold through fight logs, news text feeds, media press conferences, and promotional rivalry arcs.

## 7. Game World
*   Divided into local, national, and global circuits. Boxers navigate gyms, promotional agencies, and sanctioning bodies (e.g., ranking ladders across weight classes).

## 8. Characters and Opponents
*   **Player Avatar**: The boxing manager or the primary fighter profile.
*   **Roster**: Persistent database of user-controlled fighters alongside a dynamic pool of AI opponents that train, age, decline, and retire automatically to maintain a living eco-system.

## 9. Levels & Ranking Ecosystem
*   **Progression**: Structured via dynamic ranking ladders (Rank #15 to Champion) managed by automated sanctioning bodies.
*   **Onboarding**: Initial amateur tournaments act as a tutorial to teach stats balancing and fight planning.

## 10. User Interface & Assets
*   **UI/Control**: Strategic dashboard featuring a training planner, contract mailbox, rank ladder board, and a ticker displaying the Current Game Time and time remaining until the next World Tick.
*   **Visual Style**: Clean, metric-driven text dashboard layout optimized for mobile and desktop web browsers.

## 11. Proposed Microservice Architecture
A service-oriented, event-driven design boundary to handle high-frequency time simulation:

1.  **Gateway / Auth Service**: Handles JWT-based user session authentication and routing.
2.  **World Clock & Tick Service**: Manages the core virtual timer loop, issues global tick events, and handles distributed locking to ensure single-worker execution.
3.  **Boxer & Gym Service**: Manages state, statistics, training history, fatigue, and recovery metrics.
4.  **Matchmaking & Contract Service**: Handles fight queues, asynchronous user negotiations, and AI contract generation.
5.  **Fight Simulation Engine**: Isolated, high-performance deterministic service that runs round-by-round computations and generates full fight text replays.
6.  **Rankings & Economics Service**: Recalculates leaderboards, processes tournament brackets, distributes fight purses, and monitors gym maintenance fees.

## 12. Data & Deployment
*   **Storage Architecture**: 
    *   *PostgreSQL*: Core relational database for structured tabular entities (Users, Fights, Boxers, Financial logs).
    *   *Redis*: High-speed storage for active matchmaking queues, distributed system locks, cooldown tracking, and live clock state.
*   **Deployment**: Containers managed via Docker/Kubernetes, utilizing background cron workers for periodic execution loops.