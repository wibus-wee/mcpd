// Input: None
// Output: Empty state for when no server is selected
// Position: Placeholder in detail panel

import { ServerIcon } from 'lucide-react'
import { m } from 'motion/react'

import { Spring } from '@/lib/spring'

export function ServerDetailEmptyState() {
  return (
    <m.div
      className="flex flex-col items-center justify-center h-full text-center p-8"
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={Spring.presets.smooth}
    >
      <div className="rounded-2xl bg-muted/50 p-6 mb-6">
        <ServerIcon className="size-10 text-muted-foreground/60" />
      </div>
      <h3 className="text-lg font-medium text-foreground/80 mb-2">
        No server selected
      </h3>
      <p className="text-sm text-muted-foreground max-w-xs">
        Select a server from the list to view its details, tools, and configuration.
      </p>
    </m.div>
  )
}
