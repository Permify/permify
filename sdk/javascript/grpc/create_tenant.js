const permify = require("@permify/permify-node");

const request = new permify.grpc.payload.TenantCreateRequest();
request.setId("t1");
request.setName("tenant 1");

const client = permify.grpc.newClient({
    endpoint: "localhost:3478",
    cert: undefined
});

client.tenancy.create(request).then((response) => {
    console.log(response);

    // handle response
})