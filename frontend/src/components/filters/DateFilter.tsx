import React, { useRef, useState } from "react";
import { Popover } from "../ui/Popover";
import { DatePicker } from "../pickers/DatePicker";

export function DateFilter({
                               value,
                               onChange,
                               label = "С даты",
                           }: {
    value: string;
    onChange: (v: string) => void;
    label?: string;
}) {
    const [open, setOpen] = useState(false);
    const btnRef = useRef<HTMLButtonElement | null>(null);

    return (
        <>
            <div className="flex flex-col text-sm">
                <span className="lbl mb-1">{label}</span>
                <button
                    ref={btnRef}
                    className="field flex items-center justify-between w-[220px]"
                    onClick={() => setOpen(true)}
                    type="button"
                >
                    {new Date(value).toLocaleDateString("ru-RU", { day: "2-digit", month: "long", weekday: "long" })}
                    <span className="opacity-60">▾</span>
                </button>
            </div>

            <Popover open={open} onClose={() => setOpen(false)} anchor={btnRef.current}>
                <DatePicker
                    value={value}
                    onChange={(v) => {
                        onChange(v);
                        setOpen(false);
                    }}
                />
            </Popover>
        </>
    );
}
