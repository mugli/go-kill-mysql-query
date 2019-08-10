# kill-mysql-query

```
  _____     ____
/      \  |  o |
|        |/ ___\|
|_________/
|_|_| |_|_|
```

`kill-mysql-query` interactively shows long running queries in MySQL database and provide option kill them one by one.

👉 Great for firefighting situations 🔥🚨🚒

It can connect to MySQL server as configured, can use SSH Tunnel if necessary, and let you decide which query to kill. By default queries running for more than 10 seconds will be marked as long running queries, but it can be configured.

![screenshot](https://raw.githubusercontent.com/mugli/go-kill-mysql-query/master/screenshot.png)

---

## Installation

Download binary from [release tab](https://github.com/mugli/go-kill-mysql-query/releases).

---

## Usage

```
kill-mysql-query [config.toml]:
	Checks for long running queries in the configured server.
	If no file is given, it tries to read from config.toml
	in the current directory.

Other commands:

	generate [config.toml]:
		Generates a new empty configuration file

	init:
		Alias for generate

	help, --help, -h:
		Shows this message

```

---

## Configuration

Run `kill-mysql-query init` to generate an empty configuration file.

```
[MySQL]
  mysql_host = ""
  mysql_port = 3306
  mysql_username = ""
  mysql_password = ""

  # hosted_in_aws_rds: Optional.
  #
  # Uses `CALL mysql.rds_kill()` instead of `kill` command.
  # Useful in RDS databases or replica where
  # `mysql_username` may not have privilege to use `kill`
  hosted_in_aws_rds = false

  # db: Optional.
  #
  # If provided, filter out long running
  # queries from other databases
  db = ""

[ssh_tunnel]
  use_ssh_tunnel = false
  ssh_host = ""
  ssh_port = 22
  ssh_username = ""
  ssh_password = ""

  # ssh_private_key takes priority over ssh_password
  # if both are provided
  ssh_private_key = ""

  # ssh_key_passphrase: Optional.
  ssh_key_passphrase = ""

[long_running_query]
  # Default is 10 seconds.
  # `kill-mysql-query` will only list running queries
  # those are being executed for more than or equal to
  # this value.
  timeout_second = 10

```

---

## FAQ

**How do I simulate a long running query to test `kill-mysql-query`?**

This stackoverflow answer may come in handy:
https://stackoverflow.com/a/3892443/761555

```
select benchmark(9999999999, md5('when will it end?'));
```

---

## License

MIT
