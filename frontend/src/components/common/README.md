<!-- Once this directory changes, update this README.md -->

# Components/Common

应用级共享组件，跨多个功能模块复用。
这些组件包含应用特定逻辑，但仍需保持职责清晰。
通用 UI 原子组件请放在 `components/ui/`。

## Files

- **app-sidebar.tsx**: 应用主侧边栏与导航入口
- **app-topbar.tsx**: 顶部栏，展示核心状态与主题切换
- **main-content.tsx**: 主内容区域容器，负责布局与顶栏集成
- **universal-empty-state.tsx**: 通用空状态/错误态组件
