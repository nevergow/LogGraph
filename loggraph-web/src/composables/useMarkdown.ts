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

/**
 * Strip binding syntax (&Name, @Person, ^uuid) from text.
 * Only strips known projects/people; ^uuid is always stripped.
 */
export function stripBindingTags(text: string, knownProjects?: Set<string>, knownPeople?: Set<string>): string {
  let result = text
  if (knownProjects) {
    // Build regex dynamically from known project names (sorted longest-first to avoid partial matches)
    const names = [...knownProjects].sort((a, b) => b.length - a.length)
    if (names.length > 0) {
      const escaped = names.map(n => n.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'))
      // Strip #Name (backward compat) and &Name for known projects
      const projectRe = new RegExp(`(^|\\s)[#&](${escaped.join('|')})(?=\\s|$)`, 'g')
      result = result.replace(projectRe, '$1')
    }
  }
  if (knownPeople) {
    const names = [...knownPeople].sort((a, b) => b.length - a.length)
    if (names.length > 0) {
      const escaped = names.map(n => n.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'))
      const personRe = new RegExp(`(^|\\s)@(${escaped.join('|')})(?=\\s|$)`, 'g')
      result = result.replace(personRe, '$1')
    }
  }
  // ^uuid references are always stripped (can't be accidental)
  result = result.replace(/(^|\s)\^([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})\b/gi, '$1')
  return result
}

function capsulizeTags(text: string, knownProjects?: Set<string>, knownPeople?: Set<string>): string {
  const codeBlocks: string[] = []
  let processed = text.replace(/```[\s\S]*?```/g, (m) => {
    codeBlocks.push(m)
    return `\x00CODE${codeBlocks.length - 1}\x00`
  })
  processed = processed.replace(/`[^`]+`/g, (m) => {
    codeBlocks.push(m)
    return `\x00CODE${codeBlocks.length - 1}\x00`
  })

  // Capsulize #Name as &Name for known projects (backward compat)
  processed = processed.replace(/(^|\s)#([^\s#<>]+)/g, (match, before, name) => {
    if (knownProjects && !knownProjects.has(name)) return match
    return `${before}<span class="tag-capsule tag-project">&amp;${name}</span>`
  })
  // Capsulize &Name if known project (or if no known list, always capsulize)
  processed = processed.replace(/(^|\s)&([^\s&<>]+)/g, (match, before, name) => {
    if (knownProjects && !knownProjects.has(name)) return match
    return `${before}<span class="tag-capsule tag-project">&amp;${name}</span>`
  })
  // Only capsulize @Person if known person (or if no known list, always capsulize)
  processed = processed.replace(/(^|\s)@([^\s@<>]+)/g, (match, before, name) => {
    if (knownPeople && !knownPeople.has(name)) return match
    return `${before}<span class="tag-capsule tag-person">@${name}</span>`
  })
  // ^uuid is always capsulized
  processed = processed.replace(/(^|\s)\^([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})\b/gi,
    '$1<span class="tag-capsule tag-reference">^$2</span>')

  return processed.replace(/\x00CODE(\d+)\x00/g, (_, i) => codeBlocks[parseInt(i)])
}

export function renderMarkdown(text: string, knownProjects?: Set<string>, knownPeople?: Set<string>): string {
  const capsulized = capsulizeTags(text, knownProjects, knownPeople)
  const raw = marked.parse(capsulized) as string
  return DOMPurify.sanitize(raw)
}

export function extractTitle(text: string, knownProjects?: Set<string>, knownPeople?: Set<string>): string {
  // Strip binding tags first so &Name doesn't appear in title
  const cleaned = stripBindingTags(text, knownProjects, knownPeople)
  const tokens = marked.lexer(cleaned)
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
  return cleaned.split('\n')[0] || ''
}
