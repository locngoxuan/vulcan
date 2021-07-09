package core

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type ProjectConfig struct {
	Name      string `yaml:"name,omitempty"`
	*OnConfig `yaml:"on,omitempty"`
	Jobs      map[string]*JobConfig `yaml:"jobs,omitempty"`
}

type OnConfig struct {
	Branch          []string `yaml:"branch,omitempty"`
	Event           []string `yaml:"event,omitempty"`
	*ScheduleConfig `yaml:"schedule,omitempty"`
}

type ScheduleConfig struct {
	Cron     string `yaml:"cron,omitempty"`
	Repeated string `yaml:"repeat,omitempty"`
}

const (
	RepeatedEveryDay  = "every-day"
	RepeatedEveryHour = "every-hour"
)

type ArgsConfig map[string]string

func (a ArgsConfig) ReplaceEnv() error {
	for k, v := range a {
		a[k] = ReadEnvVariableIfHas(v)
	}
	return nil
}

type JobConfig struct {
	Id        string       `yaml:"-"`
	Name      string       `yaml:"name,omitempty"`
	RunOn     string       `yaml:"run-on,omitempty"`
	BaseDir   string       `yaml:"base-dir,omitempty"`
	OS        string       `yaml:"os,omitempty"`
	Arch      string       `yaml:"arch,omitempty"`
	Artifacts []string     `yaml:"artifacts,omitempty"`
	Args      *ArgsConfig  `yaml:"args,omitempty"`
	Steps     []StepConfig `yaml:"steps,omitempty"`
}

type StepConfig struct {
	Id   string      `yaml:"id,omitempty"`
	Name string      `yaml:"name,omitempty"`
	Run  string      `yaml:"run,omitempty"`
	Use  string      `yaml:"use,omitempty"`
	Args *ArgsConfig `yaml:"args,omitempty"`
	With *ArgsConfig `yaml:"with,omitempty"`
}

func ReadProjectConfig(configFile string) (c ProjectConfig, err error) {
	_, err = os.Stat(configFile)
	if os.IsNotExist(err) {
		err = fmt.Errorf("configuration file not found")
		return
	}

	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		err = fmt.Errorf("read application config file get error %v", err)
		return
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		err = fmt.Errorf("unmarshal application config file get error %v", err)
		return
	}
	return
}

//Docker config
type DockerConfig struct {
	Hosts      []string         `yaml:"hosts,omitempty"`
	Registries []RegistryConfig `yaml:"registries,omitempty"`
}
type RegistryConfig struct {
	Address  string `yaml:"address,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}
