Ktctl Exchange
---

Redirect all requests of specified kubernetes service to local. Basic usage:

```bash
ktctl exchange <TargetService> --expose <LocalPort>:<TargetServicePort>
```

Available options:

```
--mode value             Exchange method 'selector', 'scale' or 'ephemeral'(experimental) (default: "selector")
--expose value           Ports to expose, use ',' separated, in [port] or [local:remote] format, e.g. 7001,8080:80
--recoverWaitTime value  (scale method only) Seconds to wait for original deployment recover before turn off the shadow pod (default: 120)
```

Key options explanation:

- `--mode` provides three ways to replace services.
  The default `selector` mode has the fastest traffic switching and switching back, and there is no need to restart the Pod of the switched service, but the `selector` attribute of the target service will be modified during the switching;
  The `scale` mode will not change the properties of the target service, but the switching process will restart the Pod of the target service, and it will take a relatively long time to wait for the original Pod to restart when switching back.
  The `ephemeral` mode can combine the advantages of the above two modes, but the current function of this mode is not complete, and it can only be used for Kubernetes v1.23 and above, so it is not recommended for the time being.
- `--expose` is a required parameter, and its value should be the same as the value of the `port` attribute of the replaced Service. If the port of the locally running service is inconsistent with the value of the `port` attribute of the target Service, you should use `<LocalPort>:<ExpectedServicePort>` format to specify.
