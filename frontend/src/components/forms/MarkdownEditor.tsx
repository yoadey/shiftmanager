// WYSIWYG-by-default Markdown editor for event descriptions (V-009), with a
// toggle to the raw Markdown source. Lazy-loaded by callers (see F-009) since
// MDXEditor (built on Lexical) is a comparatively heavy dependency — this
// module is its own chunk, only fetched when an event form actually renders.
import {
  MDXEditor,
  headingsPlugin,
  listsPlugin,
  quotePlugin,
  thematicBreakPlugin,
  linkPlugin,
  linkDialogPlugin,
  markdownShortcutPlugin,
  diffSourcePlugin,
  toolbarPlugin,
  UndoRedo,
  BoldItalicUnderlineToggles,
  ListsToggle,
  CreateLink,
  DiffSourceToggleWrapper,
} from '@mdxeditor/editor';
import '@mdxeditor/editor/style.css';
import './MarkdownEditor.css';

interface MarkdownEditorProps {
  /** Initial content only — MDXEditor is uncontrolled after mount, updates flow out via onChange. */
  markdown: string;
  onChange: (markdown: string) => void;
  placeholder?: string;
}

export default function MarkdownEditor({ markdown, onChange, placeholder }: MarkdownEditorProps) {
  return (
    <div className="md-editor">
      <MDXEditor
        markdown={markdown}
        onChange={onChange}
        placeholder={placeholder}
        contentEditableClassName="md-editor-content"
        plugins={[
          headingsPlugin(),
          listsPlugin(),
          quotePlugin(),
          thematicBreakPlugin(),
          linkPlugin(),
          linkDialogPlugin(),
          markdownShortcutPlugin(),
          diffSourcePlugin({ viewMode: 'rich-text', diffMarkdown: markdown }),
          toolbarPlugin({
            toolbarContents: () => (
              <DiffSourceToggleWrapper>
                <UndoRedo />
                <BoldItalicUnderlineToggles />
                <ListsToggle />
                <CreateLink />
              </DiffSourceToggleWrapper>
            ),
          }),
        ]}
      />
    </div>
  );
}
