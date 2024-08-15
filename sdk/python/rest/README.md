# Permify Python SDK

This repository contains a sample usage for the Python REST SDK for Permify.

### Prerequisites

Ensure you have the following installed:
- Python 3.7+

### Cloning the Repository

To get started, clone the repository using the following command:

```sh
git clone https://github.com/ucatbas/permify-sdk-samples.git
cd permify-sdk-samples/python/rest
```

### Building the Project

Install permify:
```sh
pip install permify
```
(you may need to run `pip` with root permission: `sudo pip install permify`)

### Running the Application

After successfully installing the package, you can run the application using the following command:
```sh
python3 create_tenant.py
```

## For your own Projects

Here is a simple permify client:

```python
import permify
from permify import ApiException

configuration = permify.Configuration(
    host = "http://localhost:3476"
)

with permify.ApiClient(configuration) as api_client:
    api_instance = permify.TenancyApi(api_client)
```

