---
name: qa-functional-execution
description: 执行或复测范围已确认的功能用例，按实际工具记录交互、状态、网络及设备观察；区分模拟、浏览器和真机证据，不自动创设产品规则、提交Jira或把重试通过覆盖首次失败。
---
# 功能执行与复测

读取 [专业规则](../qa-test-design/references/professional-rules.md) 及 [执行规则](references/execution.md)。用例设计缺失或失效时返回 `qa-test-design`；不能拿不明确预期边测边判通过。

核对当前需求/用例修订、构建、环境、目标设备与授权。声明实际可用入口/工具；普通人工黑盒不强制代码或自动化依赖。只操作指定设备与隔离测试数据；不能扩大到其他已连接手机或正式账号。

逐条准备前置，执行真实动作，保存实际观察与原始证据。准备误入单独记，正式失败保留；复测创建新run并说明变更，不能修改旧结果。恢复页面直接观察，切应用、切后台、锁屏与杀进程不得相互冒充。

记录mode、profile、版本/设备、时间/时区和结果。fixture/浏览器模拟不能报告原生锁屏、弱网射频或热状态通过。热与性能没有可测指标/判据时只报观察或阻塞，不编阈值。

本地结果按 [数据契约](../qa-test-design/references/data-contract.md) 用 amazing-skills qa record保存；缺工具用blocked/not-run及具体原因。交付真实结果、证据、失败/阻塞和未执行。格式/脚本检查通过不等于被测功能通过。新异常有依据才进入 `jira-defect-delivery`，执行授权不等于建单授权。
