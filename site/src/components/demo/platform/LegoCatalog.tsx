'use client';

import React, { useState } from 'react';
import { motion } from 'framer-motion';
import {
  Database,
  Server,
  Network,
  Box,
  Check,
  ArrowRight,
  Zap,
  Settings,
} from 'lucide-react';

interface LegoCatalogProps {
  onComponentSelect: (componentId: string) => void;
}

interface Component {
  id: string;
  name: string;
  description: string;
  icon: React.ComponentType<{ className?: string }>;
  color: string;
}

const components: Component[] = [
  {
    id: 'aws-alb',
    name: 'AWS ALB',
    description: 'Application Load Balancer for high availability',
    icon: Network,
    color: 'from-white to-[#666]',
  },
  {
    id: 'aws-rds',
    name: 'AWS RDS',
    description: 'Managed relational database service',
    icon: Database,
    color: 'from-white to-[#666]',
  },
  {
    id: 'aws-eks',
    name: 'AWS EKS',
    description: 'Managed Kubernetes cluster on AWS',
    icon: Server,
    color: 'from-white to-[#666]',
  },
  {
    id: 'ecs-cluster',
    name: 'ECS Cluster',
    description: 'Container orchestration with ECS',
    icon: Box,
    color: 'from-[#10b981] to-[#059669]',
  },
];

