'use client';

import React from 'react';
import Image from 'next/image';
import { motion } from 'framer-motion';
import { GitBranch } from 'lucide-react';

export default function InfraChartsDAG() {
  return (
    <div className="h-full overflow-y-auto bg-gradient-to-br from-slate-50 to-gray-100">
      <div className="max-w-7xl mx-auto p-8 space-y-8">
        {/* Header */}
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6 }}
          className="text-center"
        >
          <div className="inline-flex items-center justify-center w-16 h-16 bg-gradient-to-br from-[#999] to-red-600 rounded-2xl mb-4">
            <GitBranch className="w-8 h-8 text-white" />
          </div>
          <h1 className="text-4xl md:text-5xl font-black text-gray-900 mb-4">
            Orchestrated Deployment
          </h1>
          <p className="text-xl text-gray-600 max-w-3xl mx-auto">
            9 interconnected resources deployed as a Directed Acyclic Graph (DAG)
          </p>
        </motion.div>

        {/* DAG Visualization */}
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ delay: 0.2, duration: 0.6 }}
          className="bg-white rounded-2xl border-2 border-gray-200 overflow-hidden demo-card-shadow"
        >
          <div className="bg-gradient-to-r from-[#999] to-red-600 px-8 py-6">
            <h3 className="font-bold text-2xl text-white mb-2">AWS ECS Infra Chart</h3>
            <p className="text-white/90 text-lg">
              Dependencies automatically resolved. Resources deployed in topological order.
            </p>
          </div>
          <div className="p-8 bg-gray-50">
            <Image
              src="/demo/_site/images/ecs-chart-dag.png"
              alt="ECS Infra Chart DAG showing VPC, Route53, security groups, ALB, certificates, ECS cluster, IAM roles, ECR repo, and ECS service with their dependency arrows and deployment order"
              width={2674}
              height={1430}
              className="w-full h-auto rounded-xl border-2 border-gray-200 demo-card-shadow"
            />
          </div>
        </motion.div>

        {/* Key Stats */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.4, duration: 0.6 }}
          className="grid md:grid-cols-3 gap-6"
        >
          <div className="bg-white rounded-2xl p-8 text-center border-2 border-gray-200 demo-card-shadow">
            <div className="text-5xl font-black bg-gradient-to-r from-[#999] to-red-600 bg-clip-text text-transparent mb-2">
              9
            </div>
            <div className="text-gray-600 font-medium">Resources</div>
          </div>
          <div className="bg-white rounded-2xl p-8 text-center border-2 border-gray-200 demo-card-shadow">
            <div className="text-5xl font-black bg-gradient-to-r from-[#999] to-red-600 bg-clip-text text-transparent mb-2">
              7
            </div>
            <div className="text-gray-600 font-medium">Input Parameters</div>
          </div>
          <div className="bg-white rounded-2xl p-8 text-center border-2 border-gray-200 demo-card-shadow">
            <div className="text-5xl font-black bg-gradient-to-r from-[#999] to-red-600 bg-clip-text text-transparent mb-2">
              ~20
            </div>
            <div className="text-gray-600 font-medium">Minutes to Deploy</div>
          </div>
        </motion.div>
      </div>
    </div>
  );
}

