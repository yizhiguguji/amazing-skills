# QA result summary

Status: synthetic-only
Scope: r2
Release decision: unverified; never inferred from counts alone

| Status | Count |
| --- | --- |
| passed | 1 |
| failed | 0 |
| blocked | 0 |
| deferred | 0 |
| not-run | 2 |

Total: 3
Pass-rate denominator: passed + failed; rate: 1
Evidence modes: {"synthetic-fixture":1}
Historical failures: r1-TC-003-failed
Defects (separate from failed cases): {"confirmed_total":null,"current_open":null,"source":"not-provided"}

## Current cases

- TC-001@2: not-run (<nil>)
- TC-002@2: not-run (<nil>)
- TC-003@1: passed (synthetic-fixture)

## Gaps

- G1: 昵称保存时会话失效如何处理？时限与提示未定义。

## Security scope and candidate decisions

Only the recorded targets and executed evidence are covered; this is not a system security or compliance conclusion.

```json
{"checks":[{"case_ids":["TC-001","TC-002"],"decision":"include","family":"input-integrity","id":"SC1","reason":"本期昵称有明确码点长度与保存规则","requirement_ids":["R1"],"target_ids":["T1"]},{"case_ids":["TC-003"],"decision":"include","family":"authorization","id":"SC2","reason":"同功能确认基线只允许读写自己的昵称","requirement_ids":["R2"],"target_ids":["T1"]},{"case_ids":[],"decision":"not-applicable","family":"output-context","id":"SC3","reason":"预览入口已撤回，不继续覆盖已撤回行为","requirement_ids":["R1"],"target_ids":[]},{"case_ids":["TC-005"],"decision":"clarify","family":"session","gap_ids":["G1"],"id":"SC4","reason":"保存的会话失效预期与时限未知，不造固定规则","requirement_ids":["R2"],"target_ids":["T1"]}],"exclusions":["头像/支付/上传/密码重置","其他接口和全站扫描","真实服务与真实设备"],"execution":{"allowed_actions":["observe","request-modification"],"budget":"At most 10 isolated model calls; no network","source_refs":[{"locator":"AUTH-1","source_id":"S1"}],"status":"authorized","stop_conditions":["Target escape, unexpected data, or model failure"],"target_ids":["T1"]},"mode":"specialist","scope_requirement_ids":["R1","R2"],"targets":[{"id":"T1","locator":"fixture:profile/nickname","platform":"web-model","requirement_ids":["R1","R2"]}]}
```
