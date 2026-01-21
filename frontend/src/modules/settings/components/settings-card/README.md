<!-- Once this directory changes, update this README.md -->

# SettingsCard Compound Component

A compound component pattern for building settings forms with react-hook-form.
Uses React Context to eliminate prop drilling for `form`, `canEdit`, and `isSaving` state.

## Architecture

The component provides a context-based approach where:
- Parent `SettingsCard` provides form context
- Child components consume context automatically
- Custom fields can access context via `useSettingsCardContext()`

## Files

- **index.tsx**: Main entry point, exports `SettingsCard` compound component
- **context.tsx**: React context and provider for form state
- **fields.tsx**: Field components (NumberField, SelectField, SwitchField, TextField, TextareaField)
- **layout.tsx**: Layout components (Header, Content, Section, Footer, ReadOnlyAlert, ErrorAlert)

## Usage

```tsx
import { SettingsCard } from './settings-card'

<SettingsCard form={form} canEdit={canEdit} onSubmit={handleSave}>
  <SettingsCard.Header title="Runtime" description="Configure runtime settings" />
  <SettingsCard.Content>
    <SettingsCard.ReadOnlyAlert />
    <SettingsCard.ErrorAlert error={error} />
    <SettingsCard.Section title="Core">
      <SettingsCard.SelectField
        name="bootstrapMode"
        label="Bootstrap Mode"
        options={BOOTSTRAP_MODE_OPTIONS}
      />
      <SettingsCard.NumberField
        name="routeTimeoutSeconds"
        label="Route Timeout"
        unit="seconds"
      />
    </SettingsCard.Section>
  </SettingsCard.Content>
  <SettingsCard.Footer statusLabel={statusLabel} />
</SettingsCard>
```

## Custom Fields

For custom rendering, use `SettingsCard.Field` or access context directly:

```tsx
import { useSettingsCardContext } from './settings-card'

function CustomModelField() {
  const { form, canEdit, isSaving } = useSettingsCardContext()
  // Custom implementation
}
```
