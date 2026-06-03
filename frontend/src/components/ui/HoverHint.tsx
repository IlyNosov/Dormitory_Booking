import React, { useState } from "react";
import { cn } from "../../utils/cn";

export function HoverHint({ hint, children }: { hint: string; children: React.ReactNode }) {
    const [open, setOpen] = useState(false);
    return (
        <span className="relative inline-block" onMouseEnter={() => setOpen(true)} onMouseLeave={() => setOpen(false)}>
      {children}
            <div
                className={cn(
                    "pointer-events-none absolute left-1/2 -translate-x-1/2 top-full mt-1",
                    "z-50 rounded-xl px-2 py-1 text-xs",
                    "bg-zinc-900 text-white shadow-lg transition-all duration-200",
                    open ? "opacity-100 translate-y-0" : "opacity-0 -translate-y-1"
                )}
            >
        {hint}
      </div>
    </span>
    );
}
