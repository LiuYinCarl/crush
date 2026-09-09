# Fork 维护说明（LiuYinCarl/crush）

> 本文既是给维护者的说明，也是一份可直接照做的 **发布 runbook**。
> 下次让 AI 发版时，直接让它读这个文件按步骤执行即可。

本 fork 在上游 [charmbracelet/crush](https://github.com/charmbracelet/crush)
基础上加了功能补丁（`feat/history-navigation` 分支）：

- **消息搜索跳转**：聊天区焦点下（`Tab` 切换焦点）按 `/`
  打开搜索框，输入关键词过滤历史消息，`Enter` 跳转并选中。
  搜索框内按 `Tab` 循环切换角色过滤：全部 → 仅 user → 仅 assistant。
- **用户消息跳转**：聊天区焦点下按 `K` / `J`
  跳到上一条 / 下一条用户消息（占用原“逐项滚动”的 `K`/`J` 别名，
  该功能仍可用 `Shift+↑/↓`）。
- **历史输入翻页**：输入框中按 `Ctrl+←` / `Ctrl+→`
  召回上一条 / 下一条输入（任意光标位置；`Alt+↑/↓` 为别名）。
  - 注意：macOS 默认把 `Ctrl+←/→` 分配给调度中心切换桌面，需在
    系统设置 → 键盘 → 键盘快捷键 → 调度中心 中关闭这两项。
- **Subagent 模型白名单**：配置 `options.subagent_models` 后，
  subagent 的默认模型与 `agent_dispatch` / `agent_send` 显式选择的模型
  都被限制在白名单内，`agent_dispatch` 工具描述里展示的模型目录也同步
  收窄；不传 `model` 时使用白名单里第一个可构建的模型，避免通过省略
  参数绕过限制。白名单项写成 `provider/model`（模型 ID 内的 `/` 保留）。
  `agentic_fetch` 的搜索子 agent 固定使用 small 模型，不受此限制。
  - crushrc：`option subagent-model openai/gpt-5-mini`
  - crush.json：`"options": {"subagent_models": ["openai/gpt-5-mini"]}`
  - 配置加载时会校验每项必须是 `provider/model`，写错会直接报错。

## 目录约定

- `origin` 远程 = 本 fork `LiuYinCarl/crush`
- `upstream` 远程 = 上游 `charmbracelet/crush`
- 工作目录：`/Users/kenshin/dev/github/crush`
- `gh` 已用 `LiuYinCarl` 账号登录（`gh auth status` 应显示 `repo` / `workflow` scope）

## 分支与 commit 结构

补丁始终收敛为 **三个 commit**，放在 `feat/history-navigation` 分支上：

1. `feat(ui): message search/jump and ctrl+left/right history cycle keys` —— UI 补丁
2. `feat(agent): restrict subagent models to a configured allowlist` —— subagent 模型白名单
3. `ci: add fork release workflow and maintenance doc` —— release 工作流 + 本文档

发布时**不建发布分支**，直接在新 tag 上 cherry-pick 这三个 commit，
tag 挂在 cherry-pick 出来的游离 commit 上。

> 重要：`feat/history-navigation` 是这三个 commit 的唯一事实来源。
> 发布过程中凡是对“与上游版本无关”的文件做的修改（尤其是
> `.github/workflows/fork-release.yml` 和本 `FORK.md`），一定要**回写到该分支**，
> 否则下次发版又会被旧内容覆盖。

---

## 发布流程（runbook）

### 0. 前置检查

```bash
cd /Users/kenshin/dev/github/crush
git status                 # 工作区必须干净
git remote -v              # origin=本 fork, upstream=上游
gh auth status             # 必须是 LiuYinCarl，且有 repo/workflow scope
```

### 1. 拉取上游最新 tag

```bash
git fetch upstream --tags
git tag --sort=-v:refname | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | head -5
```

取最新版本号，记为 `LATEST`（例如 `v0.101.0`）。

### 2. 找出分支上的三个补丁 commit

```bash
git log --oneline -4 feat/history-navigation
```

最上面三个就是：**UI commit**、**subagent 白名单 commit** 和 **CI commit**。
记为 `UI`、`SUB`、`CI`。

### 3. 在新 tag 上脱离 HEAD 并 cherry-pick

```bash
LATEST=v0.101.0       # ← 换成第 1 步的最新 tag
UI=<UI commit>        # ← 换成第 2 步的 hash
SUB=<SUB commit>      # ← 换成第 2 步的 hash
CI=<CI commit>        # ← 换成第 2 步的 hash

git checkout --detach "$LATEST"
git cherry-pick "$UI" "$SUB" "$CI"
```

#### 3a. 已知冲突与解法

每次上游改动后冲突点可能不同，目前出现过这几类：

**① `internal/ui/model/ui.go`（几乎每次都会出现）**

冲突在 `Chat.End` 分支附近：上游可能把
`if cmd := m.chat.ScrollToBottomAndSelectLast(); cmd != nil { ... }`
简化成 `m.chat.ScrollToBottomAndSelectLast()`。

- **保留上游（`HEAD`）那一行**，不要恢复 `if cmd := ...`。
- 紧接着补上补丁新增的三个 case：

```go
case key.Matches(msg, m.keyMap.Chat.SearchMessages):
    if cmd := m.openSearchDialog(); cmd != nil {
        cmds = append(cmds, cmd)
    }
case key.Matches(msg, m.keyMap.Chat.PrevUserMessage):
    if cmd := m.chat.SelectPrevUserMessage(); cmd != nil {
        cmds = append(cmds, cmd)
    }
case key.Matches(msg, m.keyMap.Chat.NextUserMessage):
    if cmd := m.chat.SelectNextUserMessage(); cmd != nil {
        cmds = append(cmds, cmd)
    }
```

解决后：

```bash
git add internal/ui/model/ui.go
git -c core.editor=true cherry-pick --continue
```

**② `.github/workflows/release.yml`（modify/delete 冲突）**

上游的 `release.yml` 会在 `v*.*.*` tag 上触发 goreleaser  reusable workflow
（需要一堆 secret，fork 里根本跑不通）。本 fork 用
`.github/workflows/fork-release.yml` 取代它，因此要**删除**该文件：

```bash
git rm .github/workflows/release.yml
git -c core.editor=true cherry-pick --continue
```

**③ 上游 API 改名导致编译失败**

最典型的是 `ScrollToSelectedAndAnimate` 在 v0.95.0 被移除。遇到
`undefined (type *Chat has no field or method ...)` 时：

```bash
# 查上游当前有哪些 Scroll/Select 方法
git show "$LATEST":internal/ui/model/chat.go | grep -E '^func \(m \*Chat\) (Scroll|Select)'
```

然后做等价替换（v0.95.0 起用 `ScrollToSelected()` 代替
`ScrollToSelectedAndAnimate()`）。涉及文件：

- `internal/ui/model/chat.go`（`SelectPrevUserMessage` / `SelectNextUserMessage`）
- `internal/ui/model/ui.go`（`ActionJumpToMessage` 分支）

改完把这类“为适配新上游”的改动用 fixup 折进对应的功能 commit，保持三个
commit 的结构：

```bash
git add -A
git commit --fixup "$UI"            # UI 相关改动折进 UI commit
git commit --fixup "$SUB"           # 白名单相关改动折进 subagent commit
GIT_SEQUENCE_EDITOR=true git rebase --autosquash -i "$LATEST"
```

### 4. 本地验证（必须全绿）

```bash
CGO_ENABLED=0 GOEXPERIMENT=greenteagc go build ./...
CGO_ENABLED=0 GOEXPERIMENT=greenteagc go test ./internal/ui/... ./internal/agent/... ./internal/config/... ./internal/shellconfig/...
gofmt -l internal/ui/model/chat.go internal/ui/model/ui.go internal/ui/dialog/search.go \
  internal/agent/agent_dispatch_model.go internal/agent/shared_models.go internal/config/config.go
# gofmt 无输出 = 通过
```

### 5. 回写 `feat/history-navigation` 分支

把与上游版本无关的改动（工作流、本文档）同步回分支，并 amend 掉旧 CI commit：

```bash
# 记下发布 HEAD 的 CI commit（第 3 步之后）
RELEASE_CI=$(git rev-parse HEAD)      # 此时 HEAD 就是 CI commit

git checkout feat/history-navigation
git checkout "$RELEASE_CI" -- .github/workflows/fork-release.yml FORK.md
git add .github/workflows/fork-release.yml FORK.md
git commit --amend --no-edit          # 改写分支上的 ci: 那个 commit
git push --force-with-lease origin feat/history-navigation
```

> 注意：**只回写版本无关的文件**。功能 commit 里为适配新上游做的 API 改名
> 不要带回分支（分支的基线是旧上游，带回去反而编译不过）；这些适配在
> 每次发版时现场解决。

### 6. 打 tag 并推送（推送 tag 即触发 CI 构建）

```bash
git checkout --detach "$RELEASE_CI"   # 回到发布 commit
TAG="${LATEST}-historynav.1"          # 例如 v0.101.0-historynav.1
git tag "$TAG"
git push origin "$TAG"
```

同一个上游版本需要重发时，序号递增：`...-historynav.2`、`.3`……
（见下文“重新发布”。）

### 7. 用 `gh` 验证

```bash
# 看 run 状态，run-id 从上一行输出里取
gh run list --repo LiuYinCarl/crush --limit 5
gh run watch <run-id> --repo LiuYinCarl/crush

# 看 release（关键：不能是 Pre-release）
gh release view "$TAG" --repo LiuYinCarl/crush
gh release list --repo LiuYinCarl/crush --limit 5
```

期望结果：

- run：`completed` + `success`
- release：**Pre-release 列为空**（正式 release），Latest 为当前 tag
- release 附件：5 个平台包（darwin/linux 的 amd64+arm64、windows amd64）
  各一份 `.tar.gz`/`.zip`，外加 `checksums.txt`

### 8. 回到功能分支

```bash
git checkout feat/history-navigation
```

---

## 重新发布同一上游版本

补丁有更新、或上一次 release 需要修正时，递增序号即可：

```bash
git checkout feat/history-navigation
# 在功能 commit 上改代码 / 在 CI commit 上改工作流
# 然后重复上面第 3~7 步，tag 用 -historynav.2
```

如果只是修 release 类型 / 工作流而补丁代码没变，也可以：
修好 `feat/history-navigation` 的 CI commit 并推送，然后在一个**新的**
游离 HEAD 上重新 cherry-pick 三个 commit，打 `-historynav.2` 推送。

> 已推送的 tag 尽量**不要强推移动**；优先用递增序号发新 tag。
> 仅在刚推送、CI 还没产出 release 的极短窗口内允许强推纠正。

## 同步上游非发布更新（rebase 功能分支）

`feat/history-navigation` 基于上游 `main`：

```bash
git checkout feat/history-navigation
git fetch upstream
git rebase upstream/main
git push --force-with-lease origin feat/history-navigation
```

rebase 后如果功能 commit 需要适配上游 API（如上面第 3a-③ 的情况），
一并解决并保留三个 commit 的结构。
