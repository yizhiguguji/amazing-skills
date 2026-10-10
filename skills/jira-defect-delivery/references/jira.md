# Jira交接契约

输入draft JSON：id、summary、environment、steps数组、actual、expected、basis（可读出处）、evidence项目内相对文件数组。severity/priority/assignee只有明确映射与账户依据才填；纯文本缺陷允许无媒体但如实说明。问题ID使用本地稳定标识，不伪造外部单号。

外部操作选当前AI已授权的连接器/CLI/API。配置来自用户项目，凭证不进入Skill、正文、日志或导出。没有连接器仍可查本地重复和交付稿，不能借机自动安装陌生插件。JiraCloud/Server等API与富文本格式由adapter决定，不能统一假设wiki或ADF。

回执JSON：mode（mock/live）、operation_id、operation（create/update/comment/transition）、issue_url、write_outcome（saved/unknown/failed）、readback_confirmed、expected_payload_hash、readback_payload_hash、evidence_entries（含local、uploaded、visible、remote_id）、relationship_saved（可选，仅项目需要时）、final_state（可选，流转必须有）。hash应取规范化payload；是否真正回读必须以实际连接器返回为证，布尔字段不证明网络。

只有saved、回读已确认、hash相等、需要的媒体上传且可见、必要关联保存后，当前动作才完整。unknown/failed时返回待恢复/失败，不盲目重建；真实已创建单可存在于部分成功状态。通过检查返回的是动作回执状态，不把mock变live，不替代人工复现。

复测comment、字段update、transition分别生成回执；状态变更不能代替备注。目标状态缺规则就保留不变；负责人来自当前用户或真实责任配置，不复制相邻issue的人。
