---
schema_version: 1
purpose: 帮单人使用者记住下一次客户跟进并快速记录进展
subject: 客户
roles:
  - name: 本地使用者
    scope: 当前浏览器全部记录
entities:
  - name: customer
    fields:
      id: {writer: system, type: string}
      name: {writer: user, type: string, required: true}
      nextDate: {writer: user, type: local-date, required: true}
      records: {writer: user-and-system, type: followup-array}
    relations: 一个客户有多条跟进记录
  - name: followup
    fields:
      text: {writer: user, type: string, required: true}
      at: {writer: system, type: ISO-datetime}
channels:
  - name: 待跟进
    pages: [{level: L1, shows: 按下一次日期排序的客户列表, actions: [记录跟进]}]
  - name: 全部客户
    pages: [{level: L1, shows: 可搜索客户列表, actions: [新增客户, 记录跟进]}]
home: [真实统计或冷启动, 待跟进客户, 下一动作]
hook:
  fields: [customer.nextDate, customer.records]
  writer: derived
  rule: 根据本地日期计算超期和今日客户数量
cold_start:
  day_1: 先添加一位客户，记下下一次联系日期。
  disconnected: 无外部数据源；本地存储失败显示本次修改未保存
mvp: [新增客户, 日期派生统计, 搜索, 跟进记录, 本地保存]
excluded: [团队权限, Excel导入, 登录, 云同步, 推送通知]
deferred: [后台集成]
---
# 客户跟进规格

## 事实与假设
用户事实：现在用 Excel，常忘跟进。演练假设：单人、桌面为主，手机能完成同样任务；这是示例假设，未冒充用户确认。原型范围已在演练中固定，不处理真实客户资料。

## 使用流程
首次添加客户和下一次日期。回访时先看到超期/今日的真实数量，列表按日期排序。记录一次跟进和新的日期，保存为历史记录并更新统计。全部客户可搜索姓名。

数据全由当前浏览器的用户写入，时间戳和 id 由系统写入，统计派生。未录数据不填假客户；存储损坏不静默覆盖。记录可保留为内存状态，但页面明确告知未保存，用户可重试保存。

## 验收
- AC-01：空存储打开页面，显示添加客户动作和具体冷启动，不展示假统计或客户。
- AC-02：键盘进入新增表单，空名称无法保存；输入姓名和日期后保存，新增客户出现在全部客户。
- AC-03：创建超期、今日和未来各一位客户，待跟进只显示前两位；真实统计为超期 1、今日 1、全部 3。
- AC-04：对超期客户记录跟进并改到未来，超期数量变 0；刷新后日期与历史记录仍在。
- AC-05：按姓名搜索，只显示命中项；无命中显示可解释空状态。
- AC-06：在 1440px 与 390px 宽度无水平溢出，主操作、中文长姓名、列表和表单可用；减弱动效时保留可操作状态。
- AC-07：本地存储写失败时显示未保存和重试入口；原客户内存仍可访问。损坏数据打开时提示，拒绝自动覆盖。

## 给实现方
首屏以待跟进为默认任务，不加入营销 hero。无数据先引导新增，不能放假客户。客户列表不改成无用途大图表。所有字段直接来自用户或系统，不新造跟进评分。字体、配色和组件遵循 DESIGN.md；原型不需要权限系统，不声称拥有生产权限控制。
