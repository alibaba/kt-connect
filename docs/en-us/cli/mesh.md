Ktctl Mesh
---

Redirect marked requests of specified kubernetes service to local. Basic usage:

```bash
ktctl mesh <TargetService> --expose <LocalPort>:<TargetServicePort>
```

Available options:

```
--mode value         Mesh method 'auto' or 'manual' (default: "auto")
--expose value       Ports to expose, use ',' separated, in [port] or [local:remote] format, e.g. 7001,8080:80
--versionMark value  Specify the version of mesh service, e.g. '0.0.1' or 'mark:local'
--routerImage value  (auto method only) Customize router image (default: "registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-router:vdev")
```

Key options explanation:

- `--mode` provides two ways for the service to redirect routes.
  The default `auto` mode uses Router Pod to implement automatic routing of HTTP requests without additional configuration of service mesh components, which is suitable for scenarios where no service mesh is deployed in the cluster.
  The `manual` mode only "mixes" local services into the cluster, and adds a specific version of the Label, and developers can flexibly configure routing rules through service mesh components (such as Istio).
- `--expose` is a required parameter, and its value should be the same as the value of the `port` attribute of the target Service. If the port of the local running service is inconsistent with the value of the `port` attribute of the target Service, you should use `<LocalPort>:<ExpectedServicePort>` format to specify.
- `--versionMark` is used to specify the name and value of the Header or Label to route to the local. The default value is "version:\<randomly generated value\>", you can specify only the tag value, such as `--versionMark demo`; you can specify only the tag name in the format of the tag name plus a colon, such as `--versionMark kt-mark: `; You can also specify the name and value of the tag at the same time, such as `--versionMark kt-mark:demo`.
  In `auto` mode, the value is actually the header used for routing. In `manual` mode, this value is an extra Label attached to the Shadow Pod leading to the local service.
