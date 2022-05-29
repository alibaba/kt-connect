Ktctl Config
---

List, get or set default value for command options. It contains 8 sub-commands:

- `show`: Show all available and configured options
- `get`: Fetch default value of specified option
- `set`: Customize default value of specified option
- `unset`: Restore default value of specified option
- `list-profile`: List all pre-saved profiles
- `save-profile`: Save current configured options as a profile
- `load-profile`: Load config from a profile
- `drop-profile`: Delete a profile

Basic usage:

```bash
ktctl config show
ktctl config get <arg-name>
ktctl config set <arg-name> <default-value>
ktctl config unset <arg-name>
ktctl config list-profile
ktctl config save-profile <profile-name>
ktctl config load-profile <profile-name>
ktctl config drop-profile <profile-name>
```

Format of the argument setting is "<command>.<parameter>", the parameter part should be written in all lower-case with slash separated style. For example, below command will set default value of `ktctl connect --excludeIps` parameter to `172.2.1.0/24,172.2.2.0/24`:

```bash
ktctl config set connect.exclude-ips 172.2.1.0/24,172.2.2.0/24
                |<-cmd->||<-param->| |<--- default value --->|
```

For global parameter, use `global` as the command part, for example to change the default shadow pod image:

```bash
ktctl config set global.image your-repository-address/kt-connect-shadow:latest
```

Use `config show` command with `--all` parameter will list all available parameter names:

```bash
ktctl config show --all
```

The configuration will be stored as YAML format to file ".kt/config" under user's HOME directory.

All available parameter of `config` command itself:

```
config show
--all, -a     Show all available config options

config get
N/A

config set
N/A

config unset
--all, -a     Unset all config options

config list-profile
N/A

config save-profile
N/A

config load-profile
--dryRun      Print profile content without load it

config drop-profile
N/A
```