export default function LegoCatalog({ onComponentSelect }: LegoCatalogProps) {
  const [selectedComponent, setSelectedComponent] = useState<string | null>(null);

  const handleComponentClick = (componentId: string) => {
    setSelectedComponent(componentId);
  };

  const handleContinue = () => {
    if (selectedComponent) {
      onComponentSelect(selectedComponent);
    }
  };

  return (
    <div className="h-full flex">
      {/* Left Side - Educational Content (40%) */}
      <div className="w-[40%] bg-gray-50 p-8 overflow-y-auto border-r border-gray-200">
        <motion.div
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.6 }}
          className="space-y-8"
        >
          {/* Header */}
          <div>
            <h2 className="text-3xl font-black text-gray-900 mb-3">
              Deployment Component Store
            </h2>
            <p className="text-lg text-gray-600 leading-relaxed">
              Pre-built infrastructure components built for full Terraform-provider parity
            </p>
          </div>

          {/* Section 1: Terraform Parity */}
          <div className="bg-white rounded-2xl p-6 demo-card-shadow">
            <div className="flex items-center justify-center mb-6">
              <div className="relative w-32 h-32">
                {/* Full-parity Circle Diagram */}
                <svg className="w-full h-full transform -rotate-90">
                  <circle
                    cx="64"
                    cy="64"
                    r="56"
                    fill="none"
                    stroke="#E5E7EB"
                    strokeWidth="16"
                  />
                  <circle
                    cx="64"
                    cy="64"
                    r="56"
                    fill="none"
                    stroke="url(#gradient)"
                    strokeWidth="16"
                    strokeDasharray="352 0"
                    strokeLinecap="round"
                  />
                  <defs>
                    <linearGradient id="gradient" x1="0%" y1="0%" x2="100%" y2="0%">
                      <stop offset="0%" stopColor="#ededed" />
                      <stop offset="100%" stopColor="#a0a0a0" />
                    </linearGradient>
                  </defs>
                </svg>
                <div className="absolute inset-0 flex items-center justify-center">
                  <div className="text-center">
                    <div className="text-2xl font-black text-white">100%</div>
                    <div className="text-xs text-gray-500">parity target</div>
                  </div>
                </div>
              </div>
            </div>
            <div className="space-y-3">
              <div className="bg-[#151515] rounded-lg p-3">
                <p className="text-sm font-semibold text-white">
                  Every provider resource has a home in the catalog
                </p>
              </div>
              <div className="bg-[#151515] rounded-lg p-3">
                <p className="text-sm font-semibold text-white">
                  Every provider setting has a typed home in the manifest
                </p>
              </div>
            </div>
          </div>

          {/* Section 2: What We Built */}
          <div className="bg-white rounded-2xl p-6 demo-card-shadow">
            <div className="flex items-center gap-3 mb-4">
              <div className="w-10 h-10 bg-gradient-to-r from-white to-[#666] rounded-lg flex items-center justify-center">
                <Zap className="w-5 h-5 text-white" />
              </div>
              <h3 className="font-bold text-lg text-gray-900">What We Built</h3>
            </div>
            <div className="space-y-3">
              <div className="flex items-start gap-2">
                <Check className="w-5 h-5 text-[#10b981] flex-shrink-0 mt-0.5" />
                <p className="text-sm text-gray-700">
                  Built a catalog targeting full Terraform-provider parity
                </p>
              </div>
              <div className="flex items-start gap-2">
                <Check className="w-5 h-5 text-[#10b981] flex-shrink-0 mt-0.5" />
                <p className="text-sm text-gray-700">
                  Modeled each component&apos;s full configuration surface, essentials first
                </p>
              </div>
              <div className="flex items-start gap-2">
                <Check className="w-5 h-5 text-[#10b981] flex-shrink-0 mt-0.5" />
                <p className="text-sm text-gray-700">
                  Built technology to capture these inputs from developers
                </p>
              </div>
            </div>
          </div>

          {/* Section 3: How It Works */}
          <div className="bg-white rounded-2xl p-6 demo-card-shadow">
            <div className="flex items-center gap-3 mb-4">
              <div className="w-10 h-10 bg-gradient-to-r from-white to-[#666] rounded-lg flex items-center justify-center">
                <Settings className="w-5 h-5 text-white" />
              </div>
              <h3 className="font-bold text-lg text-gray-900">How It Works</h3>
            </div>
            <div className="space-y-4">
              {/* Flow Diagram */}
              <div className="flex items-center justify-between">
                <div className="flex-1 text-center">
                  <div className="w-12 h-12 bg-[#1a1a1a] rounded-lg flex items-center justify-center mx-auto mb-2">
                    <span className="text-2xl">📝</span>
                  </div>
                  <p className="text-xs font-medium text-gray-700">Form Inputs</p>
                </div>
                <ArrowRight className="w-5 h-5 text-gray-400 flex-shrink-0" />
                <div className="flex-1 text-center">
                  <div className="w-12 h-12 bg-[#1a1a1a] rounded-lg flex items-center justify-center mx-auto mb-2">
                    <span className="text-2xl">⚙️</span>
                  </div>
                  <p className="text-xs font-medium text-gray-700">Variables</p>
                </div>
                <ArrowRight className="w-5 h-5 text-gray-400 flex-shrink-0" />
                <div className="flex-1 text-center">
                  <div className="w-12 h-12 bg-[#1a1a1a] rounded-lg flex items-center justify-center mx-auto mb-2">
                    <span className="text-2xl">🚀</span>
                  </div>
                  <p className="text-xs font-medium text-gray-700">Modules</p>
                </div>
              </div>

              <div className="space-y-2 pt-2 border-t border-gray-100">
                <div className="flex items-center gap-2">
                  <div className="w-2 h-2 bg-white rounded-full"></div>
                  <p className="text-sm text-gray-700">Pre-written Terraform & Pulumi modules</p>
                </div>
                <div className="flex items-center gap-2">
                  <div className="w-2 h-2 bg-white rounded-full"></div>
                  <p className="text-sm text-gray-700">Open source via Planton</p>
                </div>
              </div>
            </div>
          </div>

          {/* Call to Action */}
          <div className="bg-gradient-to-r from-white to-[#666] rounded-2xl p-6 text-white">
            <p className="text-lg font-bold flex items-center gap-2">
              <span>Select a component to see it in action</span>
              <ArrowRight className="w-5 h-5" />
            </p>
          </div>
        </motion.div>
      </div>

      {/* Right Side - Selectable Components (60%) */}
      <div className="w-[60%] flex flex-col bg-white">
        <div className="p-8 flex-1 overflow-y-auto">
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.6 }}
          >
            <h3 className="text-2xl font-bold text-gray-900 mb-6">
              Choose a Component
            </h3>

            {/* Component Grid */}
            <div className="grid grid-cols-2 gap-6">
              {components.map((component, index) => {
                const Icon = component.icon;
                const isSelected = selectedComponent === component.id;

                return (
                  <motion.div
                    key={component.id}
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: index * 0.1 + 0.3 }}
                    onClick={() => handleComponentClick(component.id)}
                    className={`relative cursor-pointer transition-all duration-300 ${
                      isSelected ? 'scale-105' : 'hover:scale-105'
                    }`}
                  >
                    <div
                      className={`bg-white border-2 rounded-2xl p-6 demo-card-shadow hover:demo-card-hover transition-all duration-200 h-full ${
                        isSelected
                          ? 'border-[#3a3a3a] ring-4 ring-[#3a3a3a]'
                          : 'border-gray-200'
                      }`}
                    >
                      {isSelected && (
                        <motion.div
                          initial={{ scale: 0 }}
                          animate={{ scale: 1 }}
                          className="absolute -top-3 -right-3 w-8 h-8 bg-gradient-to-r from-white to-[#666] rounded-full flex items-center justify-center shadow-lg"
                        >
                          <Check className="w-5 h-5 text-white" />
                        </motion.div>
                      )}

                      <div className="flex flex-col items-center text-center gap-4">
                        <div
                          className={`w-16 h-16 bg-gradient-to-br ${component.color} rounded-2xl flex items-center justify-center shadow-lg`}
                        >
                          <Icon className="w-8 h-8 text-white" />
                        </div>
                        <div>
                          <h4 className="font-bold text-lg text-gray-900 mb-2">
                            {component.name}
                          </h4>
                          <p className="text-sm text-gray-600 leading-relaxed">
                            {component.description}
                          </p>
                        </div>
                      </div>
                    </div>
                  </motion.div>
                );
              })}
            </div>
          </motion.div>
        </div>

        {/* Bottom CTA */}
        <div className="p-6 border-t border-gray-200 bg-gray-50">
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.8 }}
            className="flex flex-col items-center gap-4"
          >
            <button
              onClick={handleContinue}
              disabled={!selectedComponent}
              className={`px-8 py-4 rounded-xl font-bold text-lg transition-all duration-300 ${
                selectedComponent
                  ? 'bg-gradient-to-r from-white to-[#666] text-white hover:from-white hover:to-[#666] shadow-lg hover:shadow-xl transform hover:-translate-y-0.5'
                  : 'bg-gray-200 text-gray-400 cursor-not-allowed'
              }`}
            >
              Continue with{' '}
              {selectedComponent
                ? components.find((c) => c.id === selectedComponent)?.name
                : 'Selection'}
            </button>

            {selectedComponent && (
              <motion.p
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                className="text-sm text-gray-500"
              >
                We&apos;ll show you the configuration form next
              </motion.p>
            )}
          </motion.div>
        </div>
      </div>
    </div>
  );
}
