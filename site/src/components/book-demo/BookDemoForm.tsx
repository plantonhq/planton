'use client';

import { Box, Typography } from '@mui/material';
import { useState, useCallback, useRef } from 'react';
import type {
  DemoFormData,
  FieldErrors,
  SubmissionStatus,
} from './types';
import { DEMO_COPY } from '@/data/homepage';
import { trackDemo } from '@/lib/demo-analytics';
import { JOB_TITLE_OPTIONS, COMPANY_SIZE_OPTIONS, SUBMISSION_ENDPOINT } from './types';

interface BookDemoFormProps {
  onSuccess: (data: DemoFormData) => void;
}

const INITIAL_FORM: DemoFormData = {
  firstName: '',
  lastName: '',
  workEmail: '',
  company: '',
  jobTitle: '',
  companySize: '',
};

function validateField(name: keyof DemoFormData, value: string): string | undefined {
  switch (name) {
    case 'firstName':
      return value.trim() ? undefined : 'First name is required';
    case 'lastName':
      return value.trim() ? undefined : 'Last name is required';
    case 'workEmail': {
      if (!value.trim()) return 'Work email is required';
      if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value))
        return 'Enter a valid email address';
      return undefined;
    }
    case 'company':
      return value.trim() ? undefined : 'Company is required';
    case 'jobTitle':
      return value ? undefined : 'Select your job title';
    case 'companySize':
      return value ? undefined : 'Select your company size';
    default:
      return undefined;
  }
}

function validateAll(data: DemoFormData): FieldErrors {
  const errors: FieldErrors = {};
  for (const key of Object.keys(data) as (keyof DemoFormData)[]) {
    const err = validateField(key, data[key]);
    if (err) errors[key] = err;
  }
  return errors;
}

const inputClasses =
  'w-full bg-raised border border-edge rounded-lg px-4 py-3 text-fg text-sm placeholder-fg-secondary focus:border-edge-hover focus:outline-none focus:ring-1 focus:ring-fg-secondary transition-colors';

const selectClasses =
  `${inputClasses} appearance-none cursor-pointer`;

const labelClasses = 'text-sm font-medium text-fg-secondary mb-1.5 block';

