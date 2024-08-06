# Permify Python SDK

This repository contains a sample usage for the Python gRPC SDK for Permify.

### Prerequisites

Ensure you have the following installed:
- Python 3.7+

### Cloning the Repository

To get started, clone the repository using the following command:

```sh
git clone https://github.com/ucatbas/permify-sdk-samples.git
cd permify-sdk-samples/python/grpc
```

### Building the Project

Install permify:
```sh
python3 -m pip install permifyco-permify-grpc-python --extra-index-url https://buf.build/gen/python
```

### Running the Application

After successfully installing the package, you can run the application using the following command:
```sh
python3 create_tenant.py
```

## For your own Projects

Here is a simple permify client:

```python
from base.v1.service_pb2_grpc import TenancyStub
from base.v1.service_pb2 import TenantCreateRequest

from grpc import insecure_channel as Channel
from datetime import datetime
from pprint import pprint

channel = Channel("127.0.0.1:3478")
service = TenancyStub(channel)
```

