import * as permify from "@permify/permify-node";


const request = new permify.grpc.payload.TenantCreateRequest();
request.setId("t1");
request.setName("tenant 1");

const client = permify.grpc.newClient({
    endpoint: "localhost:3478",
    cert: undefined
});

client.tenancy.create(request).then((response: any) => {
    console.log(response);

    // handle response
})
