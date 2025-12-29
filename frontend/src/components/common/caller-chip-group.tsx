// Input: active callers, icons, class utility
// Output: caller chip group for contextual caller display
// Position: shared UI component for caller lists

import { MousePointer2Icon } from 'lucide-react'

import type { ActiveCaller } from '@bindings/mcpd/internal/ui'

import { cn } from '@/lib/utils'

interface CallerChipGroupProps {
  callers: ActiveCaller[]
  maxVisible?: number
  showPid?: boolean
  emptyText?: string
  className?: string
}

export const CallerChipGroup = ({
  callers,
  maxVisible = 2,
  showPid = false,
  emptyText,
  className,
}: CallerChipGroupProps) => {
  const visibleCallers = callers.slice(0, maxVisible)
  const extraCount = Math.max(callers.length - visibleCallers.length, 0)

  if (callers.length === 0) {
    if (!emptyText) {
      return null
    }
    return <span className={cn('text-[0.7rem] text-muted-foreground', className)}>{emptyText}</span>
  }

  return (
    <div className={cn('flex flex-wrap items-center gap-1', className)}>
      {visibleCallers.map(entry => (
        <span
          key={`${entry.caller}:${entry.pid}`}
          className="inline-flex items-center gap-1 rounded-full bg-background/80 px-2 py-0.5 font-mono text-[0.7rem] text-foreground shadow-xs"
          title={showPid ? `${entry.caller} (PID: ${entry.pid})` : entry.caller}
        >
          <MousePointer2Icon className="size-3 text-info" />
          {entry.caller}
          {showPid && (
            <span className="text-muted-foreground">(PID: {entry.pid})</span>
          )}
        </span>
      ))}
      {extraCount > 0 && (
        <span className="rounded-full bg-background/70 px-2 py-0.5 text-[0.7rem] text-muted-foreground">
          +{extraCount}
        </span>
      )}
    </div>
  )
}
