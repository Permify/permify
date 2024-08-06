import * as permify from "permify-typescript";

const configuration = new permify.Configuration({
    basePath: "http://localhost:3476"
})

const apiInstance = new permify.TenancyApi(configuration);
const timestamp = new Date().getTime().toString();

let tenantId = "tenant_" + timestamp; 

apiInstance.tenantsCreate({
    body: {
        id: tenantId,
        name: "Tenant 1"
    }
}).then((data) => {
    console.log(data);
}).catch((error) => console.error(error));
