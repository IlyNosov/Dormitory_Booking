import React, { useEffect } from "react";
import { X } from "lucide-react";
import { Portal } from "./Portal";

export function Modal({
                          open,
                          onClose,
                          children,
                      }: {
    open: boolean;
    onClose: () => void;
    children: React.ReactNode;
}) {
    useEffect(() => {
        document.documentElement.classList.toggle("modal-open", open);
    }, [open]);

    if (!open) return null;

    return (
        <Portal>
            <div className="fixed inset-0 z-[100]">
                <div
                    className="absolute inset-0 bg-black/40 backdrop-blur-[2px] opacity-100 transition-opacity"
                    onClick={onClose}
                />
                <div className="absolute inset-0 flex items-center justify-center p-4">
                    <div className="relative w-full max-w-lg rounded-2xl border border-zinc-200/70 bg-white/95 p-4 shadow-2xl dark:border-[color:var(--d-border)] dark:bg-[color:var(--d-card)] dark:text-zinc-100">
                        <button className="absolute right-3 top-3 icon-btn" onClick={onClose} aria-label="Закрыть">
                            <X className="h-5 w-5" />
                        </button>
                        {children}
                    </div>
                </div>
            </div>
        </Portal>
    );
}
