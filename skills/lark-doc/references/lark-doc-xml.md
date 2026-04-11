基于 HTML 子集的 XML 格式描述飞书文档内容。

# 一、标准 HTML 标签
p, h1-h9, ul, ol, li, table, thead, tbody, tr, th, td, blockquote, pre, code, hr, img, b, em, u, del, a, br, span 语义不变

# 二、扩展标签速查表
## 块级标签
|标签|说明|关键属性|
|-|-|-|
| `<title>` | 文档标题（每篇唯一）| `align` |
| `<checkbox>` | 待办项| `done="true"\|"false"` |

## 容器标签
|标签|说明|关键属性|
|-|-|-|
| `<callout>` | 高亮框，子块仅支持文本、标题、列表、待办、引用 | `emoji`(默认 bulb), `background-color`, `border-color`, `text-color` |
| `<grid>` + `<column>` | 分栏布局，各列 width-ratio 之和为 1 | `width-ratio` |
| `<whiteboard>` | 嵌入画板 | `type`: `mermaid` \| `plantuml` \| `blank` |
| `<pre>` | （代码块，内含 `code`）| `lang`, `caption` |
| `<figure>` | 视图容器 | `view-type` |

## 行内组件
| 标签 | 说明 | 关键属性 |
|-------------------------|-|-|
| `<cite type="user">` | @人 | `<cite type="user" user-id="userID"></cite>` |
| `<cite type="doc">` | @文档 | `<cite type="doc" doc-id="docx_token"></cite>` |
| `<latex>` | 行内公式 | `<latex>E = mc^2</latex>` |
| `<img>` | 图片（可独立成块或内联） | `<img width="800" height="600" caption="说明" name="图.png" href="http 或 https"/>` |
| `<source>` | 文件附件（可独立成块或内联） | `<source name="报告.pdf"/>` |
| `<a type="url-preview">` | 预览卡片 | `<a type="url-preview" href="...">标题</a>` |
| `<button>` | 操作按钮 | `background-color`,src,必须包含 `action=OpenLink|DuplicatePage|FollowPage` |
| `<time>` | 提醒 | `必包含 expire-time、notify-time=毫秒时间戳、should-notify=true|false` |

## 文本块通用属性
- `align` — `"left"`|`"center"`|`"right"`（适用于 p / h1-h9 / li / checkbox）
- 有序列表项用 `seq="auto"` 自动编号

# 三、补充规则

## 富文本样式嵌套顺序
- 行内样式标签必须按以下固定顺序嵌套（外 → 内），关闭顺序严格反转：`<a> → <b> → <em> → <del> → <u> → <code> → <span> → 文本内容`

## 列表分组
- 连续同类型列表项自动合并为一个 `<ul>` 或 `<ol>`
- 嵌套子列表放在 `<li>` 内部
- 新增列表项必须包在 `<ul>` 或 `<ol>` 内：
   ```xml
   <ul>
     <li>第一项</li>
     <li>第二项</li>
   </ul>
   ```


## 表格扩展
标准 HTML table 结构不变，扩展点：
- `<colgroup>` / `<col>` 定义列宽，紧跟 `<table>` 之后：`<col span="2" width="100"/>`
- `<th>` / `<td>` 增加 `background-color` 和 `vertical-align`（top | middle | bottom）
- 有表头时第一行在 `<thead>` 用 `<th>`，其余在 `<tbody>` 用 `<td>`
- 合并单元格仅起始格输出 `colspan` / `rowspan`，被合并的格不出现

# 四、美化系统
- 颜色优先使用命名色，也可写 `rgb(r,g,b)` / `rgba(r,g,b,a)`。**基础色（7 色）**：gray, red, orange, yellow, green, blue, purple
  | 属性 | 支持的命名色 |                                                                                                                                                                                                        
  |-|-|
  | 文字颜色 `<span text-color>` | 基础色 |
  | 高亮框字色 `<callout text-color>` | 基础色 |
  | 高亮框边框 `<callout border-color>` | 基础色 |                                                                                                                                                                                 
  | 文字背景 `<span background-color>` | 基础色 + `light-{色}` + `medium-gray` |                                                                                                                                                   
  | 高亮框填充 `<callout background-color>` | `gray` + `light-{色}` + `medium-{色}` |                                                                                                                                              
  | 单元格背景 `<th/td background-color>` | 同文字背景 |                                                                                                                                                                           
  | 按钮背景 `<button background-color>` | 同文字背景 |
- 常用 emoji： 💡(默认)✅❌⚠️📝❓❗👍❤️📌🏁⭐

# 五、**重要规则**
## 仅标签内部的文本内容需要转义，标签本身不需要转义：
| 原始字符 | 转义后 |
|-|-|
| `<` | `&lt;` |
| `>` | `&gt;` |
| `&` | `&amp;` |
| `\n`（换行符） | `<br/>` |


## 六、完整示例

```xml
<title>项目周报 - 第 12 周</title>

<h1>本周进展</h1>

<p>完成了 <b>用户认证模块</b> 重构，性能提升 <span text-color="green">40%</span>。</p>

<callout emoji="💡" background-color="light-blue" border-color="blue">
  <p>P99 延迟从 200ms 降到 120ms。</p>
</callout>

<h2>任务清单</h2>

<checkbox done="true">完成认证模块重构</checkbox>
<checkbox done="false">更新 API 文档</checkbox>

<h2>方案对比</h2>

<grid>
  <column width-ratio="0.5">
    <h3>JWT 方案</h3>
    <ul><li>无状态，扩展性好</li></ul>
  </column>
  <column width-ratio="0.5">
    <h3>Session 方案</h3>
    <ul><li>可即时失效</li></ul>
  </column>
</grid>

<table>
  <colgroup>
    <col span="3" width="100"/>
  </colgroup>
  <thead>
    <tr>
      <th background-color="light-gray">指标</th>
      <th background-color="light-gray">重构前</th>
      <th background-color="light-gray">重构后</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>P99</td>
      <td>200ms</td>
      <td><span text-color="green">120ms</span></td>
    </tr>
  </tbody>
</table>

<p>详见 <cite type="doc" doc-id="性能报告"></cite>，请 <cite type="user" user-id="张三"></cite> 复核。</p>

<ol>
  <li seq="auto">完成文档更新</li>
  <li seq="auto">灰度发布</li>
</ol>

<p>参考：<a type="url-preview" href="https://wiki.example.com/gray">灰度指南</a></p>

<blockquote>
  <p>公式：<latex>P_{99} = \mu + 2.326\sigma</latex></p>
</blockquote>

<pre lang="go" caption="核心优化"><code>func Auth(ctx context.Context, token string) (*User, error) {
    if u, ok := cache.Get(token); ok { return u, nil }
    return redis.GetUser(ctx, token)
}</code></pre>

<hr/>

<p>附件：<source name="benchmark.pdf"/></p>
<p>图片：<img width="800" height="400" caption="架构图" name="arch.png"/></p>
<p>网络图片：<img href="https://example.com/photo.png"/></p>
<p>操作：<button action="OpenLink" src="https://example.com">打开面板</button></p>
<p>提醒：<time expire-time="1775916000000" notify-time="1775912400000" should-notify="false">xx时间截止</time></p>
<p>工单：<cite type="jira-issue">PROJ-456</cite></p>
<p>引文：<cite type="citation"><a href="https://example.com">参考文献 1</a></cite></p>
```