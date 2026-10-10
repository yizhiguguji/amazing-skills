---
name: workbench-definition
description: 将个人或业务工作台的模糊需求整理为可实现的产品规格，明确核心任务、数据来源、空状态、页面和验收；不用于纯视觉修改。
---
# 工作台定义

已有 `.workbench/spec.md` 先读并修订，不重开需求访谈。输出到该文件；用户指定其他规格位置时遵循其位置并把交接指向真实文件。

finesse-brief 实际可用时读取 [references/finesse-brief.md](references/finesse-brief.md)。否则按内置方法：找到主要用户、主对象、每日/实际使用任务与现有替代工具；先给完整草案，只询问阻碍决策的缺口。不是所有工具都需要每日访问。

逐条定义：核心任务、首屏信息及下一动作、模块与页面形态、字段和关系、每个字段由用户/系统/集成/计算谁写入、数据更新时机、无数据/断开/漏记行为、MVP、排除与延后、假设和可观察验收。外部数据源未确认不能当成已有接口；无数据不能显示假统计。

仅依据明确事实或标记为假设的决策生成方案。个人页不强加复杂权限；业务系统的授权不能用前端筛选代替。每个验收给编号，例如 AC-01，描述输入、动作、预期结果而非只写“页面美观”。

末尾给实现方：首屏顺序、页面形态、真实数据缺口、范围、角色、状态和设计方向。输出草案后按用户授权交接，不自动扩大到未请求的实现。

本仓库已捆绑 finesse-brief。先通过 amazing-skills providers doctor --provider finesse-brief 定位，再实际读取 ../finesse-brief/SKILL.md 与适用引用。该提供者没有 CLI；证据记录已读文件和产出，不编造命令退出码。

工具名`amazing-skills`指skills根下`.amazing-skills/amazing-skills`（Windows加.exe）；用真实安装绝对路径调用。仓库维护时在根目录用`go run ./cmd/amazing-skills`。正常路径不需要Python，Pro Max查询按本地适配规则运行。
