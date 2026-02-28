# Go SDK Parity Matrix

Last validated: 2026-02-28 (America/New_York)

## Standards in `/standards-sdk/src` vs `/go-sdk/pkg`

| HCS Standard | TS module in `/standards-sdk/src` | Go module in `/go-sdk/pkg` | 1:1 Parity Status | Live Validation Status |
| --- | --- | --- | --- | --- |
| HCS-2 | yes (`hcs-2`) | yes (`hcs2`) | yes | pass (`TestHCS2Integration_EndToEnd`) |
| HCS-3 | yes (`hcs-3`) | no | no | n/a |
| HCS-5 | yes (`hcs-5`) | yes (`hcs5`) | yes | pass (`TestHCS5Integration_MintWithExistingHCS1Topic`) |
| HCS-6 | yes (`hcs-6`) | no | no | n/a |
| HCS-7 | yes (`hcs-7`) | no | no | n/a |
| HCS-10 | yes (`hcs-10`) | no | no | n/a |
| HCS-11 | yes (`hcs-11`) | yes (`hcs11`) | yes | pass (`TestHCS11Integration_FetchProfileByAccountID`) |
| HCS-12 | yes (`hcs-12`) | no | no | n/a |
| HCS-14 | yes (`hcs-14`) | yes (`hcs14`) | yes | pass (`TestHCS14Integration_ANSDNSWebResolution`) |
| HCS-15 | yes (`hcs-15`) | yes (`hcs15`) | yes | pass (`TestHCS15Integration_CreateBaseAndPetalAccounts`) |
| HCS-16 | yes (`hcs-16`) | yes (`hcs16`) | yes | pass (`TestHCS16Integration_CreateFloraAndPublishMessages`) |
| HCS-17 | yes (`hcs-17`) | yes (`hcs17`) | yes | pass (`TestHCS17Integration_ComputeAndPublishStateHash`) |
| HCS-18 | yes (`hcs-18`) | no | no | n/a |
| HCS-20 | yes (`hcs-20`) | no | no | n/a |
| HCS-21 | yes (`hcs-21`) | no | no | n/a |
| HCS-26 | yes (`hcs-26`) | no | no | n/a |

## Go-only additions (not in TS baseline module list)

| Go module | Notes |
| --- | --- |
| `hcs27` | Implemented and live validated (`TestHCS27Integration_CheckpointChain`) |

## Utility parity tracked alongside standards

| Utility Surface | TS Path | Go Path | Status |
| --- | --- | --- | --- |
| Inscriber | `src/inscribe` | `pkg/inscriber` | 1:1 parity for requested features (websocket default, bulk-files, skill helpers) |
| Registry Broker client | `src/services/registry-broker` | `pkg/registrybroker` | Implemented parity |
