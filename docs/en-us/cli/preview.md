Ktctl Preview
---

Expose a local service to kubernetes cluster. Basic usage:

```bash
ktctl preview <NewService> --expose <LocalPort>:<NewServicePort>
```

Available options:

```
--expose value      Ports to expose, use ',' separated, in [port] or [local:remote] format, e.g. 7001,8080:80
--external          If specified, a public, external service is created
--skipPortChecking  Do not check whether specified local ports are listened
```

Key options explanation:

- `--expose` is a required parameter, and its value should be the same as the port of the locally running service. If you want the created Service to use a different port than the local service, you should use `<LocalPort>:<ExpectedServicePort>` format to specify.
