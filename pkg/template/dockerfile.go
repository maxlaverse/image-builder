package template

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"text/template"

	"github.com/maxlaverse/image-builder/pkg/config"
	"github.com/maxlaverse/image-builder/pkg/executor"
)

const (
	// dirFriendlyTag allows to specify a suffix for the generated tag
	dirFriendlyTag = "FriendlyTag"

	// dirTagAlias allows to specify additional tags
	dirTagAlias = "TagAlias"

	// dirContentHashIgnoreNextLine tells the content ahsing algorithm to ignore the
	// next line
	dirContentHashIgnoreNextLine = "ContentHashIgnoreNextLie"

	// dirDockerIgnore automatically excludes files from the Docker context
	dirDockerIgnore = "DockerIgnore"

	// dirUseBuilderContext changes the build context for the directory where the builder
	// is defined
	dirUseBuilderContext = "UseBuilderContext"
)

var (
	// regExpDirectives is the regular expression to parse those directives
	regExpDirectives = regexp.MustCompile(`# ([a-zA-Z]+)(?: (.*))?`)
)

type dockerfile struct {
	builderContext string
	content        *bytes.Buffer
	currentContext string
	data           map[string][]string
	templateData   data
}

// Dockerfile is the interface for a parsed Dockerfile
type Dockerfile interface {
	GetBuildContext() string
	GetContent() string
	GetContentWithoutIgnoredLines() string
	GetDockerIgnores() []string
	GetFriendlyTag() string
	GetTagAliases() []string
	GetRequiredStages() []string
	Render() error
}

// NewDockerfile renders a given Dockerfile based on provided BuildData
func NewDockerfile(content []byte, buildConf config.BuildConfiguration, currentContext, builderContext string, resolver StageResolver, exec executor.Executor) Dockerfile {
	return &dockerfile{
		builderContext: builderContext,
		content:        bytes.NewBuffer(content),
		currentContext: currentContext,
		data:           map[string][]string{},
		templateData:   newTemplateData(buildConf, currentContext, resolver, exec),
	}
}

// NewDockerfileFromFile renders a given Dockerfile based on provided BuildData
func NewDockerfileFromFile(filepath string, buildConf config.BuildConfiguration, currentContext, builderContext string, resolver StageResolver, exec executor.Executor) (Dockerfile, error) {
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read the Dockerfile template: %v", err)
	}
	return NewDockerfile(content, buildConf, currentContext, builderContext, resolver, exec), nil
}

func (d *dockerfile) Render() error {
	tmpl, err := template.New("dockerfile").Funcs(d.templateData.FuncMaps()).Parse(d.content.String())
	if err != nil {
		return fmt.Errorf("failed to parse the Dockerfile template: %w, %s", err, d.content)
	}

	newContent := bytes.NewBufferString("")
	err = tmpl.Execute(newContent, d.templateData)
	if err != nil {
		return fmt.Errorf("failed to render the Dockerfile template: %w", err)
	}

	d.content = newContent
	if len(d.data) == 0 {
		d.parseDirectives(d.content.Bytes())
	}

	if d.useBuilderContext() {
		d.currentContext = d.builderContext
	}
	return nil
}

// Returns the rendered content of the Dockerfile
func (d *dockerfile) GetContent() string {
	return d.content.String()
}

// Returns the friendly rendered content
func (d *dockerfile) GetContentWithoutIgnoredLines() string {
	filteredLines := []string{}
	lines := strings.Split(d.content.String(), "\n")
	skip := false
	for _, line := range lines {
		if skip {
			skip = false
			line = "# THIS LINE HAS BEEN AUTOMATICALLY REMOVED"
		} else if strings.HasPrefix(line, "# "+dirContentHashIgnoreNextLine) {
			skip = true
		}
		filteredLines = append(filteredLines, line)
	}
	return strings.Join(filteredLines, "\n")
}

// GetDockerIgnores returns the list of files to ignore when loading the build context
func (d *dockerfile) GetDockerIgnores() []string {
	if d.data[dirDockerIgnore] == nil {
		return []string{}
	}
	return d.data[dirDockerIgnore]
}

// GetTagAliases returns the list of tag aliases
func (d *dockerfile) GetTagAliases() []string {
	if d.data[dirTagAlias] == nil {
		return []string{}
	}
	return d.data[dirTagAlias]
}

// GetFriendlyTag returns the friendly tag
func (d *dockerfile) GetFriendlyTag() string {
	if d.data[dirFriendlyTag] == nil {
		return ""
	}
	return d.data[dirFriendlyTag][0]
}

// GetRequiredStages returns the dependency of the Dockerfile
func (d *dockerfile) GetRequiredStages() []string {
	deps := []string{}
	for dep := range d.templateData.deps {
		deps = append(deps, dep)
	}
	return deps
}

// useBuilderContext returns if the Dockerfile specified the builder context should be used
func (d *dockerfile) useBuilderContext() bool {
	_, ok := d.data[dirUseBuilderContext]
	return ok
}

// GetBuildContext returns the friendly tag
func (d *dockerfile) GetBuildContext() string {
	return d.currentContext
}

func (d *dockerfile) parseDirectives(content []byte) {
	res := regExpDirectives.FindAllSubmatch(content, -1)
	for _, line := range res {
		name := string(line[1])
		if _, ok := d.data[name]; !ok {
			d.data[name] = []string{}
		}
		d.data[name] = append(d.data[name], string(line[2]))
	}
}
