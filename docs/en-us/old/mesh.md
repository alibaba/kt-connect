## Scenarios

The `Connect` and `Exchange'`are suitable for personal exclusive development test environments, in exclusive mode. Developers can coordinate with services in this environment while forwarding requests for specific services in the environment to the local.

The main solution of `Mesh` is that if the team shares a development test environment, how to ensure that each team member can independently perform the joint test on the basis of this 'unique public environment'. In this model, the most immediate benefit is to reduce the investment in infrastructure resources while supporting large-scale collaboration.

`Mesh` is similar to `Exchange`. The difference is that Exchange will completely replace the original application instance, while Mesh creates a new version based on the original instance, thus enabling users to do more based on Service Mesh capabilities. Multiple custom traffic rule definitions. This enables team members to perform local joint testing on a common development test environment.

![logo](../../media/logo-large.png)
