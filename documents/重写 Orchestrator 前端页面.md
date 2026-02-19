## 目标与边界
- 完全重写现有 [index.html](file:///Users/bytedance/GolandProjects/ai-nexus/examples/orchestrator/static/index.html)，保留后端接口与功能
- 保证会话隔离、群聊式 UI、@ 提示、流式输出、自定义模态框、工具进度与思考展示
- 继续使用现有接口（/sessions、/agents、/chat/stream）以保证功能可用

## 页面与布局
- 删除现有 DOM，重建三栏结构：左侧会话列表、中间群聊主区、右侧 Agent 面板
- 群聊主区包含：顶部会话信息、消息流、底部输入框与 @ 联想面板
- 视觉上强化“群聊”感：消息卡带“从/到”标签、不同 Agent 头像与色块

## 会话与消息逻辑
- 继续基于 /sessions 与 /sessions/:id 加载会话，实现会话切换与隔离
- 渲染 group_messages 优先；没有则渲染 messages
- 群聊消息按时间排序插入，保持实时插入

## @ 规则与输入体验
- 用户发主 AI：不需要 @，直接发送
- 用户发其他 Agent：必须 @，从输入框联想选择
- Agent 之间与 Agent → 用户：渲染时强制展示 @ 目标（体现规则）
- 联想列表中隐藏主 AI（不出现在建议里）

## 流式与进度展示
- /chat/stream 继续用 SSE
- 主 AI 回复保持真实流式；其他 Agent 消息用“逐字显示”做流式感
- 为每个 Agent 消息卡提供“内部面板”，展示思考、工具调用与结果（可折叠）

## 自定义模态与提示
- 所有弹窗用自建遮罩与卡片，不用原生 alert/confirm
- 统一 Toast 提示样式与状态

## 验证方式
- 手动打开页面，验证会话切换、@ 联想、流式输出、工具进度、模态框与删除会话
- 发送：无 @（主 AI）与有 @（其他 Agent）两类消息做检查

如确认该计划，我会直接重写 index.html 并完成上述功能。