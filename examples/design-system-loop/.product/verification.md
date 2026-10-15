# 合成演练验收记录

日期：2026-10-15（Asia/Tokyo）。范围：一句抽象的合成需求，参考提取、设计包、组件预览、实际阅读记录页与现有设计锁。环境：本机 macOS/arm64，已安装 Chrome，经 Playwright 无头执行；视口宽度 1440、390、320。没有真机或其他系统的执行声明。

## 行为检查

| 编号 | 结果 | 方法与证据 |
| --- | --- | --- |
| AC-01 | 通过 | 空存储进入，无假记录或计数；添加入口可见，无远程请求。evidence/browser-results.json，evidence/desktop-empty.png。 |
| AC-02 | 通过 | 空标题聚焦并给错误；Tab/Enter 可保存，有效数据刷新后保留。browser-results 的 AC-02。 |
| AC-03 | 通过 | 中文混排、200 字符连续 ASCII 书名与脚本样式文本均按文本处理；三个宽度无整页横溢。browser-results 的 AC-03；desktop-records.png、mobile-records.png。 |
| AC-04 | 通过 | 键盘移除指定记录后计数、列表与刷新一致。browser-results 的 AC-04。 |
| AC-05 | 通过 | 模拟浏览器存储写失败，未报成功且保留输入；恢复后重试成功；损坏历史不被覆盖。browser-results 的 AC-05。属于浏览器故障注入，不是真实设备故障。 |
| AC-06 | 通过 | 应用直接 link tokens.css，计算背景/正文/辅助色/按钮/字体声明/基础间距/操作尺寸与预览一致。evidence/computed-style-results.json；.product/tool-checks.json；.product/design-lock.json。 |
| AC-07 | 通过 | 组件表单错误、有效反馈、原生禁用与键盘焦点可用；两套分发包均在三个宽度渲染检查。browser-results 的 AC-07/PRESET-RENDERS。 |

首次运行 AC-06 失败：本地白名单服务将 tokens.css 返回为 HTML，导致应用样式未加载。已修正为按扩展名发送 CSS MIME，并增加 Go 回归测试。evidence/browser-results-first-run.json 和 evidence/first-run 保留原始失败；当前 browser-results 为修复后结果，未覆盖历史记录。

## 视觉检查

实际查看 evidence/desktop-records.png、mobile-records.png、mobile-form.png、preview-mobile.png、preview-mobile-components.png 以及分发包移动截图；桌面连续列表与侧栏有明确层次，小屏单列，长书名换行，表单无遮挡、焦点可见。标题采用平衡换行，保存反馈的连续长词允许换行，网格子项允许缩小。当前截图使用固定视口捕获；早期全页捕获出现图块重复，已改为禁用 GPU 并在相应滚动位置分别捕获，不用异常截图证明视觉通过。

实测选定文字配色对比：calm-paper/阅读页的正文约 13.14、辅助文字 6.05、按钮 7.82；precise-console 分别约 14.09、5.83、7.27。这里只证明这些已测配色，不代表完整无障碍合规。字体声明与中文可读性已检查，不声称某个未安装字体被实际加载。

截图均为合成数据展示副本。evidence/screenshot-metadata.json 检查 PNG 文本/EXIF 块，无本机用户名、联系人或私有地址；历史失败截图同样检查。可见脚本样式书名是转义测试文本，不执行脚本。

## 结构与设计锁

Go design-system check 验证包结构、证据摘要、token 引用与预览同步；project check-design 验证 DESIGN.md 与实际 tokens 基线；两者是结构检查，不能替代上述行为/视觉。命令结果记录在 .product/tool-checks.json。单点技能安装移动后可在 PATH 不含 Go/Python/Node 的环境执行包初始化、校验、应用、预览和设计锁，汇总见仓库 docs/verification/design-system-loop.json。

## 未验证与范围外

真实用户的阅读习惯和对“安静”的偏好、真实需求访谈、跨设备同步、外部书目、真机、其他浏览器/操作系统、完整屏幕阅读器审计不在本次完成范围。来源未知继续保留在 design-evidence.json；不影响这个明确为合成的演练验收，不转成客户确认。
