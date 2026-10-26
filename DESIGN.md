# Design

<!-- impeccable:design-schema 1 -->

```yaml
---
design-system:
  version: "1.0"
  
  colors:
    primary:
      main: "#3b82f6"
      light: "#60a5fa"
      dark: "#2563eb"
      
    navy:
      dark: "#0f0f23"
      medium: "#1a1a2e"
      light: "#16213e"
      
    success:
      main: "#10b981"
      light: "#34d399"
      
    danger:
      main: "#ef4444"
      dark: "#dc2626"
      
    warning:
      main: "#f59e0b"
      dark: "#d97706"
      
    victory:
      gold: "#ffd700"
      light: "#ffed4e"
      
    neutral:
      gray-100: "#f5f7fa"
      gray-200: "#f5f5f5"
      gray-300: "#d1d5db"
      gray-600: "#718096"
      gray-700: "#4a5568"
      
    stats:
      strength: "#ef4444"
      defense: "#3b82f6"
      agility: "#eab308"
      health: "#10b981"
      energy: "#eab308"

  typography:
    font-family: "'Segoe UI', Tahoma, Geneva, Verdana, sans-serif"
    scale:
      xs: "0.65rem"
      sm: "0.875rem"
      md: "1rem"
      lg: "1.25rem"
      xl: "1.5rem"
      xxl: "2rem"
    weights: [400, 500, 600, 700, 800]
    line-height:
      tight: "1.2"
      normal: "1.6"

  spacing:
    unit: "4px"
    scale: [4, 8, 12, 16, 20, 24, 32, 40, 48]
    
  rounded:
    values: [6, 8, 10, 12, 16, 20]  # px
    
  shadows:
    - name: "subtle"
      value: "0 1px 3px rgba(0,0,0,0.1)"
    - name: "card"
      value: "0 4px 20px rgba(0,0,0,0.08)"
    - name: "elevated"
      value: "0 8px 30px rgba(0,0,0,0.12)"
    - name: "glow-red"
      value: "0 0 10px rgba(239,68,68,0.5)"
    - name: "glow-blue"
      value: "0 0 10px rgba(59,130,246,0.5)"
    - name: "glow-gold"
      value: "0 0 10px rgba(234,179,8,0.5)"
      
  breakpoints:
    mobile: "480px"
    tablet: "768px"
    desktop: "1024px"

  components:
    button:
      variants: [primary, secondary, danger]
    badge:
      variants: [level-1, level-2, level-3, level-4, level-5, training]
    card:
      variants: [default, profile, fight-details, comparison]
---
```

## Overview

**Creative North Star**: Professional Boxing Management Dashboard — a serious SaaS-grade interface for managing boxer careers, treating the simulation with the visual respect of enterprise software rather than arcade game aesthetics.

**Mode**: Operate — visitors complete tasks (training decisions, fight booking, contract negotiation). Scanability, consistency, and precision drive design choices. Brand expression lives in refined details: gradient treatments, stat bar glows, activity timeline color-coding.

**Aesthetic Voice**: Refined, precise, open-minded, efficient, eye-opening. The interface respects user intelligence with information density balanced through visual hierarchy rather than empty space. Dark navy gradients signal premium boxer content (profiles, fight details); light backgrounds maintain readability for operational views (dashboard, listings).

---

## Colors

### Primary System

**Primary Blue (`#3b82f6`)** — Action and interaction: buttons, links, active states, selection indicators. Gradient variants create depth on primary CTAs.

**Navy Gradient Family** — Boxing-specific dark mode for premium content:
- Navy Dark (`#0f0f23`) — Deepest layer, outer gradient stops
- Navy Medium (`#1a1a2e`) — Mid-tone for contrast
- Navy Light (`#16213e`) — Accent and highlight layer

Used on boxer profiles, fight details pages, training views — contexts where visual drama supports the subject matter without compromising readability.

### Semantic Colors

**Success Green (`#10b981`)** — Wins, health status (high), positive outcomes
**Danger Red (`#ef4444`)** — Losses, critical fatigue, low health, destructive actions
**Warning Amber (`#f59e0b`)** — Medium fatigue, caution states, pending events
**Victory Gold (`#ffd700`)** — Champions, knockouts, exceptional achievements

### Combat Stat Colors

