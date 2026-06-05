import React from "react";
import { ArrowLeft, Clock, CalendarCheck } from "lucide-react";
import { getRoomColor } from "../utils/rooms";
import { RoomIcon } from "../components/RoomIcon";

interface RoomCard {
  numbers: number[];
  buildings: string[];
  roomLabel: string;
  description: string;
  hours: string;
}

const ROOMS: RoomCard[] = [
  {
    numbers: [21],
    buildings: ["1 корпус"],
    roomLabel: "21",
    description: "Уютная досуговая с мягкими креслами и телевизором. Идеальное место для кинопоказов, просмотров аниме и настольных игр.",
    hours: "06:00–23:00 (Пт–Сб до 01:00)",
  },
  {
    numbers: [256],
    buildings: ["1 корпус"],
    roomLabel: "256",
    description: "Просторная досуговая с фортепиано и проектором. Идеально подходит для публичных мероприятий и творческих вечеров.",
    hours: "06:00–23:00 (Пт–Сб до 01:00)",
  },
  {
    numbers: [132],
    buildings: ["2 корпус"],
    roomLabel: "132",
    description: "Комфортная досуговая с диванами и проектором. Пользуется особой популярностью у жителей второго корпуса.",
    hours: "06:00–22:00 (Пт–Сб до 23:00)",
  },
  {
    numbers: [2812, 3812],
    buildings: ["2 корпус", "3 корпус"],
    roomLabel: "812",
    description: "Коворкинги во втором и третьем корпусах. Признаться честно, автор сервиса до конца не представляет, что это за комнаты, но они есть.",
    hours: "06:00–23:00 (Пт–Сб до 01:00)",
  },
];

// SVG illustration placeholder per room — simple geometric pattern
function RoomIllustration({ number, hex }: { number: number; hex: string }) {
  const seed = number % 5;
  return (
    <svg
      viewBox="0 0 280 120"
      xmlns="http://www.w3.org/2000/svg"
      className="w-full h-full"
      aria-hidden="true"
    >
      <rect width="280" height="120" fill={hex + "18"} rx="12" />
      {/* Floor */}
      <rect x="20" y="85" width="240" height="6" rx="3" fill={hex + "30"} />
      {/* Couch or furniture silhouette */}
      {seed < 3 && (
        <>
          <rect x="50" y="65" width="80" height="22" rx="6" fill={hex + "55"} />
          <rect x="44" y="60" width="14" height="28" rx="4" fill={hex + "44"} />
          <rect x="122" y="60" width="14" height="28" rx="4" fill={hex + "44"} />
          <rect x="50" y="73" width="80" height="8" rx="3" fill={hex + "33"} />
        </>
      )}
      {/* TV / projector screen */}
      {seed % 2 === 0 && (
        <>
          <rect x="160" y="45" width="80" height="48" rx="4" fill={hex + "22"} stroke={hex + "44"} strokeWidth="1.5" />
          <rect x="194" y="93" width="12" height="6" rx="2" fill={hex + "33"} />
          <rect x="168" y="52" width="64" height="34" rx="2" fill={hex + "18"} />
        </>
      )}
      {/* Piano for room 256 */}
      {number === 256 && (
        <>
          <rect x="160" y="58" width="90" height="30" rx="4" fill={hex + "33"} />
          <rect x="165" y="54" width="80" height="8" rx="2" fill={hex + "55"} />
          {[0,1,2,3,4,5].map(i => (
            <rect key={i} x={167 + i * 12} y="55" width="8" height="6" rx="1" fill="#FDFCFA" fillOpacity={0.8} />
          ))}
        </>
      )}
      {/* Plants decoration */}
      <ellipse cx="230" cy="82" rx="10" ry="12" fill={hex + "44"} />
      <rect x="226" y="82" width="8" height="8" rx="2" fill={hex + "33"} />
      <ellipse cx="245" cy="80" rx="7" ry="9" fill={hex + "33"} />
      {/* Window */}
      <rect x="20" y="20" width="50" height="40" rx="3" fill={hex + "22"} stroke={hex + "33"} strokeWidth="1.5" />
      <line x1="45" y1="20" x2="45" y2="60" stroke={hex + "44"} strokeWidth="1" />
      <line x1="20" y1="40" x2="70" y2="40" stroke={hex + "44"} strokeWidth="1" />
      {/* Sun through window */}
      <circle cx="35" cy="30" r="6" fill="#C49A5E" fillOpacity={0.4} />
    </svg>
  );
}

