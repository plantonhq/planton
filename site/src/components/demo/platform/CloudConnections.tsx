'use client';

import React, { useState } from 'react';
import Image from 'next/image';
import { motion, AnimatePresence } from 'framer-motion';
import { Link, CheckCircle, Loader2, ExternalLink } from 'lucide-react';
import { AwsAlb, AwsCredential } from '../interfaces';
import { FormModalRegistry } from '../modals';
import { useAutoModalAndFill } from '../../../hooks/useAutoModalAndFill';

// Constant for auto-clicking the first card on page load
const AUTO_CLICK_FIRST_CARD = true;

interface CloudProvider {
  id: string;
  name: string;
  description: string;
  color: string;
  bgColor: string;
  isConnected: boolean;
  isConnecting: boolean;
  icon: string;
}

// Icon component using real SVG files from planton-cloud
const ProviderIcon = ({ providerId }: { providerId: string }) => {
  const getIconPath = (id: string) => {
    switch (id) {
      case 'aws':
        return '/_site/images/resources/aws.svg';
      case 'gcp':
        return '/_site/images/resources/gcp.svg';
      case 'azure':
        return '/_site/images/resources/azure.svg';
      case 'cloudflare':
        return '/_site/images/resources/cloudflare.svg';
      case 'digitalocean':
        return '/_site/images/resources/digital-ocean.svg';
      default:
        return '/_site/images/resources/aws.svg';
    }
  };

  return (
    <Image
      src={getIconPath(providerId)}
      alt={`${providerId} icon`}
      width={32}
      height={32}
      className="w-8 h-8"
    />
  );
};

