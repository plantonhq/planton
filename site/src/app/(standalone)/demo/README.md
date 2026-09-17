# Planton Interactive Demo

A web-based interactive demo platform for Planton, built entirely with modern web technologies to provide personalized, branching demonstrations based on company type and user needs.

## Quick Links

- **Live Demo**: [planton.ai/demo](https://planton.ai/demo)
- **Coding Guidelines**: [coding-guidelines.md](./coding-guidelines.md)
- **Concept Docs**: [concepts/](./concepts/)

## Philosophy

**We treat the demo as a full-blown web application** built with React, TypeScript, and Tailwind.

Why? As engineers, we found traditional presentation tools limiting. Building with web technologies gives us full control, enables complex branching, and keeps us in our workflow comfort zone.

**Read more**: [Why Web Technologies for Demos](./concepts/why-web.md)

## Project Status

**Current State**: Super Alpha - Continuous iteration based on demo feedback

The platform is vast - no single linear workflow can demonstrate all capabilities. We're constantly refining based on each demo session.

## Core Concepts

**Product Concepts** (What we built):
- **[Full Terraform Parity](./concepts/terraform-parity.md)**: Built for 100% Terraform-provider parity — agents always find the right field
- **[Deployment Component Store](./concepts/deployment-store.md)**: Infrastructure Lego Blocks with form-based configuration
- **[Infra Charts](./concepts/infra-charts.md)**: Orchestrated bundles of Lego Blocks deployed as a DAG

**Demo Design Concepts** (How we present):
- **[Why Web Technologies](./concepts/why-web.md)**: Philosophy behind building demo as web app
- **[Demo Narrative Structure](./concepts/demo-narrative.md)**: Three-act story with milestones
- **[Demo Journey System](./concepts/demo-journeys.md)**: Focused screens and reusable journey patterns

## Architecture

### Journey System

The demo uses a **Journey System** to break complex concepts into focused, digestible screens.

**Key Principle**: One screen = One concept

**Journey**: A reusable sequence of focused screens that tells a cohesive story.

**Example - Infra Charts Journey**:
1. `infra-charts-challenge` - The problem (production needs many resources)
2. `infra-charts-dag` - Visual demonstration (ECS DAG with 9 resources)
3. `infra-charts-concept` - Explanation (what Infra Charts are)
4. `infra-charts-impact` - Proof (customer success story)

**Benefits**:
- **Clarity**: Each screen explains one concept clearly
- **Modularity**: Journeys can be reused or reordered
- **Flexibility**: Different flows can use different journey subsets
- **Maintainability**: Update one screen without affecting others

**Implementation**: Journeys are defined in `src/components/demo/journeys.ts` and composed into company-specific flows.

See [Demo Journeys](./concepts/demo-journeys.md) for complete architecture details.

### Company Type Personas

The demo personalizes the experience based on three distinct company types, each with different pain points and value propositions:

#### 1. IT Consulting Agency

**Pain Points**:
- 3+ weeks per client project setup
- Lost deals due to slow delivery  
- Client onboarding delays

**Value Proposition**: 
- Quick infrastructure bootstrapping
- Rapid CI/CD setup
- Production-ready in minutes, not weeks

**Demo Flow Emphasis**:
- Cloud connections (quick setup)
- Lego catalog (pre-built modules with educational content)
- Component configuration (form-based, popular configuration knobs)
- Deploy logs (live progress visibility)
- Infrastructure charts (combining multiple Lego blocks)
- Service Hub (quick CI/CD)
- GitHub connection + No Dockerfile Required + Deployment

#### 2. Small Product Company

**Pain Points**:
- Cannot justify a dedicated ops team yet
- Developers spending 40% time on ops
- Infrastructure becomes scaling bottleneck

**Value Proposition**:
- Self-service platform for developers
- Run production infrastructure without growing your ops team
- Service Hub: Vercel for Backend, In Your Own Cloud

**Demo Flow Emphasis**:
- Lego catalog (no expertise needed, educational)
- Component configuration (form-based, no code required)
- Deploy logs (see deployment happen live)
- Infrastructure charts (visual composition of blocks)
- Service Hub (no CI/CD code to write)
- Service configuration (developer-friendly)
- Service deployment (one-click deploys)
- Live examples

#### 3. Established Product Company

**Pain Points**:
- 73% have compliance gaps
- 28 security incidents/year from misconfigurations
- Zero visibility into infrastructure state

**Value Proposition**:
- Governance and visibility
- Golden paths and guardrails
- Compliance and audit trails

**Demo Flow Emphasis**:
- Cloud connections
- Lego catalog (educational understanding)
- Component configuration (popular configuration knobs)
- Deploy logs (live deployment visibility)
- Infrastructure charts (combining blocks)
- Infrastructure visualization (complete visibility)
- Version history (audit trail and compliance)
- Edit flow (golden paths and standardization)
- Deploy summary (governance checks)
- Service Hub
- Live examples

### Demo Flow Structure

```
Welcome Screen
    ↓
Company Type Selection (IT Consulting | Small Product | Established Product)
    ↓
    ├─→ IT Consulting Flow (21 screens)
    ├─→ Small Product Flow (21 screens)  
    └─→ Established Product Flow (21 screens)
```

**Key Flow Pattern**: All flows follow the product's actual user journey, now organized with focused screens and reusable **Journeys**:

1. **Lego Blocks** - Individual component deployment
   - Lego Catalog (interactive selection with education)
   - Component Configuration (form-based)
   - Deploy Logs (live deployment)

2. **Infra Charts Journey** - Orchestrated infrastructure (4 focused screens)
   - The Challenge (why production is complex)
   - The DAG (visual orchestration demo)
   - The Concept (what Infra Charts are)
   - The Impact (customer success story)

3. **Chart Deployment** - Live demo with form + DAG

4. **Infrastructure Ready** - Milestone celebration and segue to Service Hub

**Journey System**: Complex concepts are broken into focused, single-concept screens organized into reusable journeys. See [Demo Journeys](./concepts/demo-journeys.md) for architecture details.

Each flow is defined as an array of screens in `DemoPage.tsx`:

```typescript
const itConsultingFlow: DemoScreen[] = [
  'welcome',
  'company-selection',
  'intro-problem',
  'intro-stakes',
  'intro-promise',
  'intro-cta',
  'cloud-connections',
  'lego-catalog',           // Interactive with education
  'component-config',        // Form screen for single component
  'deploy-logs',            // Live deployment of single component
  // Infra Charts Journey (4 focused screens)
  'infra-charts-challenge', // The problem: production needs many resources
  'infra-charts-dag',       // Visual: ECS DAG with 9 resources
  'infra-charts-concept',   // Explain: What Infra Charts are (Helm inspiration)
  'infra-charts-impact',    // Proof: Real customer success story
  'infra-chart-deploy',     // Deploy complete chart with DAG
  'infrastructure-ready',   // Infrastructure complete, segue to services
  'service-hub-intro',
  'github-connection',
  'no-dockerfile-required', // BuildPacks - automatic containerization
  'service-deployment',
  'service-success-story'
];
```

### Educational Content: Lego Catalog

The redesigned Lego Catalog screen serves dual purposes:

#### Left Side (40%) - Educational Panel
Explains key concepts while you present:

1. **Terraform Parity Visual**
   - Full-parity circular diagram
   - "Every provider resource has a home in the catalog"
   - "Every provider setting has a typed home in the manifest"

2. **What We Built**
   - Built a catalog targeting full Terraform-provider parity
   - Modeled each component's full configuration surface, essentials first
   - Built technology to capture inputs from developers

3. **How It Works**
   - Visual flow: Form Inputs → Variables → Terraform/Pulumi Modules
   - Pre-written modules in Terraform & Pulumi
   - Open source via Planton

#### Right Side (60%) - Component Selection
- 4 clickable component cards (AWS ALB, RDS, EKS, ECS)
- Visual selection state with checkmark
- "Continue" button appears when component selected
- Seamless transition to configuration form

This design lets you explain the Deployment Component Store concept while simultaneously showing real, selectable components.

### Component Configuration Flow

After selecting a component from the catalog:

1. **Configuration Screen** displays with educational header
2. User sees the **AWS ALB form** (regardless of component selected for demo simplicity)
3. Header explains: every provider setting has a typed home, essentials surface first
4. Form auto-fills to demonstrate the experience
5. Clicking "Deploy ALB" advances to **Deploy Logs** screen
6. Live deployment progress shown
7. **Then** Infra Charts introduced to show combining multiple blocks

This sequence matches the actual product experience and reinforces the educational messaging.

### Infra Charts Introduction

After users understand individual Lego Blocks and see one deploy, we introduce Infra Charts:

**Screen Content**:
1. **The Challenge**: Real environments need many interconnected resources (VPC, security groups, ALB, SSL, DNS, ECS, IAM, etc.)
2. **The Inspiration**: Kubernetes Helm Charts - package multiple resources as reusable templates
3. **The Solution**: Infra Charts - orchestrated collections of Lego Blocks deployed as a DAG
4. **Real Example**: Large, prominent ECS Chart DAG visualization
   - Full-width display showing 9 interconnected resources
   - Dependency arrows clearly visible
   - Header emphasizes: "9 resources from a simple 7-parameter form"
5. **Customer Impact**: 3 clients, ~20 min deployments, <1 hour code-to-live-URL

**Why This Works**: 
- Users already understand individual components from previous screens
- Natural progression: "You've seen one block, now see many blocks working together"
- Large DAG visualization makes dependencies immediately understandable
- Shows the complexity being handled automatically
- Real product screenshot builds credibility
- Customer success metrics demonstrate value

### Infra Chart Deployment

After explaining what charts are, we show the actual deployment experience:

**Screen Layout** (Side-by-Side):

**Left Side (40%) - Configuration Form**: 
- Parameters from values.yaml
- Fields: availability zones, domain name, service name, ECR repo, service port
- Educational header: "Your inputs will render all Lego Block configurations"
- "Deploy Chart" button at bottom

**Right Side (60%) - DAG Preview**: 
- Shows ECS chart DAG with 9 resources
- Visual representation of dependencies
- Header: "9 resources will be deployed in dependency order"
- When Deploy clicked: Stats footer appears showing "Deployment Started"
- Stats: 9 Resources, ~20 Minutes, In Progress indicator

**Key Points**:
- Form parameters (7 fields) configure entire environment (9 resources)
- DAG shows all resources and dependencies before deployment
- User sees what will be created before clicking Deploy
- When deployed: Each resource executes when dependencies are satisfied
- Parallel execution where possible
- ~20 minutes to production-ready infrastructure
- At completion, infrastructure ready for Service Hub

**Visual Impact**: 
- Side-by-side layout shows the power of templatization (7 inputs → 9 resources)
- DAG visualization makes complex dependencies understandable
- Real product screenshot builds credibility

**Segue to Service Hub**: "Infrastructure is ready. Now let's deploy services to it."

### Infrastructure Ready (Transition Screen)

A celebratory milestone screen that marks the completion of infrastructure setup:

**Visual Design**:
- Large green checkmark icon (success indicator)
- Bold headline: "Infrastructure is Production Ready"
- Subheading: "Your AWS environment is fully configured and ready. Developers can now start deploying services."

**What's Ready (3 cards)**:
1. **Networking**: VPC, subnets, security groups, load balancer
2. **Compute**: ECS cluster ready for containerized workloads
3. **Security**: SSL certificates, DNS, IAM roles configured

**Stats Panel**:
- 9 Resources Deployed
- ~20 Minutes Total Time
- ✓ Production Ready

**Segue to Service Hub**:
- Purple gradient panel with rocket icon
- "Next: Deploy Services"
- "Infrastructure is ready. Let's see how to deploy your applications to this environment."
- Clear visual cue (arrow) indicating transition

**Purpose**:
- Provides psychological "checkpoint" after complex infrastructure section
- Reinforces what was accomplished
- Creates clear mental transition: Infrastructure → Services
- Builds momentum toward next section

**Why This Works**:
- Celebrates achievement (positive reinforcement)
- Summarizes value (9 resources in 20 minutes)
- Natural break point in the narrative
- Explicit handoff between Infra Hub and Service Hub

### State Management

The demo uses React state to manage:

1. **Current Screen**: Which screen is being displayed
2. **Company Type**: Selected company type that determines the flow
3. **Selected Component**: Which component was chosen from catalog
4. **Navigation State**: Current position in the flow

**Key Pattern**: When user clicks Home, both current screen AND company type are reset, allowing them to start fresh with a different company type.

```typescript
const goToHome = useCallback(() => {
  setCurrentScreen('welcome');
  setCompanyType(null); // Reset company type when going home
  setSelectedComponent(null); // Reset selected component
}, []);
```

## Technical Stack

- **Framework**: Next.js 14+ (App Router)
- **Language**: TypeScript
- **UI Library**: React 18+
- **Styling**: Tailwind CSS
- **Animations**: Framer Motion
- **Icons**: Lucide React
- **Components**: Custom UI components + shadcn/ui patterns

## File Structure

```
src/app/(tour-demo)/demo/
├── page.tsx              # Route entry point
├── layout.tsx            # Demo-specific layout and metadata
├── demo.css              # Custom CSS (card shadows, buttons)
└── README.md            # This file

src/components/demo/
├── DemoPage.tsx          # Main orchestrator (state, navigation, flows)
├── journeys.ts           # Journey definitions and helper functions
├── demo.css              # Shared styles
│
├── intro/                # Introduction screens
│   ├── CompanyTypeSelection.tsx      # Company type picker
│   ├── IntroProblem.tsx              # Generic problem statement
│   ├── IntroProblemITConsulting.tsx  # IT consulting-specific problem
│   ├── IntroProblemSmallProduct.tsx  # Small product-specific problem
│   ├── IntroProblemEstablished.tsx   # Established product-specific problem
│   ├── IntroStakes.tsx               # What's at stake
│   ├── IntroPromise.tsx              # Our promise/solution
│   └── IntroCTA.tsx                  # Cloud provider selection
│
├── platform/             # Platform/Infrastructure screens
│   ├── CloudConnections.tsx
│   ├── LegoCatalog.tsx              # Educational + component selection
│   ├── ComponentConfiguration.tsx   # Form for selected component
│   │
│   │ # Infra Charts Journey (4 focused screens)
│   ├── InfraChartsChallenge.tsx     # Journey screen 1: The problem
│   ├── InfraChartsDAG.tsx           # Journey screen 2: Visual demo
│   ├── InfraChartsConcept.tsx       # Journey screen 3: Explanation
│   ├── InfraChartsImpact.tsx        # Journey screen 4: Customer success
│   │
│   ├── InfraChartsIntro.tsx         # Legacy (deprecated, kept for reference)
│   ├── InfraChartDeployment.tsx     # Chart deployment: Form + DAG side-by-side
│   ├── InfrastructureReady.tsx      # Milestone: Infra complete, segue to services
│   ├── DeploySummary.tsx
│   ├── DeployLogs.tsx
│   ├── InfraVersionHistory.tsx
│   ├── InfraVisualization.tsx
│   └── InfraEditFlow.tsx
│
├── service/              # Service deployment screens
│   ├── ServiceHubIntro.tsx
│   ├── GitHubConnection.tsx
│   ├── NoDockerfileRequired.tsx     # BuildPacks - automatic containerization
│   ├── ServiceConfiguration.tsx
│   ├── ServiceDeployment.tsx
│   ├── ServiceLiveExample.tsx
│   └── ServiceSuccessStory.tsx
│
├── modals/               # Modal components for forms
│   └── ...
│
├── forms/                # Form components and elements
│   └── ...
│
└── interfaces/           # TypeScript interfaces
    └── ...
```

## How to Add or Modify Demo Screens

### Adding a New Screen

1. **Create the screen component** in the appropriate directory (`intro/`, `platform/`, or `service/`)

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
        {/* Your content */}
      </motion.div>
    </div>
  );
}
```

2. **Add screen type** to `DemoPage.tsx`:

```typescript
type DemoScreen = 'welcome' | 'company-selection' | ... | 'new-feature';
```

3. **Import the component** in `DemoPage.tsx`:

```typescript
import NewFeature from './platform/NewFeature';
```

4. **Add to appropriate flow(s)**:

```typescript
const itConsultingFlow: DemoScreen[] = [
  'welcome',
  'company-selection',
  // ... other screens
  'new-feature',
  'service-success-story'
];
```

5. **Add render case** in `renderCurrentScreen()`:

```typescript
case 'new-feature':
  return <NewFeature />;
