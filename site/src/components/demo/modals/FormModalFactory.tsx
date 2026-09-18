'use client';

import React from 'react';
import { CredentialModal } from './CredentialModal';

/** What a form inside the modal takes; `T` is the record the form produces. */
export interface FormProps<T> {
  onSubmit: (data: T) => void;
  onCancel: () => void;
  initialData?: Partial<T>;
}

interface FormModalFactoryProps<T> {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: T) => void;
  FormComponent: React.ComponentType<FormProps<T>>;
  formProps?: Pick<FormProps<T>, 'initialData'>;
}

export function FormModalFactory<T>({
  isOpen,
  onClose,
  onSubmit,
  FormComponent,
  formProps = {},
}: FormModalFactoryProps<T>) {
  const handleFormSubmit = (data: T) => {
    onSubmit(data);
    onClose();
  };

  const handleFormCancel = () => {
    onClose();
  };

  return (
    <CredentialModal
      isOpen={isOpen}
      onClose={onClose}
    >
      <FormComponent
        {...formProps}
        onSubmit={handleFormSubmit}
        onCancel={handleFormCancel}
      />
    </CredentialModal>
  );
}

// Factory function that creates a modal component
export function createFormModal<T>(
  FormComponent: React.ComponentType<FormProps<T>>
): React.FC<Omit<FormModalFactoryProps<T>, 'FormComponent'>> {
  const ModalComponent = (props: Omit<FormModalFactoryProps<T>, 'FormComponent'>) => (
    <FormModalFactory
      {...props}
      FormComponent={FormComponent}
    />
  );
  
  ModalComponent.displayName = `FormModal(${FormComponent.displayName || FormComponent.name || 'Component'})`;
  
  return ModalComponent;
}
