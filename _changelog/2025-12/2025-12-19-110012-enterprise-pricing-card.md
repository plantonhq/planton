# Enterprise Pricing Card Addition

**Date**: December 19, 2025  
**Type**: Feature  
**Components**: Pricing Page, UI Components

## Summary

Added a fourth "Enterprise" pricing tier card to the pricing page, targeting organizations that need self-hosted deployments and premium support. The card links to a Cal.com scheduling page for sales conversations. Also fixed responsiveness issues with the pricing grid to ensure proper display across all screen sizes.

## Problem Statement / Motivation

The existing pricing page had three tiers (Free, Plus, Pro) but lacked an option for larger organizations with specific requirements like self-hosted deployments, custom SLAs, and dedicated support. This gap meant potential enterprise customers had no clear path to engage with sales.

### Pain Points

- No enterprise option for organizations needing self-hosted deployments
- Missing path for customers requiring custom pricing and SLAs
- Pricing grid was not responsive—cards became unreadable on smaller screens

## Solution / What's New

### Enterprise Pricing Card

Added a new Enterprise tier with:
- **Price**: "Custom" (no fixed pricing displayed)
- **Target audience**: Organizations that need full control
- **Features**:
  - Self-hosted deployments
  - Deploy to your own cloud or on-premises
  - Premium dedicated support
  - Custom SLAs
  - Security & compliance assistance
  - White-glove onboarding
- **CTA**: "Get in Touch" button linking to `https://cal.com/swarup`

### Responsive Grid Layout

Updated the pricing card grid to be fully responsive:
- **Mobile (xs)**: 1 column—full width cards
- **Tablet (sm)**: 2 columns—side by side
- **Desktop (lg)**: 4 columns—all cards visible

## Implementation Details

### File Modified

**File**: `src/components/pricing/plans.tsx`

### Key Changes

1. **Interface Update**: Added `isEnterprise?: boolean` to `IPlanCard` interface

2. **Enterprise Plan Data**:
```typescript
{
  name: 'Enterprise',
  shortDesc: '',
  amount: 'Custom',
  featureLabel: 'For organizations that need full control.',
  features: [
    'Self-hosted deployments',
    'Deploy to your own cloud or on-premises',
    'Premium dedicated support',
    'Custom SLAs',
    'Security & compliance assistance',
    'White-glove onboarding',
  ],
  isEnterprise: true,
}
```

3. **PriceCard Component**: Added helper functions for dynamic button behavior:
```typescript
const getButtonText = () => {
  if (isFreePlan) return 'Get Started';
  if (isEnterprise) return 'Get in Touch';
  return 'Subscribe';
};

const getButtonHref = () => {
  if (isEnterprise) return 'https://cal.com/swarup';
  return 'https://planton.ai';
};
```

4. **Responsive Grid**: Changed from fixed 3-column to responsive layout:
```typescript
<Grid2
  container
  spacing={4}
  className="w-full max-w-[1400px] flex items-stretch px-4"
>
  {plans.map((plan, index) => (
    <Grid2
      size={{ xs: 12, sm: 6, lg: 3 }}
      key={index}
      className="flex"
    >
      <PriceCard {...plan} />
    </Grid2>
  ))}
</Grid2>
```

## Benefits

### For Sales
- Clear enterprise engagement path via Cal.com scheduling
- Professional presentation of enterprise capabilities
- Self-hosted option highlighted for security-conscious buyers

### For Users
- Complete pricing visibility across all tiers
- Cards readable on all devices (mobile, tablet, desktop)
- Clear differentiation between tiers

### For Marketing
- Enterprise positioning with premium features messaging
- Consistent with DevOps-in-a-Box value proposition
- Supports security & compliance narrative from landing page

## Impact

### User Experience
- Pricing page now displays correctly on mobile devices
- Enterprise buyers have a clear next step (schedule call)
- All four tiers visible without horizontal scrolling on desktop

### Conversion Path
- Enterprise leads directed to Cal.com for Swarup's calendar
- Button opens in new tab to preserve pricing page context

## Related Work

- **Landing Page Redesign**: `2025-12-19-104433-landing-page-devops-in-a-box-redesign.md` - Established security & compliance messaging now reflected in Enterprise features
- **Copywriting Revamp**: `2025-12-19-104351-landing-page-copywriting-revamp.md` - Aligned enterprise messaging with overall positioning

---

**Status**: ✅ Live (development)  
**Timeline**: Single session (~15 minutes)  
**Page**: https://planton.ai/pricing
