# Impeccable 交接

来源：https://github.com/pbakaus/impeccable
复核：2026-10-10，当前入口标记 4.5.2。命令、文件和运行器以实际安装版本为准。

先输入 `.workbench/spec.md`、现有 DESIGN.md、代码、授权范围。工作台选择 Operate；工具的营销页另选 Persuade。执行其上下文加载与必要 playbook，不重新发现已知产品事实。

PRODUCT.md 是规格的上下文投影：注明来源和规格修订，包含用户、任务、约束、假设，不复制一套独立业务决定。缺失重要品牌事实可询问；已有设计事实应保留。

DESIGN.md 为视觉唯一依据。原生设计 sidecar/surface brief 如生成，应有来源。规范存在时仅做必要局部修正；要求重设计时明确记录替代的决策与范围。

构建后按实际安装技能做截图与有限批次修正；记录查看的路由、尺寸、状态与证据。audit 的检测通过不等于业务通过；随后执行规格验收。运行器失败按上游允许的回退读取项目上下文，不能伪造检测结果。

业务定义优先于默认视觉偏好；不新造统计数字、功能或权限。完整本地实施已授权时不添加重复批准步骤。外部发布仍根据实际授权判断。

## 捆绑与实际加载

本版本通过 amazing-skills providers doctor 定位并校验真实文件。路径相对于 skills 根，使用者无需设置创建者的 HOME 或 CLAUDE_PLUGIN_ROOT。必须读取实际 SKILL.md 后采用其方法；helper doctor 只做发现和完整性检查。工具通过 helper run 执行，记录真实退出码和输出。来源 commit 以 providers.lock.json 为准；这里只说明接口，不复制上游操作细节。
