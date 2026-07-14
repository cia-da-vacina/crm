"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState, type ChangeEvent, type FormEvent } from "react";
import styled from "styled-components";
import {
  Badge,
  Box,
  Button,
  EmptyState,
  Flex,
  Heading,
  Spinner,
  Stack,
  Text,
  TextField,
} from "@cia-da-vacina/design-system";
import { MessageIcon } from "@cia-da-vacina/icon-system";
import { useAuth } from "@/contexts/auth-context";
import { api, ApiError } from "@/lib/api";
import type { Pop } from "@/lib/types";

const INTENT_SUGGESTIONS = ["agendar", "precos", "duvidas", "outro"] as const;

const TextArea = styled.textarea`
  width: 100%;
  min-height: 96px;
  resize: vertical;
  padding: 10px 12px;
  font-family: inherit;
  font-size: ${({ theme }) => theme.fontSizes.md};
  line-height: 1.45;
  color: ${({ theme }) => theme.colors["input.text"]};
  background: ${({ theme }) => theme.colors["input.bg"]};
  border: 1px solid ${({ theme }) => theme.colors["input.border"]};
  border-radius: ${({ theme }) => theme.radii.sm};
  outline: none;

  &:hover {
    border-color: ${({ theme }) => theme.colors["input.border.hover"]};
  }

  &:focus {
    border-color: ${({ theme }) => theme.colors["input.border.focus"]};
    box-shadow: ${({ theme }) => theme.shadows.focus};
  }

  &::placeholder {
    color: ${({ theme }) => theme.colors["input.placeholder"]};
  }
`;

function parseTags(raw: string) {
  const seen = new Set<string>();
  return raw
    .split(/[,;\s]+/)
    .map((t) => t.trim().toLowerCase())
    .filter((t) => {
      if (!t || seen.has(t)) return false;
      seen.add(t);
      return true;
    });
}

function resetFormFields() {
  return { title: "", body: "", tags: "outro" };
}

