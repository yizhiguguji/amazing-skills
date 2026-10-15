---
name: product-workflow
description: 将模糊的个人或业务工作台需求推进为规格、固定设计规范、本地实现与验收；已有工作台从当前阶段继续。不用于独立 Jira 操作或一般测试文档。
---
# 工作台编排

运行工具为Go独立程序；下文 `amazing-skills` 指 skills 根下 `.amazing-skills/amazing-skills`（Windows加.exe）。优先使用实际安装位置的绝对路径；仓库开发可在根目录用 `go run ./cmd/amazing-skills`。仅阅读Skill不需要任何语言运行时。

技能目录在 [references/catalog.json](references/catalog.json)。以本技能目录的父目录为 skills 根，按需读取 workbench-definition、design-reference-extraction、interface-delivery、workbench-verification。若缺失，明确报告缺失并按已读契约继续可做的部分，不能声称调用成功。

先读取用户范围、目标项目适用指令、现有代码、`.product/state.json`、`.workbench/spec.md` 与 `DESIGN.md`。已有内容优先，不重开发现流程。只有工作台缺定义时进入 definition；已有规格则进入 design/implement；局部修改直接实现并针对性验证。用户仅请求定义时不进入代码阶段。参考 [references/workflow.md](references/workflow.md) 判断跳转。

需要新方案文件时使用 assets/templates；可运行 `amazing-skills project init --project <target>`，该脚本只创建缺失文件，不能完成定义或验收。在消费文件前将模板填成当前项目的事实、假设和决策。

抽象需求也能开始：先复用上下文形成工作草案与标记的假设，只有会改变核心任务、数据可行性或方向的缺口才询问。业务定义沿用 workbench-definition；视觉方向需要提取时进入 design-reference-extraction，再以设计包与组件预览交给 interface-delivery。空白项目若选用包，先 apply 再 project init；已有模板或规范按当前授权合并，不为应用包删除现有文件。组件预览与实际产品验收分开记录。

按授权范围逐步推进，每阶段结束写 `.product/state.json`：阶段、状态、选择的能力与原因、未解决事项、证据路径。调用外部 Skill 前读取对应适配器并确认实际可用；不自动下载安装。未安装使用能力技能中的内置流程，并在状态里标为 builtin。

产品定义以 `.workbench/spec.md` 为准，视觉以 `DESIGN.md` 为准；原生上下文文件是有来源的投影。后续实现先继承设计规范和已有组件，不重复搜索或改风格。定义与视觉变更分别登记 decisions，只有失效的验收重跑。

完成要有运行与验收证据。`amazing-skills project check --project <target>` 仅检查文档状态，不能证明产品可用。无法运行浏览器时标明未验证，不写 complete；参考 [references/artifacts.md](references/artifacts.md)。

首轮使用捆绑提供者并记录实际加载文件及真实工具执行。后续修改先读 spec 和 DESIGN.md，使用 amazing-skills project 的风格锁检查；原有规范不再全局生成。新电脑使用本仓库Go工具与校验过的Impeccable引擎；不依赖创建者电脑的全局配置。
