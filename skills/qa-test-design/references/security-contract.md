# 可选安全数据扩展（schema_version 1保持兼容）

已有管理系统沿用其格式；以下只用于没有系统时的便携计划。普通旧计划无需新增字段才能读取。新功能设计记录安全候选筛选；只涉及无安全影响的文字变更时可写not-applicable及理由，不生成专项用例。

`plan.security`：

```json
{
  "mode": "functional",
  "scope_requirement_ids": ["R1"],
  "targets": [{"id":"T1","platform":"web","locator":"fixture:profile/nickname","requirement_ids":["R1"]}],
  "checks": [{"id":"SC1","family":"authorization","decision":"include","reason":"本期涉及个人资料修改的归属规则","requirement_ids":["R1"],"target_ids":["T1"],"case_ids":["TC-001"]}],
  "exclusions": ["其他功能、全站扫描不在本期范围"]
}
```

mode为functional/specialist；scope是本期active需求或确认基线，不能把标准目录、风险经验单独登记为active需求。targets是实际范围内的入口，不是扫描目标清单；每个入口关联scope中的需求。checks只记相关候选，family自由命名；decision使用include/not-applicable/clarify/deferred，reason必填。include须关联目标和非retired用例；clarify/deferred须有gap_ids。未选择的family不强制全量登记。

每条安全用例额外带：

```json
{"security":{"check_id":"SC1","target_id":"T1","oracle":"confirmed","oracle_refs":[{"id":"R1","revision":1}],"action":"observe"}}
```

action为observe（普通功能操作/观察）、request-modification、scan、load、device-change。后三类和修改请求记录即使混在functional计划中，也须有已确认的目标/动作授权；不通过改mode绕过。专项ready用例须全部有关联安全元数据。

oracle为confirmed/unknown。confirmed引用该用例已引用的active当前需求修订，来源在需求source_refs中按具体位置追踪；ready必须confirmed。unknown只可proposed/deferred，expected留空，问题/观察目标放gap或探索章程，不允许passed/failed记录。工具只能验证引用结构，不能判断文字是否真正支撑每条断言，仍需模型审核。

专项有实际执行意图时补`security.execution`：status为authorized/pending，target_ids、allowed_actions、source_refs（source_id/locator）、stop_conditions、budget。纯设计可完全不设execution，ready仅代表设计可执行，不等于授权执行。实际passed/failed记录要求authorized、目标和用例action在允许集合内、授权来源accepted；authorized只记录已确认授权，不是自动生成权限。当前工具没有扫描器或攻击执行器。

无预期探索用独立观察笔记，不强行塞入正式passed/failed事件。安全相关用例继续使用原11维度、稳定ID、递增修订、冻结与追加执行链。新安全字段属于用例语义，修改必须增加revision；撤回用例retain retired，撤回的目标/决策不再声称已覆盖。

`amazing-skills qa check`核对范围、目标、候选、用例、缺口及oracle引用；不能证明真实授权、业务覆盖、安全结论或合规。它不联网、执行载荷或安装工具。

目标ID保持不变但入口、平台或目标上下文改变时，也要修订相关用例。工具比较冻结计划与新执行记录保存的target上下文，阻止旧修订继续使用变更的入口。仅撤回不相关需求关联不强制重写未变场景。旧记录没有该上下文时保留兼容，无法追溯的部分由模型明确说明。
