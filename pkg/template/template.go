package template

import (
	"bytes"
	"io/ioutil"
	"regexp"
	"text/template"

	"github.com/maxlaverse/image-builder/pkg/config"
	log "github.com/sirupsen/logrus"
)

var (
	regExpDirectives = regexp.MustCompile(`# ([a-zA-Z]+) (.*)`)
)

type TemplateContext struct {
	Build  config.BuildConfiguration
	Images map[string]string
}

func NewMinimalTemplateContext(buildConf config.BuildConfiguration) TemplateContext {
	return TemplateContext{
		Build:  buildConf,
		Images: map[string]string{},
	}
}

func NewTemplateContext(buildConf config.BuildConfiguration, image map[string]string) TemplateContext {
	return TemplateContext{
		Build:  buildConf,
		Images: image,
	}
}

func RenderDockerfile(sourcePath string, i TemplateContext) (*Dockerfile, error) {

	data, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		log.Errorf("Failed reading the template: %v", err)
		return nil, err
	}

	tmpl, err := template.New("name").Funcs(template.FuncMap{
		"GitCommitShort":     GitCommitShort,
		"Image":              Image(i.Images),
		"ExternalImage":      ExternalImage,
		"Concat":             Concat,
		"HasFile":            HasFile,
		"Parameter":          ParamOrDefault(i.Build.Spec),
		"MandatoryParameter": MandatoryParameter(i.Build.Spec),
		"ReadFile":           ReadFile,
		"ParamOrFile":        ParamOrFile(i.Build.Spec),
	}).Parse(string(data))
	if err != nil {
		log.Errorf("Fatal rendering error: %v, %s", err, string(data))
		return nil, err
	}

	buffer := bytes.NewBufferString("")
	err = tmpl.Execute(buffer, i)
	if err != nil {
		log.Errorf("Error executing template: %v, %v", err, string(data))
		return nil, err
	}

	dockerfile := DockerfileFromContent(buffer.Bytes())

	return &dockerfile, nil
}
