"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { motion } from "framer-motion";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import styled, { keyframes, type DefaultTheme } from "styled-components";
import { z } from "zod";
import {
  Button,
  Heading,
  Stack,
  Text,
  TextField,
} from "@cia-da-vacina/design-system";
import { fadeUp } from "@/lib/motion";
import { useAuth } from "@/providers/auth-provider";
import { ApiError } from "@/lib/errors";

const schema = z.object({
  email: z.email("Email inválido"),
  password: z.string().min(1, "Informe a senha"),
});

type FormValues = z.infer<typeof schema>;

function isDark(theme: DefaultTheme): boolean {
  return String(theme?.name ?? "")
    .toLowerCase()
    .includes("dark");
}

const drift = keyframes`
  from { transform: translate3d(0, 0, 0) scale(1); }
  to { transform: translate3d(12px, -18px, 0) scale(1.04); }
`;

const Shell = styled.div`
  position: relative;
  min-height: 100dvh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: max(24px, env(safe-area-inset-top))
    max(16px, env(safe-area-inset-right))
    max(24px, env(safe-area-inset-bottom))
    max(16px, env(safe-area-inset-left));
  overflow: hidden;
  background: ${({ theme }) =>
    isDark(theme)
      ? `
        radial-gradient(1200px 600px at 12% -10%, rgba(15, 107, 76, 0.22), transparent 55%),
        radial-gradient(900px 500px at 96% 110%, rgba(15, 107, 76, 0.12), transparent 50%),
        linear-gradient(165deg, ${theme.colors["bg.canvas"]} 0%, #0d1612 42%, ${theme.colors["bg.canvas"]} 100%)
      `
      : `
        radial-gradient(1200px 600px at 12% -10%, rgba(15, 107, 76, 0.2), transparent 55%),
        radial-gradient(900px 500px at 96% 110%, rgba(15, 107, 76, 0.14), transparent 50%),
        linear-gradient(165deg, #f1f6f3 0%, #e4efe9 42%, #f7faf8 100%)
      `};
`;

const Blob = styled.div<{
  $top?: string;
  $left?: string;
  $right?: string;
  $bottom?: string;
  $size: string;
  $delay?: string;
}>`
  position: absolute;
  top: ${({ $top }) => $top ?? "auto"};
  left: ${({ $left }) => $left ?? "auto"};
  right: ${({ $right }) => $right ?? "auto"};
  bottom: ${({ $bottom }) => $bottom ?? "auto"};
  width: ${({ $size }) => $size};
  height: ${({ $size }) => $size};
  border-radius: 50%;
  background: ${({ theme }) =>
    isDark(theme)
      ? "radial-gradient(circle, rgba(31, 127, 88, 0.28) 0%, rgba(31, 127, 88, 0) 70%)"
      : "radial-gradient(circle, rgba(15, 107, 76, 0.18) 0%, rgba(15, 107, 76, 0) 70%)"};
  filter: blur(2px);
  animation: ${drift} 14s ease-in-out infinite alternate;
  animation-delay: ${({ $delay }) => $delay ?? "0s"};
  pointer-events: none;
`;

const Grain = styled.div`
  position: absolute;
  inset: 0;
  opacity: ${({ theme }) => (isDark(theme) ? 0.18 : 0.32)};
  pointer-events: none;
  background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.85' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)' opacity='0.45'/%3E%3C/svg%3E");
  mix-blend-mode: soft-light;
`;

const Stage = styled.div`
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 400px;
  display: flex;
  flex-direction: column;
  gap: 28px;
`;

const BrandBlock = styled.div`
  text-align: center;
`;

const LogoWrap = styled.div`
  display: flex;
  justify-content: center;
  margin-bottom: 18px;
`;

const LogoMark = styled.div`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 88px;
  height: 88px;
  padding: 14px;
  border-radius: 22px;
  background: ${({ theme }) =>
    isDark(theme)
      ? theme.colors["bg.surface.raised"]
      : "linear-gradient(145deg, rgba(255, 255, 255, 0.95), rgba(212, 236, 228, 0.88))"};
  border: 1px solid
    ${({ theme }) =>
      isDark(theme) ? theme.colors["border.default"] : "transparent"};
  box-shadow: ${({ theme }) =>
    isDark(theme)
      ? "0 12px 32px rgba(0, 0, 0, 0.35)"
      : "0 12px 32px rgba(8, 43, 35, 0.1), inset 0 1px 0 rgba(255, 255, 255, 0.85)"};

  img {
    width: 60px;
    height: 60px;
    object-fit: contain;
  }
`;

const BrandTitle = styled(Heading)`
  font-size: clamp(1.85rem, 4.5vw, 2.35rem);
  margin: 0;
  color: ${({ theme }) => theme.colors["text.brand"]};
`;

const Panel = styled(motion.div)`
  padding: 28px 24px;
  border-radius: 18px;
  background: ${({ theme }) =>
    isDark(theme)
      ? theme.colors["bg.surface"]
      : "rgba(255, 255, 255, 0.82)"};
  border: 1px solid
    ${({ theme }) =>
      isDark(theme)
        ? theme.colors["border.default"]
        : "rgba(255, 255, 255, 0.72)"};
  backdrop-filter: blur(14px);
  box-shadow: ${({ theme }) =>
    isDark(theme)
      ? "0 18px 40px rgba(0, 0, 0, 0.4)"
      : "0 18px 40px rgba(8, 43, 35, 0.08)"};
`;

export default function LoginPage() {
  const { login, user, loading } = useAuth();
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      email: "",
      password: "",
    },
  });

  useEffect(() => {
    if (!loading && user) router.replace("/inbox");
  }, [loading, user, router]);

  const onSubmit = handleSubmit(async (values) => {
    setError(null);
    try {
      await login(values.email, values.password);
      router.replace("/inbox");
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Falha no login");
    }
  });

  return (
    <Shell>
      <Blob $top="-8%" $left="-6%" $size="340px" />
      <Blob $bottom="-12%" $right="-8%" $size="420px" $delay="2s" />
      <Blob $top="40%" $right="18%" $size="180px" $delay="4s" />
      <Grain />

      <Stage>
        <motion.div {...fadeUp}>
          <BrandBlock>
            <LogoWrap>
              <LogoMark>
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img src="/favicon.svg" alt="Cia da Vacina" />
              </LogoMark>
            </LogoWrap>
            <BrandTitle as="h1">Cia da Vacina</BrandTitle>
          </BrandBlock>
        </motion.div>

        <Panel
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.32, delay: 0.08, ease: [0.22, 1, 0.36, 1] }}
        >
          <form onSubmit={onSubmit}>
            <Stack gap={3}>
              <Heading as="h2" fontSize="lg">
                Entrar
              </Heading>
              <TextField
                label="Email"
                type="email"
                autoComplete="username"
                error={errors.email?.message}
                {...register("email")}
              />
              <TextField
                label="Senha"
                type="password"
                autoComplete="current-password"
                error={errors.password?.message}
                {...register("password")}
              />
              {error ? (
                <Text color="text.danger" fontSize="sm">
                  {error}
                </Text>
              ) : null}
              <Button type="submit" fullWidth size="lg" disabled={isSubmitting}>
                {isSubmitting ? "Entrando…" : "Entrar no CRM"}
              </Button>
            </Stack>
          </form>
        </Panel>
      </Stage>
    </Shell>
  );
}
