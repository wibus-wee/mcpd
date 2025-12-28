// Input: children, onClick, selected state
// Output: ListItem component - clickable row with hover state
// Position: Custom atomic component for list views

import { ChevronRightIcon } from 'lucide-react'
import { m } from 'motion/react'

import { cn } from '@/lib/utils'

interface ListItemProps {
  children: React.ReactNode
  onClick?: () => void
  selected?: boolean
  className?: string
  showChevron?: boolean
  index?: number
}

export function ListItem({
  children,
  onClick,
  selected = false,
  className,
  showChevron = true,
  index = 0,
}: ListItemProps) {
  return (
    <m.button
      type="button"
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.2, delay: index * 0.03 }}
      onClick={onClick}
      className={cn(
        'group flex w-full items-center justify-between gap-3 rounded-lg px-3 py-3 text-left transition-colors',
        selected
          ? 'bg-accent text-accent-foreground'
          : 'hover:bg-muted/50',
        className,
      )}
    >
      <div className="flex min-w-0 flex-1 items-center gap-3">
        {children}
      </div>
      {showChevron && (
        <ChevronRightIcon
          className={cn(
            'size-4 shrink-0 text-muted-foreground/50 transition-transform',
            'group-hover:translate-x-0.5 group-hover:text-muted-foreground',
          )}
        />
      )}
    </m.button>
  )
}
