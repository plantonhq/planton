'use client';

import { sendGAEvent } from '@next/third-parties/google';

export type DemoEvent = 'demo_cta_click' | 'demo_form_start' | 'demo_form_submit_success' | 'demo_scheduler_view' | 'demo_booking_confirmed';
export type DemoLocation = 'hero' | 'controls' | 'close' | 'booking';

/** Closed vocabulary: no form contents, calendar payloads, or identifying URLs. */
export function trackDemo(event: DemoEvent, location: DemoLocation = 'booking') {
  sendGAEvent('event', event, { location });
}
