# Fork 维护说明（LiuYinCarl/crush）

本 fork 在上游 [charmbracelet/crush](https://github.com/charmbracelet/crush)
基础上加了一个功能补丁（见 `feat/history-navigation` 分支）：

- **消息搜索跳转**：聊天区焦点下（`Tab` 切换焦点）按 `/`
  打开搜索框，输入关键词过滤历史消息，`Enter` 跳转并选中。
- **用户消息跳转**：聊天区焦点下按 `K` / `J`
  跳到上一条 / 下一条用户消息（占用原"逐项滚动"的 `K`/`J` 别名，
  该功能仍可用 `Shift+↑/↓`）。
- **历史输入翻页**：输入框中按 `Ctrl+←` / `Ctrl+→`
  召回上一条 / 下一条输入（任意光标位置；`Alt+↑/↓` 为别名）。
  - 注意：macOS 默认把 `Ctrl+←/→` 分配给调度中心切换桌面，需在
    系统设置 → 键盘 → 键盘快捷键 → 调度中心 中关闭这两项。

## 目录约定

- `origin` 远程 = 上游 charmbracelet/crush
- `fork` 远程 = 本 fork LiuYinCarl/crush

## 发布流程（每次上游发新版本后执行）

补丁收敛为两个 commit，放在 `feat/history-navigation` 分支上：

1. `feat(ui): ...` —— 功能本体
2. `ci: ...` —— 本 fork 的 release 工作流 + 本文档

上游发布新 tag（例如 `v0.93.0`）后：

```bash
git fetch origin --tags

# 从新 tag 拉一个发布分支
git checkout -b release/v0.93.0-historynav v0.93.0

# 把补丁 cherry-pick 上来（如有冲突，解决后继续）
git cherry-pick <功能commit> <CI commit>

# 打本 fork 的发布 tag 并推送，推送 tag 即触发 CI 构建
git tag v0.93.0-historynav.1
git push fork release/v0.93.0-historynav
git push fork v0.93.0-historynav.1
```

CI（`.github/workflows/fork-release.yml`）会用纯 `go build` 构建
macOS（arm64/amd64）、Linux（amd64/arm64）、Windows（amd64）五个目标，
并在 fork 的 Releases 页面创建带二进制附件的预发布版本。

如果同一个上游版本需要重新发布（例如补丁有更新），递增序号即可：
`v0.93.0-historynav.2`、`v0.93.0-historynav.3`……

## 同步上游非发布更新

`feat/history-navigation` 基于上游 `main`：

```bash
git checkout feat/history-navigation
git fetch origin
git rebase origin/main
git push --force-with-lease fork feat/history-navigation
```
