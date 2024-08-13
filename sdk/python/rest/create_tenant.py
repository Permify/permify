import permify
from permify import ApiException

from datetime import datetime
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:3476
configuration = permify.Configuration(
    host = "http://localhost:3476"
)

with permify.ApiClient(configuration) as api_client:
    api_instance = permify.TenancyApi(api_client)
    tenant_id = f'tenant_id_{datetime.now().timestamp()}' 
    tenant_name = 'tenant example' 

    req = permify.TenantCreateRequest(id=tenant_id, name=tenant_name)
    try:
        response = api_instance.tenants_create(req)
        pprint(response)
    except ApiException as e:
        print("Exception when creating tenant: %s\n" % e)
