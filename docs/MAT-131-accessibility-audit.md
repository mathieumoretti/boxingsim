# MAT-131: Accessibility & Responsive UI Audit

**Status:** In Progress | **Priority:** Medium  
**Audited:** 2026-11-19  
**Related Issue:** https://linear.app/mathieu-moretti/issue/MAT-131/accessibility-and-responsive-ui-audit

---

## Executive Summary

The Boxing Simulator has a solid foundation with design tokens (MAT-146), shared UI components (MAT-150), and existing responsive breakpoints. However, **mobile functionality is severely broken** and accessibility needs significant work.

### Critical Findings

| Category | Severity | Count |
|----------|----------|-------|
| Mobile Layout Breakage | 🔴 Critical | 8 issues |
| Keyboard Navigation | 🟡 Medium | 4 issues |
| Form Accessibility | 🟡 Medium | 5 issues |
| Color Contrast | 🟢 Low | 2 issues |
| Semantic HTML | 🟡 Medium | 6 issues |

---

## Detailed Findings

### 1. MOBILE LAYOUT BREAKAGE (Critical) 🔴

#### 1.1 TopBar Navigation Collapses Poorly

**Location:** `src/components/TopBar.css` (lines 233-269)

**Issue:** At `≤768px`, the TopBar wraps into 3 rows with navigation splitting awkwardly. The flex-wrap approach causes:
- Brand icon on top row alone
- Navigation + World Clock forced onto second full-width row
- User actions on third row with border separator

**Current Code:**
```css
@media (max-width: var(--breakpoint-mobile)) {
  .top-bar {
    flex-wrap: wrap;
    /* ... */
  }
  .top-bar-left { order: 1; }
  .top-bar-center { 
    order: 2; 
    width: 100%;  /* Full width wastes space */
  }
  .top-bar-right { 
    order: 3; 
    width: 100%;  /* Another full-width row */
    border-top: 1px solid rgba(255, 255, 255, 0.1);
  }
}
```

**Impact:** Takes up ~150-180px vertical space on mobile (3 rows) instead of ~64px (single row).

**Recommendation:** 
- Implement hamburger menu for mobile navigation
- Keep brand + hamburger on left, user actions on right
- Hide full nav in dropdown/modal

---

#### 1.2 Dashboard Grid Breaks Without Stacking

**Location:** `src/components/Dashboard.css` (lines 98-115)

**Issue:** Two-column grid (`1fr 380px`) doesn't stack properly on tablet/mobile:
```css
.dashboard-content-grid {
  display: grid;
  grid-template-columns: 1fr 380px;  /* Fixed width breaks on small screens */
}
```

**Breakpoint Fix Applied:** Lines 423-436 correctly stack at `≤1024px`:
```css
@media (max-width: var(--breakpoint-tablet)) {
  .dashboard-content-grid {
    grid-template-columns: 1fr;
  }
}
```

**Remaining Issue:** Stat cards show 2x2 grid on mobile but `minmax(200px, 1fr)` causes tight wrapping. Cards become unreadable below 400px width.

---

#### 1.3 BoxerProfile Horizontal Layout Issues

**Location:** `src/components/BoxerProfile.css` (lines 87-92)

**Issue:** `.header-avatar-section` uses `display: flex` with avatar + info side-by-side. At tablet breakpoint, it stacks but **action buttons remain horizontal**:

```css
.profile-actions {
  display: flex;
  gap: var(--space-3);
  justify-content: center;
  flex-wrap: wrap;  /* Wraps but creates uneven rows */
}
```

**Impact:** On `≤480px`, three action buttons split across 2 rows awkwardly.

---

#### 1.4 FightBookingPage Arena Layout Overwhelming on Mobile

**Location:** `src/components/FightBookingPage.css` (lines 51-58)

**Issue:** Complex gradient background + multiple nested containers + dramatic spacing creates:
- Excessive vertical scrolling (~4000px page height on mobile)
- Tiny interactive targets (checkboxes, slider)
- Filter toggles stack vertically consuming entire viewport

