# Demo Narrative Structure

## The Story Arc

Every effective demo tells a story. Our demo follows a classic narrative arc with clear progression and milestones.

## Act 1: Understanding the Problem (Intro Screens)

**Screens**: Company Selection → Problem → Stakes → Promise → CTA

**Goal**: Establish context and build desire

**Structure**:
1. **Who are you?** (Company type selection)
2. **What's your pain?** (Problem screen - persona-specific)
3. **Why does it matter?** (Stakes - what's at risk)
4. **What if there was a better way?** (Promise - our solution)
5. **Let's begin** (CTA - choose cloud provider)

**Outcome**: Prospect is engaged and ready to see the solution

---

## Act 2: Infrastructure Hub (The Foundation)

**Screens**: Cloud Connections → Lego Catalog → Component Config → Deploy Logs → **Infra Charts Journey** (4 screens) → Chart Deploy → **Infrastructure Ready**

**Goal**: Show how to deploy production infrastructure without writing IaC

### Part A: Understanding Individual Components (Lego Blocks)

**Cloud Connections** → **Lego Catalog** → **Component Configuration** → **Deploy Logs**

**Narrative**:
1. "Connect your cloud account" (Security and trust)
2. "Here's our component library" (Terraform parity education)
3. "Configure via forms, not code" (Self-service)
4. "Watch it deploy live" (Transparency and confidence)

**Key Learning**: One Lego Block = One form = One live deployment

### Part B: Understanding Orchestration (Infra Charts Journey)

**Infra Charts Journey** → **Chart Deployment**

The Infra Charts Journey is a sequence of 4 focused screens, each explaining one concept:

**Screen 1: The Challenge**
- "Real environments need many interconnected resources" (The problem)
- Shows 9 resources needed for production ECS
- One focused message: Complexity of manual deployment

**Screen 2: The DAG**
- Visual demonstration of orchestration
- ECS Chart DAG with 9 resources and dependencies
- Impact: 7 inputs → 9 resources → 20 minutes

**Screen 3: The Concept**
- "We package them as Infra Charts" (The solution)
- Helm inspiration: Same pattern for infrastructure
- How it works: Template → Values → DAG → Deployment

**Screen 4: The Impact**
- Real customer success story
- IT Consulting Agency: 3 clients, 20 minutes each
- Social proof and credibility

**Then: Chart Deployment**
- Live demonstration: Form with 7 parameters
- DAG visualization shows all 9 resources
- Deploy and watch orchestration happen

**Key Learning**: Charts = Multiple Lego Blocks + Orchestration
**New Approach**: One screen = One concept = Better comprehension

### Milestone: Infrastructure Ready ✅

**Purpose of this transition screen**:

**Psychological Function**:
- Provides a "breath" after intensive infrastructure content
- Celebrates achievement (dopamine hit)
- Reinforces what was accomplished
- Creates sense of progress

**Narrative Function**:
- Clear chapter break: Infrastructure → Services
- Explicit handoff between platform components
- Sets up expectation for next section
- Builds momentum ("infrastructure done, services next")

**Educational Function**:
- Summarizes what's deployed (9 resources)
- Emphasizes speed (20 minutes)
- Highlights production-readiness
- Makes abstract deployment concrete (networking, compute, security cards)

**Why It's Critical**:
Without this screen, the transition from Infra Hub to Service Hub feels abrupt. Users might not recognize they're entering a new section with different focus. This screen:
- Makes the transition explicit
- Provides context for what comes next
- Reinforces the value already delivered
- Creates a natural pause point in the presentation

---

## Act 3: Service Hub (The Applications)

**Screens**: Service Hub Intro → GitHub Connection → BuildPacks → Service Config → Service Deploy → Live Example → Success Story

**Goal**: Show how to deploy applications without writing CI/CD code

**Narrative** (varies slightly by persona):

**For IT Consulting**:
1. "Infrastructure ready → Now deploy client applications"
2. "Connect GitHub repo" (No CI/CD setup needed)
3. "Choose BuildPacks or Dockerfile" (Flexibility)
4. "Deploy with one click" (Speed to value)
5. "Client sees live URL in their domain" (Success)

**For Small Product**:
1. "Infrastructure ready → Developers self-serve"
2. "Developers deploy self-serve"
3. "Service Hub: Vercel for Backend, In Your Own Cloud"
4. "One-click deploys" (Developer happiness)
5. "Team scales without growing the ops team" (Success)

**For Established Product**:
1. "Infrastructure ready → Standardized deployment"
2. "Golden paths for all teams"
3. "Governance and visibility built-in"
4. "Compliant, auditable deployments"
5. "Platform engineering at scale" (Success)

**Outcome**: Complete picture - Infrastructure + Services = Full DevOps platform

---

## The Power of Milestones

The **Infrastructure Ready** screen is a milestone marker. It serves multiple purposes:

### 1. Chunking Information

Humans process information in chunks. By breaking the demo into clear sections with milestones, we:
- Prevent cognitive overload
- Make complex platform digestible
- Allow prospects to mentally "save progress"

### 2. Creating "Aha!" Moments

Each milestone is an opportunity for insight:
- "Wow, that was fast" (20 minutes for 9 resources)
- "This would have taken us weeks"
- "Our developers could actually use this"

### 3. Storytelling Pacing

Good stories have rhythm: tension → release → tension → release

**Infrastructure Section**:
- Tension: "Complex infrastructure is hard to deploy"
- Education: "Here's how we make it simple"
- Action: "Watch it happen"
- **Release**: "It's done! You're ready!" ← Infrastructure Ready screen
  
**Service Section**:
- Tension: "Now you need CI/CD for applications"
- Education: "Here's our solution"
- Action: "Watch deployment"
- Release: "Success!" ← Success Story screen

### 4. Maintaining Engagement

In longer demos (18 screens), milestones:
- Break monotony
- Provide satisfaction
- Keep energy up
- Make progress visible

---

## Applying This to Other Demos

When creating new demo flows, consider:

**Do you need milestones when**:
- Demo has 10+ screens
- Multiple distinct product sections
- Complex concepts that need mental breaks
- Transitioning between major topics

**Milestone screen characteristics**:
- Celebratory tone (achievement language)
- Visual success indicators (checkmarks, green colors)
- Summary of what was accomplished
- Clear pointer to what's next
- Energizing (not exhausting)

**Don't overuse**:
- Too many milestones become repetitive
- Each should mark significant accomplishment
- Should feel earned, not arbitrary

---

## The Complete Arc

```
Act 1: Problem & Promise
  → Establish context
  → Build desire
  
Act 2: Infrastructure (Infra Hub)
  → Individual Components (educate)
  → Orchestration (demonstrate power)
  → MILESTONE: Infrastructure Ready ✅
  
Act 3: Services (Service Hub)
  → Application Deployment (complete picture)
  → Success (prove value)
```

This three-act structure, with clear milestones, creates a compelling narrative that prospects can follow and remember.

---

## The Journey System

Starting with Infra Charts, we've introduced a **Journey System** that breaks complex concepts into focused, single-concept screens.

**Why?** Cognitive science shows people process information better in focused chunks.

**How?** Instead of one dense screen with 4 sections, we create 4 focused screens that form a journey.

**Learn more**: [Demo Journeys](./demo-journeys.md) - The architecture behind focused screen sequences

## Related Concepts

- [Demo Journeys](./demo-journeys.md) - Journey system architecture and focused screens
- [Why Web](./why-web.md) - Explains why web technologies enable this narrative approach
- [Deployment Store](./deployment-store.md) - The foundation of Act 2, Part A
- [Infra Charts](./infra-charts.md) - The foundation of Act 2, Part B

