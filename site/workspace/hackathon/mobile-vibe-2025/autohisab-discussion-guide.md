# AutoHisab - Phase 1 Discussion Guide

**App Name:** AutoHisab (Auto = Vehicle, Hisab = Calculation/Accounting in Hindi)

**Purpose:** Use AutoHisab as a concrete example to make ALL technical decisions needed for mobile development. By the end of this discussion, every question about platform, tools, workflow, and architecture should be answered.

**What is AutoHisab?** A fuel refill tracker that calculates vehicle mileage (fuel efficiency). Every time you refuel, you log the odometer reading and fuel quantity. The app calculates distance traveled and mileage between refills, showing fuel efficiency trends over time.

**How to use this guide:** Go through sections in order. Each decision builds on previous ones. Capture decisions as you go - these become the foundation for Phase 2 execution.

---

## 1. Product Scope & Requirements

**Start here to understand what we're building.**

### Core Use Case

**The fundamental workflow:**
1. User refuels their vehicle at a gas station
2. User opens AutoHisab
3. User logs: Current odometer reading + Fuel quantity added
4. App calculates:
   - Distance traveled = Current reading - Previous reading
   - Mileage = Distance / Fuel quantity (e.g., km/liter or MPG)
5. User sees mileage trends over time

**Example:**
- **Refill 1:** Odometer 44,500 km, added 40 liters
- **Refill 2:** Odometer 45,000 km, added 38 liters
- **Calculation:** Distance = 500 km, Mileage = 500/38 = 13.16 km/liter

---

### User Profile

**Q: Who is the primary user?**
- [ ] Personal use (individual tracking their own vehicle)
- [ ] Family use (multiple family members tracking shared vehicles)
- [ ] Business use (fleet manager tracking company vehicles)

**User: Avinash's use case** - Personal vehicle fuel efficiency tracking

**Decision:**

---

**Q: How many vehicles will a typical user track?**
- [ ] 1 vehicle (single car)
- [ ] 2-3 vehicles (family has multiple cars, bike)
- [ ] 5+ vehicles (small business fleet)

**This affects:** UI design (single vehicle view vs vehicle selector), data model

**Decision:**

---

### MVP Features

**Q: What's the absolute minimum viable product (MVP)?**

**Essential features (must have for Day 1):**
- [ ] Log fuel refill (odometer reading + fuel quantity)
- [ ] Auto-calculate mileage between refills
- [ ] View refill history (list of all refills)
- [ ] View mileage trends (chart showing efficiency over time)

**Nice to have (future enhancements):**
- [ ] Track fuel cost (price per liter, total cost)
- [ ] Track multiple vehicles
- [ ] Fuel type (petrol, diesel, electric charging)
- [ ] Location of refill (gas station name/location)
- [ ] Partial refills (tank not filled completely)
- [ ] Export data (CSV, PDF reports)
- [ ] Reminders (refuel when due, maintenance schedule)
- [ ] Photo of fuel receipt
- [ ] Compare mileage across vehicles or time periods
- [ ] Budget tracking (fuel expenses per month)

**MVP scope:**

**Future features:**

---

### Data Model

**Q: What data needs to be captured for each refill entry?**

**Core data (required):**
- Odometer reading (number, in km or miles)
- Fuel quantity (number, in liters or gallons)
- Date/time (auto-captured or manual?)

**Optional data:**
- Fuel cost (total amount spent)
- Price per liter/gallon
- Fuel type (petrol, diesel, CNG)
- Tank full? (yes/no - affects mileage calculation accuracy)
- Location (GPS or manual entry)
- Gas station name
- Notes (e.g., "city driving", "highway trip")
- Photo of receipt

**What to capture:**

---

**Q: How is mileage calculated?**

**Standard calculation:**
```
Distance = Current Odometer - Previous Odometer
Mileage = Distance / Fuel Quantity

Example:
Current: 45,000 km, Fuel: 40 liters
Previous: 44,500 km
Distance: 500 km
Mileage: 500/40 = 12.5 km/liter
```

