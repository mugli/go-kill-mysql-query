package configuration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/rhysd/abspath"
	"gopkg.in/go-playground/validator.v9"
)

func setDefault() Config {
	return Config{
		LongQuery: longRunningQuery{
			TimeoutSecond: 10,
		},
		MySQL: mysql{
			Port: 3306,
		},
		SSH: sshTunnel{
			Port: 22,
		},
	}
}

func validate(config Config) error {
	validate := validator.New()

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("toml"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	err := validate.Struct(config)

	if err != nil {
		errs := err.(validator.ValidationErrors)

		return errors.New(fmt.Sprintf("%s is required in configuration", errs[0].Field()))
	}

	return nil
}

func Read(file string) (Config, error) {
	fmt.Println("⚙️	Reading configuration")
	var filePath string
	config := setDefault()

	if file == "" {
		dir, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		defaultConfig := filepath.Join(dir, "config.toml")

		if fExists(defaultConfig) {
			filePath = defaultConfig
		} else {
			return config, errors.New(defaultConfig + " does not exist. Cannot read from config. \n\nRun `kill-mysql-query help` for usage details.\n")
		}
	} else {
		abs, err := abspath.ExpandFrom(file)
		if err != nil {
			return config, err
		}

		filePath = abs.String()
		if !fExists(filePath) {
			return config, errors.New(filePath + " does not exist. Cannot read from config. \n\nRun `kill-mysql-query help` for usage details.\n")
		}
	}

	_, err := toml.DecodeFile(filePath, &config)
	if err != nil {
		return config, err
	}

	err = validate(config)
	if err != nil {
		return config, err
	}

	return config, err
}

func Generate(file string) error {
	var filePath string
	if file == "" {
		dir, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		defaultConfig := filepath.Join(dir, "config.toml")

		if !fExists(defaultConfig) {
			filePath = defaultConfig
		} else {
			return errors.New(defaultConfig + " already exists")
		}
	} else {
		abs, err := abspath.ExpandFrom(file)
		if err != nil {
			return err
		}

		filePath = abs.String()

		if !strings.HasSuffix(strings.ToLower(filePath), ".toml") {
			filePath = filePath + ".toml"
		}

		if fExists(filePath) {
			return errors.New(file + " already exists")
		}
	}

	config := setDefault()
	f, err := os.Create(filePath)
	defer f.Close()
	if err != nil {
		return err
	}

	if err := toml.NewEncoder(f).Encode(config); err != nil {
		return err
	}

	fmt.Println("Saved empty config to file " + filePath)

	return nil
}

func fExists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}
