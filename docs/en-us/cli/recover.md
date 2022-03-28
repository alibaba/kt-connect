Ktctl Recover
---

Restore traffic of specified kubernetes service changed by exchange or mesh. Basic usage:

```bash
ktctl recover <TargetService>
```

No extra parameter available.

Special notice:

- This command should only use for restore traffic redirect made by KtConnect `0.3.2` or above. If there are still `0.3.1` or below version user in the cluster, it's better to wait the user quit himself or use `ktctl clean` command to automatically clean up expired resource and restore the network traffic
