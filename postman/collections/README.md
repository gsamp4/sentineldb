# Postman Collections

Collections exportadas no formato Postman v2.1 para importacao manual.

Variaveis usadas nas collections:

- `baseUrl`: URL base da API. Exemplo: `http://localhost:8080`
- `assetId`: ID de um asset existente
- `runId`: ID de um run existente
- `findingId`: ID de um finding existente

Observacoes sobre a implementacao atual da API:

- `GET /api/v1/runs/:id/jobs` tambem tenta ler o `id` no body JSON.
- `PATCH /api/v1/findings/:id/resolve` tambem tenta ler o `id` no body JSON.

Por isso, as requests de `runs/:id/jobs` e `findings/:id/resolve` enviam o parametro no path e no body para maximizar compatibilidade com o codigo atual.

No caso de `POST /api/v1/trigger/:id`, o `:id` correto e um `assetId`, porque o `runId` so existe depois que o trigger e aceito.
