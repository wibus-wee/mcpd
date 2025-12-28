// Input: Callers mapping
// Output: CallersList component - list view of caller-to-profile mappings
// Position: Tab content in config page

import { ArrowRightIcon, UsersIcon } from 'lucide-react'
import { m } from 'motion/react'

import { Badge } from '@/components/ui/badge'
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '@/components/ui/empty'
import { Skeleton } from '@/components/ui/skeleton'
import { Spring } from '@/lib/spring'

interface CallersListProps {
  callers: Record<string, string>
  isLoading: boolean
  onRefresh: () => void
}

function CallersListSkeleton() {
  return (
    <div className="space-y-2">
      {Array.from({ length: 2 }).map((_, i) => (
        <Skeleton key={i} className="h-14 w-full rounded-lg" />
      ))}
    </div>
  )
}

function CallersListEmpty() {
  return (
    <Empty className="py-12">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <UsersIcon className="size-5" />
        </EmptyMedia>
        <EmptyTitle>No caller mappings</EmptyTitle>
        <EmptyDescription>
          Define caller mappings in your configuration to route clients to specific profiles.
        </EmptyDescription>
      </EmptyHeader>
    </Empty>
  )
}

export function CallersList({
  callers,
  isLoading,
}: CallersListProps) {
  const entries = Object.entries(callers)

  if (isLoading) {
    return <CallersListSkeleton />
  }

  if (entries.length === 0) {
    return <CallersListEmpty />
  }

  return (
    <m.div
      className="space-y-2"
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={Spring.smooth(0.3)}
    >
      {entries.map(([caller, profile], index) => (
        <m.div
          key={caller}
          initial={{ opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.2, delay: index * 0.03 }}
          className="flex items-center gap-3 rounded-lg bg-muted/30 px-4 py-3"
        >
          <span className="font-mono text-sm truncate flex-1 min-w-0">
            {caller}
          </span>
          <ArrowRightIcon className="size-3.5 text-muted-foreground shrink-0" />
          <Badge variant="outline" className="shrink-0">
            {profile}
          </Badge>
        </m.div>
      ))}
    </m.div>
  )
}
