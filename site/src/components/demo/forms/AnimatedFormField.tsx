'use client';

import React, { useContext } from 'react';
import { AutoFillContext } from '../../../contexts/AutoFillContext';
import { Tooltip } from '../ui/tooltip';

interface AnimatedFormFieldProps {
  label: string;
  value: string | boolean | string[];
  onChange: (value: string | boolean | string[]) => void;
  placeholder?: string;
  type?: 'text' | 'password' | 'email' | 'number' | 'checkbox' | 'radio' | 'select' | 'textarea';
  helpText?: string;
  required?: boolean;
  disabled?: boolean;
  fieldName: string;
  isAnimating?: boolean;
  isCurrentField?: boolean;
  isCompleted?: boolean;
  animatedValue?: string | boolean | string[];
  options?: { value: string; label: string }[]; // For select/radio options
  radioName?: string; // For radio button groups
}

export const AnimatedFormField = React.forwardRef<HTMLDivElement, AnimatedFormFieldProps>(
  (
    {
      label,
      value,
      onChange,
      placeholder,
      type = 'text',
      helpText,
      required = false,
      disabled = false,
      fieldName,
      isAnimating = false,
      isCurrentField = false,
      isCompleted: _isCompleted = false,
      animatedValue,
      options = [],
      radioName,
    },
    ref
  ) => {
    // Safely get the auto-fill context - it might not be available in all forms
    const autoFillContext = useContext(AutoFillContext);
    const registerFieldRef = React.useMemo(
      () => autoFillContext?.registerFieldRef || (() => {}),
      [autoFillContext?.registerFieldRef]
    );

    // Use animated value if available and field is being animated
    const displayValue = isAnimating && animatedValue !== undefined ? animatedValue : value;

    // Determine if field should be disabled during animation
    const isDisabled = disabled || (isAnimating && !isCurrentField);

    // Register field ref for scrolling
    React.useEffect(() => {
      const element = ref && typeof ref === 'object' && 'current' in ref ? ref.current : null;
      registerFieldRef(fieldName, element);
      return () => registerFieldRef(fieldName, null);
    }, [fieldName, registerFieldRef, ref]);

    // Simple field styling without animation states
    const getFieldClassName = () => {
      return 'w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-white focus:border-[#3a3a3a] bg-white text-gray-900 disabled:bg-gray-100 disabled:cursor-not-allowed';
    };

    const getLabelClassName = () => {
      return 'block text-sm font-medium text-gray-700 mb-2';
    };

    const renderInput = () => {
      switch (type) {
        case 'checkbox':
          return (
            <div className="flex items-center">
              <input
                type="checkbox"
                checked={displayValue as boolean}
                onChange={(e) => onChange(e.target.checked)}
                disabled={isDisabled}
                className="w-4 h-4 text-white bg-gray-100 border-gray-300 rounded focus:ring-white focus:ring-2"
              />
              <span className="ml-2 text-sm text-gray-700">{label}</span>
            </div>
          );

        case 'radio':
          return (
            <div className="space-y-2">
              {options.map((option) => (
                <label key={option.value} className="flex items-center">
                  <input
                    type="radio"
                    name={radioName || fieldName}
                    value={option.value}
                    checked={displayValue === option.value}
                    onChange={(e) => onChange(e.target.value)}
                    disabled={isDisabled}
                    className="w-4 h-4 text-white bg-gray-100 border-gray-300 focus:ring-white focus:ring-2"
                  />
                  <span className="ml-2 text-sm text-gray-700">{option.label}</span>
                </label>
              ))}
            </div>
          );

        case 'select':
          return (
            <select
              value={displayValue as string}
              onChange={(e) => onChange(e.target.value)}
              disabled={isDisabled}
              className={getFieldClassName()}
            >
              <option value="">Select an option</option>
              {options.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          );

        case 'textarea':
          return (
            <textarea
              value={displayValue as string}
              onChange={(e) => onChange(e.target.value)}
              placeholder={placeholder}
              disabled={isDisabled}
              rows={4}
              className={getFieldClassName()}
            />
          );

        default:
          return (
            <input
              type={type}
              value={displayValue as string}
              onChange={(e) => onChange(e.target.value)}
              placeholder={placeholder}
              disabled={isDisabled}
              className={getFieldClassName()}
            />
          );
      }
    };

    return (
      <div ref={ref} className="relative">
        {type !== 'checkbox' && (
          <label className={getLabelClassName()}>
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
        )}
        {renderInput()}
      </div>
    );
  }
);

AnimatedFormField.displayName = 'AnimatedFormField';
