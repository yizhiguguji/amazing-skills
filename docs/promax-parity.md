# UI UX Pro Max：Go与原版的差异

基准是本仓库锁定的 `nextlevelbuilder/ui-ux-pro-max-skill` commit `50d8a7de0900119855614541f15a1a616691eb33`，不是对未来上游版本的兼容承诺。原始技能、脚本、CSV、许可证和文件摘要均不修改。

原先0.4.0的Go工具是简化适配，存在实际能力缩水：查询只做基础BM25，自动分类不完整，没有原版置信度门槛、建议重试、别名与版本规则；设计系统只有候选列表，不能生成完整主方案或页面覆盖。换语言不必造成这些损失，本次已经补齐。

| 能力 | 原先Go适配 | 当前Go移植 |
|---|---|---|
| 查询归一化 | 简单分词 | 原版同义词、停用词、最短词长、下划线规则 |
| 自动分类 | 少量关键词 | 12领域的权重和固定优先顺序、精确风格优先 |
| 排名与拒答 | BM25有分就返回 | 原版BM25参数、查询重写、分领域置信度门槛、词覆盖率 |
| 无命中 | 空结果 | 相近词建议、完整页面模式建议、明确默认回退 |
| 风格 | 排除过时行 | 风格ID/名称/别名、包含身份匹配、父风格与跨领域重定向 |
| 技术栈 | 全列检索、无代际筛选 | 22栈、固定搜索/输出列、当前/旧版筛选、迁移意图、shadcn基础库、精确API身份与UWP后继建议 |
| 设计系统 | 五组候选 | 产品类型、封闭决策规则、权威风格选择、布局、字体、16种语义颜色、暗色推导、反模式、来源追踪 |
| 设计参数 | 只记录参数 | variance影响选风格；motion选择真实GSAP片段；density生成间距，支持低/中/高档 |
| 方案文件 | 不支持 | MASTER.md、页面覆盖、原子写入、默认不覆盖、显式force |
| 输出 | 一律JSON | JSON、Markdown、纯文本、full；代码片段不截断 |

## 原版对照证据

`testdata/promax-parity.json`直接由未修改的原版运行生成，Go测试读取这些结果作对照，不是由Go代码生成期望值。参考运行使用开发环境已有的Python，只用于维护时制作基准；分发程序不执行Python、不嵌入Python解释器，也没有第三方Go运行依赖。

对照包括988个查询、808个完整设计系统组合、11种页面覆盖，总计1,807个场景。覆盖全部12查询领域、22技术栈、风格/模式的公开身份及别名、过时重定向、无命中/拼写建议、框架版本/基础库、产品类别、默认回退、暗色请求、三个设计参数的低/中/高档。比较全部结构化字段及结果顺序；排序浮点分数允许1e-10相对/绝对容差。

额外验证主方案默认不覆盖、页面新增保留主方案、force显式替换、封闭规则拒绝重复键/未知条件/未知动作、路径逃逸拒绝。完整安装测试把工作台技能安装后移动目录，在PATH中没有Go和Python的情况下生成并持久化完整方案，再用另一个需求确认主方案未被覆盖。

## 仍然存在的区别

- **排版不同。** ASCII采用纯文本而非原版彩色框；Markdown/MASTER.md按相同结果组织信息，包含原版组件CSS、间距、阴影、反模式、检查清单，并增加规则与来源记录。输出不逐字相同，不使用原版ANSI色块和生成时间戳。
- **路径检查更严格。** 在原版slug和默认不覆盖之外，拒绝通过符号链接写出目标项目。部分不安全路径会报错。
- **维护边界。** 静态契约来自锁定版，升级上游要同步Go契约、移植变更并重新生成原版对照，不可只替换CSV或抬高版本号。对照集通过不证明所有可能输入都完全等价。
- **实际页面效果。** 对照证明检查场景的推荐数据和推理没有因Go迁移变化；AI的实现、渲染和交互仍需Impeccable及项目验收检查。中文语义查询沿用原版英文关键词检索能力，编排AI先提取英文检索词；不把没有命中当成已理解需求。

## 使用和方案固化

```bash
amazing-skills search "SaaS dashboard dark mode" --design-system \
  --project-name "Example" --density 8 --motion 2 --json

amazing-skills search "SaaS dashboard dark mode" --design-system \
  --project-name "Example" --persist --output-dir /目标项目 --format markdown

amazing-skills search "dashboard keyboard navigation" --design-system \
  --project-name "Example" --persist --page dashboard --output-dir /目标项目
```

实际使用安装根 `.amazing-skills/amazing-skills` 的绝对路径。需要提供者执行证据时通过 `providers run --project /目标项目 --provider ui-ux-pro-max -- ...` 调用，implementation记录为 `amazing-skills-go-promax`。

`DESIGN.md`管理选定规范；MASTER.md/页面覆盖是它的投影。已有规范先读后改，只查相关领域，避免再生成另一套全局风格。`skipped_exists`表示磁盘主方案保留，不能把本次内存推荐当成已采用方案。Go工具支持显式force，但编排必须遵守项目变更授权和现有设计约束。

## 来源与维护

`search.go`、`promax_design.go`及`promax/contract.json`移植/派生自原版core.py、design_system.py、reasoning_contract.py；`promax/components.md`、`shadows.md`、`checklist.md`保留原版组件/规则。依原版MIT许可使用，署名和完整许可证见 `skills/ui-ux-pro-max/LICENSE`；第三方源文件与锁定摘要不改动。

维护者运行 `go test ./...`、`go vet ./...`、`go run ./cmd/amazing-skills sync`、`go run ./cmd/amazing-skills validate`，再重建六平台程序。对照结果见 [验证记录](verification/promax-parity.json)。六平台交叉构建不等于在六种系统实际运行。
