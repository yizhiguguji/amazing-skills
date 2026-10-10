# finesse-brief 适配

来源：https://github.com/mouse-lin/finesse-brief
复核：2026-10-10，SKILL.md 标记 0.5.0。实际安装版本变化时检查输出而非盲信兼容性。

输入：当前用户请求、现有 `.workbench/spec.md`、已知用户事实、授权范围。不再询问已知主对象/角色/任务。
输出：保留原生 YAML+中文正文，唯一存入 `.workbench/spec.md`。将验收编号与假设追加为正文，不要求修改上游 schema。其 subject/entities/channels.pages/hook/cold_start/mvp/excluded/deferred 传给 UI。

交接检查：数字有字段来源及写入者，未接数据有退路，页面有形态和操作，范围明确，“给实现方”解释私有术语。缺任何一项补齐再实施。

上游针对网页留存和角色的强判断按项目实际情况采用；不要把“无推送”当作平台事实，也不要把筛选当成权限控制。

若用户已明确要求完整本地实现，记录规格后继续，不重复请求同一授权；只请求定义则停在定义。未安装时使用 workbench-definition 内置方法，provider=builtin，不声称运行 finesse-brief。

## 捆绑与实际加载

本版本通过 amazing-skills providers doctor 定位并校验真实文件。路径相对于 skills 根，使用者无需设置创建者的 HOME 或 CLAUDE_PLUGIN_ROOT。必须读取实际 SKILL.md 后采用其方法；helper doctor 只做发现和完整性检查。工具通过 helper run 执行，记录真实退出码和输出。来源 commit 以 providers.lock.json 为准；这里只说明接口，不复制上游操作细节。
