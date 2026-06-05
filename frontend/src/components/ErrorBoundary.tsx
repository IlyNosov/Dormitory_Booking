import React from "react";
import { AlertTriangle, Bug, X, Send } from "lucide-react";
import { sendBugReport } from "../api/bugs";

interface State {
  hasError: boolean;
  error: Error | null;
  showReport: boolean;
  description: string;
  sent: boolean;
}

export class ErrorBoundary extends React.Component<{ children: React.ReactNode }, State> {
  constructor(props: { children: React.ReactNode }) {
    super(props);
    this.state = { hasError: false, error: null, showReport: false, description: "", sent: false };
  }

  static getDerivedStateFromError(error: Error): Partial<State> {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, info: React.ErrorInfo) {
    console.error("[ErrorBoundary] Uncaught error:", error, info.componentStack);
  }

  handleSendReport = async () => {
    const { error, description } = this.state;
    try {
      await sendBugReport({
        message: error?.message ?? "unknown",
        description: description || "(не указано)",
        url: window.location.href,
        userAgent: navigator.userAgent,
      });
      this.setState({ sent: true });
    } catch (e) {
      console.error("[BugReport] failed to send:", e);
      // Fallback to mailto
      const details = [
        `URL: ${window.location.href}`,
        `User agent: ${navigator.userAgent}`,
        `Time: ${new Date().toISOString()}`,
        `Description: ${description || "(не указано)"}`,
        `Error: ${error?.message ?? "unknown"}`,
        `Stack:\n${error?.stack ?? ""}`,
      ].join("\n");
      const subject = encodeURIComponent("Ошибка dubki.fun");
      const body = encodeURIComponent(details);
      window.location.href = `mailto:silverdenanz@gmail.com?subject=${subject}&body=${body}`;
    }
  };

  render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen flex items-center justify-center p-6" style={{ background: "var(--d-bg)" }}>
          <div
            className="max-w-md w-full rounded-2xl border p-8 text-center space-y-4"
            style={{ background: "var(--d-surface)", borderColor: "var(--d-border)" }}
          >
            <AlertTriangle className="h-10 w-10 mx-auto" style={{ color: "var(--d-error)" }} />
            <div className="text-lg font-semibold" style={{ color: "var(--d-text)" }}>
              Что-то пошло не так
            </div>
            <p className="text-sm" style={{ color: "var(--d-text-secondary)" }}>
              Произошла непредвиденная ошибка. Попробуйте обновить страницу.
            </p>
            {this.state.error?.message && (
              <p className="text-xs font-mono px-3 py-2 rounded-lg text-left" style={{
                background: "var(--d-panel)",
                color: "var(--d-error)",
                wordBreak: "break-all",
              }}>
                {this.state.error.message}
              </p>
            )}
            <div className="flex gap-3 justify-center flex-wrap">
              <button
                className="btn btn-primary"
                onClick={() => window.location.reload()}
              >
                Обновить страницу
              </button>
              <button
                className="btn"
                onClick={() => this.setState({ showReport: true, sent: false })}
              >
                <Bug className="h-4 w-4" />
                Сообщить об ошибке
              </button>
            </div>
          </div>

          {/* Bug report modal */}
          {this.state.showReport && (
            <div
              className="fixed inset-0 z-50 flex items-center justify-center p-4"
              style={{ background: "rgba(44,53,32,0.5)", backdropFilter: "blur(4px)" }}
            >
              <div
                className="relative w-full max-w-sm rounded-2xl border p-6 space-y-4 shadow-2xl"
                style={{ background: "var(--d-surface)", borderColor: "var(--d-border)" }}
              >
                <button
                  className="icon-btn absolute top-4 right-4"
                  onClick={() => this.setState({ showReport: false, sent: false })}
                  aria-label="Закрыть"
                >
                  <X className="h-4 w-4" />
                </button>

                <div className="flex items-center gap-2">
                  <Bug className="h-5 w-5" style={{ color: "var(--d-primary)" }} />
                  <span className="font-semibold" style={{ color: "var(--d-text)" }}>
                    Сообщить об ошибке
                  </span>
                </div>

                {this.state.sent ? (
                  <p className="text-sm text-center py-4" style={{ color: "var(--d-primary-h)" }}>
                    Отчёт отправлен, спасибо!
                  </p>
                ) : (
                  <>
                    <p className="text-sm" style={{ color: "var(--d-text-secondary)" }}>
                      Опишите, что произошло. Мы получим детали ошибки автоматически.
                    </p>

                    <textarea
                      className="field resize-none"
                      rows={4}
                      placeholder="Что вы делали, когда возникла ошибка?"
                      value={this.state.description}
                      onChange={(e) => this.setState({ description: e.target.value })}
                      autoFocus
                    />

                    <div className="flex gap-2">
                      <button
                        className="btn btn-primary flex-1"
                        onClick={this.handleSendReport}
                      >
                        <Send className="h-4 w-4" />
                        Отправить
                      </button>
                      <button
                        className="btn"
                        onClick={() => this.setState({ showReport: false })}
                      >
                        Отмена
                      </button>
                    </div>
                  </>
                )}
              </div>
            </div>
          )}
        </div>
      );
    }
    return this.props.children;
  }
}

// Global reportBug function that can be called from anywhere
export async function reportBug(message: string) {
  try {
    await sendBugReport({
      message,
      description: "(не указано)",
      url: window.location.href,
      userAgent: navigator.userAgent,
    });
  } catch (e) {
    console.error("[BugReport] failed:", e);
    const subject = encodeURIComponent("Ошибка dubki.fun");
    const body = encodeURIComponent(`URL: ${window.location.href}\nError: ${message}`);
    window.location.href = `mailto:silverdenanz@gmail.com?subject=${subject}&body=${body}`;
  }
}
