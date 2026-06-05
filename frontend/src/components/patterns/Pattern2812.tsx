import React from "react";

export function Pattern2812() {
    const c = "#c0915f"; // amber — coworking 2к
    return (
        <g>
            <defs>
                <radialGradient id="p2812_r" cx="70%" cy="30%" r="55%">
                    <stop stopColor={c} />
                    <stop offset="1" stopColor={c} stopOpacity="0" />
                </radialGradient>
            </defs>
            <circle cx="250" cy="22" r="95" fill="url(#p2812_r)" opacity=".22" />
            {/* Grid dots — coworking feel */}
            {[0,1,2,3].map(row => [0,1,2,3,4].map(col => (
                <circle key={`${row}-${col}`}
                    cx={200 + col * 18} cy={28 + row * 18}
                    r="2.5" fill={c} opacity=".18" />
            )))}
            {/* Corner accent */}
            <rect x="205" y="110" width="70" height="6" rx="3" fill={c} opacity=".20" />
            <rect x="215" y="120" width="50" height="4" rx="2" fill={c} opacity=".14" />
            {/* Small square */}
            <rect x="18" y="18" width="16" height="16" rx="4" fill={c} opacity=".18" />
            <rect x="40" y="22" width="10" height="10" rx="3" fill={c} opacity=".13" />
        </g>
    );
}
