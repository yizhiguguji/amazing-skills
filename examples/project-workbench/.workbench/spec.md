---
name: 项目推进工作台
domain: system
subject: 项目
purpose: 让自由职业者先找到超期项目，再快速更新状态和下一步
surface: desktop
structure: {primary: pipeline, secondary: registry}
moment: 早上打开电脑开始工作
roles:
  - {name: 项目执行者, opens_daily: true, does: 录项目、推进阶段、记下一步}
dials: {cadence: 7, input: 3, depth: 7}
hook:
  text: "{未完成数} 个待推进 · {超期数} 个超期 · 先看最早交付的项目"
  shape: imperative
  fields:
    - {name: Project.stage, writes: user, when: 修改项目阶段, day_one: 先录第一个项目}
    - {name: Project.due, writes: user, when: 新建项目一次填写, day_one: 选择交付日期}
    - {name: pending_count, writes: derived, when: 打开或修改后从未完成项目计算, day_one: 不显示示例数}
    - {name: overdue_count, writes: derived, when: 本地日期大于 due 且未完成, day_one: 不显示示例数}
entities:
  - name: Project
    fields: [id, title, due, stage, created_at, updated_at]
    written_by: {id: system, title: user, due: user, stage: user, created_at: system, updated_at: system}
    relations: [Project 1-n Activity]
  - name: Activity
    fields: [project_id, text, created_at]
    written_by: {project_id: system, text: user, created_at: system}
    relations: [Activity n-1 Project]
  - name: Preferences
    fields: [owner_name]
    written_by: {owner_name: user}
    relations: [单人浏览器设置]
depends_on: []
cold_start:
  day_1: 先录一个项目和交付日期，今天该推进什么就清楚了。
  day_2: 根据已录项目计算，不要求每天额外打卡。
  day_7: 看真实阶段分布与进展记录，不生成虚构完成率。
  re_entry: 直接更新现在的阶段，不补写过去。
home: [hook, 四阶段看板按交付日期排序, 超期项目有文字提示]
channels:
  - name: 待推进
    type: today
    weight: primary
    does: 查看四阶段项目、打开详情并更新阶段
    pages:
      - {level: L1, shows: 待开始/进行中/待确认/已完成四列看板, actions: [录项目, 打开详情]}
      - {level: L2, shows: 项目日期、阶段与进展历史, actions: [更新阶段, 记录下一步]}
  - name: 项目
    type: record
    weight: regular
    does: 搜索项目并查看详细记录
    pages: [{level: L1, shows: 可搜索列表, actions: [打开详情]}]
  - name: 进展
    type: record
    weight: regular
    does: 查看跨项目的实际时间轴
    pages: [{level: L1, shows: 时间轴, actions: [打开关联项目]}]
  - name: 复盘
    type: review
    weight: occasional
    does: 对照真实项目数查看阶段分布
    pages: [{level: L1, shows: 阶段数量横向条形图与文字标签, actions: [回看项目]}]
  - name: 设置
    type: tool
    weight: occasional
    does: 设置当前使用者称呼并查看数据边界
    pages: [{level: L1, shows: 单项称呼表单, actions: [保存称呼]}]
mvp: [待推进, 项目, 进展, 复盘, 设置]
later: []
visual: 明亮桌面办公，系统中文字体，青绿语义色，密度适中，熟悉控件
seam: {type: none, why: 自己使用的本地工具}
excluded: [团队权限, 登录, 云同步, 推送, 消息平台接入, 删除记录]
deferred: [财务工作台]
---
# 产品定义

输入与假设来源：.product/input.md 的合成需求和结构化模拟回答。对象是项目而非消息；保留真实进度与交付日期，不假设消息平台 API 已接入。

## 每天怎么用
打开看板先处理超期项目。点项目打开详情，更新阶段并记录下一步，系统自动记录时间。所有列表和复盘从同一份项目数据派生，不维护重复指标。

## 范围与数据
当前浏览器 localStorage 原型，不跨设备。日期以浏览器本地日历计算；系统写入 id 和时间，用户写入项目名、阶段、日期、记录。损坏存储不覆盖；写失败保留内存并提示重试。

## 验收
- AC-01：空数据给录入动作，不显示假项目与假指标。
- AC-02：录入项目后进入相应阶段；改状态、记录下一步后刷新仍保留；已完成不算超期。
- AC-03：项目搜索、跨项目进展、真实阶段复盘和称呼保存可用。
- AC-04：1440/390 宽度无横溢，长中文、键盘与焦点、错误提示可用。
- AC-05：锁定规范后新增交付日期筛选，筛选仅影响看板；设计文件、CSS tokens 和共享控件外观保持不变。

## 给实现方
首屏先给结论或冷启动，再显示四列项目；不是营销页面。不同模块分别是看板、列表、时间轴、数据对比和单项设置，不做五张同款表。
每个统计只读 stage、due 与本地日期，已完成不加入待推进与超期。所有项目明细和记录属于同一数据源。
单人本地原型不造登录或权限系统。交付阶段禁止添加无来源金额和完成率。
设计方向先由真实 Pro Max 查询提供候选，经 Impeccable Operate 规则选定，规范在首轮完成后固化。

## 已实施后续变更

依据 input.md 的合成后续需求，看板增加交付日期筛选（全部、已超期且未完成、今天、未来）。筛选不改变项目模块、全局待推进统计、字体、颜色或共享组件。AC-05 是本轮回归验收。
