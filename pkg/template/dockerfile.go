package template

import "strings"

const (
	dirFriendlyTag       = "FriendlyTag"
	dirContentHashIgnore = "ContentHashIgnore"
	dirDockerIgnore      = "DockerIgnore"
	dirUseBuilderContext = "UseBuilderContext"
)

type Dockerfile struct {
	content []byte
	data    map[string][]string
	deps    []string
}

func DockerfileFromContent(content []byte, deps []string) Dockerfile {
	d := Dockerfile{
		content: content,
		data:    map[string][]string{},
		deps:    deps,
	}
	d.parseDirectives()
	return d
}

func (d *Dockerfile) parseDirectives() {
	res := regExpDirectives.FindAllSubmatch(d.content, -1)
	for _, line := range res {
		name := string(line[1])
		if _, ok := d.data[name]; !ok {
			d.data[name] = []string{}
		}
		d.data[name] = append(d.data[name], string(line[2]))
	}
}

func (d *Dockerfile) GetContent() string {
	return string(d.content)
}

func (d *Dockerfile) GetFilteredContent() string {
	lines := []string{}
	curLines := strings.Split(string(d.content), "\n")
	skip := false
	for _, v := range curLines {
		if strings.Contains(v, "# ContentHashIgnore") {
			skip = true
			continue
		}
		if skip {
			skip = false
			continue
		}
		lines = append(lines, v)
	}
	return strings.Join(lines, "\n")
}

func (d *Dockerfile) UseBuilderContext() bool {
	_, ok := d.data[dirUseBuilderContext]
	return ok
}

func (d *Dockerfile) GetContentHashIgnores() []string {
	if d.data[dirContentHashIgnore] != nil {
		return d.data[dirContentHashIgnore]
	}
	return []string{}
}

func (d *Dockerfile) GetDockerIgnores() []string {
	if d.data[dirDockerIgnore] != nil {
		return d.data[dirDockerIgnore]
	}
	return []string{}
}

func (d *Dockerfile) GetFriendlyTag() string {
	if d.data[dirFriendlyTag] != nil {
		return d.data[dirFriendlyTag][0]
	}
	return ""
}

func (d *Dockerfile) GetRequiredStages() []string {
	return d.deps
}