export function BookDemoForm({ onSuccess }: BookDemoFormProps) {
  const started = useRef(false);
  const submitting = useRef(false);
  const [form, setForm] = useState<DemoFormData>(INITIAL_FORM);
  const [errors, setErrors] = useState<FieldErrors>({});
  const [touched, setTouched] = useState<Set<keyof DemoFormData>>(new Set());
  const [status, setStatus] = useState<SubmissionStatus>('idle');
  const [serverError, setServerError] = useState<string | null>(null);

  const handleChange = useCallback(
    (field: keyof DemoFormData, value: string) => {
      setForm((prev) => ({ ...prev, [field]: value }));
      if (touched.has(field)) {
        setErrors((prev) => ({
          ...prev,
          [field]: validateField(field, value),
        }));
      }
    },
    [touched],
  );

  const handleBlur = useCallback(
    (field: keyof DemoFormData) => {
      setTouched((prev) => new Set(prev).add(field));
      setErrors((prev) => ({
        ...prev,
        [field]: validateField(field, form[field]),
      }));
    },
    [form],
  );

  const submitForm = useCallback(async () => {
    if (submitting.current) return;
    setServerError(null);

    const allErrors = validateAll(form);
    setErrors(allErrors);
    setTouched(
      new Set(Object.keys(form) as (keyof DemoFormData)[]),
    );

    if (Object.values(allErrors).some(Boolean)) return;

    submitting.current = true;
    setStatus('submitting');

    const MIN_SUBMIT_MS = 800;
    const sleep = new Promise((r) => setTimeout(r, MIN_SUBMIT_MS));

    let res: Response | undefined;
    let fetchError: unknown;

    try {
      res = await fetch(SUBMISSION_ENDPOINT, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(form),
      });
    } catch (err) {
      fetchError = err;
    }

    await sleep;
    submitting.current = false;

    if (fetchError) {
      setStatus('error');
      setServerError(
        fetchError instanceof TypeError &&
          (fetchError as TypeError).message === 'Failed to fetch'
          ? 'network'
          : 'unknown',
      );
      return;
    }

    if (!res!.ok) {
      const body = await res!.json().catch(() => null);
      setStatus('error');
      setServerError('unknown');
      console.error('Submission error:', body?.error ?? res!.status);
      return;
    }

    setStatus('success');
    trackDemo('demo_form_submit_success');
    onSuccess(form);
  }, [form, onSuccess]);

  const handleFormSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();
      submitForm();
    },
    [submitForm],
  );

  return (
    <form
      className="rounded-xl bg-panel border border-edge p-6 md:p-8"
      onSubmit={handleFormSubmit}
      onFocusCapture={() => { if (!started.current) { started.current = true; trackDemo('demo_form_start'); } }}
      noValidate
    >
      <Typography className="text-lg font-semibold text-white mb-6">
        {DEMO_COPY.formTitle}
      </Typography>

      <Box className="flex flex-col gap-5">
        {/* First Name + Last Name — side by side */}
        <Box className="grid grid-cols-1 sm:grid-cols-2 gap-5">
          <Field
            label="First Name"
            name="firstName"
            placeholder="Jane"
            value={form.firstName}
            error={errors.firstName}
            required
            onChange={(v) => handleChange('firstName', v)}
            onBlur={() => handleBlur('firstName')}
          />
          <Field
            label="Last Name"
            name="lastName"
            placeholder="Doe"
            value={form.lastName}
            error={errors.lastName}
            required
            onChange={(v) => handleChange('lastName', v)}
            onBlur={() => handleBlur('lastName')}
          />
        </Box>

        <Field
          label="Work Email"
          name="workEmail"
          type="email"
          placeholder="jane@company.com"
          value={form.workEmail}
          error={errors.workEmail}
          required
          onChange={(v) => handleChange('workEmail', v)}
          onBlur={() => handleBlur('workEmail')}
        />

        <Field
          label="Company"
          name="company"
          placeholder="Acme Corp"
          value={form.company}
          error={errors.company}
          required
          onChange={(v) => handleChange('company', v)}
          onBlur={() => handleBlur('company')}
        />

        <SelectField
          label="Job Title"
          name="jobTitle"
          placeholder="Select your role"
          options={JOB_TITLE_OPTIONS}
          value={form.jobTitle}
          error={errors.jobTitle}
          onChange={(v) => handleChange('jobTitle', v)}
          onBlur={() => handleBlur('jobTitle')}
        />

        <SelectField
          label="Company Size"
          name="companySize"
          placeholder="Select company size"
          options={COMPANY_SIZE_OPTIONS}
          value={form.companySize}
          error={errors.companySize}
          onChange={(v) => handleChange('companySize', v)}
          onBlur={() => handleBlur('companySize')}
        />

        {/* Server error banner */}
        {serverError && <ErrorBanner type={serverError} />}

        {/* Submit button */}
        <button
          type="submit"
          disabled={status === 'submitting'}
          className="w-full bg-cta hover:opacity-90 text-cta-text font-medium text-sm px-5 py-3 rounded-lg transition-all duration-300 hover:-translate-y-0.5 disabled:opacity-60 disabled:hover:translate-y-0 disabled:cursor-not-allowed flex items-center justify-center gap-2 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-fg-secondary focus-visible:ring-offset-2 focus-visible:ring-offset-panel"
        >
          {status === 'submitting' ? (
            <>
              <Spinner />
              Submitting...
            </>
          ) : (
            DEMO_COPY.submit
          )}
        </button>
      </Box>
    </form>
  );
}

/* ─── Reusable field component ─── */

interface FieldProps {
  label: string;
  name: string;
  type?: string;
  placeholder: string;
  value: string;
  error?: string;
  required?: boolean;
  onChange: (value: string) => void;
  onBlur: () => void;
}

function Field({
  label,
  name,
  type = 'text',
  placeholder,
  value,
  error,
  required,
  onChange,
  onBlur,
}: FieldProps) {
  return (
    <Box>
      <label htmlFor={name} className={labelClasses}>
        {label}
      </label>
      <input
        id={name}
        name={name}
        type={type}
        className={inputClasses}
        placeholder={placeholder}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onBlur={onBlur}
        aria-required={required ? 'true' : undefined}
        aria-invalid={Boolean(error)}
        aria-describedby={error ? `${name}-error` : undefined}
      />
      {error && (
        <span id={`${name}-error`} className="text-xs text-danger mt-1 block">
          {error}
        </span>
      )}
    </Box>
  );
}

