import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeSanitize from 'rehype-sanitize';

/**
 * Renders a Markdown string (event description, V-009) as sanitized HTML.
 * rehype-sanitize strips embedded scripts/event handlers/javascript: links
 * regardless of what the source contains — the editor never emits raw HTML,
 * but a description can be edited by hand via the API too, so this doesn't
 * trust the source.
 */
export function MarkdownContent({ text, className }: { text: string; className?: string }) {
  if (!text) return null;
  return (
    <div className={className ? `md-content ${className}` : 'md-content'}>
      <ReactMarkdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeSanitize]}>
        {text}
      </ReactMarkdown>
    </div>
  );
}
