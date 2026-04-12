#!/bin/bash
curl -s -v -X POST http://localhost:8082/v1/webhooks/deposits/sepay \
  -H "Content-Type: application/json" \
  -d '{
    "gateway": "MBBank",
    "transactionDate": "2026-04-12 22:03:00",
    "accountNumber": "0327182537",
    "subAccount": null,
    "code": null,
    "content": "DEP5616cfd3447ea7b954b58bf4a34efdea   Ma giao dich  Trace287063 Trace 287063",
    "transferType": "in",
    "transferAmount": 50000,
    "referenceCode": "FT26103422070458",
    "accumulated": 0,
    "id": 50029960
  }'
echo ""
