import React from "react";
import type { Room } from "../types/bookings";
import { getRoomColor } from "../utils/rooms";

// ── Room 21 — TV / cinema room ───────────────────────────────────────────────
function Art21({ hex }: { hex: string }) {
  return (
    <>
      {/* TV on wall — shifted left so it's centred above the couch */}
      <rect x="2"  y="44" width="72" height="46" rx="3" fill={hex} fillOpacity="0.28" />
      <rect x="6"  y="48" width="64" height="38" rx="2" fill={hex} fillOpacity="0.10" />
      {/* TV stand */}
      <rect x="33" y="90" width="10" height="5"  rx="1" fill={hex} fillOpacity="0.22" />
      <rect x="24" y="95" width="28" height="3"  rx="1" fill={hex} fillOpacity="0.18" />
      {/* Couch */}
      <rect x="2"  y="110" width="72" height="22" rx="5" fill={hex} fillOpacity="0.45" />
      <rect x="0"  y="104" width="13" height="30" rx="4" fill={hex} fillOpacity="0.38" />
      <rect x="61" y="104" width="13" height="30" rx="4" fill={hex} fillOpacity="0.38" />
      <rect x="2"  y="124" width="72" height="9"  rx="3" fill={hex} fillOpacity="0.28" />
      {/* Cushions */}
      <rect x="6"  y="112" width="18" height="11" rx="4" fill={hex} fillOpacity="0.22" />
      <rect x="28" y="112" width="18" height="11" rx="4" fill={hex} fillOpacity="0.22" />
    </>
  );
}

// ── Room 256 — Piano room ────────────────────────────────────────────────────
// Key pattern: бчбчббчбчбчббчб (9 white keys × 9px = 81px, starts x=3, ends x=84)
// White key x-starts: 3,12,21,30,39,48,57,66,75
// Black key x-starts (5px wide, centred on white-key boundaries):
//   C#(10), D#(19), F#(37), G#(46), A#(55), C#2(73)
const WK_X = [3, 12, 21, 30, 39, 48, 57, 66, 75] as const;
const BK_X = [10, 19, 37, 46, 55, 73]             as const;

function Art256({ hex }: { hex: string }) {
  const keyY = 82; // top of white keys
  return (
    <>
      {/* Music note — bigger */}
      <text x="2" y="56" fontSize="46" fill={hex} fillOpacity="0.17">♪</text>
      {/* Keyboard cover / lid edge */}
      <rect x="2" y={keyY - 5} width="82" height="6" rx="2" fill={hex} fillOpacity="0.55" />
      {/* Piano body — taller */}
      <rect x="2" y={keyY + 1} width="82" height="32" rx="3" fill={hex} fillOpacity="0.45" />
      {/* White keys — span full body width */}
      {WK_X.map((x, i) => (
        <rect key={`w${i}`} x={x} y={keyY + 1} width={8} height={18} rx={1} fill="white" fillOpacity="0.58" />
      ))}
      {/* Key dividers */}
      {WK_X.slice(1).map((x, i) => (
        <line key={`d${i}`} x1={x} y1={keyY + 1} x2={x} y2={keyY + 19} stroke={hex} strokeWidth="0.5" strokeOpacity="0.35" />
      ))}
      {/* Black keys */}
      {BK_X.map((x, i) => (
        <rect key={`b${i}`} x={x} y={keyY + 1} width={5} height={11} rx={1} fill={hex} fillOpacity="0.75" />
      ))}
      {/* X-stand legs */}
      <line x1="22" y1="117" x2="62" y2="142" stroke={hex} strokeWidth="3.5" strokeLinecap="round" strokeOpacity="0.45" />
      <line x1="62" y1="117" x2="22" y2="142" stroke={hex} strokeWidth="3.5" strokeLinecap="round" strokeOpacity="0.45" />
      {/* Stand feet */}
      <rect x="10" y="141" width="22" height="3" rx="1.5" fill={hex} fillOpacity="0.38" />
      <rect x="52" y="141" width="22" height="3" rx="1.5" fill={hex} fillOpacity="0.38" />
    </>
  );
}

// ── Room 132 — Couch room (2к) ───────────────────────────────────────────────
function Art132({ hex }: { hex: string }) {
  return (
    <>
      {/* Window — bigger, lower, left edge */}
      <rect x="2"  y="26" width="52" height="44" rx="2" fill={hex} fillOpacity="0.28" />
      <rect x="3"  y="27" width="50" height="42" rx="1" fill={hex} fillOpacity="0.10" />
      <line x1="28" y1="26" x2="28" y2="70" stroke={hex} strokeWidth="1.5" strokeOpacity="0.4" />
      <line x1="2"  y1="48" x2="54" y2="48" stroke={hex} strokeWidth="1.5" strokeOpacity="0.4" />
      {/* Plant */}
      <rect x="57" y="108" width="18" height="14" rx="3" fill={hex} fillOpacity="0.42" />
      <ellipse cx="66" cy="94"  rx="12" ry="16" fill={hex} fillOpacity="0.42" />
      <ellipse cx="55" cy="98"  rx="9"  ry="12" fill={hex} fillOpacity="0.32" />
      <ellipse cx="75" cy="99"  rx="8"  ry="11" fill={hex} fillOpacity="0.32" />
      {/* Couch */}
      <rect x="2"  y="110" width="52" height="22" rx="5" fill={hex} fillOpacity="0.45" />
      <rect x="0"  y="104" width="12" height="30" rx="4" fill={hex} fillOpacity="0.38" />
      <rect x="42" y="104" width="12" height="30" rx="4" fill={hex} fillOpacity="0.38" />
      {/* Cushions */}
      <rect x="5"  y="112" width="16" height="12" rx="4" fill={hex} fillOpacity="0.22" />
      <rect x="24" y="112" width="16" height="12" rx="4" fill={hex} fillOpacity="0.22" />
    </>
  );
}

