# 设计系统包 v1

这是本仓库的可携带设计资源，不是新增技能层级或产品规格。可独立生成，也可在工作台流程中使用；实际调用只用 Go 标准库工具。

## 文件与真源

| 文件 | 所有权 |
| --- | --- |
| design-system.json | schema、稳定 id、语义版本、名称、说明、source（origin/license/revision）及固定 files 映射 |
| DESIGN.md | 设计意图、适用条件、布局、响应式、组件状态、资产与例外；章节随实际系统组织 |
| tokens.css | 精确 token 值的唯一来源；DESIGN.md 引用变量名，避免重复抄值 |
| design-evidence.json | schema_version=1、非空 claims、显式 unknowns 列表；观察/提供/推断与决策依据 |
| design-preview.html | Go 从同一包生成的组件样例及摘要，派生产物，不反向成为设计依据 |
| design-system.LICENSE | 包所用材料的实际许可；若自建沿用仓库 MIT，第三方需逐项核实 |
| sources/（可选） | 观察证据，file 相对路径及 SHA256 由 claim 声明；不含凭证或私人样本 |

最小 manifest 示例见 [../assets/systems/calm-paper/design-system.json](../assets/systems/calm-paper/design-system.json)，证据字段示例见 [../assets/systems/calm-paper/design-evidence.json](../assets/systems/calm-paper/design-evidence.json)。ID 用小写字母开头的字母/数字/连字符，版本为 x.y.z；source 必须保留来源与修订。修改设计后提高包版本，更新来源与 decisions。工具不验证来源声明的真实性，不自动获取网址。

files 固定映射为 design=DESIGN.md、tokens=tokens.css、evidence=design-evidence.json、preview=design-preview.html、license=design-system.LICENSE；不允许重定向到包外。所有被消费文件为普通文件，不接受符号链接；单文件上限 2 MiB。来源路径不得带 .git、node_modules、__pycache__、.DS_Store 或 tool-cache。observed 必须有 sources/ 内证据与匹配摘要；provided 的会话摘录、inferred 的理由及适用前提也要明确定位。未知不强制清零，但影响关键验收的未知必须保持未验证。

## Token 子集

仅一个平面 :root 块，变量名以 --ds- 开头；允许注释、常规数字/单位/颜色/字体列表及简单 var(--ds-name) 引用，引用必须存在且无循环。不能包含 URL、导入、脚本、HTML 或应用选择器。它是便携语义层；复杂媒体查询、语言/主题选择器和应用组件规则放项目 CSS，记录映射并按实际浏览器检查。无法无损表示的现有系统保留原始 CSS，本包输出一个经过说明的语义投影，不强行重写项目。

必须提供：

- 颜色：--ds-bg、--ds-surface、--ds-fg、--ds-muted、--ds-border、--ds-accent、--ds-accent-on、--ds-danger。
- 排版：--ds-font-body、--ds-font-display、--ds-text、--ds-leading；字体包含适用的中文回退。
- 几何：--ds-space、--ds-radius、--ds-content、--ds-control。

值是否适用、对比度、触控尺寸和实际字体均需浏览器验证，结构检查不是完整 CSS 解析器，也不证明无障碍合规。包预览使用同一变量；不要写第二套预览色值。应用的 tokens 必须直接导入本文件或给出受检的映射，不能仅“看起来一样”。

## 命令与接续

`amazing-skills` 指 skills 根 `.amazing-skills/amazing-skills`，Windows 用 .exe；源码工作在仓库根用 `go run ./cmd/amazing-skills`。

```sh
amazing-skills design-system list
amazing-skills design-system init --preset calm-paper --output /新目录/设计包
amazing-skills design-system check --package /新目录/设计包
amazing-skills design-system preview --package /新目录/设计包 --replace
amazing-skills design-system apply --package /新目录/设计包 --project /目标项目
amazing-skills project init --project /目标项目
```

init 的两套包为自建合成起点，不能据此声称已完成真实提取；先把方向、证据与 tokens 改成当前任务，再生成预览。也可以直接创建上述文件，不选模板。preview --replace 仅替换派生预览；apply 没有覆盖选项，遇到已有文件或设计锁时整批拒绝。已有项目先提取现状，按已授权范围直接合并，不用删除旧文件来绕过检查，也不反复请求已授予的权限。

apply 在项目根落实同一包并写 `.product/design-system-selection.json`（身份/版本/manifest 摘要）；不改产品定义、阶段或验收状态。随后 project init 只补缺失模板。已有 workbench spec 优先继承；填入真实草案、假设和可检查验收，不把模板初始化当需求完成。

打开 design-preview.html，记录路由/文件、尺寸、浏览器和日期，分别检查行为、视觉和未验证。接入实际页面后运行构建、观察计算样式及范围内状态，再锁定：

```sh
amazing-skills project lock-design --project /目标项目 --tokens tokens.css
amazing-skills project check-design --project /目标项目
```

如果实现改用其他 tokens 路径就锁定真实文件；CSS 锁仅保护平面 :root，复杂主题可以锁完整文件/JSON，按现有契约处理。后续修改先后 check-design；授权重设计才 --replace，并记 decisions。check 验证结构、证据摘要、引用和生成预览的新鲜度，不把真实产品标为通过。验收继续用现有 workbench-verification 或当前界面的等价流程。