Domain-specific semantic coding for instant stat recognition:
- **Strength**: Red (`#ef4444`) with matching glow shadow
- **Defense**: Blue (`#3b82f6`) with matching glow shadow
- **Agility**: Gold (`#eab308`) with matching glow shadow

Consistent across boxer cards, profile pages, fight details comparisons.

### Backgrounds

**Light Mode** — Dashboard and operational views:
- Page background: `#f5f7fa` (neutral gray-100)
- Cards: White (`#ffffff`) with subtle shadows

**Dark Mode** — Profile and fight content pages:
- Gradient backgrounds using navy family
- Card overlays with semi-transparent navy backgrounds (`rgba(30, 30, 50, 0.8)`)

---

## Typography

### Font Stack

`'Segoe UI', Tahoma, Geneva, Verdana, sans-serif`

System font stack prioritizes performance and familiarity over custom web fonts. Clean sans-serif supports information density without visual noise.

### Type Scale

| Name | Size | Weight | Usage |
|------|------|--------|-------|
| xs | 0.65rem | 500-600 | Badges, labels, inline indicators |
| sm | 0.875rem | 400-500 | Body text, secondary information |
| md | 1rem | 500 | Standard body, list items |
| lg | 1.25rem | 600 | Section headers, card titles |
| xl | 1.5rem | 700 | Page headers in sections |
| xxl | 2rem | 700-800 | Primary page headings |

### Hierarchy Patterns

**Page Titles**: 2rem/700 on dark navy backgrounds (profile pages) or large with gradient text treatment (FightDetails victory banners)

**Section Headers**: 1.5rem-1.25rem/600, white or high-contrast gray

**Body Content**: 0.875rem-1rem/400-500, line-height 1.6 for readability

**Badges & Labels**: 0.65rem/600 uppercase treatment, condensed spacing

---

## Layout

### Page Structure

**TopBar Navigation**: Fixed at 64px height, full width, navy gradient background with elevated shadow (`0 2px 20px rgba(0,0,0,0.3)`)

**Main Content Padding**: `padding-top: calc(64px + 24px)` compensates for fixed TopBar

**Container Widths**:
- Dashboard grid: auto-fill with minmax(315px, 340px) columns
- Profile pages: Single column, centered, max-width ~900px implied
- Fight details: Split comparison layout, full width with padding

### Responsive Breakpoints

| Name | Min Width | Behavior |
|------|-----------|----------|
| Mobile | 480px | Single column stacks, full-width cards |
| Tablet | 768px | Two-column grids appear |
| Desktop | 1024px | Full multi-column layouts, side-by-side comparisons |

### Grid Patterns

**Boxer Cards (Dashboard)**: `grid-template-columns: repeat(auto-fill, minmax(315px, 340px))` with 28px row gap

**Stat Rows**: Flexbox space-between layouts for label-value pairs

**Timeline Views** (Training History, Activity): Vertical stack with left-border accent lines

---

## Elevation & Depth

### Shadow Vocabulary

**Subtle (`0 1px 3px rgba(0,0,0,0.1)`)** — Minimal depth for inline elements, borders as primary separation

**Card (`0 4px 20px rgba(0,0,0,0.08)`)** — Standard elevation for boxer cards on light backgrounds, sections within profile pages

**Elevated (`0 8px 30px rgba(0,0,0,0.12)`)** — Hover state on cards, modals, focused interactive elements

**Glow Effects** — Stat bars and status indicators:
- Red glow (strength): `0 0 10px rgba(239,68,68,0.5)`
- Blue glow (defense): `0 0 10px rgba(59,130,246,0.5)`
- Gold glow (agility/energy): `0 0 10px rgba(234,179,8,0.5)`

### Layering Philosophy

Shadows define **focus and interactivity**, not just depth. Cards lift on hover; stat bars glow to signal active attributes; victory banners use animation plus shadows to capture attention. Dark navy backgrounds create depth through gradient layering rather than shadows alone.

---

## Shapes

### Border Radius Scale

| Value | Usage |
|-------|-------|
| 6px | Small badges, pill truncation |
| 8px | Standard cards, buttons |
| 10px | Larger cards with more content |
| 12px | Modal backgrounds, fight detail sections |
| 16px | Major profile containers, header cards |
| 20px | Avatar crops, large featured elements |
| 50% | Pill badges (training status), circular avatars |

### Geometric Language

