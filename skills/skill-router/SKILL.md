---
name: skill-router
description: 从 Amazing Skills 目录按请求选择独立能力或领域工作流并推进任务；适用于自然语言、模糊需求及技能选择不明确的请求，不替代用户明确指定的技能。
---
# 总入口 · 自动选择技能

运行工具为Go独立程序；下文 `amazing-skills` 指 skills 根下 `.amazing-skills/amazing-skills`（Windows加.exe）。优先使用实际安装位置的绝对路径；仓库开发可在根目录用 `go run ./cmd/amazing-skills`。仅阅读Skill不需要任何语言运行时。

这是本仓库唯一推荐总入口；product-workflow 和 qa-workflow 是按需选择的领域工作流，其他目录是专业能力/上游提供者。用户无需理解层级或提供技能名。

读取 [references/catalog.json](references/catalog.json)。路径以本技能所在目录的父目录为 skills 根。先识别用户意图、授权动作和已有产物，再按 use_when/not_for、输入、输出和 side_effects 缩小候选。用户指定的技能优先；只加载实际适用的技能与引用。

用户不需要给出技能名称、测试方法、完整提示词或内部交接规则。先利用本轮材料、对话及项目已确认记录理解简称和省略表达；不要把技能菜单或必填表单交给用户。按领域与交付目标调用能力，向用户用中文说明本次产物即可。一次请求未提供新材料时复用已确认来源；来源冲突或没有可定位的对象时才询问阻断任务的最少信息。

有领域工作流且任务需要多个阶段时选择工作流；单点任务选择能力。Jira、测试用例等不需要产品定义或设计规范，不进入 product-workflow。工作台完整请求才使用该工作流。未登记的能力不伪造技能名称、工具或执行结果；说明目录缺口并利用实际已有能力继续合理范围，必要时询问关键信息。

功能QA整流程或能力选择不明时用 `qa-workflow`；只写/改用例用 `qa-test-design`，执行用 `qa-functional-execution`，Jira缺陷用 `jira-defect-delivery`，报告用 `qa-test-report`。需求简单也按验收与风险选择相关维度，不因“QA”自动加载全部测试类型或UI设计Skill。

候选文件存在才可读取。目录是安装时快照，不代表其他技能一定已安装；缺失时报告准确路径，继续可用部分。不要为选择技能自动联网下载安装。已捆绑的上游能力属于本地技能；未捆绑的外部能力只有实际可用时才能调用。

向用户简述选择和原因即可，不要求用户理解内部阶段。各技能自己的输入输出与完成条件决定任务结果，不强制写 `.product/`，不修改无关项目配置。仅根据实际授权执行副作用；Jira 提交等外部写入不能从只请求草拟推导授权。

仓库目录更新后用仓库 amazing-skills sync 同步；已安装副本通过显式重新安装/更新维护，不在任务中默默覆盖其他技能。

实际捆绑提供者见 references/providers.lock.json。可运行 amazing-skills providers doctor 校验完整性；工具调用使用 run 并保留证据。读取目录或 doctor 通过不等于已执行技能。
