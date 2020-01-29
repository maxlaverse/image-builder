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

type TemplateData struct {
	Build        config.BuildConfiguration
	Images       map[string]string
	dependencies map[string]struct{}
}

func NewMinimalTemplateData(buildConf config.BuildConfiguration) TemplateData {
	return TemplateData{
		Build:        buildConf,
		Images:       map[string]string{},
		dependencies: map[string]struct{}{},
	}
}

func NewTemplateData(buildConf config.BuildConfiguration, image map[string]string) TemplateData {
	return TemplateData{
		Build:        buildConf,
		Images:       image,
		dependencies: map[string]struct{}{},
	}
}

//TODO: Extract in a util package
func (t *TemplateData) Dependencies() []string {
	result := []string{}
	for k := range t.dependencies {
		result = append(result, k)
	}
	return result
}

func RenderDockerfile(sourcePath string, i TemplateData) (*Dockerfile, error) {

	data, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		log.Errorf("Failed reading the template: %v", err)
		return nil, err
	}

	o := NewTemplateContext(i)
	tmpl, err := template.New("name").Funcs(template.FuncMap{
		"BuilderStage":       o.BuilderStage,
		"Concat":             o.Concat,
		"ExternalImage":      o.ExternalImage,
		"GitCommitShort":     o.GitCommitShort,
		"HasFile":            o.HasFile,
		"MandatoryParameter": o.MandatoryParameter,
		"Parameter":          o.ParamOrDefault,
		"ParamOrFile":        o.ParamOrFile,
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

	dockerfile := DockerfileFromContent(buffer.Bytes(), i.Dependencies())

	return &dockerfile, nil
}
