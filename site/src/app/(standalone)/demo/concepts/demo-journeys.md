# Demo Journey System

## What is a Journey?

A **Journey** is a reusable sequence of focused screens that tells a cohesive story or explores a specific feature. Each journey is a modular unit that can be composed into different demo flows.

## Why Journeys?

### The Problem Before Journeys

Originally, concept explanations were packed into single, multi-section screens. For example, `InfraChartsIntro.tsx` tried to explain:
- The challenge (why production is complex)
- The DAG visualization (showing orchestration)
- The concept (what Infra Charts are)
- Customer impact (social proof)

**Result**: Cognitive overload. Too much information on one screen.

### The Solution: Focused Screens + Journeys

**One screen = One concept**

Instead of one dense screen with 4 sections, we create 4 focused screens:
1. **The Challenge** - Just the problem statement
2. **The DAG** - Just the visual demonstration
3. **The Concept** - Just the explanation
4. **The Impact** - Just the success story

These 4 screens together form the **Infra Charts Journey**.

## Architecture

### Journey Definition

```typescript
type Journey = {
  id: string;              // Unique identifier
  name: string;            // Display name
  description: string;     // What this journey covers
  screens: DemoScreen[];   // Ordered array of screens
  tags?: string[];         // Optional categorization
};
```

### Example: Infra Charts Journey

```typescript
export const infraChartsJourney: Journey = {
  id: 'infra-charts',
  name: 'Infra Charts Journey',
  description: 'Explains Infra Charts concept through focused screens',
  screens: [
    'infra-charts-challenge',  // Screen 1: The problem
    'infra-charts-dag',        // Screen 2: Visual demo
    'infra-charts-concept',    // Screen 3: Explanation
    'infra-charts-impact',     // Screen 4: Success story
  ],
  tags: ['infrastructure', 'orchestration', 'core-product'],
};
```

## Composing Flows with Journeys

Company-specific flows compose journeys and standalone screens:

```typescript
const itConsultingFlow: DemoScreen[] = [
  'welcome',
  'company-selection',
  'intro-problem',
  'intro-stakes',
  'intro-promise',
  'intro-cta',
  'cloud-connections',
  
  // Lego Blocks screens
  'lego-catalog',
  'component-config',
  'deploy-logs',
  
  // Infra Charts Journey (4 focused screens)
  'infra-charts-challenge',
  'infra-charts-dag',
  'infra-charts-concept',
  'infra-charts-impact',
  
  // Deployment
  'infra-chart-deploy',
  'infrastructure-ready',
  
  // Services
  'service-hub-intro',
  'github-connection',
  'no-dockerfile-required',
  'service-deployment',
  'service-success-story'
];
```

## Journey Characteristics

### 1. Reusability

Journeys can be used across different company type flows:
- IT Consulting flow: Full journey (all 4 screens)
- Small Product flow: Full journey (all 4 screens)
- Established Product flow: Full journey (all 4 screens)

Future flexibility: Use subset of journey screens for different contexts.

### 2. Entry and Exit Points

Journeys have clear entry and exit points:

**Entry**: The screen before the journey (e.g., `deploy-logs`)
**Journey**: The 4 focused screens
**Exit**: The screen after the journey (e.g., `infra-chart-deploy`)

Different flows can have different exit destinations.

### 3. Self-Contained Narrative

Each journey tells a complete story:
- **Beginning**: Problem/Context
- **Middle**: Solution/Demonstration
- **End**: Proof/Impact

### 4. State Independence

Journeys use the demo's shared state (e.g., `companyType`, `selectedComponent`) but don't manage journey-specific state. This keeps them simple and reusable.

## Available Journeys

### 1. Infra Charts Journey

**Purpose**: Explain how Infra Charts orchestrate multiple infrastructure components

**Screens**:
1. `infra-charts-challenge` - Production environments need many resources
2. `infra-charts-dag` - Visual: ECS DAG with 9 resources
3. `infra-charts-concept` - What Infra Charts are (Helm inspiration)
4. `infra-charts-impact` - Real customer success story

**When to use**: After showing individual component deployment, before chart deployment

**Flow position**: Between Lego Blocks demo and Chart deployment

---

### 2. Lego Blocks Journey *(Defined but not refactored yet)*

**Purpose**: Demonstrate individual infrastructure component deployment

**Screens**:
1. `lego-catalog` - Browse component library
2. `component-config` - Configure via form
3. `deploy-logs` - Watch live deployment

**When to use**: First infrastructure demo in the flow

**Flow position**: After cloud connections, before Infra Charts

---

### 3. Service Deployment Journey *(Defined but not refactored yet)*

**Purpose**: Show application deployment to infrastructure

**Screens**:
1. `service-hub-intro` - Introduction to Service Hub
2. `github-connection` - Connect GitHub repo
3. `no-dockerfile-required` - BuildPacks explanation
4. `service-deployment` - Deploy application

**When to use**: After infrastructure is ready

**Flow position**: After `infrastructure-ready` milestone

## Design Principles

### Principle 1: One Concept Per Screen

Each screen should explain exactly ONE thing:
- ✅ Good: "The Challenge" screen shows why production is complex
- ❌ Bad: "The Challenge" screen also explains the solution

