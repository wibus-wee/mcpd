// Input: title, icon, action element
// Output: SectionHeader component for content sections
// Position: Custom atomic component for section titles

import { cn } from '@/lib/utils'

interface SectionHeaderProps {
  icon?: React.ReactNode
  title: string
  badge?: React.ReactNode
  action?: React.ReactNode
  className?: string
}

export function SectionHeader({
  icon,
  title,
  badge,
  action,
  className,
}: SectionHeaderProps) {
  return (
    <div className={cn('flex items-center justify-between', className)}>
      <div className="flex items-center gap-2">
        {icon && (
          <span className="text-muted-foreground">{icon}</span>
        )}
        <h3 className="font-medium text-sm">{title}</h3>
        {badge}
      </div>
      {action}
    </div>
  )
}
