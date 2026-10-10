# 当前功能测试用例

此稿是设计产物，不是审批或执行通过。proposed/deferred/retired单列状态，只有ready可承诺执行。

| ID@修订 | 状态 | 目标 | 模块/子功能 | 方法/影响 | 前置/数据 | 步骤 | 预期 | 需求与依据 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| TC-001@2 | ready | 保存自己的合法昵称 | 个人资料/昵称 | scenario/medium | 使用合成测试用户甲和乙及模型内昵称资源；无真实服务；{"owner":"test-user-a","other":"test-user-b"} | 保存自己的合法昵称，观察返回和对应资源 | 仅自己的昵称更新成功 | R1@2: S2/AC-1 |
| TC-002@2 | ready | 拒绝超过上限的昵称且数据不变 | 个人资料/昵称 | boundary-value/medium | 使用合成测试用户甲和乙及模型内昵称资源；无真实服务；{"max":21} | 拒绝超过上限的昵称且数据不变，观察返回和对应资源 | 拒绝且原数据不变 | R1@2: S2/AC-1 |
| TC-003@1 | ready | 跨测试用户读写昵称被拒绝 | 个人资料/昵称 | decision-table/medium | 使用合成测试用户甲和乙及模型内昵称资源；无真实服务；{"owner":"test-user-a","other":"test-user-b"} | 跨测试用户读写昵称被拒绝，观察返回和对应资源 | 读取不返回对方昵称，写入被拒绝且对方数据不变 | R2@1: S1/AC-2 |
| TC-004@1 | retired | 预览昵称为文本 | 个人资料/昵称 | scenario/medium | 使用合成测试用户甲和乙及模型内昵称资源；无真实服务；<em>示例</em> | 预览昵称为文本，观察返回和对应资源 | 标记转为文本，不作为标记解析 | R3@1: S2/AC-3-WITHDRAWN |
| TC-005@1 | proposed | 会话失效时昵称保存的观察目标 | 个人资料/昵称 | state-transition/medium | 使用合成测试用户甲和乙及模型内昵称资源；无真实服务；{"owner":"test-user-a","other":"test-user-b"} |  |  | R2@1: S1/AC-2 |

## 安全用例依据与操作边界

| 用例 | 候选/目标 | 判定来源 | 预期状态 | 操作 |
| --- | --- | --- | --- | --- |
| TC-001@2 | SC1/T1 | [{"id":"R1","revision":2}] | confirmed | observe |
| TC-002@2 | SC1/T1 | [{"id":"R1","revision":2}] | confirmed | observe |
| TC-003@1 | SC2/T1 | [{"id":"R2","revision":1}] | confirmed | observe |
| TC-004@1 | SC3/T1 | [{"id":"R3","revision":1}] | confirmed | observe |
| TC-005@1 | SC4/T1 | [] | unknown | observe |