interface Props {
  onBack: () => void;
  onBook: () => void;
}

export function RoomsPage({ onBack, onBook }: Props) {
  return (
    <div className="min-h-screen" style={{ background: "var(--d-bg)" }}>
      {/* Page header */}
      <div
        className="sticky top-0 z-10 border-b px-4 py-3"
        style={{
          background: "var(--d-bg-soft)",
          borderColor: "var(--d-border)",
          backdropFilter: "blur(10px)",
          WebkitBackdropFilter: "blur(10px)",
        }}
      >
        <div className="mx-auto max-w-4xl flex items-center gap-3">
          <button
            onClick={onBack}
            className="icon-btn"
            aria-label="Назад"
          >
            <ArrowLeft className="h-5 w-5" />
          </button>
          <div>
            <h1 className="text-xl font-semibold font-display italic leading-tight" style={{ color: "var(--d-text)" }}>
              Досуговые комнаты
            </h1>
            <div className="text-xs" style={{ color: "var(--d-text-muted)" }}>
              Дубки, ВШЭ — 4 комнаты
            </div>
          </div>
        </div>
      </div>

      <main className="mx-auto max-w-4xl px-4 py-8 space-y-6">
        <p className="text-sm leading-relaxed" style={{ color: "var(--d-text-secondary)" }}>
          Общежитие Дубки располагает досуговыми комнатами в трёх корпусах.
          Для бронирования выберите нужную комнату и время.
        </p>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
          {ROOMS.map((room) => {
            const rc = getRoomColor(room.numbers[0]);
            return (
              <div
                key={room.numbers.join("-")}
                className="rounded-2xl border overflow-hidden transition-all duration-200 hover:shadow-md"
                style={{
                  background: "var(--d-surface)",
                  borderColor: "var(--d-border)",
                  borderTop: `3px solid ${rc.hex}`,
                }}
              >
                {/* Illustration */}
                <div
                  className="w-full overflow-hidden"
                  style={{ height: 120, background: rc.hex + "0c" }}
                >
                  <RoomIllustration number={room.numbers[0]} hex={rc.hex} />
                </div>

                <div className="p-4 space-y-3">
                  <div className="flex items-start justify-between gap-2">
                    <div>
                      <div className="flex items-center gap-2 font-semibold text-base" style={{ color: "var(--d-text)" }}>
                        <span style={{ color: rc.hex }}>
                          <RoomIcon roomNumber={room.numbers[0]} size={20} />
                        </span>
                        Комната {room.roomLabel}
                      </div>
                      {/* Building badges — support multiple */}
                      <div className="flex flex-wrap gap-1.5 mt-1">
                        {room.buildings.map((b, i) => (
                          <span
                            key={b}
                            className="inline-block text-xs px-2 py-0.5 rounded-full font-medium"
                            style={{
                              background: (i === 0 ? rc.hex : getRoomColor(room.numbers[i] ?? room.numbers[0]).hex) + "20",
                              color: i === 0 ? rc.hex : getRoomColor(room.numbers[i] ?? room.numbers[0]).hex,
                            }}
                          >
                            {b}
                          </span>
                        ))}
                      </div>
                    </div>
                  </div>

                  <p className="text-sm leading-relaxed" style={{ color: "var(--d-text-secondary)" }}>
                    {room.description}
                  </p>

                  <div className="flex items-center gap-1.5 text-xs" style={{ color: "var(--d-text-muted)" }}>
                    <Clock className="h-3.5 w-3.5 shrink-0" />
                    {room.hours}
                  </div>

                  <button
                    className="btn btn-primary w-full justify-center text-sm"
                    onClick={onBook}
                  >
                    <CalendarCheck className="h-4 w-4" />
                    Забронировать
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      </main>
    </div>
  );
}
