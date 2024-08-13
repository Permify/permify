# Permify Typescript SDK

This repository contains a sample usage for the Typescript REST SDK for Permify.

### Prerequisites

Ensure you have the following installed:
- npm
- typescript
- ts-node

### Cloning the Repository

To get started, clone the repository using the following command:

```sh
git clone https://github.com/ucatbas/permify-sdk-samples.git
cd permify-sdk-samples/typescript/rest
```

### Building the Project

Install permify:
```sh
npm install permify-typescript
```

### Running the Application

After successfully installing the package, you can run the application using the following command:
```sh
ts-node create_tenant.ts
```

## For your own Projects

Here is a simple permify client:

```typescript
import * as permify from "permify-typescript";

const configuration = new permify.Configuration({
    basePath: "http://localhost:3476"
})

const apiInstance = new permify.TenancyApi(configuration);
```
