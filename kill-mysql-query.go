package main

import (
	"fmt"
	"os"
	"strings"

	"go-kill-mysql-query/configuration"
	"go-kill-mysql-query/mysql"

	"github.com/gookit/color"
	"github.com/jmoiron/sqlx"
	"github.com/manifoldco/promptui"
)

func main() {
	var config configuration.Config

	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "generate":
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
		default:
			var err error
			config, err = configuration.Read(os.Args[1])

			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	} else {
		var err error

		// Try to read from default config file
		config, err = configuration.Read("")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

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
		fmt.Println("💫	Rechecking...")
		fmt.Println()
	}
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
			Active: "🧨  DB `{{ .DB | cyan }}`,	Running Time: {{ .Time | cyan }}s,	Query: {{ .Info.String | cyan }}",
			Inactive: "   DB `{{ .DB | cyan }}`,	Running Time: {{ .Time | cyan }}s,	Query: {{ .Info.String | cyan }}",
			Selected: "💥  DB `{{ .DB | cyan }}`,	Running Time: {{ .Time | cyan }}s,	Query: {{ .Info.String | cyan }}",
			Details: `
--------- QUERY ----------
{{ "ID:" | faint }}	{{ .ID }}
{{ "DB:" | faint }}	{{ .DB }}
{{ "State:" | faint }}	{{ .State.String }}
{{ "Command:" | faint }}	{{ .Command }}
{{ "Running Time:" | faint }}	{{ .Time }} second(s)
{{ "Query:" | faint }}	{{ .Info.String }}`,
		}

		prompt := promptui.Select{
			Label:     "Which one to kill?",
			Items:     longQueries,
			Templates: templates,
			Size:      10,
		}

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
