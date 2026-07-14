# CRM Cia da Vacina

POC do CRM WhatsApp (triagem IA + handoff humano) — Next.js frontend + API Go em memória.

## Demo local

```bash
npm install --ignore-scripts
npm run build -w @cia-da-vacina/styled-system
npm run build -w @cia-da-vacina/design-system-tokens
npm run build -w @cia-da-vacina/icon-system
npm run build -w @cia-da-vacina/design-system
npm run dev -w frontend
```

Login: `admin@ciadavacina.com.br` / `admin123`

Com `NEXT_PUBLIC_USE_MOCKS=true` o front usa MSW (não precisa da API Go).

## Deploy

Frontend na Vercel (root directory `frontend`, mocks ligados).
