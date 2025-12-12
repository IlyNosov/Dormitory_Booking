import React from "react";

export function Pattern132() {
    const c = "#34d399";
    return (
        <g>
            <defs>
                <radialGradient id="p132_r" cx="50%" cy="50%" r="50%">
                    <stop stopColor={c} />
                    <stop offset="1" stopColor={c} stopOpacity="0" />
                </radialGradient>
            </defs>
            <circle cx="260" cy="25" r="110" fill="url(#p132_r)" opacity=".28" />
            <path
                d="M20 110 C40 90, 60 70, 80 90 S120 120, 150 100"
                stroke={c}
                strokeOpacity=".5"
                strokeWidth="8"
                fill="none"
            />
            <rect x="12" y="18" width="18" height="18" rx="4" fill={c} opacity=".26" />
            <circle cx="100" cy="35" r="8" fill={c} opacity=".22" />
            <path d="M170 100 c 12 -20 28 -20 40 0" stroke={c} strokeWidth="8" opacity=".3" />
        </g>
    );
}
