// SPDX-License-Identifier: Apache-2.0
//
// A lightweight contentEditable rich-text editor with a small formatting
// toolbar (bold / italic / underline / lists / link). Emits HTML via onChange.

import { useEffect, useRef } from 'react';
import { Icon } from '@/components/common/Icon';

interface RichTextEditorProps {
  value: string; // HTML
  onChange: (html: string) => void;
  placeholder?: string;
  className?: string;
}

export function RichTextEditor({ value, onChange, placeholder, className }: RichTextEditorProps) {
  const ref = useRef<HTMLDivElement>(null);

  // Sync external value into the DOM only when the editor isn't focused, so
  // programmatic updates (AI draft, reset) apply without disrupting typing.
  useEffect(() => {
    const el = ref.current;
    if (el && document.activeElement !== el && el.innerHTML !== value) {
      el.innerHTML = value;
    }
  }, [value]);

  const emit = () => onChange(ref.current?.innerHTML ?? '');

  // execCommand is deprecated but remains the simplest broadly-supported way to
  // format contentEditable selections; sufficient for composing mail.
  const exec = (command: string, arg?: string) => {
    document.execCommand(command, false, arg);
    ref.current?.focus();
    emit();
  };

  const addLink = () => {
    const url = window.prompt('リンク先 URL');
    if (url) exec('createLink', url);
  };

  return (
    <div className={className}>
      <div className="flex flex-none items-center gap-0.5 border-b border-line px-2 py-1">
        <ToolBtn icon="format_bold" label="太字" onClick={() => exec('bold')} />
        <ToolBtn icon="format_italic" label="斜体" onClick={() => exec('italic')} />
        <ToolBtn icon="format_underlined" label="下線" onClick={() => exec('underline')} />
        <div className="mx-1 h-5 w-px bg-line" />
        <ToolBtn icon="format_list_bulleted" label="箇条書き" onClick={() => exec('insertUnorderedList')} />
        <ToolBtn icon="format_list_numbered" label="番号付き" onClick={() => exec('insertOrderedList')} />
        <ToolBtn icon="link" label="リンク" onClick={addLink} />
      </div>
      <div className="relative min-h-0 flex-1 overflow-y-auto">
        {!value && placeholder && (
          <div className="pointer-events-none absolute left-4 top-3 text-[14.5px] text-content-sub/70">
            {placeholder}
          </div>
        )}
        <div
          ref={ref}
          contentEditable
          role="textbox"
          aria-multiline="true"
          onInput={emit}
          className="email-html min-h-full px-4 py-3 text-[14.5px] leading-relaxed text-content outline-none [&_a]:text-brand-strong [&_a]:underline"
          suppressContentEditableWarning
        />
      </div>
    </div>
  );
}

function ToolBtn({ icon, label, onClick }: { icon: string; label: string; onClick: () => void }) {
  return (
    <button
      type="button"
      aria-label={label}
      title={label}
      // Prevent the editor from losing its selection on mousedown.
      onMouseDown={(e) => e.preventDefault()}
      onClick={onClick}
      className="flex h-8 w-8 items-center justify-center rounded-md text-content-sub transition-colors hover:bg-hover hover:text-content"
    >
      <Icon name={icon} size={19} />
    </button>
  );
}
