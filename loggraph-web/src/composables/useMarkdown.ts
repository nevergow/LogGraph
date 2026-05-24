import { marked } from 'marked'
import DOMPurify from 'dompurify'
import hljs from 'highlight.js'

const renderer = new marked.Renderer()
renderer.code = function ({ text, lang }: { text: string; lang?: string }) {
  if (lang && hljs.getLanguage(lang)) {
    try {
      return `<pre><code class="hljs language-${lang}">${hljs.highlight(text, { language: lang }).value}</code></pre>`
    } catch { /* fall through */ }
  }
  return `<pre><code class="hljs">${hljs.highlightAuto(text).value}</code></pre>`
}

marked.use({ gfm: true, breaks: false, renderer })

DOMPurify.addHook('afterSanitizeAttributes', (node) => {
  if (node.tagName === 'A') {
    node.setAttribute('target', '_blank')
    node.setAttribute('rel', 'noopener noreferrer')
  }
})

function capsulizeTags(text: string): string {
  const codeBlocks: string[] = []
  let processed = text.replace(/```[\s\S]*?```/g, (m) => {
    codeBlocks.push(m)
    return `\x00CODE${codeBlocks.length - 1}\x00`
  })
  processed = processed.replace(/`[^`]+`/g, (m) => {
    codeBlocks.push(m)
    return `\x00CODE${codeBlocks.length - 1}\x00`
  })

  processed = processed.replace(/(^|\s)&([^\s&<>]+)/g,
    '$1<span class="tag-capsule tag-project">&amp;$2</span>')
  processed = processed.replace(/(^|\s)@([^\s@<>]+)/g,
    '$1<span class="tag-capsule tag-person">@$2</span>')
  processed = processed.replace(/(^|\s)\^([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})\b/gi,
    '$1<span class="tag-capsule tag-reference">^$2</span>')

  return processed.replace(/\x00CODE(\d+)\x00/g, (_, i) => codeBlocks[parseInt(i)])
}

export function renderMarkdown(text: string): string {
  const capsulized = capsulizeTags(text)
  const raw = marked.parse(capsulized) as string
  return DOMPurify.sanitize(raw)
}

export function extractTitle(text: string): string {
  const tokens = marked.lexer(text)
  // Find highest-level heading first
  let bestHeading = ''
  let bestLevel = 7
  for (const t of tokens) {
    if (t.type === 'heading') {
      if ((t as any).depth < bestLevel) {
        bestLevel = (t as any).depth
        bestHeading = (t as any).text
      }
    }
  }
  if (bestHeading) return bestHeading

  // Fall back to first non-empty paragraph/paragraph text
  for (const t of tokens) {
    if (t.type === 'paragraph') {
      const txt = (t as any).text
      if (txt) return txt
    }
    if (t.type === 'text' || t.type === 'code') {
      const txt = (t as any).text
      if (txt) return txt
    }
  }

  // Ultimate fallback: first line of raw content
  return text.split('\n')[0] || ''
}
