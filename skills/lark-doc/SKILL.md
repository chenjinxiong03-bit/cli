---
name: lark-doc
version: 1.0.0
description: "飞书云文档：创建和编辑飞书文档。默认使用 DocxXML 格式（也支持 Markdown）。创建文档、获取文档内容（支持 simple/with-ids/full 三种导出详细度）、更新文档（七种指令：str_replace/str_delete/block_insert_after/block_replace/block_delete/overwrite/append）、上传和下载文档中的图片和文件、搜索云空间文档。当用户需要创建或编辑飞书文档、读取文档内容、在文档中插入图片、搜索云空间文档时使用；如果用户是想按名称或关键词先定位电子表格、报表等云空间对象，也优先使用本 skill 的 docs +search 做资源发现。"
metadata:
  requires:
    bins: ["lark-cli"]
  cliHelp: "lark-cli docs --help"
---

# docs (v1)

**CRITICAL — 创建或编辑文档前，MUST 先用 Read 工具读取以下文件，缺一不可：**
1. [`../lark-shared/SKILL.md`](../lark-shared/SKILL.md) — 认证、权限处理
2. [`lark-doc-xml.md`](references/lark-doc-xml.md) — XML 语法规则（使用 Markdown 格式时改读 [`lark-doc-md.md`](references/lark-doc-md.md)）
3. [`lark-doc-style.md`](references/style/lark-doc-style.md) — 排版指南（元素选择、丰富度规则、颜色语义）

**未读完以上文件就生成内容会导致格式错误或样式不达标。**

> **⚠️ 格式选择规则（全局）：`docs +create` 和 `docs +update` 始终使用 XML 格式（`--doc-format xml`，即默认值），除非用户明确要求使用 Markdown。** XML 是飞书文档的首选格式，表达能力更强，支持 callout、grid、checkbox 等丰富 block 类型。不要因为 Markdown 更简单就自行选择 Markdown。

## 核心概念

### 文档类型与 Token

飞书开放平台中，不同类型的文档有不同的 URL 格式和 Token 处理方式。在进行文档操作（如添加评论、下载文件等）时，必须先获取正确的 `file_token`。

### 文档 URL 格式与 Token 处理

| URL 格式 | 示例                                                      | Token 类型 | 处理方式 |
|----------|---------------------------------------------------------|-----------|----------|
| `/docx/` | `https://example.larksuite.com/docx/doxcnxxxxxxxxx`    | `file_token` | URL 路径中的 token 直接作为 `file_token` 使用 |
| `/doc/` | `https://example.larksuite.com/doc/doccnxxxxxxxxx`     | `file_token` | URL 路径中的 token 直接作为 `file_token` 使用 |
| `/wiki/` | `https://example.larksuite.com/wiki/wikcnxxxxxxxxx`    | `wiki_token` | ⚠️ **不能直接使用**，需要先查询获取真实的 `obj_token` |
| `/sheets/` | `https://example.larksuite.com/sheets/shtcnxxxxxxxxx`  | `file_token` | URL 路径中的 token 直接作为 `file_token` 使用 |
| `/drive/folder/` | `https://example.larksuite.com/drive/folder/fldcnxxxx` | `folder_token` | URL 路径中的 token 作为文件夹 token 使用 |

### Wiki 链接特殊处理（关键！）

知识库链接（`/wiki/TOKEN`）背后可能是云文档、电子表格、多维表格等不同类型的文档。**不能直接假设 URL 中的 token 就是 file_token**，必须先查询实际类型和真实 token。

#### 处理流程

1. **使用 `wiki.spaces.get_node` 查询节点信息**
   ```bash
   lark-cli wiki spaces get_node --params '{"token":"wiki_token"}'
   ```

2. **从返回结果中提取关键信息**
   - `node.obj_type`：文档类型（docx/doc/sheet/bitable/slides/file/mindnote）
   - `node.obj_token`：**真实的文档 token**（用于后续操作）
   - `node.title`：文档标题