/* ─── Reusable select field component ─── */

interface SelectFieldProps {
  label: string;
  name: string;
  placeholder: string;
  options: ReadonlyArray<{ value: string; label: string }>;
  value: string;
  error?: string;
  onChange: (value: string) => void;
  onBlur: () => void;
}

function SelectField({
  label,
  name,
  placeholder,
  options,
  value,
  error,
  onChange,
  onBlur,
}: SelectFieldProps) {
  return (
    <Box>
      <label htmlFor={name} className={labelClasses}>
        {label}
      </label>
      <Box className="relative">
        <select
          id={name}
          className={selectClasses}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          onBlur={onBlur}
          aria-required="true"
          aria-invalid={Boolean(error)}
        aria-describedby={error ? `${name}-error` : undefined}
        >
          <option value="" disabled>
            {placeholder}
          </option>
          {options.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
        <Box className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-fg-secondary">
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M6 9l6 6 6-6" />
          </svg>
        </Box>
      </Box>
      {error && (
        <span id={`${name}-error`} className="text-xs text-danger mt-1 block">
          {error}
        </span>
      )}
    </Box>
  );
}

/* ─── Error banner ─── */

const ERROR_COPY: Record<string, { heading: string; body: string }> = {
  network: {
    heading: 'We hit a snag',
    body: "We weren't able to submit your request right now. Please try again in a few moments, or drop us a line at hello@planton.ai",
  },
  unknown: {
    heading: 'Something went wrong',
    body: "We weren't able to process your request. Please try again shortly, or reach out to us at hello@planton.ai",
  },
};

function ErrorBanner({ type }: { type: string }) {
  const copy = ERROR_COPY[type] ?? ERROR_COPY['unknown'];

  return (
    <Box role="alert" className="rounded-lg bg-raised border border-edge px-5 py-4 flex gap-3.5 items-start animate-[fadeIn_0.3s_ease-out]">
      <Box className="w-8 h-8 rounded-full bg-danger/10 flex items-center justify-center flex-shrink-0 mt-0.5">
        <svg
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <circle cx="12" cy="12" r="10" />
          <line x1="12" y1="8" x2="12" y2="12" />
          <line x1="12" y1="16" x2="12.01" y2="16" />
        </svg>
      </Box>
      <Box className="flex-1 min-w-0">
        <Typography className="text-sm font-medium text-white mb-0.5">
          {copy.heading}
        </Typography>
        <Typography className="text-xs text-fg-secondary leading-relaxed">
          {copy.body.split('hello@planton.ai').map((part, i, arr) =>
            i < arr.length - 1 ? (
              <span key={i}>
                {part}
                <CopyEmail />
              </span>
            ) : (
              part
            ),
          )}
        </Typography>
      </Box>
    </Box>
  );
}

/* ─── Copy email inline button ─── */

function CopyEmail() {
  const [copied, setCopied] = useState(false);

  const copy = () => {
    void navigator.clipboard.writeText('hello@planton.ai').then(() => setCopied(true)).catch(() => setCopied(false));
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <span
      role="button"
      tabIndex={0}
      onClick={copy}
      onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') copy(); }}
      className="inline-flex items-center gap-1.5 cursor-pointer group"
    >
      <span className="text-white font-medium">hello@planton.ai</span>
      {copied ? (
        <span className="inline-flex items-center gap-0.5 text-ok">
          <svg
            width="12"
            height="12"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2.5"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M5 13l4 4L19 7" />
          </svg>
          <span className="text-[10px] font-medium">copied</span>
        </span>
      ) : (
        <svg
          width="12"
          height="12"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          className="text-fg-secondary opacity-0 group-hover:opacity-100 transition-opacity"
        >
          <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
          <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" />
        </svg>
      )}
    </span>
  );
}

/* ─── Loading spinner ─── */

function Spinner() {
  return (
    <svg
      className="animate-spin h-4 w-4"
      viewBox="0 0 24 24"
      fill="none"
    >
      <circle
        className="opacity-25"
        cx="12"
        cy="12"
        r="10"
        stroke="currentColor"
        strokeWidth="4"
      />
      <path
        className="opacity-75"
        fill="currentColor"
        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
      />
    </svg>
  );
}
