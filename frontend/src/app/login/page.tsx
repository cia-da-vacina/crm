"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import styled, { keyframes } from "styled-components";
import { z } from "zod";
import {
  Button,
  Heading,
  Stack,
  Text,
  TextField,
} from "@cia-da-vacina/design-system";
import { SyringeIcon } from "@cia-da-vacina/icon-system";
import { useAuth } from "@/contexts/auth-context";
import { ApiError } from "@/lib/api";

const schema = z.object({
  email: z.email("Email inválido"),
  password: z.string().min(1, "Informe a senha"),
});

type FormValues = z.infer<typeof schema>;

const rise = keyframes`
  from {
    opacity: 0;
    transform: translateY(14px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
`;

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
  background:
    radial-gradient(1200px 600px at 12% -10%, rgba(15, 107, 76, 0.18), transparent 55%),
    radial-gradient(900px 500px at 96% 110%, rgba(15, 107, 76, 0.12), transparent 50%),
    linear-gradient(165deg, #f3f7f5 0%, #e7f0eb 42%, #f7faf8 100%);
`;

const Blob = styled.div<{ $top?: string; $left?: string; $right?: string; $bottom?: string; $size: string; $delay?: string }>`
  position: absolute;
  top: ${({ $top }) => $top ?? "auto"};
  left: ${({ $left }) => $left ?? "auto"};
  right: ${({ $right }) => $right ?? "auto"};
  bottom: ${({ $bottom }) => $bottom ?? "auto"};
  width: ${({ $size }) => $size};
  height: ${({ $size }) => $size};
  border-radius: 50%;
  background: radial-gradient(circle, rgba(15, 107, 76, 0.16) 0%, rgba(15, 107, 76, 0) 70%);
  filter: blur(2px);
  animation: ${drift} 14s ease-in-out infinite alternate;
  animation-delay: ${({ $delay }) => $delay ?? "0s"};
  pointer-events: none;
`;

const Grain = styled.div`
  position: absolute;
  inset: 0;
  opacity: 0.35;
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
  animation: ${rise} 520ms ease-out both;
`;

const Mark = styled.div`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 52px;
  height: 52px;
  margin-bottom: 16px;
  border-radius: 16px;
  color: ${({ theme }) => theme.colors["text.brand"]};
  background: linear-gradient(
    145deg,
    rgba(255, 255, 255, 0.92),
    rgba(212, 236, 228, 0.85)
  );
  box-shadow: 0 10px 28px rgba(8, 43, 35, 0.08);
`;

const BrandTitle = styled.h1`
  margin: 0;
  font-family: ${({ theme }) => theme.fonts.display};
  font-size: clamp(2rem, 5vw, 2.6rem);
  font-weight: ${({ theme }) => theme.fontWeights.semibold};
  letter-spacing: ${({ theme }) => theme.letterSpacings.tight};
  line-height: 1.1;
  color: ${({ theme }) => theme.colors["text.brand"]};
`;

const Support = styled.p`
  margin: 10px 0 0;
  font-family: ${({ theme }) => theme.fonts.body};
  font-size: ${({ theme }) => theme.fontSizes.sm};
  color: ${({ theme }) => theme.colors["text.secondary"]};
  line-height: 1.45;
`;

const Panel = styled.div`
  padding: 28px 24px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.78);
  border: 1px solid rgba(255, 255, 255, 0.7);
  backdrop-filter: blur(14px);
  box-shadow: 0 18px 40px rgba(8, 43, 35, 0.08);
  animation: ${rise} 620ms ease-out 80ms both;
`;

const Foot = styled.p`
  margin: 0;
  text-align: center;
  font-size: ${({ theme }) => theme.fontSizes.xs};
  color: ${({ theme }) => theme.colors["text.muted"]};
  animation: ${rise} 700ms ease-out 140ms both;
`;

const Demo = styled.code`
  font-family: ${({ theme }) => theme.fonts.mono};
  font-size: 11px;
  color: ${({ theme }) => theme.colors["text.secondary"]};
  background: ${({ theme }) => theme.colors["bg.brand.subtle"]};
  padding: 2px 6px;
  border-radius: 4px;
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
      email: "admin@ciadavacina.com.br",
      password: "admin123",
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
        <BrandBlock>
          <Mark aria-hidden>
            <SyringeIcon size="lg" fill="text.brand" />
          </Mark>
          <BrandTitle>Cia da Vacina</BrandTitle>
          <Support>
            CRM de atendimento WhatsApp das cinco unidades —
            triagem com IA e handoff humano.
          </Support>
        </BrandBlock>

        <Panel>
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

        <Foot>
          Demo · <Demo>admin@ciadavacina.com.br</Demo> / <Demo>admin123</Demo>
        </Foot>
      </Stage>
    </Shell>
  );
}
