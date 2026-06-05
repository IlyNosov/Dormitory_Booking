import React from "react";

export function Pattern3812() {
    const c = "#c1ac6c"; // gold — coworking 3к
    return (
        <g>
            <defs>
                <radialGradient id="p3812_r" cx="75%" cy="25%" r="55%">
                    <stop stopColor={c} />
                    <stop offset="1" stopColor={c} stopOpacity="0" />
                </radialGradient>
            </defs>
            <circle cx="255" cy="18" r="95" fill="url(#p3812_r)" opacity=".22" />
            {/* Diagonal lines — distinct from 2812 */}
            <line x1="185" y1="10" x2="295" y2="130" stroke={c} strokeWidth="14" strokeOpacity=".10" />
            <line x1="205" y1="5"  x2="300" y2="110" stroke={c} strokeWidth="7"  strokeOpacity=".10" />
            {/* Hexagon-like shapes */}
            <path d="M220 40 l12 7 l0 14 l-12 7 l-12 -7 l0 -14 z" fill={c} opacity=".18" />
            <path d="M248 55 l8 4.5 l0 9 l-8 4.5 l-8 -4.5 l0 -9 z" fill={c} opacity=".13" />
            {/* Bottom line */}
            <rect x="18" y="18" width="12" height="12" rx="6" fill={c} opacity=".18" />
        </g>
    );
}
