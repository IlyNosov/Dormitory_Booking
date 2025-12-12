import React from "react";

export function Pattern256() {
    const c = "#f59e0b";
    return (
        <g>
            <defs>
                <radialGradient id="p256_r" cx="50%" cy="50%" r="50%">
                    <stop stopColor={c} />
                    <stop offset="1" stopColor={c} stopOpacity="0" />
                </radialGradient>
            </defs>
            <circle cx="245" cy="18" r="100" fill="url(#p256_r)" opacity=".28" />
            <path d="M-10 115 A60 60 0 0 0 110 115" stroke={c} strokeOpacity=".45" strokeWidth="10" fill="none" />
            <path d="M36 24 l10 10 m0 -10 l-10 10" stroke={c} strokeWidth="4" opacity=".35" />
            <path d="M86 28 l12 20 l-24 0 z" fill={c} opacity=".24" />
            <circle cx="62" cy="95" r="9" fill={c} opacity=".28" />
        </g>
    );
}
