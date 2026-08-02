"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { AnimatePresence, motion } from "framer-motion";
import { useRouter } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import styled from "styled-components";
import {
  Badge,
  Button,
  DataList,
  DataListRow,
  EmptyState,
  Flex,
  PageHeader,
  Spinner,
  Stack,
  Text,
  TextField,
  useToast,
} from "@cia-da-vacina/design-system";
import { fadeUp, staggerContainer, staggerItem } from "@/lib/motion";
import { useAuth } from "@/providers/auth-provider";
import { usersService, unitsService } from "@/services";
import {
  USER_ROLE_LABELS,
  type Unit,
  type User,
  type UserRole,
} from "@/domain";

const EMPTY_UNITS: Unit[] = [];
const EMPTY_USERS: User[] = [];

const Layout = styled.div`
  display: grid;
  grid-template-columns: 1fr;
  gap: 16px;
  width: 100%;

  @media (min-width: 960px) {
    grid-template-columns: minmax(0, 1.1fr) minmax(280px, 0.9fr);
    align-items: start;
  }
`;

const Panel = styled.section`
  padding: 16px 18px;
  background: ${({ theme }) => theme.colors["bg.surface"]};
  border: 1px solid ${({ theme }) => theme.colors["border.default"]};
  border-radius: ${({ theme }) => theme.radii.md};
  min-width: 0;
`;

const Overlay = styled(motion.div)`
  position: fixed;
  inset: 0;
  z-index: 80;
  background: rgba(20, 30, 26, 0.35);
  backdrop-filter: blur(4px);
  display: flex;
  justify-content: flex-end;
`;

const Drawer = styled(motion.aside)`
  width: min(440px, 100vw);
  height: 100%;
  background: ${({ theme }) => theme.colors["bg.surface"]};
  border-left: 1px solid ${({ theme }) => theme.colors["border.default"]};
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
  overflow-y: auto;
`;

const CheckRow = styled.label`
  display: flex;
  align-items: flex-start;
  gap: 8px;
  font-size: ${({ theme }) => theme.fontSizes.sm};
  color: ${({ theme }) => theme.colors["text.primary"]};
  cursor: pointer;
`;

type Draft = {
  id: string | null;
  name: string;
  code: string;
  city: string;
  address: string;
  district: string;
  complement: string;
  reference: string;
  timezone: string;
  active: boolean;
};

function blankDraft(): Draft {
  return {
    id: null,
    name: "",
    code: "",
    city: "",
    address: "",
    district: "",
    complement: "",
    reference: "",
    timezone: "America/Sao_Paulo",
    active: true,
  };
}

function fromUnit(u: Unit): Draft {
  return {
    id: u.id,
    name: u.name,
    code: u.code,
    city: u.city,
    address: u.address,
    district: u.district ?? "",
    complement: u.complement ?? "",
    reference: u.reference ?? "",
    timezone: u.timezone,
    active: u.active,
  };
}

function formatAddress(u: Unit): string {
  const parts = [u.address];
  if (u.complement) parts.push(u.complement);
  if (u.district) parts.push(u.district);
  parts.push(u.city);
  return parts.join(", ");
}

function slugifyCode(name: string): string {
  return name
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "");
}

function usersForUnit(users: User[], unitId: string): User[] {
  return users.filter((u) => {
    if (!u.active) return false;
    if (u.role === "admin") return true;
    return u.unit_ids?.includes(unitId);
  });
}

