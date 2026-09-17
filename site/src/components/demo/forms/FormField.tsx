'use client';

import React from 'react';
import { Tooltip } from '../ui/tooltip';

interface FormFieldProps {
  label: string;
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  type?: 'text' | 'password' | 'email' | 'number';
  helpText?: string;
  required?: boolean;
  disabled?: boolean;
}

export const FormField: React.FC<FormFieldProps> = ({
  label,
  value,
  onChange,
  placeholder,
  type = 'text',
  helpText,
  required = false,
  disabled = false,
}) => {
  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 mb-2">
        {label}
        {required && <span className="text-[#ef4444] ml-1">*</span>}
        {label && helpText && (
          <Tooltip content={helpText} position="top">
            <svg
              className="inline w-4 h-4 ml-1 text-gray-400 hover:text-gray-600 cursor-help"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
          </Tooltip>
        )}
      </label>
      <input
        type={type}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        disabled={disabled}
        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-white focus:border-[#3a3a3a] bg-white text-gray-900 disabled:bg-gray-100 disabled:cursor-not-allowed"
      />
    </div>
  );
};