**Edge cases to handle:**
- First refill (no previous reading - can't calculate mileage)
- Partial refills (tank not full - affects accuracy)
- Odometer rollover (999,999 → 000,000)
- Invalid data (current reading < previous reading)

**Calculation rules:**

---

### Reports and Analytics

**Q: What insights does the user need?**

**Core analytics:**
- [ ] Current mileage (latest calculated value)
- [ ] Average mileage (over last N refills or time period)
- [ ] Mileage trend chart (line graph over time)
- [ ] Best mileage (highest efficiency recorded)
- [ ] Worst mileage (lowest efficiency recorded)

**Financial analytics:**
- [ ] Total fuel expenses (this month, year, all-time)
- [ ] Cost per kilometer
- [ ] Fuel expense trends

**Comparison analytics:**
- [ ] City vs highway mileage
- [ ] Summer vs winter mileage
- [ ] Compare across vehicles

**Required reports:**

---

**Q: Do users need to share or collaborate on data?**
- [ ] No - single user, personal tracking
- [ ] Yes - same user across multiple devices (sync needed)
- [ ] Yes - family members sharing vehicle data

**This affects:** Backend requirements, authentication

**Decision:**

---

## 2. User Experience & Interaction Patterns

**How will users actually interact with AutoHisab?**

### Primary User Flow

**Q: Walk through the most common user action - logging a refill**

Scenario: User just refueled at gas station, still in car

**Option A: Quick entry**
1. Open app
2. See "Add Refill" button prominently
3. Enter odometer reading (keyboard input)
4. Enter fuel quantity (keyboard input)
5. Tap "Save"
6. See calculated mileage immediately

**Option B: Photo-assisted entry**
1. Open app
2. Tap "Add Refill"
3. Take photo of odometer
4. OCR extracts reading automatically
5. Take photo of fuel receipt
6. OCR extracts fuel quantity and cost
7. Review and confirm
8. Tap "Save"

**Option C: Voice entry**
1. Open app
2. Tap microphone button
3. Say: "Odometer 45,000, fuel 40 liters"
4. App parses and fills form
5. Review and confirm

**Preferred flow:**

---

**Q: How often will users interact with the app?**
- [ ] Every refuel (weekly for most people)
- [ ] Batch entry (log multiple refills at once)
- [ ] Daily review (check mileage trends)

**This affects:** Performance requirements, offline capabilities

**Decision:**

---

**Q: What's the home screen?**

When user opens app, what do they see first?

**Option A: Dashboard**
- Current vehicle's latest mileage (big number)
- Recent refills (last 3-5 entries)
- "Add Refill" prominent button
- Quick stats (avg mileage, total expense this month)

**Option B: Refill history**
- List of all refills (newest first)
- Each entry shows: date, odometer, fuel, calculated mileage
- Floating "+" button to add refill

**Option C: Mileage chart**
- Graph showing mileage trends
- Summary stats below
- "Add Refill" button

**Preferred home screen:**

---

**Q: Input methods - how does user enter data?**

**Odometer reading:**
- [ ] Keyboard (number pad)
- [ ] Photo + OCR (extract from odometer photo)
- [ ] Voice input
- [ ] Increment last reading (45000 → tap to adjust)

**Fuel quantity:**
- [ ] Keyboard (number pad, decimal allowed)
- [ ] Photo + OCR (extract from receipt)
- [ ] Voice input
- [ ] Common presets (30L, 40L, 50L, Full Tank)

**Decision:**

---

**Q: Must the app work offline?**

Scenarios:
- User at gas station in remote area (no signal)
- User in basement parking (no connectivity)

- [ ] Must work fully offline (critical - users refuel anywhere)
- [ ] Can require internet for sync only
- [ ] Requires constant internet

**This affects:** Backend strategy, data storage

**Decision:**

---

**Q: What devices do users need to access from?**
- [ ] Phone only (primary use case - refuel on the go)
- [ ] Tablet (larger screen for reviewing reports)
- [ ] Desktop/web (detailed analysis, fleet managers)
- [ ] Smartwatch (quick logging)

**This affects:** Responsive design, multi-platform strategy

**Decision:**

---

**Q: What's the data review/analysis workflow?**

User wants to understand their fuel efficiency:
- View mileage chart (last 6 months)
- Identify trend (improving, worsening, stable)
- Export data for tax/business purposes

**Analysis features needed:**

---

## 3. Platform Choice Decision

**THE CRITICAL DECISION: What technology do we build with?**

### Platform Options

**Option A: React Native**
- **Pros:** JavaScript (team knows it), large ecosystem, Expo framework (fast start), hot reload, web developers can contribute
- **Cons:** Performance issues with complex UIs, bridge overhead, occasional native module headaches

**Option B: Flutter**
- **Pros:** Fast performance, beautiful UI out of box, single codebase for iOS/Android/Web, strong hot reload, growing ecosystem
- **Cons:** Dart language (new learning), smaller ecosystem than React Native, larger app size

**Option C: Native (Swift/Kotlin)**
- **Pros:** Best performance, full platform access, no framework limitations
- **Cons:** Two codebases (iOS and Android), slower development, steeper learning curve

### Decision Framework

**Q: Does AutoHisab need camera access?**
- [ ] Yes (odometer photo, receipt capture with OCR)
- [ ] No (manual keyboard entry only)

**Impact:** All platforms support camera, but OCR complexity varies

---

**Q: Does it need GPS/location services?**
- [ ] Yes (track refill location, gas station mapping)
- [ ] No (location not important for mileage calculation)

**Impact:** All platforms support GPS if needed

---

**Q: Does it need background processing?**
- [ ] Yes (reminders, notifications for maintenance)
- [ ] No (foreground use only - log refills when they happen)

**Impact:** Background processing is complex, especially on iOS

---

**Q: What are the performance requirements?**
- [ ] Simple forms and lists (low complexity - just refill entries)
- [ ] Complex charts and visualizations (medium - mileage graphs)
- [ ] Real-time data processing (high complexity)

**Impact:** AutoHisab is relatively simple - forms, lists, charts

---

**Q: What does the team already know?**
- JavaScript/TypeScript expertise: ___/10
- React expertise: ___/10
- Dart expertise: ___/10
- Swift/Kotlin expertise: ___/10

**Impact:** Faster learning curve with familiar tech

---

**Q: What other mobile apps do we want to build after AutoHisab?**

List 2-3 future app ideas and their complexity:
1. Construction Expense Tracker (Sai's request) - forms, calculations, reports
2. 
3. 

**Impact:** Choose platform that works for all future apps

---

**Q: Who will maintain this code long-term?**
- [ ] Team will maintain actively
- [ ] Minimal maintenance expected
- [ ] External developers may contribute

**Impact:** Popular platforms have more community support

---

**PLATFORM DECISION:**

**Rationale:**

---

## 4. Development Environment Setup

**Based on the platform choice, what tools do we need?**

### IDE and Tools

**Q: Which IDE/editor?**

For React Native:
- [ ] VS Code (most popular, great extensions)
- [ ] WebStorm (JetBrains, powerful but paid)
- [ ] Cursor (AI-powered VS Code fork)

For Flutter:
- [ ] VS Code (with Flutter extension)
- [ ] Android Studio (official, heavier)
- [ ] IntelliJ IDEA (JetBrains)

For Native:
- [ ] Xcode (iOS - required)
- [ ] Android Studio (Android - required)

**Decision:**

---

**Q: Development machines and simulators?**

Who has what:
- Mac users (can develop iOS + Android): ___
- Windows/Linux users (Android only): ___

Simulator/emulator setup:
- [ ] iOS Simulator (Mac only)
- [ ] Android Emulator (all platforms)
- [ ] Both

**Decision:**

---

**Q: Physical device testing strategy?**
- [ ] Simulators/emulators only
- [ ] Test on real devices (which ones?)
- [ ] Mix of both

**Devices available:**

---

### Development Workflow

**Q: How fast does hot reload need to be?**

Expected workflow:
1. Make code change (e.g., adjust button color)
2. Save file (Cmd+S)
3. See result in simulator in ___ seconds

**Requirement:** 

---

**Q: What debugging tools are essential?**
- Component inspector?
- Network request logging?
- Performance profiling?
- Console logging?
- Breakpoint debugging?

**Required tools:**

---

**Q: How will we handle different API endpoints (if backend)?**
- [ ] Environment variables (.env files)
- [ ] Config files
- [ ] Build-time constants

**Decision:**

---

## 5. Project Structure & Organization

**Where does the code live and how is it organized?**

### Repository Strategy

**Q: Monorepo or separate repos?**

**Option A: Monorepo** (all hackathon apps in one repo)
```
mobile-apps/
  autohisab/
  construction-expense-tracker/
  shared/
```
- **Pros:** Share code easily, consistent tooling, single place
- **Cons:** Larger repo, need monorepo tools (Nx, Turborepo)

**Option B: Separate repos** (each app independent)
```
autohisab-app/
construction-expense-tracker-app/
```
- **Pros:** Simple, independent versioning
- **Cons:** Code duplication, harder to share

**Decision:**

---

**Q: Where does the repo live?**
- [ ] GitHub plantoncloud org
- [ ] GitHub personal accounts
- [ ] New org/account

**Repo name:**

**Repo location:**

---

### Folder Structure

**Q: How should code be organized?**

**Option A: Feature-based** (group by feature)
```
src/
  features/
    refills/
      RefillForm.tsx
      RefillList.tsx
      refillService.ts
    vehicles/
      VehicleSelector.tsx
    analytics/
      MileageChart.tsx
      StatsCard.tsx
```

**Option B: Type-based** (group by type)
```
src/
  components/
    RefillForm.tsx
    MileageChart.tsx
  services/
    refillService.ts
    calculationService.ts
  screens/
    HomeScreen.tsx
    AnalyticsScreen.tsx
```

**Decision:**

---

**Q: How will shared code be managed?**

For components, utilities, types used across features:
- [ ] Shared folder in same app
- [ ] Separate shared package
- [ ] Copy-paste (acceptable for small apps)

**Decision:**

---

**Q: Asset management strategy?**

For images, icons, fonts:
```
assets/
  images/
    gas-pump-icon.png
    empty-state.png
  icons/
  fonts/
```

- Where do assets live?
- How are they imported?
- SVG vs PNG for icons?

**Decision:**

---

## 6. Git Workflow

**How do we collaborate and manage code?**

### Branching Strategy

**Q: What branching model?**

**Option A: Simple** (good for small team)
```
main (production-ready)
feature/add-refill-form
feature/mileage-chart
```

**Option B: Git Flow** (more structure)
```
main (production)
develop (integration)
feature/xyz
release/v1.0.0
hotfix/critical-bug
```

**Decision:**

---

**Q: Branch naming convention?**
- [ ] `feature/description`
- [ ] `type/scope-description` (e.g., `feat/refill-form`)
- [ ] `your-name/description`

**Convention:**

---

### Commit Standards

**Q: Commit message format?**

**Conventional Commits:**
```
feat: add fuel refill entry form
fix: correct mileage calculation for partial refills
docs: update README with setup steps
```

- [ ] Use conventional commits
- [ ] Freeform messages
- [ ] Other format

**Decision:**

---

### Pull Request Process

**Q: PR requirements?**
- [ ] At least 1 approval required
- [ ] Can merge own PRs
- [ ] CI checks must pass
- [ ] No requirements (move fast)

**Decision:**

---

**Q: PR template?**

Should PRs have a standard template with:
- What changed
- Why it changed
- Testing done
- Screenshots (for UI changes)

**Decision:**

---

## 7. State Management

**How does data flow through AutoHisab?**

### State Types

**Q: What state does AutoHisab need?**

Identify different types:

**UI State:** 
- Refill form input values
- Modal open/closed (add refill modal)
- Selected tab (home, analytics, settings)
- Loading states

**App State:**
- Current/selected vehicle (if multiple vehicles)
- User preferences (units: km vs miles, liters vs gallons)
- Theme (light/dark mode)

**Server/Persistent State:**
- Refill entries (all fuel refill records)
- Vehicles (if multiple)
- Calculated mileage data

**State inventory:**

---

### State Management Approach

**Q: How should state be managed?**

For **React Native:**
- [ ] React Context API + hooks (simple, built-in)
- [ ] Redux Toolkit (powerful, verbose)
- [ ] Zustand (simple, fast)
- [ ] MobX (reactive)
- [ ] Jotai/Recoil (atomic state)

For **Flutter:**
- [ ] Provider (recommended, simple)
- [ ] Riverpod (modern Provider)
- [ ] Bloc (structured, verbose)
- [ ] GetX (all-in-one)
- [ ] setState (built-in, local only)

**Decision:**

**Rationale:**

---

**Q: When to use local state vs global state?**

**Local state** (component-only):
- Refill form input values (odometer, fuel quantity)
- Modal open/closed
- Dropdown expanded

**Global state** (app-wide):
- Current vehicle selection
- User preferences (units)
- Cached refill data
- Theme preference

**Guidelines:**

---

### Data Synchronization

**Q: How should local data sync with backend (if using one)?**

- [ ] Optimistic updates (update UI immediately, sync in background)
- [ ] Pessimistic updates (wait for server confirmation)
- [ ] Offline-first (local is source of truth, sync when possible)

**Decision:**

---

## 8. Backend Strategy

**THE CRITICAL DECISION: Does AutoHisab need a backend?**

### Requirements Analysis

**Q: Does AutoHisab need a backend?**

Consider requirements from Section 1:

**If single device only:**
→ No backend needed (local storage is enough)

**If multi-device sync needed:**
→ Backend required (Avinash wants to see data on phone and tablet)

**If family sharing needed:**
→ Backend required (multiple users accessing same vehicle data)

**Based on requirements:**

**Backend needed?** [ ] Yes [ ] No

---

### Storage Options

**Option A: Local Storage Only** (no backend)

**For React Native:**
- AsyncStorage (simple key-value)
- SQLite (relational database - good for refill records)
- Realm (object database)
- WatermelonDB (sync-capable, offline-first)

**For Flutter:**
- shared_preferences (simple key-value)
- SQLite (sqflite package - good for refill records)
- Hive (fast NoSQL)
- ObjectBox (object database)

**Pros:** Simple, fast, works offline, free, private
**Cons:** No sync, data lost if device lost, no backup

**Decision (if local only):**

---

**Option B: Backend as a Service (BaaS)**

**Firebase (Google)**
- **Firestore** (NoSQL database - store refill records)
- **Authentication** (email, Google login)
- **Storage** (for odometer/receipt photos if needed)
- **Analytics, Crash reporting**
- **Pros:** Fast setup, generous free tier, real-time sync, offline support
- **Cons:** Vendor lock-in, costs scale with usage, NoSQL limitations

**Supabase (open source Firebase alternative)**
- **PostgreSQL** (relational database - better for structured refill data)
- **Authentication** (email, social login)
- **Storage** (for photos)
- **Real-time subscriptions**
- **Pros:** Open source, SQL database, generous free tier, can self-host
- **Cons:** Smaller ecosystem than Firebase

**AWS Amplify**
- **DynamoDB** (NoSQL)
- **Cognito** (auth)
- **S3** (storage)
- **Pros:** Scalable, integrates with AWS
- **Cons:** Complex setup, AWS learning curve

**Decision (if BaaS):**

---

**Option C: Custom Backend API**

Build own backend with:
- **Node.js/Express** (JavaScript, fast development)
- **Go** (fast, compiled, good for APIs)
- **Python/FastAPI** (rapid development)

Database:
- **PostgreSQL** (relational, battle-tested - good for refill records)
- **SQLite** (simple file-based database)

**Pros:** Full control, custom logic, own your data
**Cons:** More work, need to host/maintain, DevOps overhead

**Decision (if custom):**

---

### Data Model for Backend

**Q: If using backend, what's the database schema?**

**Tables/Collections:**

**Vehicles:**
```
id, user_id, name, make, model, year, odometer_unit (km/miles), created_at
```

**Refills:**
```
id, vehicle_id, odometer_reading, fuel_quantity, fuel_unit (liters/gallons),
fuel_cost, price_per_unit, is_full_tank, date_time, location, notes,
calculated_mileage, created_at
```

**Users (if multi-user):**
```
id, email, name, preferences, created_at
```

**Schema:**

---

### Authentication Strategy

**Q: Does AutoHisab need user accounts?**

- [ ] No accounts (local device only, anonymous usage)
- [ ] Optional accounts (can use locally, sign in to sync)
- [ ] Required accounts (must sign in to use)

**If accounts needed:**

**Auth method:**
- [ ] Email/password
- [ ] Social login (Google, Apple)
- [ ] Phone number/SMS
- [ ] Passwordless (magic link)

**Decision:**

---

### Data Backup and Export

**Q: How do users protect their data?**

Critical for AutoHisab - users don't want to lose years of refill history!

- [ ] Automatic cloud backup (if backend)
- [ ] Manual export to CSV (can import to Excel)
- [ ] Manual export to JSON (technical backup)
- [ ] Email export (send CSV to self)
- [ ] No backup (risky!)

**Decision:**

---

### Sync Strategy

**Q: If using backend, how should sync work?**

**Offline-first (recommended for AutoHisab):**
- Local SQLite is source of truth
- User can log refills offline
- Changes queued when offline
- Sync automatically when connection available
- Conflict resolution (rare, but handle timestamp)

**Online-first:**
- Backend is source of truth
- App needs internet to function
- Simpler, but requires connectivity

**Decision:**

---

**BACKEND DECISION SUMMARY:**

**Storage:**

**Authentication:**

**Sync approach:**

**Backup/Export:**

---

## 9. Common Libraries & Dependencies

**What tools do we use for common needs?**

Based on platform and backend decisions, choose libraries:

### Navigation

**For React Native:**
- [ ] React Navigation (most popular, v6 is excellent)
- [ ] React Native Navigation (native performance)

**For Flutter:**
- [ ] Navigator 2.0 (built-in, modern)
- [ ] go_router (declarative routing)
- [ ] auto_route (code generation)

**AutoHisab needs:**
- Home screen (dashboard/refill list)
- Add Refill screen
- Analytics screen (charts)
- Settings screen
- Vehicle management (if multiple vehicles)

**Decision:**

---

### Forms and Validation

**For React Native:**
- [ ] React Hook Form (performance focused, minimal re-renders)
- [ ] Formik (popular, feature-rich)
- [ ] Manual state management

**For Flutter:**
- [ ] Built-in Form widgets
- [ ] flutter_form_builder
- [ ] Manual validation

**Validation library:**
- [ ] Yup (schema validation)
- [ ] Zod (TypeScript-first)
- [ ] Built-in

**AutoHisab refill form validations:**
- Odometer reading: required, number, >= previous reading
- Fuel quantity: required, number, > 0
- Fuel cost: optional, number, >= 0

**Decision:**

---

### Date and Time

**For React Native:**
- [ ] date-fns (modern, tree-shakeable)
- [ ] Day.js (lightweight)
- [ ] Luxon (modern Moment replacement)
- [ ] Native Intl API

**For Flutter:**
- [ ] intl package (official)
- [ ] jiffy (easy API)

**AutoHisab needs:**
- Format refill date ("Jan 15, 2025", "15/01/2025")
- Calculate days between refills
- Group refills by month (for reports)

**Decision:**

---

### Charts and Visualization

**For React Native:**
- [ ] Victory Native (React-based, powerful)
- [ ] React Native Chart Kit (simple, good for basic charts)
- [ ] Recharts (if using web)

**For Flutter:**
- [ ] fl_chart (powerful, customizable)
- [ ] charts_flutter (Google)
- [ ] syncfusion_flutter_charts (commercial, feature-rich)

**AutoHisab needs:**
- Line chart (mileage over time)
- Bar chart (fuel expenses per month)
- Simple stats cards (average, best, worst mileage)

**Decision:**

---

### Camera Access

**For React Native:**
- [ ] expo-camera (Expo managed, easy)
- [ ] react-native-camera (bare workflow)
- [ ] react-native-vision-camera (modern, performant)

**For Flutter:**
- [ ] camera (official)
- [ ] image_picker (photos + camera)

**If OCR needed for odometer/receipt:**
- [ ] ML Kit (Google, on-device)
- [ ] Tesseract (open source OCR)
- [ ] Cloud Vision API (Google Cloud, paid)

**Decision:**

---

### Local Database (if needed)

**For React Native:**
- [ ] SQLite (react-native-sqlite-storage or expo-sqlite)
- [ ] Realm
- [ ] WatermelonDB (offline-first, sync)

**For Flutter:**
- [ ] sqflite (SQLite for Flutter)
- [ ] Hive (fast NoSQL)
- [ ] ObjectBox

**AutoHisab data:**
- Refill records (id, vehicle_id, odometer, fuel, cost, date)
- Vehicles (id, name, make, model)
- Calculated mileage (can be computed on-the-fly or stored)

**Decision:**

---

### HTTP Client (if backend)

**For React Native:**
- [ ] fetch (built-in)
- [ ] axios (feature-rich, interceptors)
- [ ] React Query / TanStack Query (with caching/sync)

**For Flutter:**
- [ ] http (official)
- [ ] dio (feature-rich, interceptors)

**Decision:**

---

### UI Component Library

**For React Native:**
- [ ] React Native Paper (Material Design)
- [ ] NativeBase (customizable)
- [ ] React Native Elements (community)
- [ ] Custom components (full design control)

**For Flutter:**
- [ ] Material Widgets (built-in)
- [ ] Cupertino Widgets (iOS-style)
- [ ] Custom widgets

**AutoHisab UI needs:**
- Forms (text inputs, number inputs)
- Buttons (primary, secondary)
- Lists (refill history)
- Cards (stats, refill entries)
- Charts
- Modals/dialogs

**Decision:**

---

## 10. Build System

**How do we build and run AutoHisab?**

### Local Development

**Q: What commands start the app?**

For React Native (Expo):
```bash
npx expo start
# Then press 'i' for iOS, 'a' for Android
```

For React Native (bare):
```bash
npm start
npm run ios
npm run android
```

For Flutter:
```bash
flutter run
```

**Commands to document:**

---

**Q: What's the hot reload workflow?**

Expected experience:
1. Make code change (e.g., change button text)
2. Save file (Cmd+S)
3. App updates in simulator in ___ seconds
4. State preserved? [ ] Yes [ ] No (depends on change)

**Workflow:**

---

### Build Configurations

**Q: What build configurations are needed?**

- **Development:** Fast builds, debugging enabled, mock data?, dev logs
- **Staging:** Production-like, test with real backend (if applicable)
- **Production:** Optimized, release signing, analytics enabled

**How to switch between configs:**

---

### Environment Variables

**Q: How to manage config and secrets?**

AutoHisab might need:
- API endpoints (if backend: dev, staging, prod URLs)
- API keys (if using Google Maps, OCR services)
- Feature flags (enable/disable camera, OCR)

**For React Native:**
- [ ] react-native-config
- [ ] react-native-dotenv
- [ ] expo-constants (Expo)

**For Flutter:**
- [ ] flutter_dotenv
- [ ] Build-time constants (--dart-define)

**Decision:**

---

### Asset Bundling

**Q: How are images and fonts included?**

**For React Native:**
- Static assets in `assets/` folder
- Import: `require('./assets/gas-pump.png')`
- Metro bundler handles packaging

**For Flutter:**
- Declare in `pubspec.yaml`
- Reference: `Image.asset('assets/gas-pump.png')`

**Assets needed for AutoHisab:**
- Gas pump icon
- Empty state illustrations
- App icon
- Splash screen

**Asset strategy:**

---

### Code Signing (for production)

**Q: Who manages certificates and provisioning?**

**iOS:**
- Apple Developer account needed ($99/year)
- Certificates and provisioning profiles
- Managed manually or with Fastlane?

**Android:**
- Keystore file for signing
- Keep keystore SAFE (lost = can't update app)
- Password protected

**Decision:**

---

## 11. Testing Strategy

**How do we ensure AutoHisab works correctly?**

### Testing Philosophy

**Q: What level of testing for this hackathon app?**

- [ ] No automated tests (move fast, manual testing only)
- [ ] Critical path tests only (mileage calculation works)
- [ ] Comprehensive testing (TDD approach)

**Decision:**

---

### Unit Tests

**Q: What business logic needs unit tests?**

**Critical AutoHisab calculations:**

**Mileage calculation:**
```javascript
calculateMileage(currentOdometer, previousOdometer, fuelQuantity) {
  const distance = currentOdometer - previousOdometer;
  return distance / fuelQuantity;
}
// Test: (45000, 44500, 40) => 12.5
// Test: Handle edge cases (first refill, invalid data)
```

**Other logic to test:**
- Average mileage over N refills
- Distance traveled validation (current >= previous)
- Unit conversions (km/L to MPG)
- Date formatting

**Test framework:**

**For React Native:**
- [ ] Jest (built-in with React Native)

**For Flutter:**
- [ ] flutter_test (built-in)

**Functions to test:**

---

### Component/Widget Tests

**Q: What UI components need testing?**

- Refill form (renders, validates input, submits correctly)
- Refill list (displays data, handles empty state)
- Mileage chart (renders with data)
- Stats cards (shows correct calculations)

**Decision:**

---

### Integration/E2E Tests

**Q: What user flows need end-to-end testing?**

**Critical AutoHisab flows:**
1. First-time user → Add vehicle → Log first refill
2. Existing user → Log second refill → See calculated mileage
3. View refill history → See all entries
4. View analytics → See mileage chart
5. Export data → Download CSV

**For React Native:**
- [ ] Detox (E2E testing)
- [ ] Appium
- [ ] Manual testing only

**For Flutter:**
- [ ] integration_test (official)
- [ ] patrol (modern alternative)
- [ ] Manual testing only

**Decision:**

---

### Manual Testing Checklist

**Q: What scenarios to manually verify before shipping?**

**AutoHisab testing checklist:**
- [ ] Fresh install works
- [ ] Add vehicle (if multi-vehicle)
- [ ] Log first refill (no previous reading - handles gracefully)
- [ ] Log second refill (mileage calculated correctly)
- [ ] Log third refill (average mileage updates)
- [ ] View refill history (sorted by date, newest first)
- [ ] View mileage chart (data displays correctly)
- [ ] Edit existing refill (mileage recalculated)
- [ ] Delete refill (mileage for next entry recalculated)
- [ ] App works offline (can log refill, syncs when online)
- [ ] App syncs across devices (if backend)
- [ ] Export data (CSV downloads/emails successfully)
- [ ] Units work (km/L and MPG, liters and gallons)
- [ ] Edge cases:
  - [ ] Very first refill (no previous data)
  - [ ] Odometer reading < previous (shows error)
  - [ ] Zero or negative fuel quantity (shows error)
  - [ ] Large numbers (1,000,000 km)
- [ ] Works on both iOS and Android
- [ ] Works on different screen sizes (phone, tablet)
- [ ] Handles bad network (slow, intermittent)
- [ ] App recovers from crashes

**Testing checklist:**

---

## 12. Deployment

**How do users get AutoHisab?**

### Development Distribution

**Q: How to share AutoHisab with team during development?**

**iOS:**
- [ ] TestFlight (Apple's official, requires app review for external)
- [ ] Ad-hoc distribution (limited to 100 devices)
- [ ] Expo Go (if using Expo managed)

**Android:**
- [ ] Firebase App Distribution (easy, free)
- [ ] Google Play Internal Testing
- [ ] APK files via email/Slack
- [ ] Expo Go (if using Expo managed)

**Decision:**

---

### Over-the-Air Updates

**Q: Can we update AutoHisab without app store review?**

**For React Native:**
- [ ] CodePush (Microsoft, JS bundle updates only)
- [ ] EAS Update (Expo, JS bundle updates only)
- [ ] No OTA updates

**For Flutter:**
- [ ] Shorebird (Flutter OTA updates)
- [ ] No OTA updates

**Note:** Native code changes still need app store review

**Use case for AutoHisab:** Fix mileage calculation bug without waiting for app store review

**Decision:**

---

### Production Distribution

**Q: How will real users get AutoHisab?**

**Apple App Store:**
- Requires Apple Developer account ($99/year)
- App review process (1-2 days typically)
- Strict guidelines

**Google Play Store:**
- Requires Google Play Developer account ($25 one-time)
- Faster review process
- More lenient guidelines

**Both:**
- [ ] Publish to both stores (reach all users)
- [ ] iOS only (team uses iPhones)
- [ ] Android only (easier to publish, faster iteration)
- [ ] Internal distribution only (not public)

**Decision:**

---

### Version Management

**Q: How to version AutoHisab?**

**Semantic versioning:**
- Major.Minor.Patch (e.g., 1.0.0)
- Increment major for breaking changes (data model change)
- Increment minor for new features (add vehicle comparison)
- Increment patch for bug fixes (fix mileage calculation)

**Build number:**
- Integer that increments with each build
- iOS: CFBundleVersion
- Android: versionCode

**Decision:**

---

### Release Process

**Q: What's the step-by-step process to ship a new version?**

Draft the workflow:
1. Merge all features to `main`
2. Update version number (package.json / pubspec.yaml)
3. Update changelog (what's new in this version)
4. Run tests (if automated tests exist)
5. Build production release
   - iOS: `npx expo build:ios` or Xcode archive
   - Android: `npx expo build:android` or `./gradlew bundleRelease`
6. Upload to TestFlight / Play Internal Testing
7. Internal team testing (manual checklist)
8. Submit for app store review
9. Release when approved

**Release checklist:**

---

## DECISION SUMMARY

**Fill this in as you make decisions - this becomes your team's mobile development foundation.**

### App Overview
- **App Name:** AutoHisab
- **Core Function:** Fuel refill tracker with mileage calculation
- **Primary User:** Personal vehicle owner (Avinash's use case)
- **MVP Scope:** 

### Platform & Tools
- **Platform:** 
- **IDE:** 
- **Simulators:** 

### Project & Git
- **Repository:** 
- **Structure:** 
- **Branching:** 
- **Commits:** 

### Architecture
- **State Management:** 
- **Backend:** 
- **Storage:** 
- **Authentication:** 

### Key Libraries
- **Navigation:** 
- **Forms:** 
- **Charts:** 
- **Database:** 
- **HTTP:** 
- **Camera/OCR:** 

### Build & Deploy
- **Dev Distribution:** 
- **OTA Updates:** 
- **Production:** 
- **Version Strategy:** 

### Testing
- **Unit Tests:** 
- **E2E Tests:** 
- **Manual Checklist:** 

---

## Next Steps for Phase 2

With all decisions made, Phase 2 execution is simple:

1. **Set up dev environment** (everyone follows same setup)
2. **Create project** (using chosen platform/tools)
3. **Implement AutoHisab features:**
   - Refill entry form
   - Mileage calculation logic
   - Refill history list
   - Mileage chart
   - Export functionality
4. **Test and deploy** (follow the process defined above)

The hard part is done. Now you just write code.

---

## AutoHisab Data Flow

**To visualize for the team:**

```
User refuels at gas station
         ↓
Opens AutoHisab
         ↓
Taps "Add Refill"
         ↓
Enters: Odometer = 45,000 km, Fuel = 40 liters
         ↓
App fetches previous refill: Odometer = 44,500 km
         ↓
App calculates:
  Distance = 45,000 - 44,500 = 500 km
  Mileage = 500 / 40 = 12.5 km/liter
         ↓
Saves refill record to database
         ↓
Updates mileage chart with new data point
         ↓
User sees: "Your mileage: 12.5 km/L"
```

---

**Last Question: What did we miss?**

Review the agenda from the hackathon plan:
- [x] Mobile Platform
- [x] Dev Environment
- [x] Project Structure
- [x] Git Workflow
- [x] Build System
- [x] Deployment
- [x] State Management
- [x] Testing
- [x] Backend Strategy
- [x] Common Libraries

Anything else to decide for AutoHisab?

**Additional topics:**

---

*After this discussion, bring this document + decisions to Phase 2. The transcript of this conversation + this filled-out guide will be your prompt to Cursor AI to generate the actual AutoHisab code.*

