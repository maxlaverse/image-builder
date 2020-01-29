package template

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
)

type TemplateContext struct {
	data TemplateData
}

func NewTemplateContext(data TemplateData) TemplateContext {
	return TemplateContext{
		data: data,
	}
}

func (t *TemplateContext) BuilderStage(arg ...string) string {
	t.data.dependencies[arg[0]] = struct{}{}
	return t.data.Images[arg[0]]
}

func (t *TemplateContext) ExternalImage(args ...string) string {
	//TODO: Replace with registry lookup
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

func (t *TemplateContext) HasFile(sourcePath string) bool {
	_, err := os.Stat(path.Join(t.data.localContext, sourcePath))
	return err == nil
}

func (t *TemplateContext) Concat(args ...string) string {
	return strings.Join(args, "")
}

// TODO: Use an externally provided executor
func (t *TemplateContext) GitCommitShort() string {
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

func (t *TemplateContext) MandatoryParameter(args ...string) interface{} {
	if len(args[0]) > 0 && t.data.Build.Spec[args[0]] != nil {
		return t.data.Build.Spec[args[0]]
	} else {
		return args[1]
	}
}

func (t *TemplateContext) ParamOrDefault(args ...string) interface{} {
	if len(args[0]) > 0 && t.data.Build.Spec[args[0]] != nil {
		return t.data.Build.Spec[args[0]]
	} else if len(args) > 1 {
		return args[1]
	}
	return ""
}

func (t *TemplateContext) ParamOrFile(args ...string) interface{} {
	if len(args[0]) > 0 && t.data.Build.Spec[args[0]] != nil {
		return t.data.Build.Spec[args[0]]
	} else {
		return strings.Replace(t.readFile(args[1]), "\n", "", -1)
	}
}

func (t *TemplateContext) readFile(sourcePath string) string {
	data, err := ioutil.ReadFile(path.Join(t.data.localContext, sourcePath))
	if err != nil {
		panic(err)
	}

	return string(data)
}