```

### Modifying an Existing Flow

Simply reorder or add/remove screens from the flow arrays:

```typescript
const smallProductFlow: DemoScreen[] = [
  'welcome',
  'company-selection',
  'intro-problem',
  // Remove a screen by deleting it
  // 'intro-stakes',  
  'intro-promise',
  'intro-cta',
  // Add a new screen
  'new-feature',
  'lego-catalog',
  // ... rest of flow
];
```

### Creating a New Company Type

1. Add the type to `CompanyTypeSelection.tsx`:

```typescript
export type CompanyType = 'it-consulting' | 'small-product' | 'established-product' | 'new-type';

const companyTypes = [
  // ... existing types
  {
    id: 'new-type' as CompanyType,
    name: 'New Company Type',
    icon: YourIcon,
    description: 'Value proposition',
    painPoints: ['Pain 1', 'Pain 2', 'Pain 3'],
    color: 'from-green-500 to-emerald-500',
  },
];
```

2. Create a problem screen for the new type:

```typescript
// src/components/demo/intro/IntroProblemNewType.tsx
export default function IntroProblemNewType() {
  // Customize for your persona
}
```

3. Create a flow in `DemoPage.tsx`:

```typescript
const newTypeFlow: DemoScreen[] = [
  'welcome',
  'company-selection',
  'intro-problem',
  // ... customize the journey
];
```

4. Update `getDemoScreens()` in `DemoPage.tsx`:

```typescript
const getDemoScreens = useCallback((): DemoScreen[] => {
  if (!companyType) return defaultDemoScreens;
  
  switch (companyType) {
    case 'new-type':
      return newTypeFlow;
    // ... other cases
  }
}, [companyType]);
```

5. Update problem screen rendering:

```typescript
case 'intro-problem':
  if (companyType === 'new-type') {
    return <IntroProblemNewType />;
  }
  // ... other cases
