# 持续需求的功能QA模拟

输入：“在条目详情写一句备注，保存后列表能看到。”合成需求、设计稿摘录和澄清在sources/。没有真实Figma或客户资料；没有原生App、网络设备、温度监测或外部Jira。

当前模型实际读取本仓库QA指引并形成场景、范围与三版本计划；Go replay只是把这些人工可审查的设计作为固定回归输入，不是LLM生成器或独立模型效果benchmark。

| 阶段 | 用例/规则 | 实际验证 |
| --- | --- | --- |
| r1 | 五条可执行模型用例，长度及生命周期待明确；不测温度专项 | 五条通过；未知规则保留proposed/gaps |
| r2 | 1–20码点及空白规则；成功仍留编辑区；分别确定切应用/后台/锁屏 | 13条，注入上限外仍允许提交的模型缺陷，实际断言产生一次失败 |
| 同轮重试 | 修正模型后再观察 | 首次正式失败仍计失败；不被重试通过覆盖 |
| 模拟Jira | 可审阅缺陷稿→mock创建→媒体不可见→恢复并回读 | 部分回执incomplete；复用同一mock issue，最终一个单/一个附件 |
| r2独立复测 | 修正模型再次执行 | 13条通过，保留历史失败及模拟缺陷解决证据 |
| r3 | 上限20→30，撤回取消流程 | 四条旧边界结果不继承；取消保留retired，当前12条通过 |
| 用例说明精化 | 把早期宽泛步骤/数据改为具体动作和单条断言，不改变业务范围 | 保留原r1–r3快照/结果；当前用例增加修订并新建r4-case-clarity轮，12条通过 |

最终报告仍为synthetic-only，发布结论unverified。对后台、锁屏、切应用只是状态模型验证，不是设备生命周期；断网/延迟只是响应模拟，不是射频弱网；温度专项明确范围外。

查看`.qa/cases.md`、`.qa/coverage.md`、`.qa/plan.json`及`.qa/report.md`；各阶段来源、快照和执行记录保留。`evidence/impact-r2-r3.json`列出变更候选及未变项目，`r3-before-execution.json`证明新修订没有继承旧通过，`r2-first-failure-summary.json`证明重试不抹失败。缺陷处理进度与测试进度分别记录。

## 重跑

从仓库根执行，使用新输出目录，脚本拒绝覆盖已有`.qa`：

```bash
go run ./cmd/amazing-skills replay --output /新临时目录/qa-evaluation
go run ./cmd/amazing-skills qa check --project /新临时目录/qa-evaluation
go test ./...
```

Go replay从fixtures/r1.json、r2.json、r3.json加载已精化的固定三版本用例；示例目录额外保留早期稿及r4精化审计，因此修订号可能与新replay不同。两者业务范围与12条当前场景一致。

先安装QA子包后，直接使用`/目标/.agents/skills/.amazing-skills/amazing-skills replay --output /新目录`；无需Python、网络或设计提供者。旧evidence/portability.json属于迁移前历史，Go迁移验证见docs/go-runtime.md。

这是可复现的交接/数据行为测试。单元测试、格式校验、模型断言和真人功能验证不同，不能据此声称模型必然设计全面或远程Jira集成已验收。

原.qa快照、执行记录及证据是迁移前的历史事实，保持不变；Go另生成新目录记录，不把旧Python运行追认为Go。
