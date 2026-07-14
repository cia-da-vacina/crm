# POC de demonstração — CRM Cia da Vacina

## Como rodar

```bash
# Terminal 1 — API
cd backend
go run ./cmd/api

# Terminal 2 — Frontend (MSW ligado por padrão)
cd frontend
npm run dev
```

Abra http://localhost:3000

Login: `admin@ciadavacina.com.br` / `admin123`  
Atendente: `atendente@ciadavacina.com.br` / `agent123`

`NEXT_PUBLIC_USE_MOCKS=true` (default): dados no browser via MSW.  
`NEXT_PUBLIC_USE_MOCKS=false`: usa a API Go em `:8080` (seed idêntico).

## Roteiro da demo (5–8 min)

1. **Login** — entra com admin, vê shell Cia da Vacina + seletor de unidade  
2. **Inbox** — lista Maria (IA), João (negociação), Ana (aguardando)  
3. **Maria / IA** — abre conversa, banner de triagem, clique **Assumir conversa**  
4. **POP** — insere script de agendamento no composer e **Enviar**  
5. **Pipeline** — **Mover pipeline** → Em negociação / Aguardando fechamento  
6. **Não fechado** — exige motivo (preço, sem retorno…)  
7. **Follow-ups** — fila de retomada; abrir conversa / concluir  
8. **Dashboard** — KPIs da unidade + consolidado 5 unidades  
9. **POPs** — biblioteca de procedimentos  
10. **WhatsApp** — settings Meta (token mascarado, flag IA)

## Escopo da POC

| Incluído | Simulado / fora |
|----------|-----------------|
| Auth, unidades, inbox, thread | Meta WhatsApp real |
| IA triagem (seed + banner) | LLM real |
| Claim, envio, pipeline, motivos | Webhook produção |
| Follow-ups, dashboard, POPs, settings | Persistência Postgres |
