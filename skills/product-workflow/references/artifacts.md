# 项目文件与完成契约

- `.workbench/spec.md`：唯一产品定义，含用户/角色、主对象、任务、页面与数据、字段写入者、断开/空数据行为、MVP、排除项、假设和验收编号。
- `DESIGN.md`：目标模式、字体与中英文回退、颜色角色、间距、密度、导航、状态、动效、响应式、组件与页面例外。实现后记录代码 token 来源。
- `.product/plan.md`：实施任务和规格验收的对应关系。
- `.product/decisions.md`：产品/设计/技术决策，区分用户事实、工作假设、实现决策；记录后续替代。
- `.product/verification.md`：每个验收编号的结果、方法、证据、环境、实际日期，视觉检查与行为检查分开，未运行写未验证。
- `.product/state.json`：schema_version=1；phase 为 definition/design/implement/verify/complete；status 为 draft/active/needs-input/blocked/passed；selected 为能力选择列表（id、provider、reason）；open_questions 为字符串列表；evidence 为项目内文件路径；acceptance 为验收列表（id、status=passed/failed/unverified、evidence 文件列表）；updated_at 为实际 ISO 日期时间。

complete 仅在所有范围内关键验收通过、没有未解决必要问题，且有可读取 verification 和证据时记录。文档完整性检查不替代实际验收。若证据缺失或关键失败，phase 保持 verify，不能完成。

文件均为项目相对路径。项目状态中的路径不得读取到目标项目之外。首次模板初始化不覆盖文件；规格正文中的空栏位必须在实现前填写。不要存凭证或私人样本到可分享示例。

`.product/design-lock.json` 记录规范与 tokens 的 SHA256；`provider-events.jsonl` 区分模型指令使用和真实工具调用，`provider-evidence/` 保留工具输出。`.product/tool-cache/` 仅为本机缓存，不提交，不随示例发布。新机器从锁定提供者与相应平台引擎重建，不复制本机可执行文件。
