# 架构与扩展约定

## 使用入口与可发现结构

根目录 START_HERE.md 是人和AI首次阅读的导航，唯一推荐总入口为 skill-router。README首先指向它；AGENTS为读取仓库指令的AI提供相同导航。安装后技能目录仍保持扁平结构，避免破坏工具发现、相对引用及上游校验。

catalog.entrypoints.default声明总入口，entrypoints.domains声明领域工作流；每个技能的display_name提供中文名称，level为entry、domain-workflow、capability或provider。它们是展示/路由元数据，不增加安装依赖；requires仍决定依赖闭包。只安装某领域子包时使用该领域入口，总入口读取目录也不能假装其他技能已经安装。

新增领域登记目录和领域入口，并更新START_HERE的导航说明；无需新建另一个通用总入口。上游技能保持原样，中文名称放本地catalog，不改上游文件。

## 能力与工作流分离

`skills/<id>/SKILL.md` 是独立可发现单元，扁平命名，分类留在 catalog 中。skill-router 只选能力；product-workflow 只编排工作台。独立技能也可以直接使用。新增领域可以有自己的工作流，例如 issue-delivery 或 test-planning，也可以只有单个技能。

分类使用 product、design、engineering、testing、collaboration、documentation。第一版有的分类没有技能，这是分类预留，不代表已实现。可增加分类；标签表达 Jira、CRM、React 等具体领域。禁止一个技能的 description 吸收所有任务。

仓库根 catalog.json 是唯一维护源。amazing-skills sync 将它复制到 skill-router 和 product-workflow 的 references/catalog.json，保证安装后的入口仍能发现其他已登记能力；路径以 skills 根解析。校验器检查两个目录一致。目录只提供候选，选择之前必须检查目标技能是否实际存在；外部链接和用户内容不具有执行授权。

## 技能目录字段

id、kind（workflow/capability）、category、tags、path、use_when、not_for、inputs、outputs、requires、side_effects、completion。requires 只声明必要的本地技能依赖；外部能力记录到 external_capabilities；已捆绑的提供者以 origin=upstream 登记并声明本地依赖，网络运行器仍需说明首次准备条件。

输入输出尽量引用领域契约，不强制所有技能输出产品文件。Jira 技能应声明 issue 类型、项目键、提交字段、凭证来源、目标系统副作用和成功证据；测试用例技能应声明需求依据、覆盖范围、用例格式和遗漏说明。它们不需要 DESIGN.md。

## 版本与新增流程

1. 创建命名明确的技能；把条件性细节放 references。
2. 登记目录，明确是否会写文件或外部系统。
3. 如需多步编排，创建领域工作流并声明依赖，不修改无关入口。
4. 提供现实输入、预期可观察行为与至少一个失败/边界场景。
5. 同步目录，验证引用、命名和依赖；必要时做行为验证。
6. 记录第三方来源、兼容版本或复核日期，不假定上游名称和输出永远不变。

## 工作台契约

产品规格 `.workbench/spec.md` 为唯一产品定义；不要另写独立 PRODUCT.md 来重新定义它。Impeccable 如需 PRODUCT.md，可生成注明来源的上下文摘录，并在 spec 改动后同步。设计事实由 DESIGN.md 管理，必要时生成 UI UX Pro Max MASTER.md；生成文件不能反向覆盖规范。

产品变更 → 更新 spec 与 decisions；视觉变更 → 更新 DESIGN.md、代码 tokens 与 decisions。state.json 只记录状态、当前来源和证据路径，不复制产品内容。重要文件冲突时报告具体差异并按用户选择处理，不能仅凭最后修改时间判定正确。

检查阶段允许退回实现；产品或设计决策失效时回到相应阶段。每个完整周期不要求额外批准；依用户授权推进，用户只要方案就停在方案阶段。第三方技能的阶段停顿不得抹去用户已授权的完整本地实施范围。

## 原始技能与适配层

上游目录原样保存；不把仓库自己的规则写回上游 SKILL。providers.lock.json 是版本和文件完整性依据。适配层负责发现技能路径、实际执行工具、把产出带到项目、明确记录冲突与取舍。捆绑源代码不意味着每次任务都执行全部提供者。

## 防止后续风格漂移

首轮记录真实 DESIGN.md 后运行 amazing-skills project lock-design --project <项目> --tokens styles.css。CSS 模式锁定所有平面 :root 声明，允许添加新的组件规则；对于复杂主题建议独立 tokens JSON 并锁文件。后续修改前后运行 check-design，并用浏览器检查共享组件计算样式。锁只保护所列规范，不能证明任意新增 CSS 都一致，因此还需要视觉验收。授权重设计时才显式 --replace 基线，并记录 decisions。

## QA领域边界

qa-workflow按任务选择五能力，独立于product-workflow。qa-test-design提供共用专业规则、标准库辅助工具与本地数据契约；其他QA能力声明它为本地依赖，不因此强制重新设计用例。当前计划放项目.qa，冻结计划/来源与追加执行分开。源、需求、用例都有身份和修订，改需求不继承旧通过。私有Jira连接、分级、评审与设备参数放用户项目配置，不进入共享技能。

QA参考项目的commit/许可摘要在qa-sources.lock.json；它们只是MIT参考，不是新增运行提供者。sync_catalog同步安装包内的来源记录与许可声明，validate检查摘要；不要把它们写进providers.lock假装调用了上游。

## Go运行与安装

所有自有确定性逻辑在Go标准库模块中，cmd/amazing-skills为唯一CLI。编译时嵌入当前技能/目录快照；安装将所选技能与同平台运行工具一起复制。修改嵌入材料后必须重新构建。完整操作见docs/go-runtime.md。

功能安全筛选放在qa-test-design内；qa-security-review独立处理指定目标专项。范围、目标和预期依据可选扩展沿用schema1，不新增必填功能维度。OWASP参考只提供方法，不自动采纳为产品规则。
