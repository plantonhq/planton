'use client';

import type { DoorId } from '@/data/doors';
import { sendGAEvent } from '@next/third-parties/google';

export type DemoEvent =
  | 'demo_cta_click'
  | 'demo_form_start'
  | 'demo_form_submit_success'
  | 'demo_scheduler_view'
  | 'demo_booking_confirmed';
export type DemoLocation = 'hero' | 'controls' | 'close' | 'booking' | 'navigation';

/** Closed vocabulary: no form contents, calendar payloads, or identifying URLs. */
export function trackDemo(event: DemoEvent, location: DemoLocation = 'booking') {
  sendGAEvent('event', event, { location });
}

/** A small, explicit vocabulary; never pass form fields, URLs, or SDK payloads. */
export function trackExperience(
  event: 'self_service_click' | 'product_example_select',
  context: {
    location?: DemoLocation | 'agents';
    door?: DoorId;
    selection?: 'groups' | 'whole' | 'resources';
  }
) {
  sendGAEvent('event', event, context);
}
