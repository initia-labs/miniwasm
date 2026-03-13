# Protobufs

This is the public protocol buffers API for the [Miniwasm](https://github.com/initia-labs/miniwasm).

## npm Package

TypeScript definitions are published to npm as [`@initia/miniwasm-proto`](https://www.npmjs.com/package/@initia/miniwasm-proto).

- **Tagged releases** (`v*`) are published as `latest` (e.g. `1.0.0`).
- **Main branch** pushes are published as `canary` (e.g. `0.0.0-canary.<short-sha>`).

### Installation

```bash
npm install @initia/miniwasm-proto @bufbuild/protobuf
```

### Usage

```typescript
import { MsgCreateDenomSchema } from "@initia/miniwasm-proto/miniwasm/tokenfactory/v1/tx_pb";
import { ParamsSchema } from "@initia/miniwasm-proto/miniwasm/tokenfactory/v1/params_pb";
```

The package requires `@bufbuild/protobuf` v2 as a peer dependency.
