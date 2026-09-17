import React from 'react';
import { getAllChangelogEntries } from '@/lib/changelog';
import { ChangelogTimeline } from '@/components/changelog';

export default function ChangelogPage() {
  const entries = getAllChangelogEntries();

  return (
    <div className="min-h-screen font-inter antialiased">
      <div className="max-w-4xl mx-auto px-4 py-8">
        {/* Header */}
        <header className="mb-10">
          <h1 className="text-4xl font-bold text-white mb-3">Changelog</h1>
          <p className="text-lg text-[#a0a0a0]">
            New features, improvements, and fixes across the Planton platform.
          </p>
        </header>

        {/* Timeline */}
        <ChangelogTimeline entries={entries} />
      </div>
    </div>
  );
}
