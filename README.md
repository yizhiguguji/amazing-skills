# Amazing Skills

可扩展的 AI 技能库。当前版本提供工作台从模糊需求到实现检查的流程；现已加入功能 QA 的用例、执行、Jira 缺陷和报告能力；可继续扩展其他领域。

## 从这里开始

**第一次使用只看 [START_HERE.md](START_HERE.md)。唯一推荐总入口是 [skill-router · 自动选择技能](skills/skill-router/SKILL.md)。**

不需要从十四个技能中挑一个，也不用先认识 finesse-brief、Impeccable 或 QA 的内部步骤。

```text
读取这个仓库的 START_HERE.md，帮我根据这份需求出测试用例。
```

将需求换成自己的话，例如“做个客户跟进工作台，做到实现检查”“需求改了，更新用例”。AI自动选择适用能力并沿用已有方案。调用关系、已安装用法及发现条件都在入口页。

| 你看到的文件/目录 | 用途 | 是否从这里调用 |
| --- | --- | --- |
| `START_HERE.md` | 首次使用说明和AI导航 | **是，先看这里** |
| `skills/skill-router/` | 总入口Skill | **是，安装后的入口** |
| `skills/product-workflow/`、`skills/qa-workflow/` | 领域编排 | 通常由总入口选择；也可直接调用 |
| `skills/` 其他目录 | 专业能力和上游提供者 | 按任务选用，无需逐个阅读 |
| `catalog.json` | 中文名称、层级及能力的机器目录 | AI选技能时读取 |
| `examples/` | 工作流演练和验证证据 | 示例，不是入口或运行依赖 |
| `docs/`、`cmd/`、Go源文件与测试 | 架构、安装维护、检查 | 安装或维护时使用 |

## 安装（Go，运行不需要Python）

源码构建需要Go 1.23+；其他电脑拿到对应平台的独立二进制后，运行与安装都不需要Go或Python。尚未发布远程下载包，本地六平台构建与使用见 [Go运行工具](docs/go-runtime.md)。安装器会同时复制 `.amazing-skills/` 运行工具。


在仓库根目录运行，目标目录按AI工具实际支持的位置选择：

```bash
go run ./cmd/amazing-skills install --target /你的项目/.agents/skills
```

默认复制十四个技能、相对引用及Go运行工具，不修改目标项目的AGENTS.md，拒绝覆盖现有目录。安装并被工具发现后，可直接用中文提出需求；显式调用总入口为 `$skill-router 你的需求`。其他工具可换成其支持的目录；不支持发现时让AI读取入口文件。

`--skills product-workflow` 可仅安装工作台八技能子包；QA六技能子包见下文。子包使用对应领域入口，只有实际安装的能力才能调用。单独安装 `skill-router` 只提供选择器，不会自动安装全部能力。

## 第一条工作流

`product-workflow → workbench-definition → interface-delivery → workbench-verification`

- 产品依据：`.workbench/spec.md`。
- 设计依据：`DESIGN.md`，实现后落实为代码 tokens 和组件。
- 阶段与技能选择：`.product/state.json`。
- 计划、变更和证据：`.product/plan.md`、`decisions.md`、`verification.md`。

finesse-brief、Impeccable、UI UX Pro Max 的固定版本已经捆绑，包括原始 Skill、脚本、数据与许可证。安装器校验后复制；Impeccable 引擎首次使用需网络或预装，详见 [真实接入与跨电脑使用](docs/providers.md)。不可用时显式报告或记录内置回退。Skill 是给模型的操作指南，不是可保证自治执行的程序。

## 换电脑使用

不需要客户跟进示例应用。AI 读取本仓库目录或安装后的 skill-router，自行加载适用技能。首次准备 Impeccable 引擎：

```bash
/项目/.agents/skills/.amazing-skills/amazing-skills providers doctor
/项目/.agents/skills/.amazing-skills/amazing-skills providers setup-engine --project /目标项目
```

setup-engine 只下载当前平台固定版本，按锁文件 SHA-256 校验；不安装 hooks、不修改全局技能。macOS/Linux/Windows 使用对应平台的Go独立程序。浏览器验证使用执行电脑的实际工具，没有浏览器就记录未验证。

## 示例与检查

[客户跟进示例](examples/customer-followup/README.md) 保留初版内置流程回归；[项目工作台](examples/project-workbench/README.md) 验证捆绑提供者的实际使用、工具调用及后续修改不改变规范。示例不安装到用户项目。

```bash
go run ./cmd/amazing-skills validate
go test ./...
go run ./cmd/amazing-skills serve --example customer-followup --port 8765
```

打开 http://127.0.0.1:8765 。可维护工具使用Go；浏览器测试需要 Node、Playwright 与浏览器：`NODE_PATH=/你的/node_modules node examples/customer-followup/test-browser.cjs`，测试自行启动本地服务，无需先启动上述服务。可用 `CHROME_PATH=/你的/Chrome可执行文件` 指定已有 Chrome，避免下载浏览器。

## 扩展

阅读 [架构与扩展约定](docs/architecture.md)。`catalog.json` 是仓库总目录；工作流只管理自己的步骤。添加 Jira 或测试技能无需修改工作台模板。修改目录后运行 `go run ./cmd/amazing-skills sync` 更新入口的可安装目录快照，然后验证。

原创新增内容采用 MIT；捆绑第三方目录保留各自许可证与锁定版本。版本：0.5.0。

功能安全筛选与独立专项使用见 [安全能力说明](docs/security.md)，模拟验证见 [安全范围演练](examples/qa-security-scope/README.md)。

## 功能 QA（0.5.0）

[使用与选择指南](docs/qa.md)解释简单需求如何覆盖相关行为而不扩大范围，以及持续变更如何保护旧结果。总入口会将多阶段功能测试任务交给 qa-workflow；熟悉后也可直接调用领域工作流或单点能力。

```bash
go run ./cmd/amazing-skills install --target /目标项目/.agents/skills --skills qa-workflow
/目标项目/.agents/skills/.amazing-skills/amazing-skills qa init --project /目标项目
```

QA子包只有六技能，不依赖三个设计提供者、私有QA系统、Jira凭证或网络即可生成和维护本地稿。Go独立运行工具，Windows使用对应.exe，；实际设备/浏览器执行与真实Jira写入按目标电脑能力和授权，不存在时记录未验证/未提交。原技能名称迁移见指南。

[持续需求模拟样例](examples/qa-continuous-change/README.md)包含三版本、一次实际运行的模型失败、修订隔离、撤回、模拟Jira部分失败及报告复算。它不代表真机或远程Jira验收。

隐私复查与本地元数据例外见 [分发隐私审计](docs/privacy-audit.md)。手动打包请排除 `.git/` 与工具缓存。
