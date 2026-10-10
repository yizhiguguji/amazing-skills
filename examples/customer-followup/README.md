# 客户跟进工作台：完整流程示例

原始请求：“我们现在用 Excel 管客户，经常忘记跟进，帮我做个工作台。”

这是本仓库内置流程的本地演练。未安装和执行三个外部 Skill；交接适配规则已编写，但其真实工具调用不属于本次实测。所有示例客户由浏览器测试临时创建，首次打开无预置客户。

## 演练记录

1. definition：围绕客户和下一次跟进日期定义任务；业务假设为单人本地原型，不声称支持团队权限或同步。
2. design：选择任务优先、克制的工作台规范；固定 DESIGN.md 和 styles.css 中 tokens。
3. implement：实现冷启动、新建客户、搜索、跟进记录、下一日期和本地持久化。
4. verify：运行 test-browser.cjs，在桌面与手机尺寸验证真实派生统计、持久化、错误和键盘路径，查看截图并记录证据。

## 启动

在仓库根：`go run ./cmd/amazing-skills serve --example customer-followup --port 8765`。

打开 http://127.0.0.1:8765 。本地存储随浏览器 origin 隔离；换端口或浏览器不会共享。清除站点数据会清除客户。

浏览器测试：设置 NODE_PATH 指向包含 playwright 的 node_modules，执行 `node examples/customer-followup/test-browser.cjs`。脚本使用隔离浏览器上下文和临时服务，服务端口固定为 8877，写入 evidence。不会触碰人工运行示例的 8765 origin。需要 Playwright Chromium 可用，或通过 CHROME_PATH 指定已有 Chrome。

## 文件

`.workbench/spec.md` 是产品定义；DESIGN.md 是设计依据；`.product/` 包含计划、决策、状态和实际验收。evidence 是本次机器检查与视觉记录。

限制：单人、同浏览器本地原型；没有后台、通知、登录、协作、Excel 导入或云同步。演练不证明其他模型也能稳定得到相同结果。
