# Demo Coding Guidelines

This document captures coding standards, patterns, and common pitfalls for the Planton interactive demo project.

## Table of Contents

1. [File Organization](#file-organization)
2. [Component Structure](#component-structure)
3. [TypeScript Patterns](#typescript-patterns)
4. [React Hooks Guidelines](#react-hooks-guidelines)
5. [State Management](#state-management)
6. [Styling Conventions](#styling-conventions)
7. [Animation Patterns](#animation-patterns)
8. [Common Linter Errors & Fixes](#common-linter-errors--fixes)
9. [Naming Conventions](#naming-conventions)
10. [Best Practices](#best-practices)

---

## File Organization

### Directory Structure

```
src/components/demo/
├── DemoPage.tsx           # Main orchestrator - ALWAYS modify this for flow changes
├── demo.css               # Shared custom styles
├── intro/                 # Introduction and onboarding screens
├── platform/              # Infrastructure/platform feature screens
└── service/               # Service deployment screens
```

### Component Placement Rules

1. **Intro screens** (`intro/`): Welcome, company selection, problem statements, value props, CTAs
2. **Platform screens** (`platform/`): Infrastructure-related features (connections, catalog, builder, deployment)
3. **Service screens** (`service/`): Service deployment features (GitHub, buildpacks, config, deployment)

### Creating New Components

**DO**:
```tsx
// src/components/demo/platform/NewFeature.tsx
'use client';

import React from 'react';
import { motion } from 'framer-motion';

export default function NewFeature() {
  return (
    <div className="h-full flex flex-col justify-center p-8">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6 }}
      >
        {/* Content */}
      </motion.div>
    </div>
  );
}
```

**DON'T**:
- Mix platform and service components in the same directory
- Create components outside the categorized folders
- Forget the `'use client';` directive for interactive components

---

## Component Structure

### Standard Screen Component Pattern

Every demo screen should follow one of these patterns based on its purpose:

#### 1. Standard Content Screen

For educational or informational content:

```tsx
'use client';

import React from 'react';
import { motion } from 'framer-motion';
import { Icon1, Icon2 } from 'lucide-react';

export default function ScreenName() {
  return (
    <div className="h-full flex flex-col justify-center">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6 }}
        className="text-center max-w-4xl mx-auto px-8"
      >
        <h1 className="text-5xl md:text-6xl lg:text-7xl font-black text-gray-900 mb-6 tracking-tight">
          Main Headline
        </h1>
        
        <p className="text-xl md:text-2xl text-gray-600 mb-12 max-w-3xl mx-auto leading-relaxed">
          Supporting description
        </p>

        {/* Content sections */}
      </motion.div>
    </div>
  );
}
```

#### 2. Full-Screen Welcome/Selection Pattern

For welcome or selection screens that need dark backgrounds:

```tsx
'use client';

import React from 'react';
import { motion } from 'framer-motion';

export default function WelcomeScreen() {
  return (
    <div className="min-h-screen bg-[#110D1F] flex items-center justify-center p-8">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6 }}
        className="text-center max-w-6xl mx-auto"
      >
        <h1 className="text-4xl md:text-5xl lg:text-6xl font-black text-white mb-6 tracking-tight">
          {/* White text for dark background */}
        </h1>
        
        <p className="text-xl text-gray-300 mb-12 max-w-3xl mx-auto">
          {/* Light gray text for dark background */}
        </p>
      </motion.div>
    </div>
  );
}
```

#### 3. Milestone/Transition Screen Pattern

For celebrating achievements and transitioning between major sections:

```tsx
'use client';

import React from 'react';
import { motion } from 'framer-motion';
import { CheckCircle2, Rocket, ArrowRight } from 'lucide-react';

export default function MilestoneScreen() {
  return (
    <div className="h-full flex flex-col justify-center">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6 }}
        className="text-center max-w-5xl mx-auto px-8"
      >
        {/* Success Badge */}
        <motion.div
          initial={{ scale: 0 }}
          animate={{ scale: 1 }}
          transition={{ delay: 0.2, type: "spring" }}
          className="inline-flex items-center justify-center w-24 h-24 bg-gradient-to-r from-green-500 to-emerald-500 rounded-full mb-8 shadow-xl"
        >
          <CheckCircle2 className="w-12 h-12 text-white" />
        </motion.div>

        {/* Achievement Message */}
        <h1 className="text-5xl md:text-6xl lg:text-7xl font-black text-gray-900 mb-6">
          Section Complete
        </h1>

        {/* Next Step Segue */}
        <div className="bg-gradient-to-r from-violet-600 to-purple-600 rounded-2xl p-8 text-white">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <Rocket className="w-7 h-7" />
              <div className="text-left">
                <h3 className="text-2xl font-black">Next: New Section</h3>
                <p className="text-white/90 text-sm">Transition message here</p>
              </div>
            </div>
            <ArrowRight className="w-8 h-8" />
          </div>
        </div>
      </motion.div>
    </div>
  );
}
```

**When to use milestone screens**:
- After completing a major section (Infrastructure Hub → Service Hub)
- When transitioning between different product areas
- To provide mental break in long demos (10+ screens)
- To celebrate accomplishments and build momentum

**Design elements**:
- Green colors (success, completion)
- Checkmark or success icon
- Stats summarizing what was achieved
- Clear "What's Next" section with different color (purple/violet)
- Energizing tone, not exhausting

---

## TypeScript Patterns

### Type Definitions

**DO**: Define types inline when they're specific to one component

```tsx
type CompanyType = 'it-consulting' | 'small-product' | 'established-product';

interface CompanyTypeSelectionProps {
  onCompanyTypeSelect: (type: CompanyType) => void;
}
```

**DO**: Export types when they're shared across components

```tsx
export type CompanyType = 'it-consulting' | 'small-product' | 'established-product';
```

**DON'T**: Use `any` - always provide specific types

```tsx
// ❌ BAD
const handleSelect = (value: any) => { }

// ✅ GOOD
const handleSelect = (value: CompanyType) => { }
```

### Union Types for Screen Names

Always use union types for screen names to ensure type safety:

```tsx
type DemoScreen = 'welcome' | 'company-selection' | 'intro-problem' | 'intro-cta';
```

When adding a new screen, add it to this union type in `DemoPage.tsx`.

---

## React Hooks Guidelines

### useCallback Dependencies

**CRITICAL**: Always include ALL dependencies in useCallback arrays to avoid React Hook warnings.

**❌ WRONG** (Will cause linter errors):

```tsx
const navigateNext = useCallback(() => {
  if (!isLastScreen) {
    setCurrentScreen(demoScreens[currentScreenIndex + 1]);
  }
}, [currentScreenIndex, isLastScreen]); // Missing 'demoScreens'
```

**✅ CORRECT**:

```tsx
const navigateNext = useCallback(() => {
  if (!isLastScreen) {
    setCurrentScreen(demoScreens[currentScreenIndex + 1]);
  }
}, [currentScreenIndex, isLastScreen, demoScreens]); // All dependencies included
```

### useState Patterns

**DO**: Use descriptive state variable names

```tsx
const [currentScreen, setCurrentScreen] = useState<DemoScreen>('welcome');
const [companyType, setCompanyType] = useState<CompanyType | null>(null);
```

**DO**: Initialize with proper default values

```tsx
const [selectedType, setSelectedType] = useState<CompanyType | null>(null);
```

### useEffect for Event Listeners

Always clean up event listeners:

```tsx
useEffect(() => {
  window.addEventListener('keydown', handleKeyPress);
  return () => window.removeEventListener('keydown', handleKeyPress);
}, [handleKeyPress]);
```

---

## State Management

### Company Type State Pattern

Company type should be managed at the top level (DemoPage) and passed down:

```tsx
export default function DemoPage() {
  const [companyType, setCompanyType] = useState<CompanyType | null>(null);
  
  const handleCompanyTypeSelect = useCallback((type: CompanyType) => {
    setCompanyType(type);
    navigateNext();
  }, [navigateNext]);
  
  // Reset on home
  const goToHome = useCallback(() => {
    setCurrentScreen('welcome');
    setCompanyType(null); // Always reset company type
  }, []);
}
```

### Dynamic Flow Generation

Use functions to generate flows based on state:

```tsx
const getDemoScreens = useCallback((): DemoScreen[] => {
  if (!companyType) return defaultDemoScreens;
  
  switch (companyType) {
    case 'it-consulting':
      return itConsultingFlow;
    case 'small-product':
      return smallProductFlow;
    case 'established-product':
      return establishedProductFlow;
    default:
      return defaultDemoScreens;
  }
}, [companyType]);

const demoScreens = getDemoScreens();
```

### Conditional Rendering Based on Company Type

```tsx
case 'intro-problem':
  if (companyType === 'it-consulting') {
    return <IntroProblemITConsulting />;
  } else if (companyType === 'small-product') {
    return <IntroProblemSmallProduct />;
  } else if (companyType === 'established-product') {
    return <IntroProblemEstablished />;
  }
  return <IntroProblem />;
```

---

## Styling Conventions

### Color Palette

Always use these exact colors for consistency:

```tsx
// Dark background
className="bg-[#110D1F]"

// White cards
className="bg-white"

// Primary gradient
className="bg-gradient-to-r from-violet-600 to-purple-600"

// Text on dark backgrounds
className="text-white"
className="text-gray-300"  // Secondary text
className="text-gray-400"  // Tertiary text

// Text on light backgrounds
className="text-gray-900"  // Primary text
className="text-gray-600"  // Secondary text
className="text-gray-500"  // Tertiary text
```

### Responsive Typography

Use consistent responsive patterns:

```tsx
// Headlines
className="text-4xl md:text-5xl lg:text-6xl font-black"

// Large headlines
className="text-5xl md:text-6xl lg:text-7xl font-black"

// Body text
className="text-xl md:text-2xl"

// Regular body
className="text-xl"
```

### Custom CSS Classes

Use these pre-defined classes from `demo.css`:

```tsx
// Card styling
className="demo-card-shadow"
className="demo-card-hover"

// Button styling
className="demo-button"
className="demo-button-disabled"

// Gradient button
className="btn-gradient"
```

### Layout Patterns

**Standard screen layout** (wrapped by DemoPage):
```tsx
<div className="h-full flex flex-col justify-center p-8">
```

**Full-screen layout** (welcome/selection screens):
```tsx
<div className="min-h-screen bg-[#110D1F] flex items-center justify-center p-8">
```

---

## Animation Patterns

### Standard Fade-In Pattern

Use this for most screen content:

```tsx
<motion.div
  initial={{ opacity: 0, y: 20 }}
  animate={{ opacity: 1, y: 0 }}
  transition={{ duration: 0.6 }}
>
```

### Staggered Animations

For grid items or lists:

```tsx
{items.map((item, index) => (
  <motion.div
    key={item.id}
    initial={{ opacity: 0, y: 20 }}
    animate={{ opacity: 1, y: 0 }}
    transition={{ delay: index * 0.1 + 0.2 }}
  >
    {/* Content */}
  </motion.div>
))}
```

### Scale Animations

For badges, buttons, or emphasis elements:

```tsx
<motion.div
  initial={{ scale: 0 }}
  animate={{ scale: 1 }}
  transition={{ duration: 0.5 }}
>
```

### Hover Animations

Use Tailwind classes for simple hovers:

```tsx
className="transition-all duration-300 hover:scale-105"
```

---

## Common Linter Errors & Fixes

This section documents actual linter errors encountered and their solutions.

### 1. React Hook Exhaustive Dependencies

**Error**:
```
Warning: React Hook useCallback has a missing dependency: 'demoScreens'. 
Either include it or remove the dependency array.
```

**Cause**: 
Using a variable inside useCallback without including it in the dependency array.

**Fix**:
```tsx
// ❌ WRONG
const navigateNext = useCallback(() => {
  setCurrentScreen(demoScreens[currentScreenIndex + 1]);
}, [currentScreenIndex]); // Missing demoScreens

// ✅ CORRECT
const navigateNext = useCallback(() => {
  setCurrentScreen(demoScreens[currentScreenIndex + 1]);
}, [currentScreenIndex, demoScreens]); // All dependencies included
```

**Prevention**: 
Always check ESLint warnings and add all referenced variables to dependency arrays.

---

### 2. Unescaped Entities in JSX

**Error**:
```
Error: `'` can be escaped with `&apos;`, `&lsquo;`, `&#39;`, `&rsquo;`.
Error: `"` can be escaped with `&quot;`, `&ldquo;`, `&#34;`, `&rdquo;`.
```

**Cause**: 
Using straight quotes or apostrophes directly in JSX text content.

**Fix**:

```tsx
// ❌ WRONG
<p>We'll customize this demo</p>
<p>"We need infrastructure ready"</p>

// ✅ CORRECT
<p>We&apos;ll customize this demo</p>
<p>&ldquo;We need infrastructure ready&rdquo;</p>
```

**Common Replacements**:
- `'` → `&apos;` (apostrophe)
- `"` → `&ldquo;` (opening quote)
- `"` → `&rdquo;` (closing quote)
- Or use both: `&ldquo;...&rdquo;` for symmetry

**Prevention**: 
Always use HTML entities for quotes and apostrophes in JSX text. Use `&apos;` for contractions and `&ldquo;...&rdquo;` for quoted text.

---

### 3. Missing 'use client' Directive

**Error**:
```
Error: You're importing a component that needs useState/useEffect. 
It only works in a Client Component but none of its parents are marked with "use client"
```

**Cause**: 
Using React hooks in a component without the `'use client';` directive in Next.js App Router.

**Fix**:
```tsx
// ✅ CORRECT - Always add this at the top of interactive components
'use client';

import React, { useState } from 'react';

export default function InteractiveComponent() {
  const [state, setState] = useState(null);
  // ...
}
```

**Prevention**: 
Always start demo screen components with `'use client';` as the first line.

---

### 4. Type Mismatches

**Error**:
```
Type 'string' is not assignable to type 'CompanyType'
```

**Cause**: 
Not properly typing function parameters or return values.

**Fix**:
```tsx
// ❌ WRONG
const handleSelect = (type: string) => {
  setCompanyType(type); // Error!
}

// ✅ CORRECT
const handleSelect = (type: CompanyType) => {
  setCompanyType(type);
}
```

**Prevention**: 
Use union types and interfaces consistently. Import types when needed across files.

---

## Naming Conventions

### File Names

- **Components**: PascalCase with descriptive names
  - `CompanyTypeSelection.tsx`
  - `IntroProblemITConsulting.tsx`
  - `ServiceDeployment.tsx`

- **CSS Files**: kebab-case
  - `demo.css`

- **Documentation**: kebab-case
  - `coding-guidelines.md`
  - `README.md`

### Component Names

```tsx
// ✅ GOOD - Descriptive, clear purpose
export default function CompanyTypeSelection() { }
export default function IntroProblemITConsulting() { }

// ❌ BAD - Too generic or unclear
export default function Selection() { }
export default function Problem1() { }
```

### Variable Names

```tsx
// State variables
const [currentScreen, setCurrentScreen] = useState();
const [companyType, setCompanyType] = useState();
const [selectedProvider, setSelectedProvider] = useState();

// Handlers
const handleCompanyTypeSelect = () => { };
const handleProviderClick = () => { };
const handleContinue = () => { };

// Navigation
const navigateNext = () => { };
const navigatePrev = () => { };
const goToHome = () => { };
```

### Type Names

```tsx
// Union types
type CompanyType = 'it-consulting' | 'small-product' | 'established-product';
type DemoScreen = 'welcome' | 'company-selection' | 'intro-problem';

// Interfaces
interface CompanyTypeSelectionProps {
  onCompanyTypeSelect: (type: CompanyType) => void;
}

interface IntroCTAProps {
  onProviderSelect: (provider: string) => void;
  companyType: CompanyType;
}
```

---

## Best Practices

### 1. Component Responsibilities

**DemoPage.tsx is the orchestrator**:
- Manages all state (currentScreen, companyType)
- Defines all flows (itConsultingFlow, smallProductFlow, etc.)
- Handles navigation logic
- Renders appropriate screens

**Screen components are presentational**:
- Receive props (callbacks, data)
- Display content
- Trigger callbacks on user actions
- No direct state management of flow

### 2. Adding New Screens Checklist

When adding a new screen, complete ALL these steps:

- [ ] Create component file in appropriate directory
- [ ] Add `'use client';` directive
- [ ] Add screen name to `DemoScreen` union type in `DemoPage.tsx`
- [ ] Import component in `DemoPage.tsx`
- [ ] Add to appropriate flow array(s)
- [ ] Add case to `renderCurrentScreen()` switch statement
- [ ] Test navigation to/from the screen
- [ ] Verify responsive design
- [ ] Check for linter errors

### 3. Flow Modification Guidelines

**When modifying flows**:
1. Understand the persona's journey
2. Keep flows focused (8-14 screens)
3. Maintain narrative arc (problem → solution → success)
4. Test the entire flow after changes
5. Ensure each screen transitions smoothly

### 4. Styling Consistency

**Always**:
- Use predefined color palette
- Follow responsive typography patterns
- Use custom CSS classes for cards/buttons
- Maintain consistent spacing (px-8, mb-6, etc.)
- Test on multiple screen sizes

**Never**:
- Use arbitrary colors
- Mix dark and light text on same background
- Use inline styles when Tailwind classes exist
- Break from established design patterns

### 5. Animation Guidelines

**Do**:
- Use subtle, purposeful animations
- Keep durations under 0.6s for entry animations
- Stagger list/grid items for visual interest
- Use consistent easing (default Framer Motion easing)

**Don't**:
- Over-animate (distracting)
- Use long durations (feels slow)
- Animate every element (overwhelming)
- Mix animation styles inconsistently

### 6. Accessibility Considerations

```tsx
// Use semantic HTML
<h1> for main headlines
<p> for paragraphs
<button> for interactive elements

// Provide descriptive alt text
<img src="..." alt="AWS cloud provider logo" />

// Use proper ARIA labels when needed
<button aria-label="Navigate to next screen">
```

### 7. Performance

**Do**:
- Use `useCallback` for event handlers passed to children
- Use `useMemo` for expensive computations
- Keep component trees shallow
- Lazy load images if needed

**Don't**:
- Create inline functions in render
- Compute expensive values on every render
- Create deep nesting unnecessarily

### 8. Code Comments

**When to comment**:
```tsx
// Complex business logic
const getDemoScreens = useCallback((): DemoScreen[] => {
  // Return default flow if no company type selected yet
  if (!companyType) return defaultDemoScreens;
  
  // Return personalized flow based on company type
  switch (companyType) {
    // ...
  }
}, [companyType]);

// Non-obvious patterns
// Reset company type when going home to allow re-selection
setCompanyType(null);
```

**Don't comment**:
```tsx
// ❌ Obvious things
const [state, setState] = useState(null); // Create state
```

---

## Testing Checklist

Before considering work complete:

### Functionality
- [ ] All screens render without errors
- [ ] Navigation (next/prev/home) works correctly
- [ ] Keyboard shortcuts work (arrows, Esc)
- [ ] Company type selection works
- [ ] Flow switches correctly based on company type
- [ ] Home button resets company type

### Code Quality
- [ ] No linter errors (`yarn lint` passes)
- [ ] All TypeScript types are correct
- [ ] No console errors in browser
- [ ] All dependencies included in hooks
- [ ] Quotes/apostrophes properly escaped

### Visual/UX
- [ ] Animations are smooth
- [ ] Responsive on mobile, tablet, desktop
- [ ] Colors match design system
- [ ] Typography is consistent
- [ ] Spacing/alignment is correct
- [ ] Progress bar updates correctly

### Cross-browser
- [ ] Tested in Chrome
- [ ] Tested in Firefox
- [ ] Tested in Safari

---

## Common Pitfalls to Avoid

1. **Forgetting to reset state on home**: Always reset `companyType` when navigating home
2. **Missing dependencies in useCallback**: Include ALL referenced variables
3. **Using straight quotes in JSX**: Always use HTML entities
4. **Not updating DemoPage when adding screens**: Must update type, import, render case, and flow
5. **Mixing light/dark text colors**: Check background before choosing text color
6. **Inconsistent spacing**: Use established patterns (mb-6, px-8, etc.)
7. **Not testing full flow**: Always test from welcome to end after changes
8. **Forgetting 'use client'**: Interactive components need this directive
9. **Deep nesting**: Keep component structure flat and simple
10. **Ignoring linter warnings**: Fix them immediately, don't accumulate

---

## Quick Reference

### Essential Imports

```tsx
'use client';

import React, { useState, useCallback } from 'react';
import { motion } from 'framer-motion';
import { Icon } from 'lucide-react';
```

### Standard Component Template

```tsx
'use client';

import React from 'react';
import { motion } from 'framer-motion';

export default function ScreenName() {
  return (
    <div className="h-full flex flex-col justify-center">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6 }}
        className="text-center max-w-4xl mx-auto px-8"
      >
        {/* Content */}
      </motion.div>
    </div>
  );
}
```

### HTML Entity Quick Reference

- Apostrophe: `&apos;`
- Opening quote: `&ldquo;`
- Closing quote: `&rdquo;`
- Less than: `&lt;`
- Greater than: `&gt;`
- Ampersand: `&amp;`

---

## Getting Help

If you encounter issues not covered here:

1. Check browser console for errors
2. Run `yarn lint` to catch common issues
3. Review similar existing components
4. Check Framer Motion docs for animation issues
5. Review Next.js App Router docs for SSR/Client issues
6. Update this document with new patterns/fixes discovered

---

**Last Updated**: October 23, 2025  
**Maintainers**: Planton Engineering Team