```

6. Update `IntroCTA.tsx` messaging:

```typescript
const getHeadline = () => {
  switch (companyType) {
    case 'new-type':
      return 'Your custom headline';
    // ... other cases
  }
};
```

## Navigation

### Keyboard Shortcuts

- `→` (Right Arrow): Next screen
- `←` (Left Arrow): Previous screen
- `Esc`: Return to home (resets demo)

### UI Controls

- **Next/Previous Buttons**: Navigate through flow
- **Home Button**: Top-right corner, resets demo and company selection
- **Progress Bar**: Shows position in current flow

## Styling Guidelines

### Color Palette

- **Background (Dark)**: `bg-[#110D1F]`
- **Card Background**: `bg-white`
- **Gradient (Primary)**: `from-violet-600 to-purple-600`
- **Gradient (Button)**: `btn-gradient` class
- **Text on Dark**: `text-white`, `text-gray-300`
- **Text on Light**: `text-gray-900`, `text-gray-600`

### Card Styling

Use consistent card patterns from `demo.css`:

```tsx
<div className="demo-card-shadow hover:demo-card-hover">
  {/* content */}
</div>
```

### Button Styling

```tsx
<button className="demo-button">
  {/* content */}
</button>
```

### Animations

Use Framer Motion for smooth transitions:

```tsx
<motion.div
  initial={{ opacity: 0, y: 20 }}
  animate={{ opacity: 1, y: 0 }}
  transition={{ duration: 0.6 }}
>
  {/* content */}
</motion.div>
```

## Development Workflow

### Running Locally

```bash
cd planton.ai
npm install
npm run dev
```

Navigate to `http://localhost:3000/demo`

### Testing a Flow

1. Start the demo
2. Select a company type
3. Navigate through the entire flow
4. Test keyboard shortcuts
5. Click Home and try a different company type
6. Verify all screens render correctly

### Gathering Feedback

After each demo session:

1. Note which screens resonated
2. Identify confusing or slow moments
3. Track questions that came up
4. Update flows or screens accordingly
5. Commit changes with descriptive messages

## Best Practices

### Screen Design

1. **Keep it focused**: One main concept per screen
2. **Visual hierarchy**: Clear headline → subheading → content → CTA
3. **Animations**: Use subtle motion to guide attention
4. **Consistency**: Maintain spacing, colors, and typography patterns

### Flow Design

1. **Tell a story**: Each flow should have a clear narrative arc
2. **8-12 screens**: Keep flows concise and engaging
3. **Persona-specific**: Emphasize features relevant to that company type
4. **Reuse components**: Share screens across flows when appropriate

### Content Guidelines

1. **Clear headlines**: Use strong, benefit-driven language
2. **Specific numbers**: "3+ weeks" is better than "a long time"
3. **Pain → Solution**: Lead with pain points, follow with solutions
4. **Real quotes**: Use actual customer feedback when possible

## Future Enhancements

### Short-term

- [ ] Add analytics to track screen engagement
- [ ] A/B test different messaging per persona
- [ ] Add video backgrounds for key screens
- [ ] Create shareable demo links with pre-selected company type
- [ ] Add "Try it yourself" interactive elements

### Medium-term

- [ ] Multi-language support
- [ ] Industry-specific sub-personas (FinTech, HealthTech, etc.)
- [ ] Custom demo paths for specific prospects
- [ ] Screen recording capability
- [ ] Email capture for follow-up

### Long-term

- [ ] Real-time collaboration (walk through demo with prospect)
- [ ] Integration with CRM to track demo engagement
- [ ] AI-driven flow recommendations based on prospect profile
- [ ] Interactive sandbox environments
- [ ] Customer success stories per persona

## Troubleshooting

### Screen not rendering

1. Check console for errors
2. Verify screen is imported in `DemoPage.tsx`
3. Ensure screen name is in `DemoScreen` type
4. Confirm case statement in `renderCurrentScreen()`

### Flow navigation broken

1. Verify screen names in flow array match type definitions
2. Check for typos in screen names
3. Ensure flow array is being returned from `getDemoScreens()`

### Styling issues

1. Check for conflicting Tailwind classes
2. Verify dark background screens use light text colors
3. Ensure card screens have proper padding and spacing
4. Test responsive breakpoints (sm, md, lg)

## Contributing

This demo is maintained by the Planton engineering team. When adding or modifying screens:

1. Test thoroughly on desktop and mobile
2. Ensure animations are smooth
3. Match existing design patterns
4. Update this README if adding new patterns
5. Get feedback from team before major changes

## Resources

- **Live Demo**: [planton.ai/demo](https://planton.ai/demo)
- **Main Site**: [planton.ai](https://planton.ai)
- **Product**: [planton.ai](https://planton.ai)
- **Framer Motion Docs**: [framer.com/motion](https://www.framer.com/motion/)
- **Tailwind CSS**: [tailwindcss.com](https://tailwindcss.com/)

---

**Remember**: This demo is a continuous work in progress. Every demo session is an opportunity to learn and improve. Don't hesitate to iterate rapidly based on feedback.
