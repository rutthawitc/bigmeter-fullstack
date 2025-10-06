import { useMemo } from "react";
import { useAuth } from "./auth";

/**
 * Hook to check if the current user is an admin
 * Admin authorization is based on username matching against VITE_ADMIN_USERNAMES env var
 * @returns boolean indicating if current user is admin
 */
export function useIsAdmin(): boolean {
  const { user } = useAuth();

  return useMemo(() => {
    if (!user?.username) return false;

    const adminUsernames =
      import.meta.env.VITE_ADMIN_USERNAMES as string | undefined;

    // Debug logging in development
    if (import.meta.env.DEV) {
      console.log("[useIsAdmin] user.username:", user.username, "type:", typeof user.username);
      console.log("[useIsAdmin] VITE_ADMIN_USERNAMES:", adminUsernames);
    }

    if (!adminUsernames) return false;

    // Parse comma-separated list and trim whitespace
    const allowedUsernames = adminUsernames
      .split(",")
      .map((u) => u.trim())
      .filter((u) => u.length > 0);

    // Convert username to string to ensure type match
    const usernameStr = String(user.username);
    const isAdmin = allowedUsernames.includes(usernameStr);

    if (import.meta.env.DEV) {
      console.log("[useIsAdmin] allowedUsernames:", allowedUsernames);
      console.log("[useIsAdmin] usernameStr:", usernameStr);
      console.log("[useIsAdmin] isAdmin:", isAdmin);
    }

    return isAdmin;
  }, [user?.username]);
}
