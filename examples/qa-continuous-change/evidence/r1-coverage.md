# 范围、维度与追踪

| 验收点@修订 | 状态 | 用例 |
| --- | --- | --- |
| R-SAVE@1 | active | TC-001@1:ready |
| R-EMPTY@1 | active | TC-002@1:ready |
| R-FAIL@1 | active | TC-003@1:ready；TC-012@1:ready |
| R-CANCEL@1 | active | TC-004@1:ready |
| R-LIMIT@1 | uncertain | TC-005@1:proposed；TC-006@1:proposed；TC-007@1:proposed；TC-008@1:proposed |
| R-RESUME@1 | uncertain | TC-009@1:proposed；TC-010@1:proposed；TC-011@1:proposed |

| 维度 | 决策 | 理由 | 场景/缺口 |
| --- | --- | --- | --- |
| interaction | include | 直接验证备注操作的输入、保存或恢复；不扩展相邻业务 | TC-001；TC-002；TC-004 |
| happy | include | 直接验证备注操作的输入、保存或恢复；不扩展相邻业务 | TC-001 |
| negative | include | 直接验证备注操作的输入、保存或恢复；不扩展相邻业务 | TC-002；TC-003；TC-004 |
| boundary | clarify | 需求未定义对应状态或边界，不固定预期 | G-LIMIT |
| state | include | 直接验证备注操作的输入、保存或恢复；不扩展相邻业务 | TC-001；TC-003；TC-004；TC-012 |
| app-switch | clarify | 需求未定义对应状态或边界，不固定预期 | G-RESUME |
| background | clarify | 需求未定义对应状态或边界，不固定预期 | G-RESUME |
| lock | clarify | 需求未定义对应状态或边界，不固定预期 | G-RESUME |
| offline | include | 直接验证备注操作的输入、保存或恢复；不扩展相邻业务 | TC-003 |
| weak-network | include | 直接验证备注操作的输入、保存或恢复；不扩展相邻业务 | TC-012 |
| thermal | not-applicable | 短时文本保存无温度风险线索，任务明确排除温度专项；不编造阈值 |  |

## 待明确

- G-LIMIT: 长度上限、单位及空白规则？ 关闭条件：有明确范围、计数与空白说明
- G-RESUME: 同进程切应用、后台、锁屏后的输入规则？ 关闭条件：三个事件的规则及进程范围明确
