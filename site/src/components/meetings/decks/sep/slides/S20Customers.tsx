'use client';

import { Slide, SlideHeader, Card, CardTitle, TwoCol } from '@/components/deck';

const topStats = [
  { value: '100%', label: 'Retention (2023+)' },
];

export default function S20Customers() {
  return (
    <Slide>
      <SlideHeader title="Customer Feedback" />

      <div className="max-w-4xl mx-auto mb-8">
        <div className="grid grid-cols-1 gap-4 max-w-xs mx-auto">
          {topStats.map((stat, index) => (
            <div
              key={index}
              className="text-center p-4 sm:p-5 bg-white/[0.03] rounded-xl border border-white/10"
            >
              <div className="text-2xl sm:text-3xl md:text-4xl font-bold text-white mb-2">
                {stat.value}
              </div>
              <div className="text-xs sm:text-sm text-white/50">{stat.label}</div>
            </div>
          ))}
        </div>
      </div>

      <TwoCol className="max-w-4xl mx-auto text-left">
        <Card>
          <CardTitle className="text-white mb-3">
            TynyBay → Odwen (Client)
          </CardTitle>
          <ul className="space-y-2 text-sm text-white/60 mb-4">
            <li>
              <strong className="text-white">Use Case:</strong> Online
              warehousing platform on GCP
            </li>
            <li>
              <strong className="text-white">Team:</strong> 1 DevOps engineer
              managing 8+ client projects
            </li>
            <li>
              <strong className="text-white">Before:</strong> 1-2 weeks
              infrastructure setup per client
            </li>
            <li>
              <strong className="text-white">With Planton:</strong> &lt;1 hour
              infrastructure setup
            </li>
          </ul>
          <p className="text-white italic text-sm">
            &quot;I no longer have to rewrite Terraform between client projects.&quot;
          </p>
        </Card>

        <Card>
          <CardTitle className="text-white mb-3">iorta TechNext</CardTitle>
          <ul className="space-y-2 text-sm text-white/60 mb-4">
            <li>
              <strong className="text-white">Use Case:</strong> SalesVerse - BFSI
              sales platform (India)
            </li>
            <li>
              <strong className="text-white">Team:</strong> 7 developers, 1
              junior DevOps
            </li>
          </ul>
          <p className="text-[#10b981] text-sm">
            Runs production infrastructure without growing their ops team.
          </p>
        </Card>
      </TwoCol>
    </Slide>
  );
}
