package mysql

import (
	"bytes"
	"database/sql"
	"fmt"
	"regexp"
	"text/template"

	"github.com/mugli/go-kill-mysql-query/configuration"

	"github.com/jmoiron/sqlx"
)

const (
	killRegularPrefix = "kill "
	killRegularSuffix = ";"
	killRdsPrefix     = "CALL mysql.rds_kill("
	killRdsSuffix     = ");"

	baseQuery = `
	SELECT 
		ID, 
		CONCAT('{{.KillPrefix}}', ID, '{{.KillSuffix}}') AS KILL_COMMAND, 
		DB, 
		STATE, 
		COMMAND, 
		TIME, 
		INFO	
	FROM information_schema.PROCESSLIST
	WHERE TRUE
		AND COMMAND NOT IN ('Sleep', 'Killed')
		AND DB IS NOT NULL
		{{.DBFilter}}
		{{.TimeFilter}}
		AND INFO NOT LIKE '%PROCESSLIST%'
	ORDER BY TIME DESC
	LIMIT 10;
	`
)

type queryParams struct {
	KillPrefix string
	KillSuffix string
	DBFilter   string
	TimeFilter string
}

type MysqlProcess struct {
	ID             int            `db:"ID"`
	KillCommand    string         `db:"KILL_COMMAND"`
	DB             string         `db:"DB"`
	State          sql.NullString `db:"STATE"`
	Command        string         `db:"COMMAND"`
	Time           int            `db:"TIME"`
	Info           sql.NullString `db:"INFO"`
	TruncatedQuery string
}

func generateQuery(config configuration.Config) (string, error) {
	tmpl := template.New("query")

	tmpl, err := tmpl.Parse(baseQuery)
	if err != nil {
		return "", err
	}

	params := queryParams{}
	if config.MySQL.AwsRds {
		params.KillPrefix = killRdsPrefix
		params.KillSuffix = killRdsSuffix
	} else {
		params.KillPrefix = killRegularPrefix
		params.KillSuffix = killRegularSuffix
	}

	params.TimeFilter = fmt.Sprintf("AND TIME >= %d", config.LongQuery.TimeoutSecond)

	if config.MySQL.DB != "" {
		params.DBFilter = fmt.Sprintf("AND DB = '%s'", config.MySQL.DB)
	}

	var queryBytes bytes.Buffer
	err = tmpl.Execute(&queryBytes, params)
	if err != nil {
		return "", err
	}

	return queryBytes.String(), nil
}

func truncateString(str string, num int) string {
	// Remove newlines
	re := regexp.MustCompile(`\r?\n`)
	retval := re.ReplaceAllString(str, " ")

	if len(retval) > num {
		if num > 3 {
			num -= 3
		}
		retval = retval[0:num] + "..."
	}
	return retval
}

func GetLongRunningQueries(dbConn *sqlx.DB, config configuration.Config) ([]MysqlProcess, error) {
	fmt.Println("🕴	Looking for slow queries...")

	longQueries := make([]MysqlProcess, 0)
	query, err := generateQuery(config)
	if err != nil {
		return nil, err
	}

	if rows, err := dbConn.Queryx(query); err == nil {
		for rows.Next() {
			longQ := MysqlProcess{}
			rows.StructScan(&longQ)

			longQ.TruncatedQuery = truncateString(longQ.Info.String, 50)

			longQueries = append(longQueries, longQ)
		}
		rows.Close()
	} else {
		return nil, err
	}

	return longQueries, nil
}

func KillMySQLProcess(killCommand string, dbConn *sqlx.DB) error {
	fmt.Println("☠️	Sending kill command...")

	if _, err := dbConn.Queryx(killCommand); err != nil {
		return err
	}

	fmt.Println("☠️	Killed it!")
	return nil
}
