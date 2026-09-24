'use client';

import { Slide, SlideHeader, Card, CardTitle, Callout } from '@/components/deck';

const topStats = [
  { value: '30d → 30m', label: 'Environment setup time' },
  { value: '$1M/year', label: 'Cloud savings (GCP)' },
  { value: 'Days → Hrs', label: 'Process modernization' },
];

export default function S19ROI() {
  return (
    <Slide>
      <SlideHeader title="The Business Case" />

      <div className="max-w-3xl mx-auto mb-8">
        <div className="grid grid-cols-3 gap-4">
          {topStats.map((stat, index) => (
            <div
              key={index}
              className="text-center p-4 sm:p-5 bg-white/[0.03] rounded-xl border border-white/10"
            >
              <div className="text-xl sm:text-2xl md:text-3xl font-bold text-white mb-2">
                {stat.value}
              </div>
              <div className="text-xs sm:text-sm text-white/50">{stat.label}</div>
            </div>
          ))}
        </div>
      </div>

      <Card variant="success" className="max-w-3xl mx-auto text-left">
        <CardTitle className="text-[#10b981] mb-4">
          Example ROI Calculation
        </CardTitle>
        <p className="text-white/70 text-sm mb-4">
          If SEP takes on <strong className="text-white">20 new clients per year</strong>{' '}
          and Planton saves <strong className="text-white">4 weeks per client</strong>:
        </p>

        <div className="grid grid-cols-3 gap-4 text-center mb-4">
          <div>
            <div className="text-2xl sm:text-3xl font-bold text-white">
              80 weeks
            </div>
            <p className="text-xs text-white/50">reclaimed</p>
          </div>
          <div>
            <div className="text-2xl sm:text-3xl font-bold text-white">
              3,200 hours
            </div>
            <p className="text-xs text-white/50">saved</p>
          </div>
          <div>
            <div className="text-2xl sm:text-3xl font-bold text-white">
              $640,000
            </div>
            <p className="text-xs text-white/50">opportunity cost recovered</p>
          </div>
        </div>

        <Callout className="bg-white/5">
          <p className="text-white/60 text-sm text-center">
            Even a <strong className="text-white">10% improvement</strong> = $64K/year
            in reclaimed time
          </p>
        </Callout>
      </Card>
    </Slide>
  );
}
