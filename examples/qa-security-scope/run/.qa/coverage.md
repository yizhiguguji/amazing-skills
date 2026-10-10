# 范围、维度与追踪

| 验收点@修订 | 状态 | 用例 |
| --- | --- | --- |
| R1@2 | active | TC-001@2:ready；TC-002@2:ready |
| R2@1 | active | TC-003@1:ready；TC-005@1:proposed |
| R3@2 | withdrawn | TC-004@1:retired |

| 维度 | 决策 | 理由 | 场景/缺口 |
| --- | --- | --- | --- |
| interaction | not-applicable | 本模拟需求无此平台事件/行为，未扩大为专项 |  |
| happy | include | 本期确认规则的直接场景 | TC-001 |
| negative | include | 本期确认规则的直接场景 | TC-003 |
| boundary | include | 本期确认规则的直接场景 | TC-002 |
| state | clarify | 会话失效行为未定义 | G1 |
| app-switch | not-applicable | 本模拟需求无此平台事件/行为，未扩大为专项 |  |
| background | not-applicable | 本模拟需求无此平台事件/行为，未扩大为专项 |  |
| lock | not-applicable | 本模拟需求无此平台事件/行为，未扩大为专项 |  |
| offline | not-applicable | 本模拟需求无此平台事件/行为，未扩大为专项 |  |
| weak-network | not-applicable | 本模拟需求无此平台事件/行为，未扩大为专项 |  |
| thermal | not-applicable | 本模拟需求无此平台事件/行为，未扩大为专项 |  |

## 待明确

- G1: 昵称保存时会话失效如何处理？时限与提示未定义。 关闭条件：项目确认该行为基线及其版本；再修订用例为ready。

## 安全候选筛选与范围

```json
{"checks":[{"case_ids":["TC-001","TC-002"],"decision":"include","family":"input-integrity","id":"SC1","reason":"本期昵称有明确码点长度与保存规则","requirement_ids":["R1"],"target_ids":["T1"]},{"case_ids":["TC-003"],"decision":"include","family":"authorization","id":"SC2","reason":"同功能确认基线只允许读写自己的昵称","requirement_ids":["R2"],"target_ids":["T1"]},{"case_ids":[],"decision":"not-applicable","family":"output-context","id":"SC3","reason":"预览入口已撤回，不继续覆盖已撤回行为","requirement_ids":["R1"],"target_ids":[]},{"case_ids":["TC-005"],"decision":"clarify","family":"session","gap_ids":["G1"],"id":"SC4","reason":"保存的会话失效预期与时限未知，不造固定规则","requirement_ids":["R2"],"target_ids":["T1"]}],"exclusions":["头像/支付/上传/密码重置","其他接口和全站扫描","真实服务与真实设备"],"execution":{"allowed_actions":["observe","request-modification"],"budget":"At most 10 isolated model calls; no network","source_refs":[{"locator":"AUTH-1","source_id":"S1"}],"status":"authorized","stop_conditions":["Target escape, unexpected data, or model failure"],"target_ids":["T1"]},"mode":"specialist","scope_requirement_ids":["R1","R2"],"targets":[{"id":"T1","locator":"fixture:profile/nickname","platform":"web-model","requirement_ids":["R1","R2"]}]}
```
