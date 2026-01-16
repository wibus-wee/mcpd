// Input: Active callers hook, icons, class utility
// Output: Active callers indicator component for headers
// Position: Shared header indicator for active caller registrations

import { MousePointer2Icon } from 'lucide-react'

import { useActiveCallers } from '@/hooks/use-active-callers'
import { cn } from '@/lib/utils'

const maxVisibleCallers = 2

export const ActiveCallersIndicator = ({ className }: { className?: string }) => {
  const { data: activeCallers } = useActiveCallers()
  const entries = activeCallers ?? []
  const hasActive = entries.length > 0
  const visibleCallers = entries.slice(0, maxVisibleCallers)
  const extraCount = Math.max(entries.length - visibleCallers.length, 0)
  const title = hasActive
    ? entries.map(entry => `${entry.caller} (PID: ${entry.pid})`).join(', ')
    : 'No active callers'

  return (
    <div
      className={cn(
        'flex items-center gap-2 rounded-full border border-border/60 bg-muted/30 px-2.5 py-1 text-xs',
        className,
      )}
      title={title}
    >
      <span
        className={cn(
          'size-2 rounded-full',
          hasActive ? 'bg-success animate-pulse' : 'bg-muted',
        )}
      />
      <span className="text-muted-foreground">Active Callers</span>
      {hasActive ? (
        <div className="flex items-center gap-1">
          {visibleCallers.map(entry => (
            <span
              key={`${entry.caller}:${entry.pid}`}
              className="inline-flex items-center gap-1 rounded-full bg-background/80 px-2 py-0.5 font-mono text-[0.7rem] text-foreground shadow-xs"
            >
              <MousePointer2Icon className="size-3 text-info" />
              {entry.caller}
              <span className="text-muted-foreground">(PID: {entry.pid})</span>
            </span>
          ))}
          {extraCount > 0 && (
            <span className="rounded-full bg-background/70 px-2 py-0.5 text-[0.7rem] text-muted-foreground">
              +{extraCount}
            </span>
          )}
        </div>
      ) : (
        <span className="text-muted-foreground">None</span>
      )}
    </div>
  )
}
