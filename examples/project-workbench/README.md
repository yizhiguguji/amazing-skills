# 项目工作台：真实接入回归样例

模糊需求：“我接了几个小项目，消息和进度散在不同地方，总不知道今天该先推进哪个，帮我弄个工作台。”

这是合成评估：澄清回答与后续需求均为测试输入，不是用户访谈。见 `.product/input.md`。finesse-brief 通过模型实际读取生成 `.workbench/spec.md`；Pro Max 实际运行设计系统和图表查询；Impeccable 实际执行 context/detect，并按上游要求使用独立 finish reviewer 与 documenter。每次工具运行的版本、退出码及原始输出在 `.product/provider-events.jsonl`。包含一次真实失败及恢复，不伪造全程成功。

首轮实现五模块并通过 AC-01–04 后，固化 DESIGN.md、设计 sidecar 和设计锁。后续需求增加仅影响看板的交付日期筛选，复用原控件。验证 AC-01–05，并比较 DESIGN.md、整个 styles.css、tokens 和共享按钮计算样式。`baseline/` 保存修改前代码，两个阶段均可重跑。

```bash
go run ./cmd/amazing-skills serve --example project-workbench --port 8894
node examples/project-workbench/test-browser.cjs
node examples/project-workbench/test-browser.cjs --revision
go run ./cmd/amazing-skills project check-design --project examples/project-workbench
```

以上从仓库根运行。浏览器测试需要 Node 与 Playwright；可以设置 NODE_PATH 指向依赖目录、CHROME_PATH 指向当前电脑的 Chrome，也可使用 Playwright 自带 Chromium。预览只公开应用资源，不公开规范或工具缓存。

测试详情见 [.product/verification.md](.product/verification.md)。实际验证发生在 macOS 和临时迁移安装目录；没有在实体 Windows/Linux 电脑上运行。样例不进入技能安装包，也不需要作为服务部署。
