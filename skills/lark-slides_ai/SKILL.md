---
name: lark-slides_ai
version: 1.0.0
description: "飞书幻灯片AI：以XML格式创建和管理PPT。创建空白演示文稿、读取PPT全文信息、创建和删除幻灯片页面。当用户需要以XML格式创建PPT、读取PPT内容、管理幻灯片页面时使用。"
metadata:
  requires:
    bins: ["lark-cli"]
  cliHelp: "lark-cli slides_ai --help"
---

# slides_ai (v1)

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../lark-shared/SKILL.md`](../lark-shared/SKILL.md)，其中包含认证、权限处理**

## 执行前必做

> **重要**：`references/slides_xml_schema_definition.xml` 是此 skill 唯一正确的 XML 协议来源；其他 md 仅是对它和 CLI schema 的摘要。

| 命令 | 必须先阅读的文档 |
|------|-----------------|
| `xml_presentations.create` | [lark-slides_ai-xml-presentations-create.md](references/lark-slides_ai-xml-presentations-create.md) |
| `xml_presentations.get` | [lark-slides_ai-xml-presentations-get.md](references/lark-slides_ai-xml-presentations-get.md) |
| `xml_presentation.sildes.create` | [lark-slides_ai-xml-presentation-slides-create.md](references/lark-slides_ai-xml-presentation-slides-create.md) |
| `xml_presentation.sildes.delete` | [lark-slides_ai-xml-presentation-slides-delete.md](references/lark-slides_ai-xml-presentation-slides-delete.md) |

**涉及 XML 格式时，阅读顺序：**
1. [xml-schema-quick-ref.md](references/xml-schema-quick-ref.md) — **首选：Schema 精简速查**
2. [xml-format-guide.md](references/xml-format-guide.md) — 详细结构、属性与示例
3. [examples.md](references/examples.md) — CLI 调用示例
4. [slides_demo.xml](references/slides_demo.xml) — 真实 XML 示例
5. [slides_xml_schema_definition.xml](references/slides_xml_schema_definition.xml) — 完整 Schema

## API Resources

```bash
lark-cli schema slides_ai.<resource>.<method>   # 调用 API 前必须先查看参数结构
lark-cli slides_ai <resource> <method> [flags] # 调用 API
```

> **重要**：使用原生 API 时，必须先运行 `schema` 查看 `--data` / `--params` 参数结构，不要猜测字段格式。

### xml_presentations

  - `create` — 创建空白 PPT（当前仅支持标题和长宽）
  - `get` — 读取ppt全文信息，xml格式返回

### xml_presentation.sildes

  - `create` — 在指定 xml 演示文稿下创建页面
  - `delete` — 删除指定 xml 演示文稿下的页面

## 意图 → 命令索引

| 意图 | 推荐命令 | 备注 |
|------|---------|------|
| 创建空白 PPT | `lark-cli slides_ai xml_presentations create` | 当前仅用于创建空白 PPT，建议传 `<presentation ...><title>...</title></presentation>` |
| 读取 PPT 内容 | `lark-cli slides_ai xml_presentations get` | `--params` 传入 `{"xml_presentation_id":"..."}` |
| 添加幻灯片页面 | `lark-cli slides_ai xml_presentation.sildes create` | 创建空白 PPT 后，用它逐页添加 slide |
| 删除幻灯片页面 | `lark-cli slides_ai xml_presentation.sildes delete` | `--params` 传入 `xml_presentation_id` 和 `slide_id` |

## 核心规则

1. **先查 schema**：调用前先运行 `lark-cli schema slides_ai.<resource>.<method>`
2. **命名空间建议**：协议标准写法应带 `xmlns`，例如 `<presentation xmlns="http://www.larkoffice.com/sml/2.0" ...>`；当前服务端实现可能兼容不带 `xmlns` 的输入，但不作为协议保证
3. **根结构固定**：`<presentation>` 直接子元素只有 `<title>`、`<theme>`、`<slide>`
4. **slide 结构固定**：`<slide>` 直接子元素只有 `<style>`、`<data>`、`<note>`
5. **文本通过 content 表达**：页面正文通常放在 `shape/table/note` 内的 `<content>` 中
6. **创建流程要分两步**：先用 `xml_presentations.create` 创建空白 PPT，再用 `xml_presentation.sildes.create` 逐页添加 slide
7. **保存关键 ID**：后续操作需要 `xml_presentation_id`、`slide_id`、`revision_id`
8. **删除谨慎**：删除操作不可逆，且至少保留一页幻灯片

## 权限表

| 方法 | 所需 scope |
|------|-----------|
| `xml_presentations.create` | `slides:presentation:create` |
| `xml_presentations.get` | `slides:presentation:read` |
| `xml_presentation.sildes.create` | `slides:presentation:update` 或 `slides:presentation:write_only` |
| `xml_presentation.sildes.delete` | `slides:presentation:update` 或 `slides:presentation:write_only` |

## 常见错误速查

| 错误码 | 含义 | 解决方案 |
|--------|------|----------|
| 400 | XML 格式错误 | 检查 XML 语法，确保标签闭合 |
| 400 | create 内容超出支持范围 | `xml_presentations.create` 仅用于创建空白 PPT，不要在这里传完整 slide 内容 |
| 400 | 请求包装错误 | 检查 `--data` 是否按 schema 传入 `xml_presentation.content` 或 `slide.content` |
| 404 | 演示文稿不存在 | 检查 `xml_presentation_id` 是否正确 |
| 404 | 幻灯片不存在 | 检查 `slide_id` 是否正确 |
| 403 | 权限不足 | 检查是否拥有对应的 scope |
| 400 | 无法删除唯一幻灯片 | 演示文稿至少保留一页幻灯片 |

## 参考文档

### 快速参考
- [xml-schema-quick-ref.md](references/xml-schema-quick-ref.md) — **XML Schema 精简速查**
- [xml-format-guide.md](references/xml-format-guide.md) — XML 结构、内容模型、常用元素
- [examples.md](references/examples.md) — 常见 CLI 调用示例
- [slides_demo.xml](references/slides_demo.xml) — 真实 XML 示例

### 命令参考

- [lark-slides_ai-xml-presentations-create.md](references/lark-slides_ai-xml-presentations-create.md) — 创建空白 PPT
- [lark-slides_ai-xml-presentations-get.md](references/lark-slides_ai-xml-presentations-get.md) — 读取 PPT
- [lark-slides_ai-xml-presentation-slides-create.md](references/lark-slides_ai-xml-presentation-slides-create.md) — 添加幻灯片
- [lark-slides_ai-xml-presentation-slides-delete.md](references/lark-slides_ai-xml-presentation-slides-delete.md) — 删除幻灯片

### Schema 定义

- [xml-schema-quick-ref.md](references/xml-schema-quick-ref.md) — Schema 精简参考
- [slides_xml_schema_definition.xml](references/slides_xml_schema_definition.xml) — **完整 Schema 定义**（唯一协议依据）

> **注意**：如果 md 内容与 `slides_xml_schema_definition.xml` 或 `lark-cli schema slides_ai.<resource>.<method>` 输出不一致，以后两者为准。
