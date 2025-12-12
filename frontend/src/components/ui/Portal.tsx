import React, { useEffect, useRef } from "react";
import { createPortal } from "react-dom";

export function Portal({ children }: { children: React.ReactNode }) {
    const elRef = useRef<HTMLDivElement | null>(null);
    if (!elRef.current) elRef.current = document.createElement("div");

    useEffect(() => {
        document.body.appendChild(elRef.current!);
        return () => {
            document.body.removeChild(elRef.current!);
        };
    }, []);

    return elRef.current ? createPortal(children, elRef.current) : null;
}
