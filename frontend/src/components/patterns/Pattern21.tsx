import React from "react";

export function Pattern21() {
    const c = "#a78bfa";
    return (
        <g>
            <defs>
                <radialGradient id="p21_r" cx="50%" cy="50%" r="50%">
                    <stop stopColor={c} />
                    <stop offset="1" stopColor={c} stopOpacity="0" />
                </radialGradient>
            </defs>
            <circle cx="260" cy="20" r="110" fill="url(#p21_r)" opacity=".28" />
            <path d="M-20 120 Q60 60 140 120 T300 120" stroke={c} strokeOpacity=".45" strokeWidth="10" fill="none" />
            <rect x="18" y="18" width="22" height="22" rx="6" fill={c} opacity=".26" />
            <path d="M90 20 l14 14 l-14 14 l-14 -14 z" fill={c} opacity=".26" />
            <circle cx="60" cy="95" r="10" fill={c} opacity=".3" />
        </g>
    );
}
