# 提供者：原始技能与Go运行适配

固定来源和文件摘要见providers.lock.json；第三方目录保留原始内容与许可证。技能是模型读取的操作指南，工具执行是另一个动作，二者分别记录。

## 每个提供者实际做什么

- finesse-brief：模型读取SKILL.md及适用引用，输出工作台规格，没有独立可执行程序。
- Impeccable：模型读取原技能；Go工具执行锁定、校验过的官方引擎context/detect等命令。必须显式准备引擎，不会因读取技能静默下载。
- UI UX Pro Max：模型读取原技能和适配交接；Go移植原版检索与设计系统逻辑，读取原CSV，不调用上游Python。对照范围与剩余差异见 [原版兼容说明](promax-parity.md)。

Go的`--design-system`按原版推理规则选择产品、风格、颜色、字体、页面布局与可选动效/间距，输出来源和激活规则；支持`--persist --page --output-dir --force`。这是设计建议，不是产品验收。模型将适用方案固化到DESIGN.md及代码tokens；MASTER.md/页面覆盖保持为规范投影。已有规范仅进行相关领域查询，禁止重新全局生成覆盖原风格。

## 准备和执行

在仓库根目录（构建需要Go 1.23+）：

```bash
go run ./cmd/amazing-skills providers doctor
go run ./cmd/amazing-skills providers setup-engine --project /目标项目
go run ./cmd/amazing-skills providers run --project /目标项目 --provider impeccable -- context
go run ./cmd/amazing-skills providers run --project /目标项目 --provider impeccable -- detect --json index.html styles.css app.js
go run ./cmd/amazing-skills providers run --project /目标项目 --provider ui-ux-pro-max -- "project dashboard restrained" --design-system --density 8 --motion 2 --json
```

使用已安装独立程序时将`go run ./cmd/amazing-skills`替换为`/实际skills目录/.amazing-skills/amazing-skills`，Windows加.exe。程序会从自身位置发现skills根；也可显式传`--skills-root`。检查目标技能真实存在且摘要通过之后才能执行。

setup-engine使用Go标准库从官方GitHub固定release下载，按锁文件SHA-256校验，落到目标项目`.product/tool-cache/`。支持的平台由锁文件摘要表决定，Windows ARM64按上游规则使用x64引擎（实际平台兼容仍需验证）。有同版本缓存且摘要正确时不联网；缺失缓存时run明确要求先准备，不通过launcher隐式下载。无需Python、Node或摘要命令来准备/调用引擎；具体项目构建和浏览器检查可能有自己的工具需求。

## 证据与边界

Go工具写provider-events.jsonl及provider-evidence输出。事件记录implementation：Impeccable为pinned-engine，Pro Max为amazing-skills-go-promax。本机项目/安装根在日志里替换为占位符，并注明是脱敏展示副本；内容中其他凭证或私有参数仍需分发前检查。

doctor只是发现和完整性检查，不等于调用了Skill或产品验收。历史示例中的Python工具日志保留为历史事实，不能重新标成Go执行。新Go验证见 [迁移记录](go-runtime.md)。示例应用不参与安装或日常技能运行。
