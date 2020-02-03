package template

import (
	"regexp"
	"strings"
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
	content string
	data    map[string][]string
	deps    []string
}

// Dockerfile is the interface for a parsed Dockerfile
type Dockerfile interface {
	GetContent() string
	GetContentWithoutIgnoredLines() string
	GetDockerIgnores() []string
	GetFriendlyTag() string
	GetTagAliases() []string
	GetRequiredStages() []string
	UseBuilderContext() bool
}

// DockerfileFromContent returns a new parsed Dockerfile
func DockerfileFromContent(content []byte, deps []string) Dockerfile {
	return &dockerfile{
		content: string(content),
		data:    parseDirectives(content),
		deps:    deps,
	}
}

// Returns the rendered content of the Dockerfile
func (d *dockerfile) GetContent() string {
	return d.content
}

// Returns the friendly rendered content
func (d *dockerfile) GetContentWithoutIgnoredLines() string {
	filteredLines := []string{}
	lines := strings.Split(d.content, "\n")
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
	return d.deps
}

// UseBuilderContext returns if the Dockerfile specified the builder context should be used
func (d *dockerfile) UseBuilderContext() bool {
	_, ok := d.data[dirUseBuilderContext]
	return ok
}

func parseDirectives(content []byte) map[string][]string {
	data := map[string][]string{}
	res := regExpDirectives.FindAllSubmatch(content, -1)
	for _, line := range res {
		name := string(line[1])
		if _, ok := data[name]; !ok {
			data[name] = []string{}
		}
		data[name] = append(data[name], string(line[2]))
	}
	return data
}
