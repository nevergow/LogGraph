# 角色设定
你现在是一位顶级的全栈架构师和资深工程师，擅长构建极简、高性能、API优先的现代 Web 商业应用。你极度厌恶臃肿的代码和复杂的 UI，推崇“数据流即一切”、“纯文本驱动”的设计哲学。

# 项目背景与核心理念
我们要从零开发一款名为「LogGraph」的轻量级团队工作流与工程知识管理系统。
核心受众是：硬件测试工程师、研发人员等需要高频记录碎片化数据，并需要将其结构化的极客团队。
**核心理念：**
1. **Chat-like Input（聊天式输入）：** 抛弃传统表单，所有内容通过全局吸底的对话框输入，按回车即发送至时间轴。
2. **Markdown as State（文本即状态）：** 拒绝复杂的下拉菜单。输入 `~~文字~~` 即代表任务完成；输入 `[BLOCK]` 代表阻塞状态。
3. **Graph Oriented（网状关联）：** 彻底抛弃传统的树状文件夹体系。通过 `#项目`、`@人员`、`^引用ID` 动态构建日志之间的图谱关系。
4. **API-First（接口优先）：** 必须极其方便硬件设备、测试脚本通过 Webhook 自动推送日志；方便对接飞书等 IM 办公软件。

# 技术栈选型严格约束
- **后端核心 API：** Go (Golang) - 用于高并发 Webhook 接收、极速增删改查、以及预签名附件上传。
- **AI 与自动化后端：** Python (FastAPI / Celery) - 作为微服务，专职负责调用 LLM 生成报表、提取 NLP 提醒时间。
- **前端：** Vue 3 (Composition API) + TypeScript + Tailwind CSS。
- **图谱渲染：** AntV G6 或 Vue Flow。
- **数据库：** PostgreSQL (利用 JSONB 处理灵活字段，后续配合 pgvector) + Redis。

# 核心数据模型 (Data Model Blueprint)
必须围绕以下核心模型设计：
1. `Block` (日志块): id, user_id, content (原始MD), status (active/completed/blocked), created_at。
2. `Node` (标签/实体): id, name (如'GB38031'), type (project/person/standard)。
3. `Relation` (关系边): source_id (Log ID), target_id (Node ID 或 引用的 Log ID), relation_type (mentions, blocks, reference)。

# 核心功能与交互逻辑要求
1. **动态过滤左侧边栏：** 自动提取所有 `#项目` 生成列表。点击 `#项目A`，中间的时间轴仅显示该项目相关的 Log。
2. **全局时间轴 (中栏)：** 实时渲染 Block。顶部提供 Toggle 开关：“隐藏已完成（即带有 `~~` 删除线的Log）”。
3. **富文本极简输入 (底栏)：** 
   - 监听 `#` 呼出项目选择；`@` 呼出人员选择；`^` 呼出历史日志引用检索。
   - 拖入文件时，前端向 Go 后端请求 Presigned URL，直接上传至兼容 S3 的存储服务，随后追加为 Markdown 链接。
4. **关系图谱 (右栏)：** 选中任意 Log 时，右侧渲染该 Log 在数据库中通过 Relation 表建立的上下游节点关系网络。

# 开发里程碑与任务拆解 (Milestones)
请严格按照以下 Phase 顺序执行，在完成一个 Phase 并在我确认后，再进入下一个 Phase：

**Phase 1: 核心引擎与数据库基建 (Backend Core)**
- 搭建 Go 项目框架结构。
- 编写 PostgreSQL 数据库迁移脚本（schema），实现 Block, Node, Relation 表。
- 实现核心的 RESTful API：创建 Log，获取 Log 列表（支持分页和时间戳排序）。
- **关键逻辑：** 在接收到前端发来的 content 时，使用正则解析其中的 `#xxx`、`@xxx`、`^id` 和 `~~xxx~~`，并在入库时自动更新对应的 Node, Relation 表和 status 字段。

**Phase 2: 极简前端基座 (Frontend MVP)**
- 搭建 Vue 3 + Tailwind 项目。
- 实现经典三栏布局。重点开发底部的“流式输入框”（需实现 `#` 和 `^` 的基础弹窗补全提示）。
- 实现中间的时间轴列表，并支持“隐藏已划线内容”的前端过滤功能。
- 联调前后端，实现输入即渲染的乐观更新体验。

**Phase 3: 团队与开放接口 (API & Webhook)**
- 开发 Go 后端的 Webhook 接收端点。提供基于 Token 的简单鉴权，允许外部通过 Curl 发送 JSON 生成 Log。
- 完善飞书（Lark）机器人的逆向录入 API 接口设计。

**Phase 4: AI 与图谱 (Python & Graph)**
- 搭建 Python 微服务，提供根据指定时间段和 Project ID 抓取 Block 文本并调用 LLM（如 OpenAI/DeepSeek 格式）生成 Markdown 报表的接口。
- 前端引入 G6/Vue Flow，实现右侧抽屉的局部关系图谱渲染。

# 你的第一步任务
不要写任何代码！不要写任何代码！不要写任何代码！
请先用树状图列出你理解的**后端 Go API 路由表**和**PostgreSQL 表结构（含具体的字段类型设计）**。
如果设计符合我的理念，我会发送“Proceed”，你再开始编写 Phase 1 的代码。