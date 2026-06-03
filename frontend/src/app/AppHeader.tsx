import {LayoutList, LogOut, RefreshCw, Shield, Table} from "lucide-react";
import type {ViewMode} from "../types/bookings";
import {cn} from "../utils/cn";

export function AppHeader({
                              loading,
                              view,
                              isAdmin,
                              onToggleAdd,
                              onRefresh,
                              onToggleView,
                              onAdminClick,
                              onAdminLogout,
                              onRulesClick,
                          }: {
    loading: boolean;
    view: ViewMode;
    isAdmin: boolean;
    onToggleAdd: () => void;
    onRefresh: () => void;
    onToggleView: () => void;
    onAdminClick: () => void;
    onAdminLogout: () => void;
    onRulesClick: () => void;
}) {
    return (
        <header className="sticky top-0 z-20 border-b border-zinc-700 bg-[#1e1f22]/80 backdrop-blur">
            <div className="mx-auto max-w-6xl px-4 py-3 flex items-center justify-between">
                <div>
                    <h1 className="text-xl font-semibold">Бронирование досуговых</h1>
                    <div className="text-xs text-zinc-400">Общежитие • комнаты 21, 132, 256</div>
                </div>

                <div className="flex items-center gap-2">
                    <button onClick={onRulesClick} className="btn">
                        Правила
                    </button>

                    <button onClick={onToggleAdd} className="btn">+ Бронь</button>

                    <button onClick={onRefresh} className="btn">
                        <RefreshCw className={cn("h-4 w-4 mr-2", loading && "animate-spin")}/>
                        Обновить
                    </button>

                    {/*
                    <button onClick={onToggleView} className="btn">
                        {view === "cards" ? <Table className="h-4 w-4 mr-2"/> : <LayoutList className="h-4 w-4 mr-2"/>}
                        {view === "cards" ? "Таблица" : "Списком"}
                    </button>
                    */}

                    {!isAdmin ? (
                        <button onClick={onAdminClick} className="btn btn-primary">
                            <Shield className="h-4 w-4 mr-2"/>
                            Админка
                        </button>
                    ) : (
                        <button onClick={onAdminLogout} className="btn">
                            <LogOut className="h-4 w-4 mr-2"/>
                            Админ: выйти
                        </button>
                    )}
                </div>
            </div>
        </header>
    );
}