export default function PopsPage() {
  const qc = useQueryClient();
  const { user } = useAuth();
  const canWrite = user?.role === "admin" || user?.role === "supervisor";

  const { data, isLoading } = useQuery({
    queryKey: ["pops"],
    queryFn: () => api.listPops(),
  });

  const [editingId, setEditingId] = useState<string | null>(null);
  const [openForm, setOpenForm] = useState(false);
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [tags, setTags] = useState("outro");
  const [formError, setFormError] = useState<string | null>(null);

  function closeForm() {
    setOpenForm(false);
    setEditingId(null);
    setFormError(null);
    const blank = resetFormFields();
    setTitle(blank.title);
    setBody(blank.body);
    setTags(blank.tags);
  }

  function startCreate() {
    setEditingId(null);
    const blank = resetFormFields();
    setTitle(blank.title);
    setBody(blank.body);
    setTags(blank.tags);
    setFormError(null);
    setOpenForm(true);
  }

  function startEdit(pop: Pop) {
    setEditingId(pop.id);
    setTitle(pop.title);
    setBody(pop.body);
    setTags(pop.intent_tags.join(", "));
    setFormError(null);
    setOpenForm(true);
  }

  const save = useMutation({
    mutationFn: () => {
      const payload = {
        title: title.trim(),
        body: body.trim(),
        intent_tags: parseTags(tags),
      };
      return editingId
        ? api.updatePop(editingId, payload)
        : api.createPop(payload);
    },
    onSuccess: async () => {
      closeForm();
      await qc.invalidateQueries({ queryKey: ["pops"] });
    },
    onError: (err) => {
      setFormError(
        err instanceof ApiError ? err.message : "Não foi possível salvar o POP.",
      );
    },
  });

  const remove = useMutation({
    mutationFn: (id: string) => api.deletePop(id),
    onSuccess: async (_data, id) => {
      if (editingId === id) closeForm();
      await qc.invalidateQueries({ queryKey: ["pops"] });
    },
  });

  function toggleSuggestion(tag: string) {
    const current = parseTags(tags);
    const next = current.includes(tag)
      ? current.filter((t) => t !== tag)
      : [...current, tag];
    setTags(next.join(", "));
  }

  function onSubmit(e: FormEvent) {
    e.preventDefault();
    if (!title.trim() || !body.trim()) {
      setFormError("Título e texto são obrigatórios.");
      return;
    }
    save.mutate();
  }

  return (
    <Stack gap={3} width="100%">
      <Flex alignItems="flex-end" justifyContent="space-between" gap={2}>
        <Stack gap={1}>
          <Flex gap={2} alignItems="center">
            <MessageIcon size="lg" fill="text.brand" />
            <Heading as="h1" display>
              POPs
            </Heading>
          </Flex>
          <Text muted fontSize="sm">
            Scripts padrão para padronizar o atendimento
          </Text>
        </Stack>
        {canWrite ? (
          <Button
            variant={openForm && !editingId ? "ghost" : "primary"}
            size="sm"
            onClick={() => {
              if (openForm && !editingId) closeForm();
              else startCreate();
            }}
          >
            {openForm && !editingId ? "Cancelar" : "Novo POP"}
          </Button>
        ) : null}
      </Flex>

      {openForm && canWrite ? (
        <Box
          p={3}
          bg="bg.surface"
          borderWidth="hairline"
          borderStyle="solid"
          borderColor="border.default"
          borderRadius="md"
          width="100%"
        >
          <form onSubmit={onSubmit}>
            <Stack gap={3}>
              <Heading as="h3">
                {editingId ? "Editar POP" : "Cadastrar POP"}
              </Heading>
              <TextField
                label="Título"
                name="title"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="Ex.: Orientação pós-vacina"
                required
              />
              <Stack gap={1}>
                <Text
                  as="label"
                  htmlFor="pop-body"
                  fontSize="xs"
                  fontWeight="medium"
                  muted
                >
                  Texto do script
                </Text>
                <TextArea
                  id="pop-body"
                  name="body"
                  value={body}
                  onChange={(e: ChangeEvent<HTMLTextAreaElement>) =>
                    setBody(e.target.value)
                  }
                  rows={4}
                  required
                  placeholder="Mensagem que o atendente pode inserir na conversa…"
                />
              </Stack>
              <Stack gap={1}>
                <TextField
                  label="Tags de intenção"
                  name="tags"
                  value={tags}
                  onChange={(e) => setTags(e.target.value)}
                  hint="Separadas por vírgula. Usadas para sugerir o POP na conversa."
                  placeholder="agendar, precos"
                />
                <Flex gap={1} flexWrap="wrap">
                  {INTENT_SUGGESTIONS.map((tag) => {
                    const active = parseTags(tags).includes(tag);
                    return (
                      <Button
                        key={tag}
                        type="button"
                        size="sm"
                        variant={active ? "primary" : "secondary"}
                        onClick={() => toggleSuggestion(tag)}
                      >
                        {tag}
                      </Button>
                    );
                  })}
                </Flex>
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
                      : "Salvar POP"}
                </Button>
              </Flex>
            </Stack>
          </form>
        </Box>
      ) : null}

      {isLoading && (
        <Flex gap={2} alignItems="center">
          <Spinner />
          <Text muted>Carregando…</Text>
        </Flex>
      )}

      {!isLoading && (data?.items.length ?? 0) === 0 && (
        <EmptyState
          title="Nenhum POP"
          description={
            canWrite
              ? "Crie o primeiro script com Novo POP."
              : "Peça a um admin ou supervisor para cadastrar scripts."
          }
          action={
            canWrite ? (
              <Button size="sm" onClick={startCreate}>
                Novo POP
              </Button>
            ) : undefined
          }
        />
      )}

      <Stack gap={2} width="100%">
        {(data?.items ?? []).map((pop) => (
          <Box
            key={pop.id}
            p={3}
            bg="bg.surface"
            borderWidth="hairline"
            borderStyle="solid"
            borderColor={editingId === pop.id ? "border.brand" : "border.default"}
            borderRadius="md"
            width="100%"
          >
            <Flex justifyContent="space-between" alignItems="flex-start" gap={2}>
              <Stack gap={1} style={{ flex: 1, minWidth: 0 }}>
                <Text fontWeight="semibold">{pop.title}</Text>
                <Text fontSize="sm" muted style={{ whiteSpace: "pre-wrap" }}>
                  {pop.body}
                </Text>
                <Flex gap={1} marginTop={2} flexWrap="wrap">
                  {pop.intent_tags.map((t) => (
                    <Badge key={t} tone="brand">
                      {t}
                    </Badge>
                  ))}
                </Flex>
              </Stack>
              {canWrite ? (
                <Flex gap={1} flexShrink={0}>
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={() => startEdit(pop)}
                  >
                    Editar
                  </Button>
                  <Button
                    type="button"
                    variant="danger"
                    size="sm"
                    disabled={remove.isPending}
                    onClick={() => {
                      if (
                        window.confirm(
                          `Excluir o POP “${pop.title}”? Essa ação remove o script da biblioteca.`,
                        )
                      ) {
                        remove.mutate(pop.id);
                      }
                    }}
                  >
                    Excluir
                  </Button>
                </Flex>
              ) : null}
            </Flex>
          </Box>
        ))}
      </Stack>
    </Stack>
  );
}