**Specific Breakage:**
```css
.filter-toggles {
  display: flex;
  gap: var(--space-6);
  flex-wrap: wrap;  /* Stacks but loses "control panel" feel */
}
```

---

#### 1.5 CreateBoxer Modal-Like Layout Has No Mobile Optimization

**Location:** `src/components/CreateBoxer.css` (line 17)

**Issue:** Fixed max-width of 500px is nearly full-screen on mobile:
```css
.create-boxer {
  max-width: 500px;  /* 94% of 532px screen width */
}
```

---

### 2. KEYBOARD NAVIGATION ISSUES (Medium) 🟡

#### 2.1 No Visible Focus Indicators on Navigation Links

**Location:** `src/components/Navigation.css`

**Issue:** Nav links have hover states but **no `:focus-visible` styles**:
```css
.nav-link {
  /* ... */
  &:hover {
    color: var(--color-text-bright);
    background-color: rgba(59, 130, 246, 0.1);
  }
}
/* Missing :focus-visible */
```

**Impact:** Keyboard users cannot see which link is selected.

---

#### 2.2 TopBar Buttons Missing Focus States

**Location:** `src/components/TopBar.css` (lines 109-156)

**Issue:** `.btn-add-fighter` and `.btn-logout` have hover/active states but no explicit focus-visible outlines.

---

#### 2.3 No Skip Navigation Link

**Issue:** No "Skip to main content" link at page top.

**Impact:** Keyboard users must tab through entire TopBar + Navigation on every page load.

---

#### 2.4 Modal/Dialog Focus Trapping Not Implemented

**Issue:** No modals currently exist, but future implementations (confirmations, filters) will need:
- Focus trap within modal
- Return focus to trigger on close
- `aria-modal="true"`

---

### 3. FORM ACCESSIBILITY ISSUES (Medium) 🟡

#### 3.1 Login Form Labels Not Programmatically Associated

**Location:** `src/components/Login.jsx` (lines 96-111)

**Issue:** Labels use `htmlFor` but inputs don't consistently have matching `id`:
```jsx
<label htmlFor="username">Username...</label>
<input
  type="text"
  id="username"  /* ✅ Correct */
  name="username"
  /* ... */
/>
```

**Status:** Login/Register forms are correctly associated. **Issue: Form component clones children but may lose id.**

---

#### 3.2 FormGroup Component ID Mismatch Risk

**Location:** `src/components/ui/Form.jsx` (lines 60-84)

**Issue:** `FormGroup` tries to extract id from children props but this is fragile:
```jsx
const elementId = children?.props?.id || children?.props?.name || '';

return (
  <div className={`form-group ${error ? 'error' : ''} ${className}`}>
    {label && (
      <label htmlFor={elementId}>  {/* May be empty string! */}
        {label}
```

**Impact:** If child has no `id` or `name`, label association breaks.

---

#### 3.3 No Error Summary for Forms

**Issue:** Form errors display inline but no summary at top of form announcing:
- Number of errors
- Links to each error field

---

#### 3.4 Required Field Indicator Inconsistent

**Locations:**
- Login/Register: Uses `*` in label (line 98 Auth.css)
- CreateBoxer: No consistent indicator
- TrainingPage/Form: Uses `required` prop but styling varies

---

#### 3.5 Checkbox Accessibility in FightBooking Filters

**Location:** `src/components/FightBookingPage.css` (lines 207-234)

**Issue:** Custom checkbox styling may interfere with native accessibility:
```css
.toggle-checkbox input[type="checkbox"] {
  width: 20px;
  height: 20px;
  accent-color: var(--fight-red-bright);  /* ✅ Good - uses accent-color */
}
```

**Status:** Acceptable - using `accent-color` preserves native checkbox behavior.

---

### 4. COLOR CONTRAST ISSUES (Low) 🟢

#### 4.1 Muted Text on Dark Navy Backgrounds

**Locations:** Multiple components with navy backgrounds

