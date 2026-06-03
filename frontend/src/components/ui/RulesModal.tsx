import React from "react";
import { Modal } from "./Modal";

export function RulesModal({
                               open,
                               onClose,
                           }: {
    open: boolean;
    onClose: () => void;
}) {
    return (
        <Modal open={open} onClose={onClose}>
            <div className="space-y-4">
                <div className="text-lg font-semibold">Правила бронирования</div>

                <div className="text-sm text-zinc-300 leading-6">
                    <p>
                        События бывают <span className="text-zinc-100 font-medium">публичными</span> и{" "}
                        <span className="text-zinc-100 font-medium">частными</span>. На публичное может прийти любой.
                        От посещения частного рекомендуем воздержаться, если вас на него не приглашали.
                    </p>
                </div>

                <div className="rounded-2xl border border-[color:var(--d-border)] bg-[color:var(--d-panel)] p-4">
                    <div className="text-sm font-semibold text-zinc-100 mb-3">
                        Ограничения для частных посиделок
                    </div>

                    <ol className="space-y-2 text-sm text-zinc-200 leading-6 list-decimal pl-5">
                        <li>
                            Не более <span className="font-medium text-zinc-100">3 часов</span> на одну посиделку
                            для каждой досуговой.
                        </li>
                        <li>
                            Нет частных посиделок в ночь с{" "}
                            <span className="font-medium text-zinc-100">пятницы на субботу</span> и с{" "}
                            <span className="font-medium text-zinc-100">субботы на воскресенье</span> для всех досуговых
                            с <span className="font-medium text-zinc-100">23:00</span> до{" "}
                            <span className="font-medium text-zinc-100">06:00</span>.
                        </li>
                        <li>
                            Не более <span className="font-medium text-zinc-100">3</span> частных посиделок в день и
                            не более <span className="font-medium text-zinc-100">одной</span> в вечернее время
                            (после <span className="font-medium text-zinc-100">18:00</span>) в каждой досуговой.
                        </li>
                    </ol>
                </div>

                <div className="flex justify-end">
                    <button className="btn btn-primary" onClick={onClose}>
                        Понятно
                    </button>
                </div>
            </div>
        </Modal>
    );
}
