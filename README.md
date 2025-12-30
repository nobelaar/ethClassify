# ethClassify

Herramienta en Go que obtiene el ultimo bloque de Ethereum mainnet via el endpoint RPC que indiques y clasifica cada transaccion. Imprime hash, destino anotado, valor, datos en hex y un tipo detectado; con `-with-logs` puede resolver eventos ERC20/721 usando recibos.

## Requisitos
- Go 1.24+
- Endpoint RPC de Ethereum mainnet (Infura, Alchemy, nodo propio, etc). Para `-with-logs` se necesitan recibos (`eth_getTransactionReceipt`).

## Uso rapido
1. Compila ejecutando `go build main.go`
2. Lanza la clasificacion con `./main -url https://mainnet.infura.io/v3/<project-id>`.
3. Agrega `-with-logs` si quieres traer recibos/logs y detectar transferencias/aprobaciones ERC20 o ERC721.

### Flags
- `-url` (obligatorio): URL del endpoint RPC.
- `-with-logs` (opcional): solicita recibos/logs para enriquecer la clasificacion de llamadas a contratos (mas llamadas RPC).
- `-h` / `--help`: imprime el mensaje de ayuda.

### Salida
Se muestra numero y hash del bloque y, por cada transaccion, hash, destino (con etiqueta si esta en `internal/infrastructure/labeler/static_labeler.go`), valor en wei/ETH, datos en hex, tipo detectado y selector de funcion si aplica.

## Tipos detectados
- `DEPLOY`
- `TRANSFER`
- `CONTRACT_CALL`
- `DEX_SWAP` (Uniswap V2/V3 via logs)
- `SANDWICH_SUSPECT` (heurística simple sobre swaps consecutivos en el mismo pool)
- `ERC20_TRANSFER`
- `ERC20_APPROVE`
- `ERC20_TRANSFER_FROM`
- `ERC721_TRANSFER`
- `ERC721_APPROVAL`
- `ERC721_APPROVAL_FOR_ALL`
- `UNKNOWN`

## Estructura
- `main.go`: parseo de flags, construccion de dependencias y ejecucion de la clasificacion.
- `internal/infrastructure/ethereum/block_reader.go`: conexion RPC y lectura del bloque mas reciente (con o sin logs).
- `internal/usecase/classify_block.go`: orquesta los clasificadores y resolvedores de logs.
- `internal/infrastructure/classifier/ethereum_classifiers.go`: reglas para tipos base y deteccion ERC20/721 via logs.
- `internal/interface/cli/presenter.go`: imprime los resultados en la consola.
- `internal/infrastructure/labeler/static_labeler.go`: etiquetas estaticas para contratos conocidos (USDT, USDC, DAI, WETH).