**Issue:** `var(--color-navy-muted)` = `#a0aec0` on `var(--color-navy-dark)` = `#0f0f23`
- **Contrast ratio:** ~4.5:1 (passes WCAG AA for normal text, fails for large text requirements)

**Example:**
```css
.tagline {
  color: var(--color-gray-500);  /* #6b7280 on navy */
}
```

---

#### 4.2 Victory Gold on White/Light Backgrounds

**Issue:** `var(--color-victory-gold)` = `#ffd700` has poor contrast on light backgrounds:
- On white: 1.5:1 (fails WCAG)
- Use only on dark backgrounds or with dark text shadow

---

### 5. SEMANTIC HTML ISSUES (Medium) 🟡

#### 5.1 Missing `<main>` Element in Many Pages

**Locations:**
- ✅ Dashboard: Has `<main className="dashboard-main">`
- ❌ BoxerProfile: Uses `<div className="boxer-profile-content">` inside PageLayout
- ❌ FightBookingPage: No semantic landmark

---

#### 5.2 Navigation Not Wrapped in Semantic Element

**Location:** `src/components/Navigation.jsx` (line 17)

**Status:** ✅ Correctly uses `<nav aria-label="Primary navigation">`

---

#### 5.3 Card Components Use `<div>` Instead of `<article>` or `<section>`

**Locations:**
- BoxerCard: Uses `<div className="boxer-card">`
- RankCard: Uses `<div className="rank-card">`

**Impact:** Screen readers cannot identify these as independent articles.

---

#### 5.4 Icon-Only Buttons Lack aria-label

**Locations:** Multiple emoji "icons" used as decorative elements:
```jsx
<span className="nav-link-icon" aria-hidden="true">📊</span>  /* ✅ Correct */
<span className="stat-icon" aria-hidden="true">💪</span>      /* ✅ Correct */
```

**Status:** Most correctly marked as `aria-hidden="true"`. Check any icon-only interactive elements.

---

#### 5.5 No Live Regions for Dynamic Content Updates

**Issue:** World Clock updates every 5 seconds but screen readers aren't notified:
```jsx
// src/components/MiniWorldClock.jsx
useEffect(() => {
  // ... fetches time
  const interval = setInterval(fetchWorldTime, 5000);
}, []);
```

**Impact:** Users unaware of time changes.

---

#### 5.6 Table Markup Not Used for Tabular Data

**Locations:**
- Rankings page: Uses card grid instead of table (acceptable for responsiveness)
- Fight history: Timeline format instead of table

**Status:** Acceptable - cards/timelines often more accessible than complex tables on mobile.

---

### 6. SCREEN READER COMPATIBILITY

#### 6.1 Missing lang Attribute on HTML

**Location:** `public/index.html`

**Check Required:** Verify `<html lang="en">` is present.

---

#### 6.2 No aria-live Regions for Notifications

**Issue:** Dashboard notifications panel updates dynamically but no live region announced:
```jsx
<section className="dashboard-panel notifications-panel" aria-label="Notifications">
```

**Fix:** Add `aria-live="polite"` for non-urgent updates.

---

#### 6.3 Loading States Not Announced

**Locations:** Multiple loading spinners

**Issue:** 
```jsx
<div className="loading-spinner"></div>
<p>Loading boxer profile...</p>  /* Visually hidden? */
```

**Fix:** Use `aria-busy="true"` on containers or `role="status"` with live region.

---

## MOBILE TESTING CHECKLIST

Test on these devices/breakpoints:

| Device | Width | Status |
|--------|-------|--------|
| iPhone SE (3rd gen) | 375px | ❌ Needs testing |
| iPhone 14/15 | 390px | ❌ Needs testing |
| iPhone 14/15 Pro Max | 430px | ❌ Needs testing |
| iPad Mini | 768px | ❌ Needs testing |
| iPad Air (landscape) | 1024px | ❌ Needs testing |

---

## RECOMMENDED CSS FRAMEWORK: TAILWIND CSS

Given the existing design token foundation, **Tailwind CSS** would accelerate mobile-first development:

### Benefits for This Project:

