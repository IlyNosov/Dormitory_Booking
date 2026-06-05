import React, { useLayoutEffect, useState } from "react";
import { Portal } from "./Portal";

export function Popover({
  open,
  onClose,
  anchor,
  children,
  align = "left",
}: {
  open: boolean;
  onClose: () => void;
  anchor: HTMLElement | null;
  children: React.ReactNode;
  align?: "left" | "right";
}) {
  const [pos, setPos] = useState<{ top: number; left: number }>({ top: 0, left: 0 });

  useLayoutEffect(() => {
    if (!open || !anchor) return;

    const calc = () => {
      const r = anchor.getBoundingClientRect();
      const top = r.bottom + 8;
      const left = align === "right" ? r.right : r.left;
      setPos({ top, left });
    };

    calc();
    window.addEventListener("scroll", calc, true);
    window.addEventListener("resize", calc);
    return () => {
      window.removeEventListener("scroll", calc, true);
      window.removeEventListener("resize", calc);
    };
  }, [open, anchor, align]);

  if (!open || !anchor) return null;

  return (
    <Portal>
      <div
        className="fixed inset-0 z-50"
        onMouseDown={onClose}
        onTouchStart={onClose}
      >
        <div
          className="fixed"
          style={{
            top: pos.top,
            left: pos.left,
            transform: align === "right" ? "translateX(-100%)" : undefined,
          }}
          onMouseDown={(e) => e.stopPropagation()}
          onTouchStart={(e) => e.stopPropagation()}
        >
          <div
            className="rounded-2xl border p-3 shadow-xl"
            style={{
              background: "var(--d-surface)",
              borderColor: "var(--d-border)",
              color: "var(--d-text)",
              boxShadow: "0 8px 32px rgba(0,0,0,0.12)",
            }}
          >
            {children}
          </div>
        </div>
      </div>
    </Portal>
  );
}
