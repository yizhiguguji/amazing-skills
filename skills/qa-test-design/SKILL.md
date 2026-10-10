---
name: qa-test-design
description: 从需求文档及可选Figma生成、审查和持续维护范围受控的功能测试用例；选择正反流程、交互、边界、状态、环境与功能相关安全维度，维护来源追踪和变更影响，不自动执行测试或提交Jira。
---
# 功能测试用例设计与维护

运行工具为Go独立程序；下文 `amazing-skills` 指 skills 根下 `.amazing-skills/amazing-skills`（Windows加.exe）。优先使用实际安装位置的绝对路径；仓库开发可在根目录用 `go run ./cmd/amazing-skills`。仅阅读Skill不需要任何语言运行时。

读取 [专业规则](references/professional-rules.md)。首次设计按 [设计方法](references/design-methods.md)，有文档/Figma或版本冲突时读 [来源和变更](references/sources-and-change.md)，选择后台/网络/温度等维度时读 [功能维度](references/functional-dimensions.md)。只加载任务相关引用。每次设计按 [功能安全筛选](references/functional-security.md) 审查本期相关的身份、权限、输入、敏感信息和数据完整性风险；不要求所有类别都有用例。

1. 明确本期交付、纳入/排除、平台、验收点、依据优先级和已有基线。不要把正式依据缺失的入口、时限、默认值、账号、等级或阈值补成产品规则。
2. 将需求映射为测试条件：触发、前置状态、数据边界、观察位置和判定依据。先选适用方法，再形成具体场景；不按测试类型无差别复制用例。
3. 用维度决策表记录交互、正向、反向、边界、状态、切应用、后台、锁屏、断网、弱网、温度的纳入/不适用/待明确/延期及理由。纳入的维度必须有相关需求或确认基线，以及对应场景；未知行为不写固定预期。
安全候选记录include/not-applicable/clarify/deferred及范围理由；有确认依据才纳入正式用例，未知预期保留缺口和proposed，不编造账号体系、错误码或策略阈值。原11维度保持不变，安全作为相关场景标签，复用已有用例ID。明确要安全专项时交接qa-security-review；普通“补全”不触发全站扫描。

4. 每条用例有稳定ID、修订、单一测试目标、模块/子功能、类别、影响级别、需求ID及修订、来源、前置、数据、步骤和可观察预期。一个目标可有多步操作和必要断言；不同目标/预期才拆分。平台相同用参数化环境，正式差异才拆编号。
5. 检查可准备数据、真实入口、重复、追踪和范围。普通设计不要求用户批准方法栈；不明确的业务预期进入缺口，已明确内容继续交付。评审准备不冒充审批；免评审、编号、优先级映射由项目决定。
6. 需求细化/推翻按来源变更逐条比较，保留编号与历史；改语义增加用例修订，撤回用例标记 retired；旧结果不算新修订通过。每次交付写变更理由、受影响和未受影响项。

无既有系统时使用本目录 [数据契约](references/data-contract.md) 和 [项目配置模板](assets/config.json)：

```bash
/安装目录/.amazing-skills/amazing-skills qa init --project /目标项目
/安装目录/.amazing-skills/amazing-skills qa check --project /目标项目
/安装目录/.amazing-skills/amazing-skills qa export-cases --project /目标项目
/安装目录/.amazing-skills/amazing-skills qa snapshot --project /目标项目 --label r1
/安装目录/.amazing-skills/amazing-skills qa impact --project /目标项目 --before r1
```

Windows使用对应.exe，Go独立运行工具，无Python依赖，不依赖私有系统或第三方Skill。init只补缺文件，snapshot拒绝覆盖；check是结构/一致性检查，不证明业务覆盖、审批或真实执行。脚本不替模型生成需求和预期。

交付用例、需求追踪、方法依据、维度决策、缺口和版本差异；需要表格时按用户格式导出，不强制固定列数。无法同步系统时明确“已生成本地稿，未入库”。

来源与许可说明见 [开源参考](references/open-source-sources.md)，仅查归属时读取，不是运行依赖。
