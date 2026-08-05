# Mocks locais (só frontend)

Camada temporária para rodar a UI **sem backend**. Todo o código fica nesta pasta.

## Ligar

Em `frontend/.env.local`:

```env
USE_MOCKS=true
COOKIE_SECURE=false
```

(`API_URL` é ignorado enquanto `USE_MOCKS=true`.)

```bash
npm run dev
```

Login demo:

- `admin@ciadavacina.com.br` / `admin123`
- `atendente@ciadavacina.com.br` / `agent123`

OTP de telefone no mock: `123456`.

## Desligar

```env
USE_MOCKS=false
API_URL=http://localhost:8080/api/v1
```

## Remover de vez

1. Apagar a pasta `frontend/src/mocks/`
2. Remover o bloco `USE_MOCKS` em:
   - `frontend/src/server/env.ts`
   - `frontend/src/server/backend.ts`
   - `frontend/src/app/api/proxy/[...path]/route.ts`
3. Tirar `USE_MOCKS` de `.env.example` / `.env.local`
4. Ajustar o README do frontend (seção de mocks)
