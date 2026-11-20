# MAT-131 Phase 3: Accessibility Testing & CSS Framework Evaluation

**Status:** Planning | **Priority:** Medium  
**Created:** 2026-11-19  
**Related:** https://linear.app/mathieu-moretti/issue/MAT-131

---

## Overview

Phase 3 focuses on comprehensive accessibility testing and evaluating CSS framework adoption (Tailwind CSS) to accelerate future responsive development. Unlike Phases 1-2 which were implementation-focused, Phase 3 is research and evaluation work that informs future decisions.

---

## Completed Work Summary

### Phase 1 ✅ - Mobile Navigation Foundation
- Hamburger menu with `MobileNavToggle` component
- Slide-out drawer with `MobileNavigationDrawer`
- Skip link for keyboard navigation
- Focus-visible states on interactive elements
- Body scroll prevention during mobile nav open

### Phase 2 ✅ - Semantic HTML & Form Accessibility
- `<article>`, `<section>`, `<nav>` landmarks on key pages
- WCAG 4.5:1 color contrast improvements
- FormGroup ID association fix
- ARIA labels and `role="alert"` for forms

---

## Phase 3 Tasks

### Task 3.1: Tailwind CSS Evaluation

**Goal:** Determine if Tailwind CSS integration would benefit the project's responsive development workflow.

#### Current State Analysis
- ✅ Design tokens (MAT-146) provide centralized colors, spacing, breakpoints
- ✅ Shared UI components (MAT-150) reduce duplication
- ⚠️ Mobile responsiveness requires custom media queries per component
- ⚠️ No utility classes - all styles in separate CSS files

#### Evaluation Criteria

**Pros of Tailwind Integration:**
1. **Mobile-First Utilities**: Built-in `md:`, `lg:` prefixes match existing breakpoints
2. **Design Token Config**: Can map `design-tokens.css` values to Tailwind config
3. **Reduced CSS Bloat**: PurgeCSS removes unused utilities in production
4. **Accessibility Plugins**: `@tailwindcss/forms` provides accessible form defaults
5. **Development Speed**: Less context switching between JSX and CSS files

**Cons / Concerns:**
1. **Learning Curve**: Team must learn utility class names
2. **Migration Effort**: Converting existing components takes time
3. **Bundle Size**: Larger initial CSS bundle (before purging)
4. **Hybrid Complexity**: Mixing Tailwind + traditional CSS can be confusing

#### Recommended Approach: Hybrid Pilot

**Step 1 - Install & Configure (1 day):**
```bash
npm install -D tailwindcss postcss autoprefixer
npx tailwindcss init
```

**tailwind.config.js:**
```js
module.exports = {
  content: ["src/**/*.{js,jsx}"],
  theme: {
    extend: {
      // Map existing design tokens to Tailwind
      colors: {
        'navy': {
          dark: '#0f0f23',
          medium: '#1a1a3e',
          light: '#252561',
        },
      },
      screens: {
        'mobile': '768px',
        'tablet': '1024px',
        'desktop': '1440px',
      }
    }
  },
  plugins: [
    require('@tailwindcss/forms'),
  ]
}
```

**Step 2 - Pilot on New Components (2-3 days):**
- Create ONE new component using Tailwind utilities only
- Compare development time vs. traditional CSS approach
- Measure final CSS bundle size impact

**Step 3 - Hybrid Migration (Optional, Future Sprint):**
- Add Tailwind responsive classes to existing components for mobile fixes only
- Example: `<div className="boxer-card md:flex-row flex-col">`
- Keep desktop styles in CSS files

**Decision Gate:** After pilot, decide whether to:
1. **Adopt fully** - Gradually migrate all components over 2-3 sprints
2. **Hybrid only** - Use Tailwind for responsive utilities, keep CSS for styling
3. **Decline** - Continue with traditional CSS (valid choice if not needed)

---

### Task 3.2: Axe-Core Automated Testing

**Goal:** Add automated accessibility testing to CI pipeline.

#### Implementation Steps

**Step 1 - Install Dependencies:**
```bash
npm install -D @axe-core/react jest-axe @testing-library/jest-dom
```

**Step 2 - Configure Jest (jest.config.js):**
```js
setupFilesAfterEnv: ['<rootDir>/src/setupTests.js']
```

**Step 3 - Create Test Setup (src/setupTests.js):**
```js
import '@testing-library/jest-dom';
import { configureAxe, toHaveNoViolations } from 'jest-axe';

configureAxe({ rules: { 'avoid-inline-spacing': { enabled: false } } });
expect.extend(toHaveNoViolations);
```

