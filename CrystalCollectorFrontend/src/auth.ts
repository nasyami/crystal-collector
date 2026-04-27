export interface UserFromToken {
  email: string;
  sub: string;
}

export function getCurrentUserFromToken(): UserFromToken | null {
  try {
    const token = localStorage.getItem('accessToken');
    if (!token) return null;
    const parts = token.split('.');
    if (parts.length < 2) return null;
    const payload = JSON.parse(atob(parts[1])) as Record<string, unknown>;
    if (!payload?.sub) return null;
    return {
      email:
        (payload.email as string) ||
        (payload.username as string) ||
        'Unknown',
      sub: payload.sub as string,
    };
  } catch {
    return null;
  }
}
