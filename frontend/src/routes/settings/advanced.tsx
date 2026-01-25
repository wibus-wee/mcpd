// Input: TanStack Router, debug utilities
// Output: Advanced settings page
// Position: /settings/advanced route

import { createFileRoute } from '@tanstack/react-router'

import { Badge } from '@/components/ui/badge'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

export const Route = createFileRoute('/settings/advanced')({
  component: AdvancedSettingsPage,
})

function AdvancedSettingsPage() {
  return (
    <div className="space-y-6 p-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-sm">
            Debug & Diagnostics
            <Badge variant="secondary" size="sm">
              Coming Soon
            </Badge>
          </CardTitle>
          <CardDescription className="text-xs">
            Debug logs, telemetry settings, and diagnostic tools will be available here.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            Advanced debugging features are currently under development.
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
