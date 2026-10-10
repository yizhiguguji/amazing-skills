# 验证记录

输入和澄清均为合成评估设置，见 input.md。实际浏览器与工具执行，不把模拟回答冒充真实访谈。

## 实际提供者

- finesse-brief：读取原始 SKILL.md、system-domain/handoff/day-two，产出 `.workbench/spec.md`；没有 CLI，不报告虚假程序退出码。
- UI UX Pro Max：两次 design-system 查询、一次 chart 查询实际退出 0；营销页面结构不适用，记录拒绝理由；采用 teal 色族和数量对比方法。
- Impeccable：首轮 context 退出 127；改用 helper 下载 engine 0.1.14，核对锁定官方 SHA-256 后 context 退出 0。detect 退出 0，返回空违规列表。按上游 new-work 要求，独立 reviewer 读取原始 TOML 并审查截图，documenter 写 DESIGN.md 与 sidecar。当前工具环境未自动加载 TOML，采用代理自己读取的适配方式，已记录限制。
- 原始执行证据见 provider-events.jsonl 和 provider-evidence/；失败日志保留。

## 首轮与后续

| 验收 | 内容 | 结果 | 证据 |
| --- | --- | --- | --- |
| AC-01 | 无预置数据、真实空状态 | 通过 | evidence/baseline-results.json 与 revision-results.json |
| AC-02 | 录入、阶段更新、历史、本地持久化、已完成不算超期 | 通过 | 两轮 results |
| AC-03 | 五模块、搜索、进展、阶段图、称呼保存 | 通过 | 两轮 results |
| AC-04 | 键盘/对话框、必填、390/1440宽度、长中文、减少动效、损坏/写入失败保护 | 通过 | 两轮 results 与截图 |
| AC-05 | 交付筛选仅影响看板，项目清单与全局统计不变；计算样式相同 | 通过 | evidence/revision-results.json |

首轮四项通过，后续五项通过，两轮 pageerrors 为空。DESIGN.md 和整个 styles.css 的 SHA-256 前后相同，锁检查通过；见 evidence/style-regression.json 和 baseline-design-lock.json。哈希锁能发现规范/tokens 漂移，不能代替实际渲染审查。后续新增 select 使用已有字段样式，没有重生成全局方案。

测试首次比较到了不同 hover 状态，统一鼠标/焦点后重跑。后续截图曾出现浏览器 resize 重绘重复条带；独立 reviewer 要求重新采集，测试加入两帧绘制等待并重跑。未通过修改 CSS 掩盖证据问题。

独立视觉记录：finish-review.md（首轮），followup-review.md（后续）。DESIGN sidecar 预览块未在上游 live editor 逐一验证；原生控件跨系统外观未测。

## 可移植性与边界

当前电脑 macOS、Python 3.9、Node+Playwright 驱动已有 Chrome，时区 Asia/Tokyo。新临时路径实际安装八技能、校验原始上游哈希、执行离线 Pro Max 查询均通过，见 evidence/portability.json。仓库目录/引用校验与12项单元测试通过。

没有在实体 Windows/Linux 或另一台电脑执行，不宣称所有平台已验收。Impeccable 首次准备需要网络或预装引擎，浏览器验收需要执行环境自行提供浏览器。Skill 是否被模型正确遵循仍需要项目证据审查；方案锁不保证任意模型永远不会乱风格。
