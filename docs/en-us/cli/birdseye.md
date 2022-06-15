Ktctl Birdseye
---

View routing status of each service in the current namespace. Basic usage:

```bash
ktctl birdseye
````

Available options:

```
--sortBy string        Sort service by 'status' or 'name' (default "status")
--showConnector        Also show name of users who connected to cluster
--hideNaturalService   Only show exchanged / meshed and previewing services
```

> The username displayed by the command is the login name of the developer's local computer

Key options explanation:

- `--sortBy` parameter is used to specify the order of the services displayed. Default value `status` will show services in order of "exchanged services" -> "meshed services" -> "natural services" -> "previewing services". Optional value `name` will display the services in alphabetical order by its name.
