import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import rehypeHighlight from 'rehype-highlight'
import { Typography } from 'antd'
import 'highlight.js/styles/github.css' // 代码高亮样式

const { Text, Paragraph } = Typography

/**
 * Markdown渲染器组件
 * 功能：
 * 1. 渲染Markdown格式的文本
 * 2. 支持GitHub风格的Markdown (GFM)
 * 3. 代码块语法高亮
 * 4. 表格、任务列表等扩展语法
 * 5. 与Ant Design样式系统兼容
 */

interface MarkdownRendererProps {
  content: string           // 要渲染的Markdown内容
  isStreaming?: boolean     // 是否正在流式传输（显示光标）
  className?: string        // 自定义CSS类名
}

/**
 * 自定义Markdown组件渲染器
 * 将Markdown元素映射到Ant Design组件
 */
const markdownComponents = {
  // 段落：使用Ant Design的Paragraph组件
  p: ({ children, ...props }: any) => (
    <Paragraph style={{ margin: '8px 0', lineHeight: 1.6, color: 'inherit' }} {...props}>
      {children}
    </Paragraph>
  ),

  // 代码块：添加样式和背景
  code: ({ inline, className, children, ...props }: any) => {
    if (inline) {
      // 行内代码：使用Ant Design的Code样式
      return (
        <Text 
          code 
          style={{ 
            backgroundColor: '#f6f8fa',
            padding: '2px 4px',
            borderRadius: '3px',
            fontSize: '0.9em'
          }}
          {...props}
        >
          {children}
        </Text>
      )
    } else {
      // 代码块：使用pre标签包装
      return (
        <pre 
          className={className}
          style={{
            backgroundColor: '#f6f8fa',
            border: '1px solid #e1e4e8',
            borderRadius: '6px',
            padding: '16px',
            overflow: 'auto',
            fontSize: '14px',
            lineHeight: 1.45,
            margin: '12px 0'
          }}
          {...props}
        >
          <code>{children}</code>
        </pre>
      )
    }
  },

  // 标题：使用Typography.Title，继承颜色
  h1: ({ children, ...props }: any) => (
    <Typography.Title level={1} style={{ margin: '16px 0 8px 0', color: 'inherit' }} {...props}>
      {children}
    </Typography.Title>
  ),
  h2: ({ children, ...props }: any) => (
    <Typography.Title level={2} style={{ margin: '16px 0 8px 0', color: 'inherit' }} {...props}>
      {children}
    </Typography.Title>
  ),
  h3: ({ children, ...props }: any) => (
    <Typography.Title level={3} style={{ margin: '16px 0 8px 0', color: 'inherit' }} {...props}>
      {children}
    </Typography.Title>
  ),
  h4: ({ children, ...props }: any) => (
    <Typography.Title level={4} style={{ margin: '16px 0 8px 0', color: 'inherit' }} {...props}>
      {children}
    </Typography.Title>
  ),
  h5: ({ children, ...props }: any) => (
    <Typography.Title level={5} style={{ margin: '16px 0 8px 0', color: 'inherit' }} {...props}>
      {children}
    </Typography.Title>
  ),

  // 列表：添加适当的间距
  ul: ({ children, ...props }: any) => (
    <ul style={{ margin: '8px 0', paddingLeft: '20px' }} {...props}>
      {children}
    </ul>
  ),
  ol: ({ children, ...props }: any) => (
    <ol style={{ margin: '8px 0', paddingLeft: '20px' }} {...props}>
      {children}
    </ol>
  ),
  li: ({ children, ...props }: any) => (
    <li style={{ margin: '4px 0' }} {...props}>
      {children}
    </li>
  ),

  // 引用：使用Ant Design的样式
  blockquote: ({ children, ...props }: any) => (
    <blockquote 
      style={{
        borderLeft: '4px solid #d1d5da',
        backgroundColor: '#f6f8fa',
        padding: '8px 16px',
        margin: '12px 0',
        borderRadius: '0 4px 4px 0'
      }} 
      {...props}
    >
      {children}
    </blockquote>
  ),

  // 表格：添加边框和样式
  table: ({ children, ...props }: any) => (
    <table 
      style={{
        borderCollapse: 'collapse',
        width: '100%',
        margin: '12px 0',
        border: '1px solid #d1d5da'
      }} 
      {...props}
    >
      {children}
    </table>
  ),
  th: ({ children, ...props }: any) => (
    <th 
      style={{
        border: '1px solid #d1d5da',
        padding: '8px 12px',
        backgroundColor: '#f6f8fa',
        fontWeight: 600,
        textAlign: 'left'
      }} 
      {...props}
    >
      {children}
    </th>
  ),
  td: ({ children, ...props }: any) => (
    <td 
      style={{
        border: '1px solid #d1d5da',
        padding: '8px 12px'
      }} 
      {...props}
    >
      {children}
    </td>
  ),

  // 链接：使用Ant Design的Link样式
  a: ({ children, href, ...props }: any) => (
    <Text 
      style={{ color: '#1890ff' }}
      {...props}
    >
      <a href={href} target="_blank" rel="noopener noreferrer">
        {children}
      </a>
    </Text>
  ),

  // 强调：粗体和斜体
  strong: ({ children, ...props }: any) => (
    <Text strong {...props}>{children}</Text>
  ),
  em: ({ children, ...props }: any) => (
    <Text italic {...props}>{children}</Text>
  ),

  // 分隔线
  hr: ({ ...props }: any) => (
    <hr 
      style={{
        border: 'none',
        borderTop: '1px solid #e1e4e8',
        margin: '16px 0'
      }} 
      {...props} 
    />
  )
}

export function MarkdownRenderer({ content, isStreaming, className }: MarkdownRendererProps) {
  return (
    <div className={className} style={{ wordBreak: 'break-word' }}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}           // 支持GitHub风格Markdown
        rehypePlugins={[rehypeHighlight]}     // 代码高亮
        components={markdownComponents}       // 自定义组件渲染
      >
        {content}
      </ReactMarkdown>
      {/* 流式传输时显示光标 */}
      {isStreaming && <span className="blinking-cursor">|</span>}
    </div>
  )
}

export default MarkdownRenderer