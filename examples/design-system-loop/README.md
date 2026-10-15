# 抽象需求到实际界面的闭环演练

输入只有一句合成需求：“想做个读书记录页，安静一点，每次打开知道接着读哪本，样式还没想好。” 没有参考图、品牌或服务。此例是仓库维护中的合成演练，不是真实客户访谈或用户偏好验证。

## 流程与产物

1. [原始需求](sources/request.md) → [分级依据](design-evidence.json)：把用户提供的合成目标与暖色/排版/存储等推断分开，未知继续保留。
2. [设计包](design-system.json) → [DESIGN.md](DESIGN.md) 与 [tokens.css](tokens.css)：设计意图和精确值分别管理，基于自建 calm-paper 起点，没有外部品牌资产。
3. Go preview → [组件样例](design-preview.html)：颜色、混排、长文本、表单错误/焦点/禁用、加载与空状态；生成摘要检查过期。
4. Go apply → 本项目根，再 project init 补缺失文件；[选择回执](.product/design-system-selection.json) 不改变验收状态。[页面](index.html) 直接导入同一 tokens，未另建主题色。
5. [规格](.workbench/spec.md) → [实现](app.js) → [验收记录](.product/verification.md)，沿用现有设计锁及状态文件。

## 查看与复验

仓库根运行：

```sh
go run ./cmd/amazing-skills serve --example design-system-loop --port 8765
go run ./cmd/amazing-skills design-system check --package examples/design-system-loop
go run ./cmd/amazing-skills project check-design --project examples/design-system-loop
go run ./cmd/amazing-skills project check --project examples/design-system-loop
```

打开本地 http://127.0.0.1:8765 ，链接可进入组件预览。预览服务仅暴露 HTML/CSS/JS 白名单，不提供来源、状态或仓库文件。

浏览器回归脚本 `test-browser.cjs` 为可选维护检查，不属于安装技能或Go运行时；使用本机已有 Playwright/Chrome，不自动安装。先编译当前平台程序，再设置 AMAZING_SKILLS_BIN（绝对路径）、NODE_PATH、CHROME_PATH 后用 Node 运行该脚本。

## 检查结果与边界

- 行为：真实无头 Chrome 执行 8 个场景；覆盖持久化、键盘、输入错误、存储失败、损坏数据、连续 ASCII/中文长标题、同源 tokens 计算样式、组件状态和两个分发包。结果在 [浏览器记录](evidence/browser-results.json)。
- 首次失败：服务将 tokens.css 错发为 HTML 类型，导致应用未加载主题；计算样式比较捕获失败。原始记录 [保留](evidence/browser-results-first-run.json)，原始截图在 evidence/first-run。修复服务 MIME 后回归通过，未用重试覆盖历史失败。
- 视觉：实际查看桌面、移动端应用与组件截图，检查主次、中文字体回退、长文本、表单和焦点。截图均为合成数据的展示副本；PNG 文本/EXIF 元数据检查见 [记录](evidence/screenshot-metadata.json)。当前截图为固定视口捕获，不宣称一张图覆盖所有滚动区域。
- 未验证：真实用户对“安静”的偏好、真实需求访谈、真机、其他浏览器/操作系统及跨设备同步。组件截图不能证明完整无障碍合规，包检查不能证明业务通过。

设计包来源和设计锁都属于示例事实，不能把本例暖色、字号、字段或存储约束应用到其他技能。
