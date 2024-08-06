# Permify Javascript SDK

This repository contains a sample usage for the Javascript REST SDK for Permify.

### Prerequisites

Ensure you have the following installed:
- node
- npm

### Cloning the Repository

To get started, clone the repository using the following command:

```sh
git clone https://github.com/ucatbas/permify-sdk-samples.git
cd permify-sdk-samples/javascript/rest
```

### Building the Project

Install permify:
```sh
npm install permify-javascript
```

### Running the Application

After successfully installing the package, you can run the application using the following command:
```sh
node create_tenant.ts
```

## For your own Projects

Here is a simple permify client:

```javascript
const permify = require("permify-javascript");

const apiInstance = new permify.ApiClient("http://127.0.0.1:3476"); // Fixed URL format
const tenantApi = new permify.TenancyApi(apiInstance);
```