3. **根据 `obj_type` 使用对应的 API**

   | obj_type | 说明 | 使用的 API |
   |----------|------|-----------|
   | `docx` | 新版云文档 | `drive file.comments.*`、`docx.*` |
   | `doc` | 旧版云文档 | `drive file.comments.*` |
   | `sheet` | 电子表格 | `sheets.*` |
   | `bitable` | 多维表格 | `bitable.*` |
   | `slides` | 幻灯片 | `drive.*` |
   | `file` | 文件 | `drive.*` |
   | `mindnote` | 思维导图 | `drive.*` |

#### 查询示例

```bash
# 查询 wiki 节点
lark-cli wiki spaces get_node --params '{"token":"wiki_token"}'
```

返回结果示例：
```json
{
   "node": {
      "obj_type": "docx",
      "obj_token": "xxxx",
      "title": "标题",
      "node_type": "origin",
      "space_id": "12345678910"
   }
}
```

### 资源关系

```
Wiki Space (知识空间)
└── Wiki Node (知识库节点)
    ├── obj_type: docx (新版文档)
    │   └── obj_token (真实文档 token)
    ├── obj_type: doc (旧版文档)
    │   └── obj_token (真实文档 token)
    ├── obj_type: sheet (电子表格)
    │   └── obj_token (真实文档 token)
    ├── obj_type: bitable (多维表格)
    │   └── obj_token (真实文档 token)
    └── obj_type: file/slides/mindnote
        └── obj_token (真实文档 token)

Drive Folder (云空间文件夹)
└── File (文件/文档)
    └── file_token (直接使用)
```

## 重要说明：画板编辑
> **⚠️ 不能直接编辑已有画板。**

### 创建画板：`<whiteboard>` 标签、
在 `docs +create` / `docs +update` 中用 `<whiteboard type="mermaid|plantuml">语法</whiteboard>` 直接创建带内容的画板。

### 高级用法：DSL 画板
Mermaid / PlantUML 无法满足时（架构图、对比图等），按以下步骤使用 DSL：
1. **创建空白画板** — 用 `<whiteboard type="blank"></whiteboard>` 通过`docs +update` 或 `docs +create`创建
2. **提取画板 token** — 从响应 `data.document.newblocks` 中取画板的 `token`
3. **设计并上传 DSL** — 参考 [`lark-whiteboard`](../lark-whiteboard/SKILL.md) skill 编写 DSL。

## 文档可视化建议
> **💡 在撰写文档时，当需要表达较为复杂的时序、架构层次、逻辑关系、数据流程等内容时，建议使用画板绘制可视化图表以显著提升文档的可阅读性。**
> 
> 请参考 [`../lark-whiteboard/SKILL.md`](../lark-whiteboard/SKILL.md) 了解如何绘制画板内容。

## 快速决策
- 用户说“看一下文档里的图片/附件/素材”“预览素材”，优先用 `lark-cli docs +media-preview`。
- 用户明确说“下载素材”，再用 `lark-cli docs +media-download`。
- 如果目标明确是画板 / whiteboard / 画板缩略图，只能用 `lark-cli docs +media-download --type whiteboard`，不要用 `+media-preview`。
- 用户说“找一个表格”“按名称搜电子表格”“找报表”“最近打开的表格”，先用 `lark-cli docs +search` 做资源发现。
- `docs +search` 不是只搜文档 / Wiki；结果里会直接返回 `SHEET` 等云空间对象。
- 拿到 spreadsheet URL / token 后，再切到 `lark-sheets` 做对象内部读取、筛选、写入等操作。
- 用户说“给文档加评论”“查看评论”“回复评论”“给评论加表情 / reaction”“删除评论表情 / reaction”，**不要留在 `lark-doc`**，直接切到 `lark-drive` 处理。

## 画板需求挖掘（主动识别）

> **用户很少主动提"画板"。创建文档时应主动识别适合可视化的内容，用画板呈现。**

### 🔴 关键要求（必须遵守）

**创建空白画板 ≠ 完成任务**。创建空白画板后，**必须继续使用 lark-whiteboard 技能填充实际内容**。

### 语义与画板类型映射

创建/编辑文档时，文档主题涉及以下语义，应**主动**创建画板，无需用户指定：

