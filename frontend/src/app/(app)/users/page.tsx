"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import {
  Badge,
  Box,
  Button,
  EmptyState,
  Flex,
  Heading,
  SelectField,
  Spinner,
  Stack,
  Text,
  TextField,
} from "@cia-da-vacina/design-system";
import { UsersIcon } from "@cia-da-vacina/icon-system";
import { useAuth } from "@/contexts/auth-context";
import { api, ApiError } from "@/lib/api";
import type { User, UserRole } from "@/lib/types";

const ROLE_LABELS: Record<UserRole, string> = {
  admin: "Administrador",
  manager: "Gerente",
  supervisor: "Supervisor",
  agent: "Atendente",
};

const ROLE_OPTIONS = (Object.keys(ROLE_LABELS) as UserRole[]).map((value) => ({
  value,
  label: ROLE_LABELS[value],
}));

type FormState = {
  name: string;
  email: string;
  password: string;
  role: UserRole;
  active: boolean;
  unit_ids: string[];
};

const blankForm = (): FormState => ({
  name: "",
  email: "",
  password: "",
  role: "agent",
  active: true,
  unit_ids: [],
});

export default function UsersPage() {
  const router = useRouter();
  const qc = useQueryClient();
  const { user: me, loading: authLoading } = useAuth();
  const isAdmin = me?.role === "admin";

  const usersQuery = useQuery({
    queryKey: ["users"],
    queryFn: () => api.listUsers(),
    enabled: isAdmin,
  });
  const unitsQuery = useQuery({
    queryKey: ["units"],
    queryFn: () => api.listUnits(),
    enabled: isAdmin,
  });

  const [openForm, setOpenForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState<FormState>(blankForm);
  const [formError, setFormError] = useState<string | null>(null);

  useEffect(() => {
    if (!authLoading && me && !isAdmin) {
      router.replace("/inbox");
    }
  }, [authLoading, me, isAdmin, router]);

  const unitName = useMemo(() => {
    const map = new Map(
      (unitsQuery.data?.items ?? []).map((u) => [u.id, u.name] as const),
    );
    return (id: string) => map.get(id) ?? id;
  }, [unitsQuery.data?.items]);

  function closeForm() {
    setOpenForm(false);
    setEditingId(null);
    setForm(blankForm());
    setFormError(null);
  }

  function startCreate() {
    setEditingId(null);
    setForm(blankForm());
    setFormError(null);
    setOpenForm(true);
  }

  function startEdit(u: User) {
    setEditingId(u.id);
    setForm({
      name: u.name,
      email: u.email,
      password: "",
      role: u.role,
      active: u.active,
      unit_ids: [...(u.unit_ids ?? [])],
    });
    setFormError(null);
    setOpenForm(true);
  }

  function toggleUnit(id: string) {
    setForm((prev) => ({
      ...prev,
      unit_ids: prev.unit_ids.includes(id)
        ? prev.unit_ids.filter((x) => x !== id)
        : [...prev.unit_ids, id],
    }));
  }

  const save = useMutation({
    mutationFn: async () => {
      if (!form.name.trim()) throw new ApiError(400, "bad_request", "Nome obrigatório");
      if (editingId) {
        await api.updateUser(editingId, {
          name: form.name.trim(),
          role: form.role,
          active: form.active,
          ...(form.password ? { password: form.password } : {}),
        });
        await api.setUserUnits(editingId, form.unit_ids);
        return;
      }
      if (!form.email.trim() || form.password.length < 8) {
        throw new ApiError(
          400,
          "bad_request",
          "Email e senha (mín. 8 caracteres) são obrigatórios",
        );
      }
      const created = await api.createUser({
        name: form.name.trim(),
        email: form.email.trim(),
        password: form.password,
        role: form.role,
        unit_ids: form.unit_ids,
      });
      if (form.unit_ids.length) {
        await api.setUserUnits(created.id, form.unit_ids);
      }
    },
    onSuccess: async () => {
      closeForm();
      await qc.invalidateQueries({ queryKey: ["users"] });
    },
    onError: (err) => {
      setFormError(
        err instanceof ApiError ? err.message : "Não foi possível salvar o usuário.",
      );
    },
  });

  const remove = useMutation({
    mutationFn: (id: string) => api.deleteUser(id),
    onSuccess: async (_d, id) => {
      if (editingId === id) closeForm();
      await qc.invalidateQueries({ queryKey: ["users"] });
    },
    onError: (err) => {
      window.alert(
        err instanceof ApiError ? err.message : "Não foi possível remover o usuário.",
      );
    },
  });

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
    <Stack gap={3} width="100%">
      <Flex alignItems="flex-end" justifyContent="space-between" gap={2}>
        <Stack gap={1}>
          <Flex gap={2} alignItems="center">
            <UsersIcon size="lg" fill="text.brand" />
            <Heading as="h1" display>
              Usuários
            </Heading>
          </Flex>
          <Text muted fontSize="sm">
            Contas, papéis e acesso às unidades
          </Text>
        </Stack>
        <Button
          variant={openForm && !editingId ? "ghost" : "primary"}
          size="sm"
          onClick={() => {
            if (openForm && !editingId) closeForm();
            else startCreate();
          }}
        >
          {openForm && !editingId ? "Cancelar" : "Novo usuário"}
        </Button>
      </Flex>

      {openForm ? (
        <Box
          p={3}
          bg="bg.surface"
          borderWidth="hairline"
          borderStyle="solid"
          borderColor="border.default"
          borderRadius="md"
          width="100%"
        >
          <form
            onSubmit={(e) => {
              e.preventDefault();
              save.mutate();
            }}
          >
            <Stack gap={3}>
              <Heading as="h3">
                {editingId ? "Editar usuário" : "Criar conta"}
              </Heading>
              <TextField
                label="Nome"
                value={form.name}
                onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                required
              />
              <TextField
                label="Email"
                type="email"
                value={form.email}
                onChange={(e) => setForm((f) => ({ ...f, email: e.target.value }))}
                disabled={Boolean(editingId)}
                required={!editingId}
              />
              <TextField
                label={editingId ? "Nova senha (opcional)" : "Senha"}
                type="password"
                value={form.password}
                onChange={(e) => setForm((f) => ({ ...f, password: e.target.value }))}
                hint={editingId ? "Deixe em branco para manter." : "Mínimo 8 caracteres."}
                required={!editingId}
              />
              <SelectField
                label="Papel / acesso"
                value={form.role}
                onChange={(e) =>
                  setForm((f) => ({ ...f, role: e.target.value as UserRole }))
                }
                options={ROLE_OPTIONS}
              />
              {editingId ? (
                <label style={{ display: "flex", gap: 8, alignItems: "center" }}>
                  <input
                    type="checkbox"
                    checked={form.active}
                    onChange={(e) =>
                      setForm((f) => ({ ...f, active: e.target.checked }))
                    }
                  />
                  <Text fontSize="sm">Conta ativa (pode entrar no CRM)</Text>
                </label>
              ) : null}

              <Stack gap={1}>
                <Text fontSize="xs" fontWeight="medium" muted>
                  Unidades com acesso
                </Text>
                <Flex gap={2} flexWrap="wrap">
                  {(unitsQuery.data?.items ?? []).map((u) => {
                    const on = form.unit_ids.includes(u.id);
                    return (
                      <Button
                        key={u.id}
                        type="button"
                        size="sm"
                        variant={on ? "primary" : "secondary"}
                        onClick={() => toggleUnit(u.id)}
                      >
                        {u.name}
                      </Button>
                    );
                  })}
                </Flex>
                <Text fontSize="xs" muted>
                  Admin costuma ter acesso a todas; atendentes só às unidades selecionadas.
                </Text>
              </Stack>

              {formError ? <Text color="text.danger">{formError}</Text> : null}

              <Flex gap={2} justifyContent="flex-end">
                <Button type="button" variant="ghost" size="sm" onClick={closeForm}>
                  Cancelar
                </Button>
                <Button type="submit" size="sm" disabled={save.isPending}>
                  {save.isPending
                    ? "Salvando…"
                    : editingId
                      ? "Salvar alterações"
                      : "Criar conta"}
                </Button>
              </Flex>
            </Stack>
          </form>
        </Box>
      ) : null}

      {usersQuery.isLoading ? (
        <Flex gap={2} alignItems="center">
          <Spinner />
          <Text muted>Carregando usuários…</Text>
        </Flex>
      ) : null}

      {!usersQuery.isLoading && (usersQuery.data?.items.length ?? 0) === 0 ? (
        <EmptyState
          title="Nenhum usuário"
          description="Crie a primeira conta com Novo usuário."
          action={
            <Button size="sm" onClick={startCreate}>
              Novo usuário
            </Button>
          }
        />
      ) : null}

      <Stack gap={2} width="100%">
        {(usersQuery.data?.items ?? []).map((u) => (
          <Box
            key={u.id}
            p={3}
            bg="bg.surface"
            borderWidth="hairline"
            borderStyle="solid"
            borderColor={editingId === u.id ? "border.brand" : "border.default"}
            borderRadius="md"
            width="100%"
          >
            <Flex justifyContent="space-between" alignItems="flex-start" gap={2}>
              <Stack gap={1} style={{ flex: 1, minWidth: 0 }}>
                <Flex gap={2} alignItems="center" flexWrap="wrap">
                  <Text fontWeight="semibold">{u.name}</Text>
                  <Badge tone={u.active ? "success" : "neutral"}>
                    {u.active ? "Ativo" : "Inativo"}
                  </Badge>
                  <Badge tone={u.role === "admin" ? "brand" : "info"}>
                    {ROLE_LABELS[u.role]}
                  </Badge>
                </Flex>
                <Text fontSize="sm" muted>
                  {u.email}
                </Text>
                <Flex gap={1} flexWrap="wrap" marginTop={1}>
                  {(u.unit_ids ?? []).length === 0 ? (
                    <Text fontSize="xs" muted>
                      Sem unidades vinculadas
                    </Text>
                  ) : (
                    (u.unit_ids ?? []).map((id) => (
                      <Badge key={id} tone="neutral">
                        {unitName(id)}
                      </Badge>
                    ))
                  )}
                </Flex>
              </Stack>
              <Flex gap={1} flexShrink={0}>
                <Button type="button" variant="ghost" size="sm" onClick={() => startEdit(u)}>
                  Editar
                </Button>
                <Button
                  type="button"
                  variant="danger"
                  size="sm"
                  disabled={remove.isPending || u.id === me.id}
                  onClick={() => {
                    if (
                      window.confirm(
                        `Remover a conta “${u.name}”? Essa pessoa perde o acesso ao CRM.`,
                      )
                    ) {
                      remove.mutate(u.id);
                    }
                  }}
                >
                  Remover
                </Button>
              </Flex>
            </Flex>
          </Box>
        ))}
      </Stack>
    </Stack>
  );
}
