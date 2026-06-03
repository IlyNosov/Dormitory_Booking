import React, { useLayoutEffect, useMemo, useState } from "react";
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

    useLayoutEffect(() => {
        if (!open) return;
        document.documentElement.classList.add("modal-open");
        return () => document.documentElement.classList.remove("modal-open");
    }, [open]);

    if (!open || !anchor) return null;

    return (
        <Portal>
            <div
                className="fixed inset-0 z-50"
                onMouseDown={onClose}
                onWheel={(e) => {
                    e.preventDefault();
                    e.stopPropagation();
                }}
                onTouchMove={(e) => {
                    e.preventDefault();
                    e.stopPropagation();
                }}
            >
                <div
                    className="fixed"
                    style={{
                        top: pos.top,
                        left: pos.left,
                        transform: align === "right" ? "translateX(-100%)" : undefined,
                    }}
                    onMouseDown={(e) => e.stopPropagation()}
                    onWheel={(e) => {
                        e.stopPropagation();
                    }}
                >
                    <div className="rounded-2xl border border-zinc-200/70 bg-white/95 p-3 shadow-xl
                          text-zinc-900
                          dark:border-[color:var(--d-border)] dark:bg-[color:var(--d-card)] dark:text-zinc-100">
                        {children}
                    </div>
                </div>
            </div>
        </Portal>
    );
}
