# Jira交接契约

输入draft JSON：id、summary、environment、steps数组、actual、expected、basis（可读出处）、evidence项目内相对文件数组。expected 兼容原非空字符串，也可用非空字符串数组表达多项预期；字符串保留已有换行/格式，数组多项生成有序列表。可选 observations 为非空字符串数组，表示不与主异常混淆的补充观察。evidence 必须显式提供，可为空数组；空数组时正文省略整个证据区，缺口保留在命令回执的 evidence_count/evidence_status 中。severity/priority/assignee只有明确映射与账户依据才填。问题ID使用本地稳定标识，不伪造外部单号。

## 正文表达与排版

- 环境、复现步骤、实际与预期应让执行人能定位对象、操作、观察时机及比较基准；不写未确认推测。异常先陈述，补充观察另列；多项独立预期分行。具体编号、加粗、颜色与模板服从用户及项目规则，颜色不能成为唯一标识，不强制强调色或统一格式。
- 保存必要的真实执行环境、来源及限制，模拟不伪装为真实执行。是否对外注明执行者或工具由用户、项目与实际判断需要决定；不自动添加整理过程，也不一律删除执行方式。来源/方式等追溯记录仍须保留。
- 只有实际证据才增加证据区；无附件不在正文生成空标题或“暂无证据”等占位。若缺证据影响异常是否成立，应保持待确认，不能因正文省略而隐去缺口。现有媒体保留顺序、比例与可访问性；日志或抓包按实际项目规则处理。
- Go 助手生成 Markdown 本地稿，不负责目标平台的富文本编码。连接器使用其支持的列表/富文本格式，不能把未解析的 Markdown、HTML 或 ADF 源码当成可见正文。平台不支持某样式时用清楚的分段或文字交付，不因样式差异阻塞内容。

提交后回读实际内容、排版及媒体，不只看接口成功。只改授权字段，保留原内容与有效媒体。新格式不会改写已存在草稿：若旧正文与新版不同但完全匹配相同输入的旧版生成结果，重跑保留旧正文及其 payload_hash，返回 legacy-preserved；如要升级正文另建明确修订，不能自动覆盖或改变既有回执摘要。

## 写入与回执

外部操作选当前AI已授权的连接器/CLI/API。配置来自用户项目，凭证不进入Skill、正文、日志或导出。没有连接器仍可查本地重复和交付稿，不能借机自动安装陌生插件。JiraCloud/Server等API与富文本格式由adapter决定，不能统一假设wiki或ADF。

回执JSON：mode（mock/live）、operation_id、operation（create/update/comment/transition）、issue_url、write_outcome（saved/unknown/failed）、readback_confirmed、expected_payload_hash、readback_payload_hash、evidence_entries（含local、uploaded、visible、remote_id）、relationship_saved（可选，仅项目需要时）、final_state（可选，流转必须有）。hash应取规范化payload；是否真正回读必须以实际连接器返回为证，布尔字段不证明网络。

只有saved、回读已确认、hash相等、需要的媒体上传且可见、必要关联保存后，当前动作才完整。unknown/failed时返回待恢复/失败，不盲目重建；真实已创建单可存在于部分成功状态。通过检查返回的是动作回执状态，不把mock变live，不替代人工复现。

复测comment、字段update、transition分别生成回执；状态变更不能代替备注。目标状态缺规则就保留不变；负责人来自当前用户或真实责任配置，不复制相邻issue的人。