**Step 4 - Add Accessibility Test Suite:**
```js
// src/components/MobileNavToggle.test.jsx
import { render, screen } from '@testing-library/react';
import { MobileNavToggle } from './MobileNavToggle';

describe('MobileNavToggle', () => {
  test('has no accessibility violations', async () => {
    const { container } = render(
      <MobileNavToggle isOpen={false} onToggle={() => {}} />
    );
    expect(container).toHaveNoViolations();
  });
});
```

**Step 5 - Add to CI Pipeline (.github/workflows/test.yml):**
```yaml
- name: Test with accessibility checks
  run: npm run test:a11y
```

**Expected Violations to Fix:**
- Color contrast failures (already addressed in Phase 2)
- Missing alt text on decorative icons (add `aria-hidden="true"`)
- Form label associations (already addressed in Phase 2)
- Focus trap in modals (future work)

---

### Task 3.3: Manual Screen Reader Audit

**Goal:** Test real-world accessibility with screen readers.

#### Testing Setup

**Tools Required:**
- **NVDA** (Windows, free): https://www.nvaccess.org/
- **VoiceOver** (macOS/iOS, built-in): Cmd+Ctrl+F5 to toggle
- **JAWS** (Windows, paid trial): https://www.freedomscientific.com/

#### Test Scenarios

1. **Dashboard Navigation:**
   - Can user reach "Add Fighter" button without getting lost?
   - Are stat cards announced as distinct items?
   - Is the navigation structure logical?

2. **Boxer Profile Page:**
   - Does screen reader announce upcoming fight if present?
   - Are action buttons grouped logically?
   - Can user navigate between stats, history, training sections?

3. **Fight Booking Page:**
   - Are filter controls announced with their purpose?
   - Does opponent selection provide meaningful context?
   - Are match warnings announced as alerts?

4. **Form Interactions (Login, Create Boxer):**
   - Are required fields marked as required?
   - Do error messages get announced on submit?
   - Can user navigate between form fields easily?

#### Checklist Template

```markdown
[ ] NVDA + Chrome: Full dashboard navigation
[ ] VoiceOver + Safari: Form completion flow
[ ] NVDA + Firefox: Fight booking with filters
[ ] Keyboard-only: All pages navigable without mouse
```

---

### Task 3.4: Contrast Ratio Verification

**Goal:** Verify all text meets WCAG 2.1 AA standards (4.5:1 for normal text, 3:1 for large text).

#### Tools

- **WebAIM Contrast Checker**: https://webaim.org/resources/contrastchecker/
- **axe DevTools Browser Extension**: Automated contrast checking
- **Stark Design Plugin**: Design-time contrast verification (Figma/Sketch)

#### Current Status

**Fixed in Phase 2:**
- ✅ `--color-text-muted`: Changed to `#d1d5db` (~7:1 on navy)
- ✅ `--color-navy-muted`: Changed to `#c4b5fd` (~6:1 on navy)

**Remaining Verification:**
- [ ] Victory gold (`#ffd700`) on white/light backgrounds - **FAILS** (needs text shadow or dark background only)
- [ ] Warning colors in stats - verify all variants
- [ ] Link colors on card backgrounds
- [ ] Disabled button states maintain contrast

---

## Success Metrics

Phase 3 is complete when:

1. **Tailwind Decision Made**: Clear recommendation documented (adopt/hybrid/decline) with pilot data
2. **Axe Tests Passing**: No critical violations in CI pipeline
3. **Screen Reader Report**: Complete test scenarios documented with findings
4. **Contrast Audit**: All UI states verified against WCAG 2.1 AA

---

## Related Issues

- ✅ MAT-131 Phases 1-2: Implementation complete
- ✅ MAT-146: Design Tokens Foundation
- ⏭️ TBD: Tailwind CSS Migration (if approved)
- ⏭️ TBD: Accessibility Testing in CI

---

## Next Steps

1. **Discuss Phase 3 scope** - Does user want to proceed with evaluation or consider MAT-131 complete?
2. **Prioritize tasks** - Which is most valuable: Tailwind eval, axe testing, screen reader audit?
3. **Timebox effort** - Each task can be done in 1-2 days; decide if worth the investment

---

## Decision Framework

**If adopting Tailwind CSS:**
- Pilot on ONE new component first (not existing codebase)
- Keep traditional CSS for desktop styles initially
- Use Tailwind only for responsive breakpoints until comfortable

**If declining Tailwind CSS:**
- Continue with design token system (it's working well)
- Focus on axe-core testing instead (high value, low effort)
- Manual screen reader audit still recommended regardless
