Ktctl Forward
---

Redirect local port to a service or any remote address. Basic usage:

```bash
ktctl forward <TargetService> <LocalPort>:<TargetServicePort>
```

Available options:

```
None
```

Key options explanation:

- When the first parameter is the name of a service which defines only one port, then the second parameter can be omitted (means forward the port of service to the same local port) or only specify local port (means forward the port of service to the specified local port)
