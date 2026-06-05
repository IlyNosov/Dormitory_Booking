import React from "react";

export function RoomIcon({ roomNumber, size = 24 }: { roomNumber: number; size?: number }) {
  // Piano keys (no stand) for room 256 — pattern: 2 black, gap, 3 black
  if (roomNumber === 256) return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      {/* White keys body */}
      <rect x="1" y="6" width="22" height="13" rx="1.5"/>
      {/* White key dividers */}
      <line x1="4.2"  y1="6" x2="4.2"  y2="19" strokeOpacity="0.5" />
      <line x1="7.3"  y1="6" x2="7.3"  y2="19" strokeOpacity="0.5" />
      <line x1="10.6" y1="6" x2="10.6" y2="19" strokeOpacity="0.5" />
      <line x1="13.8" y1="6" x2="13.8" y2="19" strokeOpacity="0.5" />
      <line x1="17"   y1="6" x2="17"   y2="19" strokeOpacity="0.5" />
      <line x1="20.2" y1="6" x2="20.2" y2="19" strokeOpacity="0.5" />
      {/* Black keys — group of 2 */}
      <rect x="2.8" y="6" width="2.4" height="7" rx="0.5" fill="currentColor" stroke="none"/>
      <rect x="6"   y="6" width="2.4" height="7" rx="0.5" fill="currentColor" stroke="none"/>
      {/* Black keys — group of 3 */}
      <rect x="11.3" y="6" width="2.4" height="7" rx="0.5" fill="currentColor" stroke="none"/>
      <rect x="14.5" y="6" width="2.4" height="7" rx="0.5" fill="currentColor" stroke="none"/>
      <rect x="17.7" y="6" width="2.4" height="7" rx="0.5" fill="currentColor" stroke="none"/>
    </svg>
  );

  // TV for room 21
  if (roomNumber === 21) return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <rect x="2" y="4" width="20" height="14" rx="2"/>
      <line x1="8" y1="22" x2="16" y2="22"/>
      <line x1="12" y1="18" x2="12" y2="22"/>
    </svg>
  );

  // Couch/sofa for room 132
  if (roomNumber === 132) return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <path d="M3 10a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2v4H3v-4z"/>
      <path d="M1 10h2v6H1z"/>
      <path d="M21 10h2v6h-2z"/>
      <path d="M3 14h18v2H3z"/>
      <line x1="5" y1="20" x2="5" y2="16"/>
      <line x1="19" y1="20" x2="19" y2="16"/>
    </svg>
  );

  // Generic monitor/desktop for coworking (2812, 3812)
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <rect x="3" y="5" width="18" height="12" rx="1"/>
      <line x1="3" y1="19" x2="21" y2="19"/>
      <line x1="8" y1="19" x2="8" y2="17"/>
      <line x1="16" y1="19" x2="16" y2="17"/>
    </svg>
  );
}
