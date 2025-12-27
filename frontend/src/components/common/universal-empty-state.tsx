// Input: Empty components from ui/empty, Button from ui/button, icons from lucide-react, Motion
// Output: UniversalEmptyState component for displaying empty/error states
// Position: Reusable empty state component for common module

import type { LucideIcon } from 'lucide-react'
import { m } from 'motion/react'
import type { ReactNode } from 'react'

import { Button } from '@/components/ui/button'
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '@/components/ui/empty'
import { Spring } from '@/lib/spring'

export interface UniversalEmptyStateProps {
  icon: LucideIcon
  title: string
  description: string
  action?: {
    label: string
    onClick: () => void
  }
  children?: ReactNode
}

export function UniversalEmptyState({
  icon: Icon,
  title,
  description,
  action,
  children,
}: UniversalEmptyStateProps) {
  return (
    <m.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={Spring.smooth(0.4)}
      className="flex size-full items-center justify-center p-6"
    >
      <Empty>
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <Icon className="size-5 text-muted-foreground" />
          </EmptyMedia>
          <EmptyTitle>{title}</EmptyTitle>
          <EmptyDescription>{description}</EmptyDescription>
        </EmptyHeader>

        {(action || children) && (
          <EmptyContent>
            {action && (
              <Button onClick={action.onClick} size="sm">
                {action.label}
              </Button>
            )}
            {children}
          </EmptyContent>
        )}
      </Empty>
    </m.div>
  )
}
