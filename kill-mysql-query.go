package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mugli/go-kill-mysql-query/configuration"
	"github.com/mugli/go-kill-mysql-query/mysql"

	"github.com/gookit/color"
	"github.com/jmoiron/sqlx"
	"github.com/manifoldco/promptui"
)

func main() {
	var config configuration.Config

	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "generate", "init":
			generateConfig()
		case "help", "-h", "--help":
			showHelp()
		default:
			config = readConfig(os.Args[1])
		}
	} else {
		config = readConfig("")
	}

	killQueries(config)
}

func readConfig(filePath string) configuration.Config {
	config, err := configuration.Read(filePath)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return config
}

func killQueries(config configuration.Config) {
	dbConn, err := mysql.Connect(config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer mysql.Disconnect()

	for {
		longQueries, err := mysql.GetLongRunningQueries(dbConn, config)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		showKillPrompt(longQueries, dbConn, config)
		fmt.Println()
		fmt.Println()
		fmt.Println()
		fmt.Println("-----------------------------------")
		fmt.Println("💫	Rechecking...")
		fmt.Println("-----------------------------------")
	}
}

func generateConfig() {
	var file string
	if len(os.Args) > 2 {
		file = os.Args[2]
	}

	err := configuration.Generate(file)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}

func showHelp() {
	help := `
  _____     ____
 /      \  |  o | 
|        |/ ___\| 
|_________/     
|_|_| |_|_|

kill-mysql-query interactively shows long running queries in MySQL database
and provides option to kill them one by one. Great for firefighting. 🔥🚨🚒

It can connect to MySQL server as configured, using SSH Tunnel if necessary 
and let you decide which query to kill. By default queries running for more
than 10 seconds will be marked as long running queries, but it can be configured.

------
Usage:

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
`
	fmt.Println(help)
	os.Exit(0)
}

func showKillPrompt(longQueries []mysql.MysqlProcess, dbConn *sqlx.DB, config configuration.Config) {
	if len(longQueries) == 0 {
		fmt.Printf("✨	No queries are running for more than %d second(s). Quitting! 👋\n", config.LongQuery.TimeoutSecond)
		os.Exit(0)
	}

	if len(longQueries) == 1 {
		query := longQueries[0]
		cyan := color.FgCyan.Render
		label := fmt.Sprintf("🐢	This query is running for %s second(s) in `%s` database:\n\n%s\n\n", cyan(query.Time), cyan(query.DB), cyan(query.Info.String))

		fmt.Println()
		fmt.Println()
		fmt.Println(label)
		prompt := promptui.Prompt{
			Label:     "🧨  Kill it?",
			IsConfirm: true,
			Default:   "n",
		}

		result, _ := prompt.Run()

		if strings.TrimSpace(strings.ToLower(result)) == "y" {
			err := mysql.KillMySQLProcess(query.KillCommand, dbConn)

			if err != nil {
				fmt.Printf("😓	There was an error killing the query: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Println("Quitting! 👋")
			os.Exit(0)
		}
	}

	if len(longQueries) > 1 {
		templates := &promptui.SelectTemplates{
			Label: "{{ . }}?",
			Active: "👉	DB `{{ .DB | cyan }}`,	Running Time: {{ .Time | cyan }}s,	Query: {{ .TruncatedQuery | cyan }}",
			Inactive: "	DB `{{ .DB }}`,	Running Time: {{ .Time }}s,	Query: {{ .TruncatedQuery }}",
			Selected: "💥	DB `{{ .DB | cyan }}`,	Running Time: {{ .Time | cyan }}s,	Query: {{ .TruncatedQuery | cyan }}",
			Details: `
--------- QUERY ----------
{{ "ID:" | faint }}	{{ .ID }}
{{ "DB:" | faint }}	{{ .DB }}
{{ "State:" | faint }}	{{ .State.String }}
{{ "Command:" | faint }}	{{ .Command }}
{{ "Running Time:" | faint }}	{{ .Time }} second(s)
{{ "Query:" | faint }}	{{ .TruncatedQuery }}`,
		}

		prompt := promptui.Select{
			Label:     "Press enter to select. Which one to kill?",
			Items:     longQueries,
			Templates: templates,
			Size:      10,
		}

		fmt.Println()
		fmt.Println()
		i, _, err := prompt.Run()

		if err != nil {
			fmt.Printf("😓 There was an error: %v\n", err)
			os.Exit(1)
		}

		selectedQuery := longQueries[i]
		err = mysql.KillMySQLProcess(selectedQuery.KillCommand, dbConn)

		if err != nil {
			fmt.Printf("😓 There was an error killing the query: %v\n", err)
			os.Exit(1)
		}
	}
}