Rounded corners throughout signal approachability within the serious management context. Sharp angles appear only in data visualizations (stat bar fill animations use rounded caps). No neobrutalist hard edges; no kicked corners or extreme radii beyond 20px.

---

## Components

### Cards

**BoxerCard** — Primary information density pattern:
- Avatar circle (20px radius, top-left) with boxer initials
- Level badge (16px pill, gradient background) in top-right corner
- Name and nickname (large/700, small/500)
- Health/Energy/Fatigue bars with semantic colors
- Combat stat rows (label-value flex layout)
- Training status badge with pulse animation when active
- Hover lift: 8px shadow increase + 4px translateY

**Profile Sections** — Semi-transparent navy backgrounds (`rgba(30, 30, 50, 0.8)`), 16px radius, white border at 10% opacity

### Badges

**Level System**: Five gradient tiers based on boxer level:
- Level 1 (Gray): `#718096 → #4a5568`
- Level 2 (Teal): `#38b2ac → #319795`
- Level 3 (Orange): `#dd6b20 → #c05621`
- Level 4 (Purple): `#805ad5 → #704ada`
- Level 5 (Gold): `#d69e2e → #fbbf24`

**Training Status**: Blue pill badge with CSS keyframe pulse animation

### Stat Bars

Gradient fill bars with matching glow shadows:
- Height: Typically 8-12px
- Background: Gray at low opacity for track
- Fill: Gradient left-to-right (e.g., `#ef4444 → #dc2626`)
- Glow: Box-shadow matching fill color at 50% opacity

### Progress Bars (Health, Energy, Fatigue)

Similar to stat bars with status-based color shifts:
- Health: Green → Yellow → Red based on percentage thresholds
- Energy: Gold gradient constant
- Fatigue: Green → Orange → Red (ok → warning → critical)

### Activity Timeline

Vertical stack of cards with left-border color-coding:
- Win: Green border (`#10b981`)
- Loss: Red border (`#ef4444`)
- Draw: Gray border (`#718096`)
- Training: Blue border (`#3b82f6`)
- Upcoming Fight: Amber border (`#f59e0b`)

### Buttons

**Primary**: Navy gradient background, white text, 8px radius, hover brightness increase
**Secondary**: Outline treatment or neutral background
**Danger**: Red gradient for destructive actions (delete boxer, cancel fight)

---

## Do's and Don'ts

### Do

- Use navy gradients for profile/fight content to create visual drama
- Code combat stats with their semantic colors consistently (strength=red, defense=blue, agility=gold)
- Apply glow shadows to stat bars to make them pop on dark backgrounds
- Keep badge sizes at 0.65rem/600 for consistency across level indicators and status tags
- Use the auto-fill grid pattern for boxer listings to ensure responsive behavior
- Create timeline borders with semantic colors for quick activity scanning
- Maintain TopBar compensation in all page layouts (64px + 24px padding-top)

### Don't

- Don't use navy dark backgrounds on the dashboard or operational views — reserve them for premium content pages
- Don't create custom border radius values outside the established scale (6/8/10/12/16/20px)
- Don't add kickers, hard edges, or extreme shadows — this is refined SaaS, not neobrutalist
- Don't use monochrome stat bars — always apply semantic color + glow
- Don't exceed 2rem for headings unless it's a victory banner requiring emphasis
- Don't stack multiple gradient treatments on the same element (gradients for backgrounds OR buttons, not both)
- Don't forget responsive breakpoints — test at 480px, 768px, 1024px minimum

---

## Boxing-Specific Patterns

### Next Fight Indicator

Urgency signals through color-coded countdown and status badges (upcoming/scheduled). Visual prominence on boxer profile page.

### Fatigue Warning System

Three-tier visual coding with bar color shifts + tooltip explanations. Critical fatigue triggers forced rest UI patterns.

### Fight Booking Validation

Past fight dates are visually filtered from opponent selection — users cannot book fights that have already occurred in game time.

### Victory Banner Treatment

Slide-in animation with gradient gold treatment and knockout variant for KO/TKO wins. Highest visual emphasis on fight details page.

---

**Evidence Source**: Scanned 26 CSS files totaling ~5000+ lines including PageLayout.css, BoxerCard.css, BoxerProfile.css, FightDetails.css, Dashboard.css, TrainingPage.css, TopBar.css, and supporting component styles. All tokens extracted from committed code at project root `src/components/` and `src/`.
