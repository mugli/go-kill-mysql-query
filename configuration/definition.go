package configuration

type Config struct {
	MySQL     mysql
	SSH       sshTunnel        `toml:"ssh_tunnel"`
	LongQuery longRunningQuery `toml:"long_running_query"`
}

type mysql struct {
	Host     string `toml:"mysql_host" validate:"required"`
	Port     int    `toml:"mysql_port"`
	Username string `toml:"mysql_username" validate:"required"`
	Password string `toml:"mysql_password"`
	AwsRds   bool   `toml:"hosted_in_aws_rds"`
	DB       string `toml:"db"`
}

type sshTunnel struct {
	UseTunnel     bool   `toml:"use_ssh_tunnel"`
	Host          string `toml:"ssh_host" validate:"required_with=UseTunnel"`
	Port          int    `toml:"ssh_port"`
	Username      string `toml:"ssh_username" validate:"required_with=UseTunnel"`
	Password      string `toml:"ssh_password"`
	Key           string `toml:"ssh_private_key"`
	KeyPassphrase string `toml:"ssh_key_passphrase"`
}

type longRunningQuery struct {
	TimeoutSecond int `toml:"timeout_second"`
}
