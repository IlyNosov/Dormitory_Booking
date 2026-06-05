import React, { useEffect, useState } from "react";
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
    // Keep in DOM during the exit animation, then unmount.
    const [mounted, setMounted] = useState(open);
    const [visible, setVisible] = useState(false);

    useEffect(() => {
        if (open) {
            setMounted(true);
            // Double rAF ensures the element is painted before the transition starts.
            const id = requestAnimationFrame(() =>
                requestAnimationFrame(() => setVisible(true))
            );
            return () => cancelAnimationFrame(id);
        } else {
            setVisible(false);
            const t = setTimeout(() => setMounted(false), 280);
            return () => clearTimeout(t);
        }
    }, [open]);

    useEffect(() => {
        if (!mounted) return;
        document.body.style.overflow = "hidden";
        return () => { document.body.style.overflow = ""; };
    }, [mounted]);

    if (!mounted) return null;

    return (
        <Portal>
            <div className="fixed inset-0 z-[100]">
                {/* Backdrop */}
                <div
                    className="absolute inset-0 modal-backdrop"
                    data-visible={visible}
                    style={{ background: "rgba(38,101,41,0.18)", backdropFilter: "blur(3px)" }}
                    onClick={onClose}
                />

                {/* Sheet container */}
                <div className="absolute inset-0 flex items-end sm:items-center justify-center p-0 sm:p-4">
                    <div
                        className="modal-sheet relative w-full sm:max-w-lg rounded-t-2xl sm:rounded-2xl shadow-2xl flex flex-col overflow-hidden"
                        data-visible={visible}
                        style={{
                            background: "var(--d-surface)",
                            border: "1px solid var(--d-border)",
                            color: "var(--d-text)",
                            maxHeight: "90dvh",
                        }}
                    >
                        <button
                            className="icon-btn absolute top-4 right-4 z-10 shrink-0"
                            onClick={onClose}
                            aria-label="Закрыть"
                        >
                            <X className="h-5 w-5" />
                        </button>
                        <div className="overflow-y-auto px-5 pt-4 pb-6 pr-12">
                            {children}
                        </div>
                    </div>
                </div>
            </div>
        </Portal>
    );
}
