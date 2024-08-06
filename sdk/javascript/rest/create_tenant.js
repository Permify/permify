const permify = require("permify-javascript");

const apiInstance = new permify.ApiClient("http://127.0.0.1:3476"); // Fixed URL format
const tenantApi = new permify.TenancyApi(apiInstance);

const timestamp = new Date().getTime().toString();
const tenantId = "tenant_" + timestamp;
const body = { 
    id: tenantId,
    name: "Tenant 1"
};

const callback = (error, response) => {
    if (error) {
        console.error("Error:", error);
    } else {
        console.log(response);
    }
};

tenantApi.tenantsCreate(body, callback);
