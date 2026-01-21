// Input: react-hook-form, canEdit state
// Output: SettingsCard context and provider
// Position: Context for compound settings card components

import { createContext, useContext, useMemo, type ReactNode } from 'react'
import type { FieldValues, UseFormReturn } from 'react-hook-form'

interface SettingsCardContextValue<T extends FieldValues = FieldValues> {
  form: UseFormReturn<T>
  canEdit: boolean
  isSaving: boolean
}

const SettingsCardContext = createContext<SettingsCardContextValue | null>(null)

export function useSettingsCardContext<T extends FieldValues = FieldValues>(): SettingsCardContextValue<T> {
  const context = useContext(SettingsCardContext)
  if (!context) {
    throw new Error('SettingsCard compound components must be used within a SettingsCard')
  }
  return context as SettingsCardContextValue<T>
}

interface SettingsCardProviderProps<T extends FieldValues> {
  form: UseFormReturn<T>
  canEdit: boolean
  children: ReactNode
}

export function SettingsCardProvider<T extends FieldValues>({
  form,
  canEdit,
  children,
}: SettingsCardProviderProps<T>) {
  const isSaving = form.formState.isSubmitting

  const value = useMemo(
    () => ({ form, canEdit, isSaving }),
    [form, canEdit, isSaving],
  )

  return (
    <SettingsCardContext value={value as SettingsCardContextValue}>
      {children}
    </SettingsCardContext>
  )
}
