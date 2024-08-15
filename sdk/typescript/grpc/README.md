# Permify Typescript SDK

This repository contains a sample usage for the Typescript gRPC SDK for Permify.

### Prerequisites

Ensure you have the following installed:
- npm
- typescript
- ts-node

### Cloning the Repository

To get started, clone the repository using the following command:

```sh
git clone https://github.com/ucatbas/permify-sdk-samples.git
cd permify-sdk-samples/typescript/grpc
```

### Building the Project

Install permify:
```sh
npm config set @buf:registry https://buf.build/gen/npm/v1/
npm install @permify/permify-node
```

### Running the Application

Create a typescript configuration file, `tsconfig.json` to let running javascript libraries: 

```json
{
    "compilerOptions": {
      "allowJs": true
    }
}
```

After successfully installing the package and configuration, you can run the application using the following command:
```sh
ts-node create_tenant.ts
```

## For your own Projects

Here is a simple permify client:

```typescript
import * as permify from "@permify/permify-node";

const client = permify.grpc.newClient({
    endpoint: "localhost:3478",
    cert: undefined
});
```
