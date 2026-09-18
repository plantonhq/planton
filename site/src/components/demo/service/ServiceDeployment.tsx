'use client';

import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { GitCommit, Package, Upload, Rocket, CheckCircle, Loader2 } from 'lucide-react';

interface DeploymentStep {
  id: string;
  title: string;
  description: string;
  icon: React.ComponentType<{ className?: string }>;
  status: 'pending' | 'in-progress' | 'completed';
  duration?: string;
}

export default function ServiceDeployment() {
  const [currentStep, setCurrentStep] = useState(0);
  const [allComplete, setAllComplete] = useState(false);

  const steps: DeploymentStep[] = [
    {
      id: 'git',
      title: 'Git Commit Detected',
      description: 'New commit to main branch: "Add payment processing endpoint"',
      icon: GitCommit,
      status: 'pending',
    },
    {
      id: 'build',
      title: 'Building Container Image',
      description: 'Using Java BuildPack to create optimized container',
      icon: Package,
      status: 'pending',
      duration: '2m 15s',
    },
    {
      id: 'push',
      title: 'Pushing to ECR',
      description: 'Uploading image to Amazon Elastic Container Registry',
      icon: Upload,
      status: 'pending',
      duration: '45s',
    },
    {
      id: 'deploy',
      title: 'Deploying to EKS',
      description: 'Rolling update to production Kubernetes cluster',
      icon: Rocket,
      status: 'pending',
      duration: '1m 30s',
    },
  ];

  const [deploymentSteps, setDeploymentSteps] = useState(steps);

  useEffect(() => {
    if (currentStep < steps.length) {
      // Update current step to in-progress
      const timer1 = setTimeout(() => {
        setDeploymentSteps(prev => 
          prev.map((step, index) => 
            index === currentStep 
              ? { ...step, status: 'in-progress' as const }
              : step
          )
        );
      }, 500);

      // Complete current step and move to next
      const timer2 = setTimeout(() => {
        setDeploymentSteps(prev => 
          prev.map((step, index) => 
            index === currentStep 
              ? { ...step, status: 'completed' as const }
              : step
          )
        );
        
        if (currentStep === steps.length - 1) {
          setTimeout(() => setAllComplete(true), 1000);
        } else {
          setCurrentStep(currentStep + 1);
        }
      }, 2500);

      return () => {
        clearTimeout(timer1);
        clearTimeout(timer2);
      };
    }
  }, [currentStep, steps.length]);

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="px-8 py-6 border-b border-gray-200">
        <h2 className="text-2xl font-bold text-gray-900">Service Deployment</h2>
        <p className="text-gray-600 mt-1">
          Automated build and deployment in progress
        </p>
      </div>

      <div className="flex-1 overflow-auto bg-gradient-to-br from-gray-50 to-gray-100">
        <div className="max-w-4xl mx-auto p-8">
          {/* Deployment timeline */}
          <div className="bg-white rounded-2xl shadow-xl p-8 border border-gray-200">
            <div className="mb-8">
              <h3 className="text-xl font-bold text-gray-900 mb-2">Deployment Pipeline</h3>
              <p className="text-gray-600 text-sm">
                Automatic deployment triggered by git push to main branch
              </p>
            </div>

            {/* Steps */}
            <div className="space-y-6">
              {deploymentSteps.map((step, index) => {
                const Icon = step.icon;
                const isCompleted = step.status === 'completed';
                const isInProgress = step.status === 'in-progress';

                return (
                  <div key={step.id} className="relative">
                    {/* Connector line */}
                    {index < deploymentSteps.length - 1 && (
                      <div className="absolute left-6 top-16 bottom-0 w-0.5 bg-gray-200">
                        <motion.div
                          initial={{ height: 0 }}
                          animate={{ 
                            height: isCompleted ? '100%' : '0%',
                          }}
                          transition={{ duration: 0.5 }}
                          className="bg-gradient-to-b from-white to-[#666] w-full"
                        />
                      </div>
                    )}

                    <motion.div
                      initial={{ opacity: 0, x: -20 }}
                      animate={{ opacity: 1, x: 0 }}
                      transition={{ delay: index * 0.1 }}
                      className={`
                        relative flex items-start gap-6 p-6 rounded-xl border-2 transition-all
                        ${isCompleted 
                          ? 'bg-[#151515] border-[#2a2a2a]' 
                          : isInProgress 
                            ? 'bg-[#151515] border-[#3a3a3a] shadow-lg' 
                            : 'bg-gray-50 border-gray-200'
                        }
                      `}
                    >
                      {/* Icon */}
                      <div className={`
                        relative flex-shrink-0 w-12 h-12 rounded-full flex items-center justify-center
                        ${isCompleted 
                          ? 'bg-[#10b981]' 
                          : isInProgress 
                            ? 'bg-[#111]' 
                            : 'bg-gray-400'
                        }
                      `}>
                        <AnimatePresence mode="wait">
                          {isCompleted ? (
                            <motion.div
                              key="completed"
                              initial={{ scale: 0 }}
                              animate={{ scale: 1 }}
                              exit={{ scale: 0 }}
                            >
                              <CheckCircle className="w-6 h-6 text-white" />
                            </motion.div>
                          ) : isInProgress ? (
                            <motion.div
                              key="in-progress"
                              initial={{ scale: 0 }}
                              animate={{ scale: 1 }}
                              exit={{ scale: 0 }}
                            >
                              <Loader2 className="w-6 h-6 text-white animate-spin" />
                            </motion.div>
                          ) : (
                            <Icon className="w-6 h-6 text-white" />
                          )}
                        </AnimatePresence>
                      </div>

                      {/* Content */}
                      <div className="flex-1">
                        <div className="flex items-start justify-between mb-2">
                          <h4 className={`
                            font-bold text-lg
                            ${isCompleted 
                              ? 'text-[#10b981]' 
                              : isInProgress 
                                ? 'text-white' 
                                : 'text-gray-700'
                            }
                          `}>
                            {step.title}
                          </h4>
                          {step.duration && isCompleted && (
                            <span className="text-sm text-gray-500 font-mono">
                              {step.duration}
                            </span>
                          )}
                        </div>
                        <p className={`
                          text-sm
                          ${isCompleted 
                            ? 'text-[#10b981]' 
                            : isInProgress 
                              ? 'text-white' 
                              : 'text-gray-600'
                          }
                        `}>
                          {step.description}
                        </p>

                        {/* Progress bar for in-progress step */}
                        {isInProgress && (
                          <motion.div
                            initial={{ opacity: 0, y: 10 }}
                            animate={{ opacity: 1, y: 0 }}
                            className="mt-3"
                          >
                            <div className="h-2 bg-[#1a1a1a] rounded-full overflow-hidden">
                              <motion.div
                                initial={{ width: '0%' }}
                                animate={{ width: '100%' }}
                                transition={{ duration: 2, ease: 'linear' }}
                                className="h-full bg-gradient-to-r from-white to-[#666]"
                              />
                            </div>
                          </motion.div>
                        )}
                      </div>
                    </motion.div>
                  </div>
                );
              })}
            </div>

            {/* Completion message */}
            <AnimatePresence>
              {allComplete && (
                <motion.div
                  initial={{ opacity: 0, y: 20, scale: 0.9 }}
                  animate={{ opacity: 1, y: 0, scale: 1 }}
                  className="mt-8 bg-gradient-to-r from-[#151515] to-[#151515] border-2 border-[#2a2a2a] rounded-xl p-8 text-center"
                >
                  <motion.div
                    initial={{ scale: 0 }}
                    animate={{ scale: 1 }}
                    transition={{ delay: 0.2, type: 'spring' }}
                    className="w-16 h-16 bg-[#10b981] rounded-full flex items-center justify-center mx-auto mb-4"
                  >
                    <CheckCircle className="w-10 h-10 text-white" />
                  </motion.div>
                  <h3 className="text-2xl font-bold text-gray-900 mb-2">
                    Deployment Successful!
                  </h3>
                  <p className="text-gray-600 mb-6">
                    Your service is now live and serving traffic at{' '}
                    <span className="font-mono text-white font-semibold">
                      https://billing-api.acmecorp.com
                    </span>
                  </p>
                  <div className="flex items-center justify-center gap-4 text-sm text-gray-600">
                    <div className="flex items-center gap-2">
                      <div className="w-2 h-2 bg-[#10b981] rounded-full animate-pulse" />
                      <span>Health checks passing</span>
                    </div>
                    <div className="w-px h-4 bg-gray-300" />
                    <div>Total time: 4m 30s</div>
                  </div>
                </motion.div>
              )}
            </AnimatePresence>
          </div>

          {/* Info banner */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.5 }}
            className="mt-6 bg-gradient-to-r from-[#151515] to-[#151515] border border-[#2a2a2a] rounded-xl p-6"
          >
            <h4 className="font-bold text-gray-900 mb-2">Fully Automated Pipeline</h4>
            <p className="text-sm text-gray-600">
              Every git push triggers automatic build, test, and deployment. No Jenkins configuration, 
              no GitHub Actions workflows. Just commit your code and it&apos;s deployed to production.
            </p>
          </motion.div>
        </div>
      </div>
    </div>
  );
}

