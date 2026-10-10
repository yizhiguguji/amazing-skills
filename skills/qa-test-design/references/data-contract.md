# 本地数据契约（schema_version 1）

Go独立运行工具，全部UTF-8，项目路径动态解析。没有现成系统时使用`.qa/`；既有系统保持其来源与历史，不擅自迁移。

`config.json`：项目策略与工具信息，模板不放私有地址、账户或凭证。实际私有配置留在使用者项目，分发共享技能前需另行脱敏。

`plan.json`：

```json
{
  "schema_version": 1,
  "sources": [{"id":"S1","version":"1","kind":"document","status":"accepted","locator":"sources/r1.md","sha256":"文件摘要"}],
  "requirements": [{"id":"R1","revision":1,"status":"active","text":"已确认验收点","source_refs":[{"source_id":"S1","locator":"AC-1"}]}],
  "cases": [{"id":"TC-001","revision":1,"status":"ready","objective":"目标","module":"模块","subfunction":"单一子功能","category":"positive","impact":"medium","requirement_refs":[{"id":"R1","revision":1}],"dimensions":["happy"],"method":"scenario","preconditions":["状态"],"data":"准备的数据","steps":["操作"],"expected":["可判定预期"]}],
  "dimensions": [{"id":"happy","decision":"include","reason":"与R1直接相关","requirement_ids":["R1"],"case_ids":["TC-001"]}],
  "gaps": []
}
```

sources可以用相对文件或外部URL；本地必须提供真实SHA256并存在，脚本不访问外部URL，模型须另核读入事实。一个已接受来源的旧版本可仍支撑未变化验收点，但被替代要求必须更新revision及引用。sources状态accepted/unverified/withdrawn；requirements状态active/uncertain/superseded/withdrawn；cases状态ready/proposed/deferred/retired。

完整设计记录11个维度，decision为include/not-applicable/clarify/deferred，原因不可空。include必须引用active需求和非retired用例；clarify/deferred必须关联gaps（ID及关闭条件）。proposed不是可执行承诺，ready不得引用uncertain需求或unverified来源。

`executions.jsonl`是追加记录：execution_id、run_id、case_id、case_revision、status（passed/failed/blocked/deferred/not-run）、mode（synthetic-fixture/mock/browser-emulation/real-device/manual）、environment、actual、evidence（项目内文件列表）、started_at/finished_at。passed/failed必须有实际观察和证据。记录自动带当时用例快照；相同execution_id重复同内容不再写，不同内容拒绝。所有记录有前序摘要链，check检测篡改/截断（项目最新checkpoint亦需保留）。无法抵御有权同时篡改记录及checkpoint的攻击，不替代系统权限或审计存储。

`snapshots/<label>/plan.json`与config、manifest是冻结计划，不是审批。脚本保存文件摘要，check检测历史快照被改。相同case ID+revision语义不可变，修订必须递增；源文件变化未同步plan摘要会失败。label仅字母数字连字符，拒绝路径逃逸/覆盖。

```bash
amazing-skills qa record --project /项目 --input /项目/evidence/result.json
amazing-skills qa summary --project /项目 --run r2
amazing-skills qa summary --project /项目
```

--run用当前计划中的用例修订对该轮统计。跨轮次每个当前修订取最近有正式结果的轮次；not-run不覆盖旧执行。同轮正式failed不被重试passed覆盖，独立新run复测可呈现最新passed，同时列历史失败。retired/proposed不进当前承诺分母；deferred用例单列延期。状态之和=总数；通过率以passed/(passed+failed)，零分母null。人工/实机、浏览器和模拟分别统计，不能混成实机结论。

脚本只验证结构/引用/版本和复算事实，不确认业务正确、全面覆盖、审批或外部保存。summary自动写report.json/report.md并注明当前来源与范围限制；报告前需模型审查风险、缺口及证据内容。
