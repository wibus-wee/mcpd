// Input: TanStack Router, Outlet, Settings layout components
// Output: Settings layout route with sidebar navigation
// Position: /settings layout route for nested settings pages

import { createFileRoute, Outlet } from '@tanstack/react-router'
import { BugIcon, PaletteIcon, ServerIcon, SettingsIcon } from 'lucide-react'
import { m } from 'motion/react'

import { NavItem } from '@/components/common/nav-item'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import { Spring } from '@/lib/spring'

export const Route = createFileRoute('/settings')({
  component: SettingsLayout,
})

const navItems: NavItem[] = [
  {
    path: '/settings/runtime',
    label: 'Runtime',
    icon: ServerIcon,
    description: 'Timeouts, retries, and global defaults',
  },
  {
    path: '/settings/subagent',
    label: 'SubAgent',
    icon: SettingsIcon,
    description: 'AI assistant configuration',
  },
  {
    path: '/settings/appearance',
    label: 'Appearance',
    icon: PaletteIcon,
    description: 'Theme and UI preferences',
  },
  {
    path: '/settings/advanced',
    label: 'Advanced',
    icon: BugIcon,
    description: 'Debug logs and telemetry',
  },
]

function SettingsLayout() {
  return (
    <div className="flex h-full flex-col">
      {/* Header */}
      <m.div
        className="px-6 pt-6 pb-4"
        initial={{ opacity: 0, y: 10, filter: 'blur(8px)' }}
        animate={{ opacity: 1, y: 0, filter: 'blur(0px)' }}
        transition={Spring.presets.smooth}
      >
        <div className="flex items-center gap-2">
          <SettingsIcon className="size-4 text-muted-foreground" />
          <h1 className="text-2xl font-bold tracking-tight">Settings</h1>
        </div>
        <p className="text-sm text-muted-foreground">
          Configure runtime defaults and preferences
        </p>
      </m.div>

      <Separator />

      <div className="flex min-h-0 flex-1">
        {/* Sidebar Navigation */}
        <nav className="w-56 shrink-0 border-r">
          <ScrollArea className="h-full">
            <div className="space-y-1 p-3">
              {navItems.map((item, index) => (
                <NavItem key={item.path} item={item} index={index} variant="inline" />
              ))}
            </div>
          </ScrollArea>
        </nav>

        {/* Content Area */}
        <div className="min-w-0 flex-1 overflow-auto">
          <Outlet />
        </div>
      </div>
    </div>
  )
}
