<!-- Once this directory changes, update this README.md -->

# Atoms

用于 Jotai 的全局 UI 状态，只承载跨页面的轻量状态。
这里不存放后端数据，后端数据由 SWR 统一管理。
命名保持 `{domain}Atom` 约定，避免跨域耦合。

## Files

- **navigation.ts**: 侧边栏与导航相关的 UI 状态 atoms
- **logs.ts**: 日志流相关的 UI 状态 atoms
