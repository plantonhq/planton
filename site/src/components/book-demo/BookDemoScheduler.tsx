'use client';

import { useEffect, useRef, useState } from 'react';
import dynamic from 'next/dynamic';
import type { DemoFormData } from './types';
import { DEMO_COPY as C } from '@/data/homepage';
import { trackDemo } from '@/lib/demo-analytics';

const Cal = dynamic(() => import('@calcom/embed-react').then((mod) => mod.default), {
  ssr: false,
  loading: () => (
    <div
      className="min-h-[600px] flex items-center justify-center text-sm text-fg-secondary"
      role="status"
    >
      {C.loading}
    </div>
  ),
});

interface BookDemoSchedulerProps {
  formData: DemoFormData;
  onBooked: () => void;
}

export function BookDemoScheduler({ formData, onBooked }: BookDemoSchedulerProps) {
  const viewed = useRef(false);
  const confirmed = useRef(false);
  const [failed, setFailed] = useState(false);
  useEffect(() => {
    let active = true;
    let dispose: (() => void) | undefined;
    if (!viewed.current) {
      trackDemo('demo_scheduler_view');
      viewed.current = true;
    }
    void import('@calcom/embed-react')
      .then(async ({ getCalApi }) => {
        const cal = await getCalApi({ namespace: '60min' });
        if (!active) return;
        // SDK event data remains local. Pending/payment-required bookings are not confirmations.
        const booked = (
          event: CustomEvent<{ data: { status?: string; paymentRequired: boolean } }>
        ) => {
          if (
            !active ||
            confirmed.current ||
            event.detail.data.status !== 'ACCEPTED' ||
            event.detail.data.paymentRequired
          )
            return;
          confirmed.current = true;
          trackDemo('demo_booking_confirmed');
          onBooked();
        };
        const linkFailed = () => {
          if (active) setFailed(true);
        };
        cal('on', { action: 'bookingSuccessfulV2', callback: booked });
        cal('on', { action: 'linkFailed', callback: linkFailed });
        cal('ui', { hideEventTypeDetails: false, layout: 'month_view', theme: 'light' });
        dispose = () => {
          cal('off', { action: 'bookingSuccessfulV2', callback: booked });
          cal('off', { action: 'linkFailed', callback: linkFailed });
        };
      })
      .catch(() => {
        if (active) setFailed(true);
      });
    return () => {
      active = false;
      dispose?.();
    };
  }, [onBooked]);

  return (
    <div>
      {!failed && (
        <div className="rounded border border-edge bg-panel">
          <Cal
            namespace="60min"
            calLink={C.calLink}
            style={{ width: '100%', minHeight: '600px' }}
            config={{
              layout: 'month_view',
              theme: 'light',
              useSlotsViewOnSmallScreen: 'true',
              name: `${formData.firstName} ${formData.lastName}`.trim(),
              email: formData.workEmail,
            }}
          />
        </div>
      )}
      <p
        className="mt-5 text-xs text-fg-secondary leading-relaxed"
        role={failed ? 'alert' : undefined}
      >
        {failed ? C.calendarFailure : C.fallback}{' '}
        <a
          href={C.calUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="underline underline-offset-4 text-fg focus-visible:outline focus-visible:outline-2"
        >
          {C.calendarLink} ↗
        </a>
      </p>
    </div>
  );
}
