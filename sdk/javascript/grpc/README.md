# Permify Javascript SDK

This repository contains a sample usage for the Javascript gRPC SDK for Permify.

### Prerequisites

Ensure you have the following installed:
- node
- npm

### Cloning the Repository

To get started, clone the repository using the following command:

```sh
git clone https://github.com/ucatbas/permify-sdk-samples.git
cd permify-sdk-samples/javascript/grpc
```

### Building the Project

Install permify:
```sh
npm config set @buf:registry https://buf.build/gen/npm/v1/
npm install @permify/permify-node
```

### Running the Application

After successfully installing the package, you can run the application using the following command:
```sh
node create_tenant.ts
```

## For your own Projects

Here is a simple permify client:

```javascript
const permify = require("@permify/permify-node");

const client = permify.grpc.newClient({
    endpoint: "localhost:3478",
    cert: undefined
});
```