export default function UnitsPage() {
  const router = useRouter();
  const qc = useQueryClient();
  const toast = useToast();
  const { user: me, loading: authLoading } = useAuth();
  const isAdmin = me?.role === "admin";

  const unitsQuery = useQuery({
    queryKey: ["units"],
    queryFn: () => unitsService.list(),
    enabled: isAdmin,
  });
  const usersQuery = useQuery({
    queryKey: ["users"],
    queryFn: () => usersService.list(),
    enabled: isAdmin,
  });

  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [draft, setDraft] = useState<Draft | null>(null);
  const [formError, setFormError] = useState<string | null>(null);
  const [accessIds, setAccessIds] = useState<string[]>([]);

  useEffect(() => {
    if (!authLoading && me && !isAdmin) {
      router.replace("/inbox");
    }
  }, [authLoading, me, isAdmin, router]);

  const units = unitsQuery.data?.items ?? EMPTY_UNITS;
  const users = usersQuery.data?.items ?? EMPTY_USERS;

  const selected = useMemo(
    () => units.find((u) => u.id === selectedId) ?? null,
    [units, selectedId],
  );

  const accessUsers = useMemo(() => {
    if (!selected) return [];
    return usersForUnit(users, selected.id).sort((a, b) =>
      a.name.localeCompare(b.name, "pt-BR"),
    );
  }, [users, selected]);

  useEffect(() => {
    if (!selected) {
      setAccessIds((prev) => (prev.length === 0 ? prev : []));
      return;
    }
    const next = users
      .filter((u) => u.role !== "admin" && u.unit_ids?.includes(selected.id))
      .map((u) => u.id);
    setAccessIds((prev) => {
      if (
        prev.length === next.length &&
        prev.every((id, i) => id === next[i])
      ) {
        return prev;
      }
      return next;
    });
  }, [selected, users]);

  const saveUnit = useMutation({
    mutationFn: async (d: Draft) => {
      const payload = {
        name: d.name.trim(),
        code: d.code.trim().toLowerCase(),
        city: d.city.trim(),
        address: d.address.trim(),
        district: d.district.trim() || null,
        complement: d.complement.trim() || null,
        reference: d.reference.trim() || null,
        timezone: d.timezone.trim() || "America/Sao_Paulo",
        active: d.active,
      };
      if (d.id) return unitsService.update(d.id, payload);
      return unitsService.create(payload);
    },
    onSuccess: async (unit) => {
      toast.push(draft?.id ? "Unidade atualizada" : "Unidade criada", "success");
      setDraft(null);
      setFormError(null);
      setSelectedId(unit.id);
      await qc.invalidateQueries({ queryKey: ["units"] });
    },
    onError: () => toast.push("Não foi possível salvar a unidade", "danger"),
  });

  const saveAccess = useMutation({
    mutationFn: async () => {
      if (!selected) return;
      const assignable = users.filter((u) => u.role !== "admin");
      await Promise.all(
        assignable.map((u) => {
          const has = accessIds.includes(u.id);
          const current = new Set(u.unit_ids ?? []);
          if (has) current.add(selected.id);
          else current.delete(selected.id);
          const next = [...current];
          const prev = u.unit_ids ?? [];
          if (
            next.length === prev.length &&
            next.every((id) => prev.includes(id))
          ) {
            return Promise.resolve();
          }
          return usersService.setUnits(u.id, next);
        }),
      );
    },
    onSuccess: async () => {
      toast.push("Acessos atualizados", "success");
      await qc.invalidateQueries({ queryKey: ["users"] });
    },
    onError: () => toast.push("Não foi possível atualizar acessos", "danger"),
  });

  function openCreate() {
    setFormError(null);
    setDraft(blankDraft());
  }

  function openEdit(u: Unit) {
    setFormError(null);
    setDraft(fromUnit(u));
  }

  function submitDraft() {
    if (!draft) return;
    if (!draft.name.trim() || !draft.code.trim() || !draft.city.trim() || !draft.address.trim()) {
      setFormError("Nome, código, cidade e endereço são obrigatórios.");
      return;
    }
    saveUnit.mutate(draft);
  }

  if (authLoading || !me) {
    return (
      <Flex gap={2} alignItems="center">
        <Spinner />
        <Text muted>Carregando…</Text>
      </Flex>
    );
  }

  if (!isAdmin) return null;

  return (
    <motion.div {...fadeUp} style={{ width: "100%" }}>
      <Stack gap={3} width="100%">
        <PageHeader
          title="Unidades"
          description="Clínicas da Cia da Vacina: endereço, status e quem acessa o CRM de cada uma."
          actions={
            <Button type="button" onClick={openCreate}>
              Nova unidade
            </Button>
          }
        />

        {unitsQuery.isLoading || usersQuery.isLoading ? (
          <Flex gap={2} alignItems="center">
            <Spinner />
            <Text muted>Carregando unidades…</Text>
          </Flex>
        ) : units.length === 0 ? (
          <EmptyState
            title="Nenhuma unidade"
            description="Cadastre a primeira clínica para começar."
          />
        ) : (
          <Layout>
            <Panel>
              <Text fontWeight="semibold" fontSize="sm" style={{ marginBottom: 12 }}>
                {units.length} unidade(s)
              </Text>
              <motion.div variants={staggerContainer} initial="initial" animate="animate">
                <DataList>
                  {units.map((u) => {
                    const count = usersForUnit(users, u.id).length;
                    return (
                      <motion.div key={u.id} variants={staggerItem}>
                        <DataListRow
                          interactive
                          onClick={() => setSelectedId(u.id)}
                          leading={
                            <Stack gap={1} style={{ minWidth: 0 }}>
                              <Flex gap={2} alignItems="center" flexWrap="wrap">
                                <Text fontWeight="semibold">{u.name}</Text>
                                <Badge tone={u.active ? "success" : "neutral"}>
                                  {u.active ? "Ativa" : "Inativa"}
                                </Badge>
                                {selectedId === u.id ? (
                                  <Badge tone="brand">Selecionada</Badge>
                                ) : null}
                              </Flex>
                              <Text fontSize="sm" muted>
                                {formatAddress(u)}
                              </Text>
                              <Text fontSize="xs" muted>
                                Código {u.code}, {count} acesso(s) no CRM
                              </Text>
                            </Stack>
                          }
                          trailing={
                            <Button
                              type="button"
                              size="sm"
                              variant="ghost"
                              onClick={(e) => {
                                e.stopPropagation();
                                openEdit(u);
                              }}
                            >
                              Editar
                            </Button>
                          }
                        />
                      </motion.div>
                    );
                  })}
                </DataList>
              </motion.div>
            </Panel>

            <Panel>
              {!selected ? (
                <Text muted fontSize="sm">
                  Selecione uma unidade para ver quem tem acesso ao CRM.
                </Text>
              ) : (
                <Stack gap={3}>
                  <div>
                    <Text fontWeight="semibold">{selected.name}</Text>
                    <Text fontSize="sm" muted style={{ marginTop: 4 }}>
                      {formatAddress(selected)}
                    </Text>
                    {selected.reference ? (
                      <Text fontSize="xs" muted style={{ marginTop: 4 }}>
                        Ref.: {selected.reference}
                      </Text>
                    ) : null}
                  </div>

                  <div>
                    <Text fontWeight="medium" fontSize="sm" style={{ marginBottom: 8 }}>
                      Acessos no CRM
                    </Text>
                    <Text fontSize="xs" muted style={{ marginBottom: 10 }}>
                      Admins entram em todas as unidades. Demais papéis só nas marcadas.
                    </Text>

                    <Stack gap={2}>
                      {users
                        .filter((u) => u.role === "admin")
                        .map((u) => (
                          <Flex key={u.id} gap={2} alignItems="center">
                            <Badge tone="brand">Admin</Badge>
                            <Text fontSize="sm">
                              {u.name} (acesso total)
                            </Text>
                          </Flex>
                        ))}

                      {users
                        .filter((u) => u.role !== "admin")
                        .map((u) => (
                          <CheckRow key={u.id}>
                            <input
                              type="checkbox"
                              checked={accessIds.includes(u.id)}
                              onChange={(e) => {
                                setAccessIds((prev) =>
                                  e.target.checked
                                    ? [...prev, u.id]
                                    : prev.filter((id) => id !== u.id),
                                );
                              }}
                            />
                            <span>
                              <strong>{u.name}</strong>
                              <Text as="span" muted fontSize="xs">
                                {" "}
                                ({USER_ROLE_LABELS[u.role as UserRole]}, {u.email})
                              </Text>
                            </span>
                          </CheckRow>
                        ))}
                    </Stack>

                    <Flex gap={2} style={{ marginTop: 14 }}>
                      <Button
                        type="button"
                        onClick={() => saveAccess.mutate()}
                        disabled={saveAccess.isPending}
                      >
                        {saveAccess.isPending ? "Salvando…" : "Salvar acessos"}
                      </Button>
                      <Button
                        type="button"
                        variant="secondary"
                        onClick={() => openEdit(selected)}
                      >
                        Editar unidade
                      </Button>
                    </Flex>
                  </div>

                  {accessUsers.length > 0 ? (
                    <div>
                      <Text fontSize="xs" muted style={{ marginBottom: 6 }}>
                        Com acesso agora ({accessUsers.length})
                      </Text>
                      <Flex gap={1} flexWrap="wrap">
                        {accessUsers.map((u) => (
                          <Badge key={u.id} tone={u.role === "admin" ? "brand" : "neutral"}>
                            {u.name}
                          </Badge>
                        ))}
                      </Flex>
                    </div>
                  ) : null}
                </Stack>
              )}
            </Panel>
          </Layout>
        )}
      </Stack>

      <AnimatePresence>
        {draft ? (
          <Overlay
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={() => !saveUnit.isPending && setDraft(null)}
          >
            <Drawer
              initial={{ x: 40, opacity: 0 }}
              animate={{ x: 0, opacity: 1 }}
              exit={{ x: 24, opacity: 0 }}
              transition={{ duration: 0.22, ease: [0.22, 1, 0.36, 1] }}
              onClick={(e) => e.stopPropagation()}
            >
              <Flex justifyContent="space-between" alignItems="center">
                <Text fontWeight="semibold">
                  {draft.id ? "Editar unidade" : "Nova unidade"}
                </Text>
                <Button
                  type="button"
                  size="sm"
                  variant="ghost"
                  onClick={() => setDraft(null)}
                  disabled={saveUnit.isPending}
                >
                  Fechar
                </Button>
              </Flex>

              <TextField
                label="Nome"
                value={draft.name}
                onChange={(e) => {
                  const name = e.target.value;
                  setDraft((d) =>
                    d
                      ? {
                          ...d,
                          name,
                          code: d.id ? d.code : slugifyCode(name),
                          city: d.id || d.city ? d.city : name,
                        }
                      : d,
                  );
                }}
                placeholder="Ex.: Marau"
                disabled={saveUnit.isPending}
              />
              <TextField
                label="Código"
                value={draft.code}
                onChange={(e) =>
                  setDraft((d) => (d ? { ...d, code: e.target.value } : d))
                }
                placeholder="marau"
                disabled={saveUnit.isPending}
              />
              <TextField
                label="Cidade"
                value={draft.city}
                onChange={(e) =>
                  setDraft((d) => (d ? { ...d, city: e.target.value } : d))
                }
                disabled={saveUnit.isPending}
              />
              <TextField
                label="Endereço"
                value={draft.address}
                onChange={(e) =>
                  setDraft((d) => (d ? { ...d, address: e.target.value } : d))
                }
                placeholder="Rua, número"
                disabled={saveUnit.isPending}
              />
              <TextField
                label="Bairro"
                value={draft.district}
                onChange={(e) =>
                  setDraft((d) => (d ? { ...d, district: e.target.value } : d))
                }
                disabled={saveUnit.isPending}
              />
              <TextField
                label="Complemento"
                value={draft.complement}
                onChange={(e) =>
                  setDraft((d) =>
                    d ? { ...d, complement: e.target.value } : d,
                  )
                }
                placeholder="Sala, andar…"
                disabled={saveUnit.isPending}
              />
              <TextField
                label="Referência"
                value={draft.reference}
                onChange={(e) =>
                  setDraft((d) =>
                    d ? { ...d, reference: e.target.value } : d,
                  )
                }
                placeholder="Ponto de referência"
                disabled={saveUnit.isPending}
              />
              <TextField
                label="Fuso"
                value={draft.timezone}
                onChange={(e) =>
                  setDraft((d) => (d ? { ...d, timezone: e.target.value } : d))
                }
                disabled={saveUnit.isPending}
              />

              <CheckRow>
                <input
                  type="checkbox"
                  checked={draft.active}
                  disabled={saveUnit.isPending}
                  onChange={(e) =>
                    setDraft((d) =>
                      d ? { ...d, active: e.target.checked } : d,
                    )
                  }
                />
                Unidade ativa
              </CheckRow>

              {formError ? <Text fontSize="sm">{formError}</Text> : null}

              <Flex gap={2} justifyContent="flex-end" style={{ marginTop: "auto" }}>
                <Button
                  type="button"
                  variant="secondary"
                  onClick={() => setDraft(null)}
                  disabled={saveUnit.isPending}
                >
                  Cancelar
                </Button>
                <Button
                  type="button"
                  onClick={submitDraft}
                  disabled={saveUnit.isPending}
                >
                  {saveUnit.isPending ? "Salvando…" : "Salvar"}
                </Button>
              </Flex>
            </Drawer>
          </Overlay>
        ) : null}
      </AnimatePresence>
    </motion.div>
  );
}
