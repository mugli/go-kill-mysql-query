package configuration

var baseConfig =`

[MySQL]
  mysql_host = ""
  mysql_port = 3306
  mysql_username = ""
  mysql_password = ""

  # hosted_in_aws_rds: Optional.
  #
  # Uses "CALL mysql.rds_kill()" instead of "kill" command.
  # Useful in RDS databases or replica where
  # "mysql_username" may not have privilege to use "kill"
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
  # kill-mysql-query will only list running queries
  # those are being executed for more than or equal to
  # this value.
  timeout_second = 10

`