### Principle 2: Progressive Disclosure

Reveal information in logical order:
1. Problem (The Challenge)
2. Visual (The DAG)
3. Explanation (The Concept)
4. Proof (The Impact)

This follows a natural learning progression.

### Principle 3: Visual Prominence

When showing visuals (like DAGs), let them dominate:
- Large images
- Minimal text
- Maximum impact

### Principle 4: Composability Over Inheritance

Journeys are composed into flows, not nested:
- ✅ Good: Flow contains Journey A screens, then Journey B screens
- ❌ Bad: Journey A contains Journey B

Keep it flat and simple.

## Helper Functions

The journey system provides helper functions in `journeys.ts`:

### Get Journey by ID
```typescript
const journey = getJourneyById('infra-charts');
```

### Compose Flows
```typescript
const flow = composeFlow([
  'welcome',
  'company-selection',
  infraChartsJourney,  // Expands to 4 screens
  'infrastructure-ready',
  serviceDeploymentJourney,  // Expands to 4 screens
  'service-success-story'
]);
```

### Check Screen Membership
```typescript
const isPartOfJourney = isScreenInJourney('infra-charts-dag', 'infra-charts');
// Returns: true
```

## Creating New Journeys

When creating a new journey:

### Step 1: Identify the Narrative

What story are you telling? What's the progression?

**Example**: Governance Journey
- Screen 1: The compliance challenge
- Screen 2: Built-in audit trails
- Screen 3: Policy enforcement
- Screen 4: Reporting and dashboards

### Step 2: Create Focused Screen Components

Each screen = one `.tsx` file with one concept:
- `GovernanceChallenge.tsx`
- `GovernanceAuditTrails.tsx`
- `GovernancePolicies.tsx`
- `GovernanceReporting.tsx`

### Step 3: Define the Journey

In `journeys.ts`:

```typescript
export const governanceJourney: Journey = {
  id: 'governance',
  name: 'Governance Journey',
  description: 'Explains compliance and governance features',
  screens: [
    'governance-challenge',
    'governance-audit-trails',
    'governance-policies',
    'governance-reporting',
  ],
  tags: ['governance', 'compliance', 'enterprise'],
};
```

### Step 4: Add to Flows

Insert the journey into appropriate company flows:

```typescript
const establishedProductFlow: DemoScreen[] = [
  // ... intro screens
  // ... infrastructure journey
  'infrastructure-ready',
  
  // Governance Journey for enterprise customers
  'governance-challenge',
  'governance-audit-trails',
  'governance-policies',
  'governance-reporting',
  
  // ... service deployment
];
```

### Step 5: Update DemoScreen Type

Add new screen types to `DemoPage.tsx`:

```typescript
type DemoScreen = 
  // ... existing screens
  | 'governance-challenge'
  | 'governance-audit-trails'
  | 'governance-policies'
  | 'governance-reporting';
```

### Step 6: Add Rendering Logic

In `renderCurrentScreen()`:

```typescript
case 'governance-challenge':
  return <GovernanceChallenge />;

case 'governance-audit-trails':
  return <GovernanceAuditTrails />;

// ... etc
```

## Benefits of the Journey System

### 1. Cognitive Clarity

One screen = One concept = Easier to understand

Prospects can process information without overwhelm.

### 2. Flexible Pacing

Different flows can:
- Use full journey (all screens)
- Use subset of journey (skip screens)
- Reorder journey screens for different narratives

### 3. Easier Maintenance

Update one journey screen without affecting others:
- Changed customer stats? Update `InfraChartsImpact.tsx` only
- New DAG visualization? Update `InfraChartsDAG.tsx` only

### 4. A/B Testing Ready

Test different orderings:
- Does Challenge → DAG → Concept work better?
- Or Concept → DAG → Challenge?

Easy to experiment with journey screen order.

### 5. Analytics Insights

Track engagement per journey screen:
- Which screens have highest drop-off?
- Which screens correlate with conversions?
- Which journeys resonate most?

## Future Journey Ideas

As the demo evolves, consider these journeys:

### Multi-Cloud Journey
1. The challenge: Managing multiple clouds
2. Unified interface across clouds
3. Cloud-agnostic abstractions
4. Real multi-cloud deployment

### Developer Experience Journey
1. The pain: Complex DevOps requirements
2. Self-service platform
3. One-click deployments
4. Developer testimonials

### Security & Compliance Journey
1. Security requirements in production
2. Built-in security controls
3. Compliance frameworks supported
4. Audit and reporting

### Cost Optimization Journey
1. Cloud cost challenges
2. Resource right-sizing
3. Cost visibility and alerts
4. Savings achieved

## Related Concepts

- [Demo Narrative Structure](./demo-narrative.md) - Overall story arc and milestones
- [Why Web Technologies](./why-web.md) - Why we can build complex journey systems
- [Infra Charts](./infra-charts.md) - The feature explained in our first refactored journey

---

## Summary

**Journeys transform the demo from monolithic screens to focused, reusable narrative units.**

Each journey:
- Tells one cohesive story
- Uses focused, single-concept screens
- Can be composed into different flows
- Maintains clear entry/exit points
- Enables flexible, personalized demos

This architecture makes the demo more maintainable, testable, and effective at explaining complex concepts.

