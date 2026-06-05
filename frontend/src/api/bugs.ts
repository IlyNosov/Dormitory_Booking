export async function sendBugReport(report: {
  description: string;
  url: string;
  userAgent: string;
}): Promise<void> {
  const res = await fetch("/api/bugs", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(report),
  });
  if (!res.ok) throw new Error("send failed");
}
