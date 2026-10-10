# Go运行工具（0.5.0）

统一程序`amazing-skills`负责安装、目录校验、快照同步、项目初始化与风格锁、QA冻结/记录/报告、Jira草稿/回执、提供者适配、合成QA演练和受限示例预览。只用Go标准库，无go.sum及第三方Go依赖。

仅让AI读取Markdown技能不需要运行时。源码构建需Go 1.23+；独立程序安装及运行都无需Go或Python。上游原Python源码及测试原样保留作为第三方来源，正常路径不执行它们；自有Python工具和测试已由Go替换。

## 当前电脑：拿本地构建程序使用

本次为macOS/Linux/Windows的arm64与amd64各构建一份，在仓库dist目录（Git忽略，不自动发布）。macOS Apple Silicon：

```bash
./dist/amazing-skills-darwin-arm64 install --target /目标项目/.agents/skills
/目标项目/.agents/skills/.amazing-skills/amazing-skills qa check --project /目标项目
```

Mac Intel改为darwin-amd64，Linux按架构选linux-amd64/linux-arm64，Windows选windows-amd64.exe/windows-arm64.exe。先核对dist/SHA256SUMS。将对应程序复制到其他电脑即可安装：程序嵌入技能、上游数据、锁文件及模板，不需要另外复制整个源仓库。安装出来的skills与`.amazing-skills`一起移动即可继续使用。

独立程序不含示例Web页面和完整维护文档；serve/validate/sync/build是仓库维护命令，需要源仓库或`--repo`指向它。replay嵌入明确的合成QA资料，可以独立运行。

## 从源码构建

```bash
go test ./...
go vet ./...
go run ./cmd/amazing-skills validate
go run ./cmd/amazing-skills build --output dist
```

build以CGO_ENABLED=0、trimpath和去除调试符号的参数交叉构建六平台，生成SHA256SUMS及manifest.json。不运行其他平台的程序，不声称已经验证跨系统行为。

日常开发可以直接运行：

```bash
go run ./cmd/amazing-skills install --target /项目/.agents/skills --skills qa-workflow
go run ./cmd/amazing-skills project init --project /项目
go run ./cmd/amazing-skills qa init --project /项目
go run ./cmd/amazing-skills qa snapshot --project /项目 --label r1
go run ./cmd/amazing-skills qa record --project /项目 --input /项目/result.json
go run ./cmd/amazing-skills qa summary --project /项目
go run ./cmd/amazing-skills jira draft --project /项目 --input /项目/defect.json
go run ./cmd/amazing-skills replay --output /新目录/qa-demo
```

`--skills`支持一个ID或逗号分隔的多个ID；安装依赖闭包。默认十四技能，QA子包六技能，工作台子包八技能；都附带同平台Go程序。已有技能目录或运行工具拒绝覆盖，重装需选择新目录或自行明确管理旧安装。

目录更新先运行sync，再重新构建，避免二进制中的嵌入快照过期。修改源文件不会自动更新已经安装的副本或已经构建的dist文件。

## 迁移保证与限制

保持项目规格、设计锁、QA计划、快照、执行链及缺陷数据schema1。Go可读取原Python生成的历史执行哈希、证据与快照；旧历史未改写。JSON展示格式可能变化，但canonical SHA语义兼容已有示例。工具采用串行单写，不提供多人写入协调。

通过测试覆盖来源变化、范围约束、未明确规则、修订隔离、冻结身份保留、同轮首次失败、复测历史、缺陷证据、重复操作、零分母、路径与符号链接逃逸、旧记录兼容、安装和移动。新增合成演练仍是模拟模型，没有原生应用/网络/温度事件、Figma连接或真实Jira提交。

Pro Max已从简化适配升级为原版查询与结构化设计系统逻辑的Go移植，1,807个原版对照场景通过；文本排版不逐字一致，详见 [兼容说明](promax-parity.md)。Impeccable仍调用校验过的官方引擎，首次准备需要网络；用户项目的前端构建/浏览器工具不是由Go工具替代。

## 本次验证结果

最初迁移的31项Go测试与go vet通过，另有11项独立程序检查、7项客户示例浏览器场景通过。移动安装目录且PATH不含Python/Go的情况下，完整合成QA演练、当前计划检查、提供者校验及Go查询都通过。实际调用了校验过的Impeccable官方引擎和原生Pro Max适配器。具体结果见 [验证记录](verification/go-runtime.json)。六平台构建结果及摘要见dist/manifest.json和SHA256SUMS，尚未发布远程release。

后续Pro Max补齐及最新验证见 [兼容说明](promax-parity.md) 和 [对照记录](verification/promax-parity.json)。

0.5.0新增通用功能安全筛选、独立qa-security-review与可选安全计划约束；完整安装14技能，QA子包6技能。`security-replay --output <全新目录>`运行合成安全范围演练；最新证据见docs/verification/security.json，不代表真实应用扫描。
