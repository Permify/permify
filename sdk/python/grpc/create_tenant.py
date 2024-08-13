from base.v1.service_pb2_grpc import TenancyStub
from base.v1.service_pb2 import TenantCreateRequest

from grpc import insecure_channel as Channel
from datetime import datetime
from pprint import pprint

channel = Channel("127.0.0.1:3478")
service = TenancyStub(channel)
tenant_id = f'tenant_id_{datetime.now().timestamp()}' 
tenant_name = 'tenant example' 
req = TenantCreateRequest(id=tenant_id, name=tenant_name)

response = service.Create(req)

pprint(response)
channel.close()
