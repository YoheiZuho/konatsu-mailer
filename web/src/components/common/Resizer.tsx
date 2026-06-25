// SPDX-License-Identifier: Apache-2.0
//
// A vertical drag handle for resizing a left-anchored pane (design §9.4).

import { useCallback, useRef } from 'react';

interface ResizerProps {
  /** Current width of the pane immediately to the left. */
  width: number;
  /** Called with the new width while dragging. */
  onWidthChange: (w: number) => void;
}

export function Resizer({ width, onWidthChange }: ResizerProps) {
  const startX = useRef(0);
  const startWidth = useRef(0);

  const onMouseMove = useCallback(
    (e: MouseEvent) => {
      onWidthChange(startWidth.current + (e.clientX - startX.current));
    },
    [onWidthChange],
  );

  const onMouseUp = useCallback(() => {
    document.removeEventListener('mousemove', onMouseMove);
    document.removeEventListener('mouseup', onMouseUp);
    document.body.style.cursor = '';
    document.body.style.userSelect = '';
  }, [onMouseMove]);

  const onMouseDown = (e: React.MouseEvent) => {
    e.preventDefault();
    startX.current = e.clientX;
    startWidth.current = width;
    document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';
    document.addEventListener('mousemove', onMouseMove);
    document.addEventListener('mouseup', onMouseUp);
  };

  return (
    <div
      role="separator"
      aria-orientation="vertical"
      onMouseDown={onMouseDown}
      className="group relative w-px flex-none cursor-col-resize bg-line"
      title="ドラッグして幅を調整"
    >
      {/* Wider invisible hit area + accent on hover. */}
      <div className="absolute inset-y-0 -left-1 -right-1 z-10 transition-colors group-hover:bg-brand/40" />
    </div>
  );
}
