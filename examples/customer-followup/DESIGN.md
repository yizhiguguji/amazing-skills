# 客户跟进设计规范

状态：已选定，规范修订 1
模式：Operate。目标是看到下一位该联系的人并记录进展。

## 字体与颜色
字体：系统 sans-serif，中文回退 PingFang SC、Microsoft YaHei；界面字号最低 14px，表单 16px。标题 28–38px，数字使用 tabular-nums。
背景 --canvas #f4f6f2，表面 --surface #ffffff，文字 --ink #182b25，次要文字 --muted #52665d，主色 --accent #24634d，警示 --danger #9d3b25，边界 --line #d7e0d7。含义由文字与颜色共同表达。

## 布局与组件
spacing tokens 为 4/8/12/16/24/32px；内容最大宽度 1100px。桌面侧栏宽 220px；小屏 <=760px 变为顶部两项导航。列表是一条条任务，按钮保持最小 44px 操作尺寸。边界半径 10/16px，轻边框，不层层嵌卡片。

主按钮统一深绿，次按钮浅表面。当前导航 aria-pressed；键盘焦点 3px outline。native dialog 包含明确标题、取消/关闭与提交；Esc 关闭，焦点返回触发者。记录保存后关闭表单。

## 状态与动效
无数据给添加动作；无搜索命中说明原因；存储错误常驻提示和重试。只有 120ms 颜色过渡，不让频繁操作等待动画；prefers-reduced-motion 关闭过渡。

## 实现来源
styles.css :root 是唯一代码 token 来源；app.js 不生成配色或字体。三个页面状态均复用同一组件。中文长名称换行而非横向溢出。

## 例外
无页面风格例外。演练未接网络，没有网络加载动画；本地同步操作不虚构 loading。
