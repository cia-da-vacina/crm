import { AuthGuard } from "@/components/auth-guard";
import { AppNav } from "@/components/app-nav";

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <AuthGuard>
      <AppNav>{children}</AppNav>
    </AuthGuard>
  );
}
