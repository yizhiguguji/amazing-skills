---
name: jira-defect-delivery
description: 将确认异常整理成可复现的Jira缺陷，按授权查重、创建、更新、复测备注或状态流转并回读；没有连接器时交付草稿，执行测试或写报告不自动授权外部写入。
---
# Jira缺陷交付

运行工具为Go独立程序；下文 `amazing-skills` 指 skills 根下 `.amazing-skills/amazing-skills`（Windows加.exe）。优先使用实际安装位置的绝对路径；仓库开发可在根目录用 `go run ./cmd/amazing-skills`。仅阅读Skill不需要任何语言运行时。

读取 [专业规则](../qa-test-design/references/professional-rules.md) 与 [Jira交接](references/jira.md)。独立确认问题不是需求疑问、建议或覆盖缺口；核对本轮实际观察、预期来源、版本/环境、步骤与证据。查重按触发、状态、现象、观察位置和根因线索，不能仅同标题。

先完成可审阅草稿。仅按当前授权及实际连接器执行外部写入；无连接器明确未提交。当前项目决定issue type、字段、severity/priority、真实负责人及可用transition，不能猜账号或默认状态。

正文按环境、前置/步骤、实际、预期、依据组织，有实际证据再增加证据区。异常与补充观察分开，多项预期清楚对应操作；无附件不生成证据占位。字体、颜色、编号细节及执行方式披露服从项目要求，不强制红色或默认隐去真实/模拟来源。具体规则及兼容输入见 [Jira交接](references/jira.md)。只保留帮助复现/判断内容；内部主键、绝对路径、令牌不进正文。附件上传、平台支持的可见媒体、正文回读、系统关联逐项核对，部分成功不能报全部完成。

结果不明先查已存在问题、评论与附件，再恢复；复用issue identity、附件及幂等标识。只更新授权字段，保留用户的新内容与已有有效标签。复测备注来自新run，再按授权流转并回读状态，不把已合入当测试通过。

便携助手只检查草稿与回执，不包含私有Jira服务或自动写入：

```bash
/安装目录/.amazing-skills/amazing-skills jira draft --project /项目 --input /项目/defect.json
/安装目录/.amazing-skills/amazing-skills jira check-receipt --project /项目 --input /项目/receipt.json
```

交付具体真实issue/comment链接、实际修改、媒体及状态；mock回执必须标mock且不能宣称远程Jira提交。字段结构校验不等于网络回读成功。