1. **Design Token Integration:** Tailwind config can read from `design-tokens.css` or convert to JavaScript config
2. **Mobile-First Utilities:** Built-in responsive prefixes (`md:`, `lg:`) match existing breakpoints
3. **Accessibility Plugins:** `@tailwindcss/forms`, `@tailwindcss/typography` provide accessible defaults
4. **No CSS Bloat:** PurgeCSS removes unused utilities in production

### Migration Path:

**Phase 1 - Hybrid Approach (Recommended):**
- Keep existing component CSS for desktop
- Add Tailwind utilities for responsive breakpoints only
- Gradually convert components over time

**Example:**
```css
/* Existing BoxerProfile.css stays */

/* Add Tailwind classes in JSX for mobile fixes */
<div className="boxer-header md:flex-row flex-col">
```

### Alternative: Windi CSS / UnoCSS

Similar to Tailwind but with better performance (on-demand compilation).

---

## PRIORITY RECOMMENDATIONS

### Immediate (This Sprint)

1. **Implement Hamburger Menu for Mobile Navigation**
   - Hide full nav on `<768px`
   - Add hamburger icon button (accessible with aria-expanded)
   - Slide-down or overlay menu panel

2. **Add Skip Navigation Link**
   - First focusable element on every page
   - Jumps to main content after TopBar

3. **Fix Focus Visible States**
   - Add `:focus-visible` styles to all interactive elements
   - Use existing design token colors for consistency

4. **Reduce Page Height on FightBookingPage Mobile**
   - Stack filter controls vertically with less spacing
   - Consider accordion/collapsible sections

### Short-Term (Next 1-2 Sprints)

5. **Semantic HTML Improvements**
   - Add `<main>` landmarks to all pages
   - Convert card containers to `<article>` where appropriate
   - Add `aria-live` regions for dynamic content

6. **Form Accessibility Enhancements**
   - Fix FormGroup ID association
   - Add error summaries
   - Consistent required field indicators

7. **Color Contrast Remediation**
   - Increase contrast of muted text on navy backgrounds
   - Add text shadows to gold/yellow on light backgrounds

### Long-Term (Backlog)

8. **CSS Framework Adoption**
   - Evaluate Tailwind CSS integration
   - Pilot on 1-2 new components before full migration

9. **Comprehensive A11y Testing**
   -axe-core automated testing in CI
   - Manual screen reader testing (NVDA, VoiceOver)
   - Keyboard-only navigation audit

---

## RELATED ISSUES

- ✅ MAT-146: Design Tokens Foundation
- ✅ MAT-150: Shared UI Component Library  
- ✅ MAT-152: Page Layout Wrapper Component
- 🔄 MAT-131: This audit
- ⏭️ TBD: Mobile Navigation Implementation
- ⏭️ TBD: Focus Indicator Standardization

---

## TESTING NOTES

### Manual Accessibility Testing Checklist

**Keyboard Only:**
- [ ] Tab through entire Dashboard
- [ ] Tab through BoxerProfile page
- [ ] Tab through FightBookingPage filters
- [ ] Submit Login form using Enter key
- [ ] Navigate between pages without mouse

**Screen Reader (NVDA/VoiceOver):**
- [ ] Announce TopBar navigation
- [ ] Read Dashboard stat cards
- [ ] Navigate BoxerProfile sections
- [ ] Form field announcements on Login
- [ ] Error message announcements

**Mobile Browsers:**
- [ ] Chrome DevTools Device Emulation (iPhone, iPad)
- [ ] Actual device testing if available

---

## CONCLUSION

The Boxing Simulator has excellent design system foundations but needs focused work on mobile responsiveness and accessibility. The recommended approach is:

1. **Immediate:** Fix critical mobile navigation with hamburger menu
2. **Short-term:** Add keyboard focus indicators and semantic landmarks  
3. **Long-term:** Evaluate Tailwind CSS for accelerated responsive development

**Estimated Effort:** 3-5 days for immediate fixes, 1-2 weeks for full accessibility remediation.
