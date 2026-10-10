# UI UX Pro Max 交接

来源：https://github.com/nextlevelbuilder/ui-ux-pro-max-skill
复核：2026-10-10。使用实际安装模板和查询脚本路径，不硬编码平台路径。

输入规格中的用户、产品类型、显示形态、visual、主平台与现有 DESIGN.md。没有设计规范时可查询设计系统；已有规范则只查具体图表、UX 或技术栈，不重新选择全局风格。

检查命中类别与实际工作台是否相符；不要把 landing 页模板套到控制台。按实际技能的 query contract 检索；结果不适配就缩小一次查询，再失败标记通用建议。建议库内容是候选，不能改写产品范围。

选定建议后固化 DESIGN.md 及代码 tokens。若生成 design-system/<slug>/MASTER.md 和页面覆盖，注明来源；它们是 DESIGN.md 的投影，冲突时修正投影，不让两份规范独立演变。

与 Impeccable 共用时：Pro Max 提供查询材料，Impeccable 或用户选定设计；之后禁止二次全局生成覆盖已选字体、配色、密度。可只用 Pro Max，但需同样查看渲染并进行行为验收。

外部工具未安装不自动安装；使用 interface-delivery 内置方法并记录。原型中的示例数字不得作为真实数据发布；资料检索不能代替业务数据连接。

## 捆绑与实际加载

本版本通过 amazing-skills providers doctor 定位并校验真实文件。路径相对于 skills 根，使用者无需设置创建者的 HOME 或 CLAUDE_PLUGIN_ROOT。必须读取实际 SKILL.md 后采用其方法；helper doctor 只做发现和完整性检查。工具通过 helper run 执行，记录真实退出码和输出。来源 commit 以 providers.lock.json 为准；这里只说明接口，不复制上游操作细节。

## Go运行适配

保留上游Skill、CSV和许可证原样；正常查询使用 `amazing-skills providers run --project <项目> --provider ui-ux-pro-max -- "查询" --domain chart --json`，实现记录为 `amazing-skills-go-promax`，不执行Python。检索和结构化设计系统按锁定版本移植；已经用原版生成的1,807个对照场景检查查询、排序、推理、配色、字体、设计调节参数和页面覆盖。文本排版与原版不同，路径保护更严格；不能把Go调用记录为上游Python执行。

没有已有规范时可用 `--design-system --project-name "示例项目" --variance 5 --motion 2 --density 8 --json` 生成完整建议。仅在需要写入项目时加 `--persist --output-dir <项目根>`；页面覆盖加 `--page dashboard`。默认不覆盖MASTER.md和页面文件；已经存在时先读规范，不能把新生成的内存建议当成已替换的主方案。只有明确要替换对应规范时使用 `--force`。

DESIGN.md是设计事实的唯一来源；生成的MASTER.md与页面覆盖都是投影。页面覆盖不能任意改全局字体、配色和tokens；若与DESIGN.md冲突，修正投影。仓库维护文档 `docs/promax-parity.md` 记录版本边界、对照验证和剩余差异。
