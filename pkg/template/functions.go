package template

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

func Image(i map[string]string, arg ...string) tttt {
	return func(arg ...string) interface{} {
		return i[arg[0]]
	}
}

func ExternalImage(args ...string) string {
	out := bytes.Buffer{}
	cmd := exec.Command("docker", append([]string{"inspect", "--format='{{index .RepoDigests 0}}'"}, args...)...)
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(out.String())
	}
	log.Debugf("Replaced external image '%s' with '%s'", args[0], strings.Replace(out.String(), "\n", "", -1))
	return strings.Replace(out.String(), "\n", "", -1)
}

// TODO: Take context into account
func HasFile(sourcePath string) bool {
	_, err := os.Stat(sourcePath)
	return err == nil
}

func Concat(args ...string) string {
	return strings.Join(args, "")
}

func GitCommitShort() string {
	out := bytes.Buffer{}
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(out.String())
	}
	return strings.Replace(out.String(), "\n", "", -1)
}

type tttt func(arg ...string) interface{}

func MandatoryParameter(i map[string]interface{}, arg ...string) tttt {
	return func(arg ...string) interface{} {
		if len(arg[0]) > 0 && i[arg[0]] != nil {
			return i[arg[0]]
		} else {
			return arg[1]
		}
	}
}

func ParamOrDefault(i map[string]interface{}, arg ...string) tttt {
	return func(arg ...string) interface{} {
		if len(arg[0]) > 0 && i[arg[0]] != nil {
			return i[arg[0]]
		} else if len(arg) > 1 {
			return arg[1]
		}
		return ""
	}
}

func ParamOrFile(i map[string]interface{}, arg ...string) tttt {
	return func(arg ...string) interface{} {
		if len(arg[0]) > 0 && i[arg[0]] != nil {
			return i[arg[0]]
		} else {
			return strings.Replace(ReadFile(arg[1]), "\n", "", -1)
		}
	}
}

func ReadFile(sourcePath string) string {
	data, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		panic(err)
	}

	return string(data)
}
