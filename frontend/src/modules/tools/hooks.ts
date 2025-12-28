// Input: WailsService bindings, SWR, runtime status hook
// Output: useToolsByServer hook for grouping tools by server
// Position: Data layer for tools module

import { useMemo } from "react";
import useSWR from "swr";

import { WailsService, type ToolEntry } from "@bindings/mcpd/internal/ui";

import { useRuntimeStatus } from "@/modules/config/hooks";

interface ServerGroup {
  id: string;
  specKey: string;
  serverName: string;
  tools: ToolEntry[];
}

export function useToolsByServer() {
  const {
    data: tools,
    isLoading,
    error,
  } = useSWR<ToolEntry[]>("tools", () => WailsService.ListTools());

  const { data: runtimeStatus } = useRuntimeStatus();

  const runtimeServerNames = useMemo(() => {
    if (!runtimeStatus) return new Map<string, string>();

    return runtimeStatus.reduce<Map<string, string>>((map, status) => {
      if (status.specKey) {
        map.set(status.specKey, status.serverName || status.specKey);
      }
      return map;
    }, new Map());
  }, [runtimeStatus]);

  const serverMap = useMemo(() => {
    if (!tools) return new Map<string, ServerGroup>();

    const map = new Map<string, ServerGroup>();

    tools.forEach((tool) => {
      const specKey = tool.specKey || tool.serverName || tool.name;
      if (!specKey) return;

      const displayName =
        runtimeServerNames.get(specKey) || tool.serverName || specKey;

      if (!map.has(specKey)) {
        map.set(specKey, {
          id: specKey,
          specKey,
          serverName: displayName,
          tools: [],
        });
      }
      map.get(specKey)!.tools.push(tool);
    });

    return map;
  }, [tools, runtimeServerNames]);

  const servers = useMemo(() => {
    return Array.from(serverMap.values()).sort((a, b) =>
      a.serverName.localeCompare(b.serverName)
    );
  }, [serverMap]);

  return {
    servers,
    serverMap,
    isLoading,
    error,
    runtimeStatus,
  };
}
