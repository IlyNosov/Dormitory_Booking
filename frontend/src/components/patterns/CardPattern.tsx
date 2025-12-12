import React from "react";
import type { Room } from "../../types/bookings";
import { cn } from "../../utils/cn";
import { Pattern21 } from "./Pattern21";
import { Pattern132 } from "./Pattern132";
import { Pattern256 } from "./Pattern256";

export function CardPattern({ room, active }: { room: Room; active?: boolean }) {
    return (
        <svg
            className={cn(
                "pattern absolute inset-0 z-0 pointer-events-none transition-all duration-600 ease-[cubic-bezier(.2,.65,.3,1)]",
                active ? "opacity-90 scale-[1.015]" : "opacity-70 group-hover:opacity-95 group-hover:scale-105",
            )}
            viewBox="0 0 300 160"
            preserveAspectRatio="none"
        >
            {room === 21 ? <Pattern21 /> : room === 132 ? <Pattern132 /> : <Pattern256 />}
        </svg>
    );
}

