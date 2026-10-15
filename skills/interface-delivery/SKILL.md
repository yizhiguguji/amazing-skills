---
name: interface-delivery
description: 依据产品目标和现有规范设计、实现或修正 Web 界面，可将模糊视觉意图接入参考提取与组件预览；不重新定义业务范围。
---
# 界面交付

输入规格、已有代码与 DESIGN.md。如核心数据源或任务仍不明确，报告具体缺口；不要自行增加模块填补空白。

规格与视觉输入可以不完整。已有上下文能支持的部分先推进：将抽象诉求转成标记为推断的可撤销方向，只询问会改变任务、方向或范围的缺口。工作台缺业务定义时交给 workbench-definition；仅视觉模糊时不重开产品访谈。

需要提取参考、建立或整理设计系统时读取 ../design-reference-extraction/SKILL.md 及其包契约，形成来源明确的 DESIGN.md、tokens 与组件预览。已有规范只补受影响部分，无参考也可从任务推导，不强迫用户选样例。空白项目在 project init 前 apply 资源包；已有 DESIGN.md 或代码先继承并按授权合并，apply 拒绝覆盖不是要求用户重新授权已明确的局部修改。

先在浏览器查看包的 design-preview.html，再把同一 tokens.css 接入实际页面（例如 link/import），不要复制一套色值。必须记录实际导入路径；生成预览不是产品实现。若应用需要变量映射，记录源变量与目标变量关系，并在实际界面核对计算样式。组件预览通过后仍执行产品状态与行为验收。无浏览器时记录视觉未验证，继续不依赖它的本地工作。

使用 Impeccable 前读 [references/impeccable.md](references/impeccable.md)，使用 UI UX Pro Max 前读 [references/ui-ux-pro-max.md](references/ui-ux-pro-max.md)。确认实际可用后执行对应技能指令；未安装时使用内置方法，明确记录 provider=builtin。

内置方法：从实际场景确定布局、密度、字阶、语义色、间距、组件、状态与响应式，写入 DESIGN.md；业务工作台以任务完成为目标。代码复用现有 tokens/组件，无现成规范时建立最小一套。已存在设计系统的项目继承其事实，不因为缺 DESIGN.md 就重设计。

实现每个范围内页面与操作，接真实授权的数据或已说明的本地原型存储。冷启动文案来自规格；测试样本只放测试中。加载、错误、空数据、长文本和键盘焦点按相关功能实现。不把前端隐藏当成服务端授权。

中英文混排需要中文字体回退，不直接套仅覆盖拉丁字符的字体效果。不为视觉对称改变页面形态、任务或数据模型；必要变更更新相应依据并记录原因。

实现后运行项目适用构建和行为检查，查看实际渲染，交给 workbench-verification 或等价验收。局部修改只检查受影响路径；保留既定字体、配色与组件。

本仓库已捆绑两个提供者。使用 amazing-skills providers doctor 定位，然后读取各实际 SKILL.md。真实工具通过 helper 的 run 子命令执行；Pro Max 路径来自技能目录，不能依赖 CLAUDE_PLUGIN_ROOT。

工具名`amazing-skills`指skills根下`.amazing-skills/amazing-skills`（Windows加.exe）；用真实安装绝对路径调用。仓库维护时在根目录用`go run ./cmd/amazing-skills`。正常路径不需要Python，Pro Max查询按本地适配规则运行。
