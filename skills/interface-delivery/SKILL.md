---
name: interface-delivery
description: 依据明确产品规格和固定设计规范实现或修正 Web 界面，支持 Impeccable 与 UI UX Pro Max 适配；不重新定义业务范围。
---
# 界面交付

输入规格、已有代码与 DESIGN.md。如核心数据源或任务仍不明确，报告具体缺口；不要自行增加模块填补空白。

使用 Impeccable 前读 [references/impeccable.md](references/impeccable.md)，使用 UI UX Pro Max 前读 [references/ui-ux-pro-max.md](references/ui-ux-pro-max.md)。确认实际可用后执行对应技能指令；未安装时使用内置方法，明确记录 provider=builtin。

内置方法：从实际场景确定布局、密度、字阶、语义色、间距、组件、状态与响应式，写入 DESIGN.md；业务工作台以任务完成为目标。代码复用现有 tokens/组件，无现成规范时建立最小一套。已存在设计系统的项目继承其事实，不因为缺 DESIGN.md 就重设计。

实现每个范围内页面与操作，接真实授权的数据或已说明的本地原型存储。冷启动文案来自规格；测试样本只放测试中。加载、错误、空数据、长文本和键盘焦点按相关功能实现。不把前端隐藏当成服务端授权。

中英文混排需要中文字体回退，不直接套仅覆盖拉丁字符的字体效果。不为视觉对称改变页面形态、任务或数据模型；必要变更更新相应依据并记录原因。

实现后运行项目适用构建和行为检查，查看实际渲染，交给 workbench-verification 或等价验收。局部修改只检查受影响路径；保留既定字体、配色与组件。

本仓库已捆绑两个提供者。使用 amazing-skills providers doctor 定位，然后读取各实际 SKILL.md。真实工具通过 helper 的 run 子命令执行；Pro Max 路径来自技能目录，不能依赖 CLAUDE_PLUGIN_ROOT。

工具名`amazing-skills`指skills根下`.amazing-skills/amazing-skills`（Windows加.exe）；用真实安装绝对路径调用。仓库维护时在根目录用`go run ./cmd/amazing-skills`。正常路径不需要Python，Pro Max查询按本地适配规则运行。