export default function CloudConnections() {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedProvider, setSelectedProvider] = useState<CloudProvider | null>(null);
  
  const [providers, setProviders] = useState<CloudProvider[]>([
    {
      id: 'aws',
      name: 'AWS',
      description: 'Amazon Web Services',
      color: 'from-white to-[#666]',
      bgColor: 'bg-[#151515]',
      isConnected: false,
      isConnecting: false,
      icon: 'aws',
    },
    {
      id: 'gcp',
      name: 'Google Cloud',
      description: 'Google Cloud Platform',
      color: 'from-white to-[#666]',
      bgColor: 'bg-[#151515]',
      isConnected: false,
      isConnecting: false,
      icon: 'gcp',
    },
    {
      id: 'azure',
      name: 'Azure',
      description: 'Microsoft Azure',
      color: 'from-[#999] to-[#666]',
      bgColor: 'bg-[#151515]',
      isConnected: false,
      isConnecting: false,
      icon: 'azure',
    },
    {
      id: 'cloudflare',
      name: 'Cloudflare',
      description: 'Edge network & security',
      color: 'from-[#999] to-[#666]',
      bgColor: 'bg-[#1a1a1a]',
      isConnected: false,
      isConnecting: false,
      icon: 'cloudflare',
    },
    {
      id: 'digitalocean',
      name: 'DigitalOcean',
      description: 'Simple cloud infrastructure',
      color: 'from-[#999] to-[#666]',
      bgColor: 'bg-[#151515]',
      isConnected: false,
      isConnecting: false,
      icon: 'digitalocean',
    },
  ]);

  const handleConnect = async (providerId: string) => {
    setProviders(prev =>
      prev.map(p =>
        p.id === providerId ? { ...p, isConnecting: true } : p
      )
    );

    // Simulate connection delay
    await new Promise(resolve => setTimeout(resolve, 2000));

    setProviders(prev =>
      prev.map(p =>
        p.id === providerId
          ? { ...p, isConnecting: false, isConnected: true }
          : p
      )
    );
  };

  const handleCardClick = (provider: CloudProvider) => {
    setSelectedProvider(provider);
    setIsModalOpen(true);
  };

  const handleModalClose = () => {
    setIsModalOpen(false);
    setSelectedProvider(null);
  };

  const handleModalSubmit = async (formData?: AwsCredential | AwsAlb) => {
    if (!selectedProvider) return;
    
    try {
      console.log('Modal submitted for:', selectedProvider.name);
      if (formData) {
        console.log('AWS Credential Form Data:', formData);
      }
      
      // Simulate connection process
      setProviders(prev =>
        prev.map(p =>
          p.id === selectedProvider.id ? { ...p, isConnecting: true } : p
        )
      );
      
      // Simulate API call delay
      await new Promise(resolve => setTimeout(resolve, 1500));
      
      // Update connection state after successful submission
      setProviders(prev =>
        prev.map(p =>
          p.id === selectedProvider.id
            ? { ...p, isConnecting: false, isConnected: true }
            : p
        )
      );
    } catch (error) {
      console.error('Error submitting form:', error);
      // Reset connecting state on error
      setProviders(prev =>
        prev.map(p =>
          p.id === selectedProvider.id ? { ...p, isConnecting: false } : p
        )
      );
    } finally {
      // Always close the modal after submission (success or error)
      handleModalClose();
    }
  };

  const connectedCount = providers.filter(p => p.isConnected).length;

  // Auto-click the first card on page load using the reusable hook
  useAutoModalAndFill({
    enabled: AUTO_CLICK_FIRST_CARD,
    autoClickDelay: 1500, // 1.5 seconds for better responsiveness
    onAutoClick: () => {
      const firstProvider = providers[0]; // AWS is the first provider
      handleCardClick(firstProvider);
    },
    debugPrefix: 'Cloud Connections'
  });

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="px-8 py-6 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold text-gray-900">Cloud Connections</h2>
            <p className="text-gray-600 mt-1">
              Connect your cloud providers to start deploying infrastructure
            </p>
          </div>
          <div className="flex items-center gap-4">
            <div className="text-right">
              <p className="text-sm text-gray-500">Connected</p>
              <p className="text-2xl font-bold text-gray-900">
                {connectedCount}/{providers.length}
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Provider Grid */}
      <div className="flex-1 overflow-auto p-8">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 max-w-6xl mx-auto">
          {providers.map((provider, index) => (
            <motion.div
              key={provider.id}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.1 }}
              className={`bg-white rounded-2xl border border-gray-200 demo-card-shadow hover:demo-card-hover transition-all duration-200 h-full ${
                provider.id === 'aws' ? 'cursor-pointer' : ''
              }`}
              onClick={() => {
                if (provider.id === 'aws') {
                  handleCardClick(provider);
                }
              }}
            >
              <div className="p-6 h-full flex flex-col">
                {/* Header with icon */}
                <div className="flex items-start justify-between mb-4">
                  <div className="w-12 h-12 bg-white rounded-xl flex items-center justify-center shadow-sm">
                    <ProviderIcon providerId={provider.icon} />
                  </div>
                  <AnimatePresence>
                    {provider.isConnected && (
                      <motion.div
                        initial={{ scale: 0 }}
                        animate={{ scale: 1 }}
                        className="bg-[#1a1a1a] rounded-full p-2"
                      >
                        <CheckCircle className="w-5 h-5 text-[#10b981]" />
                      </motion.div>
                    )}
                  </AnimatePresence>
                </div>

                {/* Title and description */}
                <div className="flex-1 mb-4">
                  <h3 className="font-semibold text-lg text-gray-900 mb-2">
                    {provider.name}
                  </h3>
                  <p className="text-sm text-gray-600 leading-relaxed">
                    {provider.description}
                  </p>
                </div>

                {/* Action button */}
                {!provider.isConnected ? (
                  <button
                    onClick={(e) => {
                      e.stopPropagation(); // Prevent card click
                      if (provider.id === 'aws') {
                        handleCardClick(provider); // Open modal for AWS
                      } else {
                        handleConnect(provider.id); // Direct connect for others
                      }
                    }}
                    disabled={provider.isConnecting}
                    className={`w-full demo-button ${
                      provider.isConnecting ? 'demo-button-disabled' : ''
                    }`}
                  >
                    {provider.isConnecting ? (
                      <span className="flex items-center justify-center gap-2">
                        <Loader2 className="w-4 h-4 animate-spin" />
                        Connecting...
                      </span>
                    ) : (
                      <span className="flex items-center justify-center gap-2">
                        <Link className="w-4 h-4" />
                        Connect
                      </span>
                    )}
                  </button>
                ) : (
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-[#10b981] font-medium">
                      ✓ Connected
                    </span>
                    <button className="text-gray-500 hover:text-gray-700 flex items-center gap-1">
                      Manage
                      <ExternalLink className="w-3 h-3" />
                    </button>
                  </div>
                )}
              </div>
            </motion.div>
          ))}
        </div>

        {connectedCount > 0 && (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.5 }}
            className="mt-8 text-center"
          >
            <p className="text-gray-600">
              Great! You&apos;ve connected {connectedCount} provider{connectedCount !== 1 ? 's' : ''}.
              {connectedCount < providers.length && ' You can always add more later.'}
            </p>
          </motion.div>
        )}
      </div>

      {/* Credential Modal Registry */}
      <FormModalRegistry
        providerId={selectedProvider?.id || ''}
        isOpen={isModalOpen && !!selectedProvider}
        onClose={handleModalClose}
        onSubmit={handleModalSubmit}
      />
    </div>
  );
}