| 语义            | 画板类型  | 参考指南                                                                                        |
|---------------|-------|---------------------------------------------------------------------------------------------|
| 架构/分层/技术方案    | 架构图   | [lark-whiteboard-cli/scenes/architecture.md](../lark-whiteboard-cli/scenes/architecture.md) |
| 流程/审批/部署/业务流转 | 流程图   | [lark-whiteboard-cli/scenes/flowchart.md](../lark-whiteboard-cli/scenes/flowchart.md)       |
| 组织/层级/汇报关系    | 组织架构图 | [lark-whiteboard-cli/scenes/organization.md](../lark-whiteboard-cli/scenes/organization.md) |
| 时间线/里程碑/版本规划  | 里程碑图  | [lark-whiteboard-cli/scenes/milestone.md](../lark-whiteboard-cli/scenes/milestone.md)       |
| 因果/复盘/根因分析    | 鱼骨图   | [lark-whiteboard-cli/scenes/fishbone.md](../lark-whiteboard-cli/scenes/fishbone.md)         |
| 方案对比/技术选型     | 对比图   | [lark-whiteboard-cli/scenes/comparison.md](../lark-whiteboard-cli/scenes/comparison.md)     |
| 循环/飞轮/闭环      | 飞轮图   | [lark-whiteboard-cli/scenes/flywheel.md](../lark-whiteboard-cli/scenes/flywheel.md)         |
| 层级占比/能力模型     | 金字塔图  | [lark-whiteboard-cli/scenes/pyramid.md](../lark-whiteboard-cli/scenes/pyramid.md)           |
| 模块依赖/调用关系     | 架构图   | [lark-whiteboard-cli/scenes/architecture.md](../lark-whiteboard-cli/scenes/architecture.md) |
| 分类梳理/知识体系     | 思维导图  | [lark-whiteboard-cli/scenes/mermaid.md](../lark-whiteboard-cli/scenes/mermaid.md)           |
| 数据分布/占比       | 饼图    | [lark-whiteboard-cli/scenes/mermaid.md](../lark-whiteboard-cli/scenes/mermaid.md)           |

创建画板前，务必先阅读 [`lark-whiteboard-cli`](../lark-whiteboard-cli/SKILL.md) 和 [`lark-whiteboard`](../lark-whiteboard/SKILL.md) 这两个 Skill，了解画板的创建流程。

### 完整执行流程（必须完整执行）

1. **创建空白画板占位**：创建场景用 `docs +create`、编辑场景用 `docs +update` 插入空白画板
2. **获取画板 token**：从 `docs +update` 响应的 `data.board_tokens` 获取画板 token 列表
3. **填充画板内容**：切换到 [`lark-whiteboard-cli`](../lark-whiteboard-cli/SKILL.md) 创建画板内容，并填入画板
4. **验证完成**：确认所有画板都有实际内容，不是空白

**不适用**：纯文字记录（日志/备忘）、数据密集型内容（用表格）、用户明确只要文字。

> ⚠️ **警告**：如果只创建空白画板而不填充内容，任务将被视为未完成！

## 补充说明 
`docs +search` 除了搜索文档 / Wiki，也承担“先定位云空间对象，再切回对应业务 skill 操作”的资源发现入口角色；当用户口头说“表格 / 报表”时，也优先从这里开始。

## Shortcuts（推荐优先使用）

Shortcut 是对常用操作的高级封装（`lark-cli docs +<verb> [flags]`）。有 Shortcut 的操作优先使用。

| Shortcut | 说明 |
|----------|------|
| [`+search`](references/lark-doc-search.md) | Search Lark docs, Wiki, and spreadsheet files (Search v2: doc_wiki/search) |
| [`+create`](references/lark-doc-create.md) | Create a Lark document (XML / Markdown) |
| [`+fetch`](references/lark-doc-fetch.md) | Fetch Lark document content (XML / Markdown) |
| [`+update`](references/lark-doc-update.md) | Update a Lark document (str_replace / block_insert_after / block_replace / ...) |
| [`+media-insert`](references/lark-doc-media-insert.md) | Insert a local image or file at the end of a Lark document (4-step orchestration + auto-rollback) |
| [`+media-download`](references/lark-doc-media-download.md) | Download document media or whiteboard thumbnail (auto-detects extension) |
| [`+whiteboard-update`](references/lark-doc-whiteboard-update.md) | Update an existing whiteboard in lark document with whiteboard dsl. Such DSL input from stdin. refer to lark-whiteboard skill for more details. |
