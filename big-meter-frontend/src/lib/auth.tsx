import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";

export type AuthUser = {
  status: string;
  status_desc: string;
  username: string;
  firstname: string;
  lastname: string;
  costcenter: string;
  ba: string;
  part: string;
  area: string;
  job_name: string;
  level: string;
  div_name: string;
  dep_name: string;
  org_name: string;
  email: string;
  position: string;
};

type StoredAuth = {
  user: AuthUser;
  loginTime: number;
};

type AuthContextValue = {
  user: AuthUser | null;
  hydrated: boolean;
  login: (user: AuthUser) => void;
  logout: () => void;
};

const STORAGE_KEY = "big-meter.auth.user";
const SESSION_TIMEOUT_MS = 8 * 60 * 60 * 1000; // 8 hours
const AuthContext = createContext<AuthContextValue | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [hydrated, setHydrated] = useState(false);

  // Check if session is still valid
  const checkSession = () => {
    if (typeof window === "undefined") return;

    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) return;

    try {
      const parsed = JSON.parse(raw) as StoredAuth;
      const now = Date.now();
      const elapsed = now - parsed.loginTime;

      if (elapsed > SESSION_TIMEOUT_MS) {
        // Session expired
        window.localStorage.removeItem(STORAGE_KEY);
        setUser(null);
        if (import.meta.env.DEV) {
          console.log("Session expired, logged out automatically");
        }
      }
    } catch (error) {
      if (import.meta.env.DEV) {
        console.error("Failed to check session", error);
      }
    }
  };

  // Initial hydration: load user from storage
  useEffect(() => {
    if (typeof window === "undefined") {
      setHydrated(true);
      return;
    }
    try {
      const raw = window.localStorage.getItem(STORAGE_KEY);
      if (raw) {
        const parsed = JSON.parse(raw) as StoredAuth;
        const now = Date.now();
        const elapsed = now - parsed.loginTime;

        // Check if session has expired
        if (elapsed > SESSION_TIMEOUT_MS) {
          // Session expired, clear storage
          window.localStorage.removeItem(STORAGE_KEY);
          if (import.meta.env.DEV) {
            console.log("Session expired, logged out automatically");
          }
        } else {
          // Session still valid
          setUser(parsed.user);
        }
      }
    } catch (error) {
      // Only log in development
      if (import.meta.env.DEV) {
        console.error("Failed to parse auth user", error);
      }
      window.localStorage.removeItem(STORAGE_KEY);
    } finally {
      setHydrated(true);
    }
  }, []);

  // Periodic session check (every 5 minutes)
  useEffect(() => {
    if (!user) return;

    const interval = setInterval(checkSession, 5 * 60 * 1000);
    return () => clearInterval(interval);
  }, [user]);

  const login = (next: AuthUser) => {
    setUser(next);
    if (typeof window !== "undefined") {
      const storedAuth: StoredAuth = {
        user: next,
        loginTime: Date.now(),
      };
      window.localStorage.setItem(STORAGE_KEY, JSON.stringify(storedAuth));
    }
  };

  const logout = () => {
    setUser(null);
    if (typeof window !== "undefined") {
      window.localStorage.removeItem(STORAGE_KEY);
    }
  };

  const value = useMemo(
    () => ({ user, hydrated, login, logout }),
    [user, hydrated],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return context;
}
