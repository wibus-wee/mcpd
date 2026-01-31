import { Card } from '@/components/ui/card'

import { PluginListTable } from './components/plugin-list-table'
import { usePluginList } from './hooks'

export function PluginPage() {
  const { data: plugins, isLoading, error } = usePluginList()

  return (
    <div className="flex flex-col gap-6 p-6">
      <div className="flex flex-col gap-2">
        <h1 className="text-3xl font-bold tracking-tight">Plugins</h1>
        <p className="text-muted-foreground">
          Manage governance plugins for request and response processing
        </p>
      </div>

      {error && (
        <Card className="border-destructive/50 bg-destructive/5 p-4">
          <p className="text-destructive text-sm">
            Failed to load plugins:
            {' '}
            {error instanceof Error ? error.message : 'Unknown error'}
          </p>
        </Card>
      )}

      <Card className="p-0">
        <PluginListTable
          plugins={plugins || []}
          isLoading={isLoading}
        />
      </Card>
    </div>
  )
}
