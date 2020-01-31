package template

import (
	"bytes"
	"io/ioutil"
	"text/template"

	"github.com/maxlaverse/image-builder/pkg/executor"
	log "github.com/sirupsen/logrus"
)

// RenderDockerfile renders a given Dockerfile based on provided BuildData
func RenderDockerfile(sourcePath string, data BuildData) (Dockerfile, error) {
	content, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		log.Errorf("Failed to read the Dockerfile template: %v", err)
		return nil, err
	}

	templateData := newTemplateData(data, executor.New())
	tmpl, err := template.New("dockerfile").Funcs(templateData.FuncMaps()).Parse(string(content))
	if err != nil {
		log.Errorf("Fatal to parse the Dockerfile template: %v, %s", err, string(content))
		return nil, err
	}

	buffer := bytes.NewBufferString("")
	err = tmpl.Execute(buffer, data)
	if err != nil {
		log.Errorf("Error to render the Dockerfile template: %s, %s", err, string(content))
		return nil, err
	}

	dockerfile := DockerfileFromContent(buffer.Bytes(), data.dependencies())
	return dockerfile, nil
}
