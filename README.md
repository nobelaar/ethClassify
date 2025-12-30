# ethClassify

Herramienta en Go que obtiene el ultimo bloque de Ethereum mainnet via Infura y clasifica cada transaccion. Imprime hash, destino anotado, valor y datos en hex y asigna un tipo simple (deploy, transferencias nativas y llamadas ERC20 basicas).

## Requisitos
- Go 1.24+
- Acceso a un endpoint RPC de Ethereum mainnet. El codigo usa Infura con el project id que hoy esta codificado en `core/block_getter.go`.

## Configuracion rapida
- Actualiza la URL RPC en `core/block_getter.go` (llamada a `ethclient.Dial`) con tu endpoint. Si prefieres no exponer el project id, cambialo para leer una variable de entorno antes de compilar.

## Uso
- Ejecuta `go run .` para traer el ultimo bloque y listar sus transacciones clasificadas.
- La salida por cada transaccion incluye hash, direccion destino con etiqueta cuando aplica, valor, datos en hex y el tipo detectado. Ejemplo de formato:

```text
Tx Hash:  0x...
Tx To:  0xdac17f... (USDT Contract)
Tx Value:  0
Tx Data:  a9059cbb...
Tx Type: ERC_TRANSFER
----------------
```

## Tipos detectados
- `DEPLOY`: `tx.To()` es nil, despliegue de contrato.
- `TRANSFER`: sin datos (`len(data)==0`) y valor mayor a cero.
- `ERC_TRANSFER`: selector `a9059cbb` (transfer de ERC20).
- `ERC_APPROVE`: selector `095ea7b3`.
- `ERC_TRANSFER_FROM`: selector `23b872dd`.
- `UNKNOWN`: cualquier otro caso. Se imprimen igual los detalles para poder inspeccionarlos.

Ademas `core/addr_classifier.go` etiqueta direcciones conocidas; hoy reconoce el contrato de USDT.

## Estructura
- `main.go`: orquestacion; obtiene un bloque y recorre las transacciones.
- `core/block_getter.go`: conexion al nodo y descarga del bloque mas reciente.
- `core/classifier.go`: reglas basicas de clasificacion por selector y tipo de transaccion.
- `core/addr_classifier.go`: etiquetas de direcciones.

## Limitaciones y posibles mejoras
- El endpoint RPC esta hardcodeado y no hay manejo de variables de entorno ni reintentos.
- Solo se cubren tres metodos ERC20 comunes; no se decodifican parametros ni se detectan otros protocolos.
- No hay pruebas automatizadas ni manejo de logging estructurado.