// ── Room 2812 — Coworking 2к ─────────────────────────────────────────────────
function Art2812({ hex }: { hex: string }) {
  return (
    <>
      {/* Dot grid */}
      {[0,1,2,3].map(row => [0,1,2,3,4].map(col => (
        <circle key={`${row}-${col}`} cx={6 + col * 14} cy={20 + row * 13} r="2.2" fill={hex} fillOpacity="0.28" />
      )))}
      {/* Desk surface */}
      <rect x="2"  y="90" width="82" height="6" rx="2" fill={hex} fillOpacity="0.48" />
      <rect x="8"  y="96" width="6"  height="24" rx="1" fill={hex} fillOpacity="0.32" />
      <rect x="70" y="96" width="6"  height="24" rx="1" fill={hex} fillOpacity="0.32" />
      {/* Laptop */}
      <rect x="18" y="62" width="42" height="28" rx="3" fill={hex} fillOpacity="0.38" />
      <rect x="21" y="65" width="36" height="22" rx="1" fill={hex} fillOpacity="0.15" />
      <rect x="12" y="89" width="54" height="4"  rx="2" fill={hex} fillOpacity="0.42" />
      {/* Coffee mug */}
      <rect x="66" y="70" width="13" height="15" rx="2" fill={hex} fillOpacity="0.35" />
      <path d="M79 74 Q86 74 86 80 Q86 86 79 86" stroke={hex} strokeWidth="2" fill="none" strokeOpacity="0.35" />
      <rect x="66" y="68" width="13" height="4"  rx="1" fill={hex} fillOpacity="0.25" />
    </>
  );
}

// ── Room 3812 — Coworking 3к ─────────────────────────────────────────────────
function Art3812({ hex }: { hex: string }) {
  return (
    <>
      {/* Monitor — lower, no extra circle */}
      <rect x="18" y="24" width="46" height="34" rx="3" fill={hex} fillOpacity="0.38" />
      <rect x="21" y="27" width="40" height="28" rx="1" fill={hex} fillOpacity="0.15" />
      <rect x="36" y="58" width="10" height="8"  rx="1" fill={hex} fillOpacity="0.32" />
      <rect x="26" y="66" width="30" height="5"  rx="2" fill={hex} fillOpacity="0.38" />
      {/* Keyboard */}
      <rect x="14" y="86" width="58" height="8" rx="2" fill={hex} fillOpacity="0.40" />
      {[0,1,2,3,4,5].map(i => (
        <rect key={i} x={16 + i * 8} y={87.5} width={6} height={5} rx={1} fill={hex} fillOpacity="0.18" />
      ))}
      {/* Chair */}
      <rect x="22" y="108" width="30" height="16" rx="4" fill={hex} fillOpacity="0.40" />
      <rect x="26" y="124" width="5"  height="16" rx="1" fill={hex} fillOpacity="0.28" />
      <rect x="43" y="124" width="5"  height="16" rx="1" fill={hex} fillOpacity="0.28" />
      <rect x="20" y="137" width="16" height="5"  rx="2" fill={hex} fillOpacity="0.22" />
      <rect x="38" y="137" width="16" height="5"  rx="2" fill={hex} fillOpacity="0.22" />
    </>
  );
}

// ── CardRoomArt — positioned on right, starts below icon/title row ────────────
export function CardRoomArt({ room }: { room: Room }) {
  const { hex } = getRoomColor(room);

  return (
    <div
      className="absolute right-0 bottom-0 pointer-events-none overflow-hidden"
      style={{
        width: 86,
        top: 36, // starts below title + status icons row
        WebkitMaskImage: "linear-gradient(to right, transparent 0%, black 32%)",
        maskImage: "linear-gradient(to right, transparent 0%, black 32%)",
      }}
    >
      <svg
        className="absolute inset-0 h-full w-full"
        viewBox="0 0 86 160"
        preserveAspectRatio="xMidYMax meet"
        aria-hidden="true"
      >
        {room === 21   && <Art21   hex={hex} />}
        {room === 256  && <Art256  hex={hex} />}
        {room === 132  && <Art132  hex={hex} />}
        {room === 2812 && <Art2812 hex={hex} />}
        {room === 3812 && <Art3812 hex={hex} />}
        {![21,256,132,2812,3812].includes(room) && <Art21 hex={hex} />}
      </svg>
    </div>
  );
}
