# 范围、维度与追踪

| 验收点@修订 | 状态 | 用例 |
| --- | --- | --- |
| R-SAVE@2 | active | TC-001@4:ready |
| R-EMPTY@2 | active | TC-002@4:ready；TC-013@4:ready |
| R-FAIL@1 | active | TC-003@4:ready；TC-012@4:ready |
| R-CANCEL@1 | withdrawn | TC-004@1:retired |
| R-LIMIT@3 | active | TC-005@4:ready；TC-006@4:ready；TC-007@4:ready；TC-008@4:ready |
| R-RESUME@2 | active | TC-009@4:ready；TC-010@4:ready；TC-011@4:ready |

| 维度 | 决策 | 理由 | 场景/缺口 |
| --- | --- | --- | --- |
| interaction | include | 直接验证备注操作的输入、保存或恢复；不扩展相邻业务 | TC-001；TC-002 |
| happy | include | 直接验证备注操作的输入、保存或恢复；不扩展相邻业务 | TC-001 |
| negative | include | 直接验证备注操作的输入、保存或恢复；不扩展相邻业务 | TC-002；TC-003；TC-013 |
| boundary | include | 直接验证备注操作的输入、保存或恢复；不扩展相邻业务 | TC-005；TC-006；TC-007；TC-008 |
| state | include | 直接验证备注操作的输入、保存或恢复；不扩展相邻业务 | TC-001；TC-003；TC-012；TC-009；TC-010；TC-011 |
| app-switch | include | 直接验证备注操作的输入、保存或恢复；不扩展相邻业务 | TC-009 |
| background | include | 直接验证备注操作的输入、保存或恢复；不扩展相邻业务 | TC-010 |
| lock | include | 直接验证备注操作的输入、保存或恢复；不扩展相邻业务 | TC-011 |
| offline | include | 直接验证备注操作的输入、保存或恢复；不扩展相邻业务 | TC-003 |
| weak-network | include | 直接验证备注操作的输入、保存或恢复；不扩展相邻业务 | TC-012 |
| thermal | not-applicable | 短时文本保存无温度风险线索，任务明确排除温度专项；不编造阈值 |  |

## 待明确

