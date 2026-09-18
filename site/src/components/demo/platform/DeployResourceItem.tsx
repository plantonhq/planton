'use client';

import * as React from 'react';
import { motion } from 'framer-motion';

interface DeployResourceItemProps {
  item: {
    id: string;
    name: string;
    icon: React.ComponentType<{ className?: string }>;
    color: string;
    config: Record<string, unknown>;
    cost: string;
  };
  index: number;
}

export default function DeployResourceItem({ item, index }: DeployResourceItemProps) {
  const Icon = item.icon;

  return (
    <motion.div
      key={item.id}
      initial={{ opacity: 0, x: -20 }}
      animate={{ opacity: 1, x: 0 }}
      transition={{ delay: index * 0.1 }}
      className="bg-white border border-gray-200 rounded-xl p-6"
    >
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Icon className="w-16 h-16 flex-shrink-0" />
          <div>
            <h4 className="font-bold text-gray-900">{item.name}</h4>
            <div className="flex items-center gap-4 mt-1">
              {Object.entries(item.config).map(([key, value]) => (
                <span key={key} className="text-sm text-gray-600">
                  {key}: <span className="font-medium">{String(value)}</span>
                </span>
              ))}
            </div>
          </div>
        </div>
        <div className="text-right">
          <p className="text-lg font-bold text-gray-900">{item.cost}</p>
          <p className="text-xs text-gray-500">estimated</p>
        </div>
      </div>
    </motion.div>
  );
}
