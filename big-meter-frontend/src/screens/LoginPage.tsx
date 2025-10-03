import { FormEvent, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth, type AuthUser } from "../lib/auth";

const LOGIN_ENDPOINT = "/auth/login";

type LoginResponse = AuthUser & {
  status: string;
  status_desc: string;
};

export default function LoginPage() {
  const navigate = useNavigate();
  const { user, login, hydrated } = useAuth();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (hydrated && user) {
      navigate("/details", { replace: true });
    }
  }, [hydrated, user, navigate]);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(null);
    setLoading(true);

    try {
      const response = await fetch(LOGIN_ENDPOINT, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ username, pwd: password }),
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }

      const data = (await response.json()) as LoginResponse;

      if (data.status !== "success") {
        const message = data.status_desc?.trim() || "ไม่สามารถเข้าสู่ระบบได้";
        throw new Error(message);
      }

      login(data);
      navigate("/details", { replace: true });
    } catch (fetchError) {
      // Only log in development
      if (import.meta.env.DEV) {
        console.error("Login failed", fetchError);
      }
      setError(
        fetchError instanceof Error
          ? fetchError.message
          : "เกิดข้อผิดพลาดในการเชื่อมต่อ กรุณาลองใหม่อีกครั้ง",
      );
    } finally {
      setLoading(false);
    }
  };

  if (!hydrated) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-slate-100 text-slate-600">
        กำลังโหลด…
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-slate-100 px-4 py-8">
      <div className="mx-auto flex min-h-[calc(100vh-4rem)] max-w-5xl flex-col items-center justify-center">
        <div className="w-full max-w-md rounded-2xl bg-white p-8 shadow-xl shadow-blue-900/5">
          <header className="text-center">
            <p className="text-sm font-medium text-blue-600">ระบบแสดงผลผู้ใช้น้ำรายใหญ่</p>
            <h1 className="mt-2 text-3xl font-bold text-slate-800">
              ลงชื่อเข้าใช้ PWA Intranet
            </h1>
            <p className="mt-2 text-sm text-slate-500">
              ใช้บัญชีผู้ใช้งานเดียวกับระบบ Intranet ของ กปภ.
            </p>
          </header>

          <form className="mt-8 space-y-5" onSubmit={handleSubmit}>
            <div>
              <label htmlFor="username" className="block text-sm font-medium text-slate-600">
                ชื่อผู้ใช้
              </label>
              <input
                id="username"
                name="username"
                type="text"
                autoComplete="username"
                required
                value={username}
                onChange={(event) => setUsername(event.target.value)}
                className="mt-1 w-full rounded-lg border border-slate-300 bg-slate-50 px-3 py-2 text-sm text-slate-800 shadow-sm focus:border-blue-500 focus:bg-white focus:outline-none focus:ring-2 focus:ring-blue-200"
              />
            </div>

            <div>
              <label htmlFor="password" className="block text-sm font-medium text-slate-600">
                รหัสผ่าน
              </label>
              <input
                id="password"
                name="password"
                type="password"
                autoComplete="current-password"
                required
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                className="mt-1 w-full rounded-lg border border-slate-300 bg-slate-50 px-3 py-2 text-sm text-slate-800 shadow-sm focus:border-blue-500 focus:bg-white focus:outline-none focus:ring-2 focus:ring-blue-200"
              />
            </div>

            {error && (
              <div className="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-600">
                {error}
              </div>
            )}

            <button
              type="submit"
              disabled={loading}
              className="flex w-full items-center justify-center rounded-lg bg-blue-600 py-2 text-sm font-semibold text-white transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:bg-blue-300"
            >
              {loading ? "กำลังตรวจสอบ…" : "เข้าสู่ระบบ"}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
