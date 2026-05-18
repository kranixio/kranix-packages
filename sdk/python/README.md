# kranix-io-sdk (Python)

First-class Python client for the Kranix API — aimed at ML and data engineering workflows.

```bash
pip install kranix-io-sdk
```

```python
from kranix_sdk import KraneClient

client = KraneClient(
    server_url="http://localhost:8080",
    api_key="krane_dev",
    skip_auth=True,  # with kranix-mock-api -skip-auth
)

wl = client.workloads.deploy({
    "name": "training",
    "image": "pytorch/pytorch:latest",
    "namespace": "default",
    "replicas": 1,
    "backend": "docker",
})

for line in client.pods.stream_logs(pod_id, follow=True, tail=50):
    print(line)

for frame in client.subscribe_workload_events(namespaces=["default"]):
    print(frame.event, frame.data)
```

See the parent [kranix-packages README](../../README.md) for ecosystem layout.
