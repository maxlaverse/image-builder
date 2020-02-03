package template

import (
	"bytes"
	"io/ioutil"
	"math"
	"os"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/maxlaverse/image-builder/pkg/executor"
	"github.com/maxlaverse/image-builder/pkg/registry"
	"github.com/maxlaverse/image-builder/pkg/utils"
	log "github.com/sirupsen/logrus"
)

// data represents the buildData provided to the Go templating
// engine when rendering Dockerfiles
type data struct {
	buildData BuildData
	exec      executor.Executor
}

// newTemplateData returns a new instance of Data
func newTemplateData(buildData BuildData, exec executor.Executor) data {
	return data{
		buildData: buildData,
		exec:      exec,
	}
}

// FuncMaps returns all functions that are available from
// inside a Dockerfile template
func (d *data) FuncMaps() template.FuncMap {
	return template.FuncMap{
		"BuilderStage":       d.BuilderStage,
		"Concat":             d.Concat,
		"ExternalImage":      d.ExternalImage,
		"ImageAgeGeneration": d.ImageAgeGeneration,
		"GitCommitShort":     d.GitCommitShort,
		"HasFile":            d.HasFile,
		"MandatoryParameter": d.MandatoryParameter,
		"Parameter":          d.ParameterWithOptionalDefault,
		"File":               d.File,
	}
}

// BuilderStage returns the imageURL corresponding to a given stage
func (d *data) BuilderStage(stageName string) string {
	d.buildData.deps[stageName] = struct{}{}
	return d.buildData.images[stageName]
}

// ExternalImage returns an imageURL referenced by its sha256
func (d *data) ExternalImage(imageURL string) string {
	digest, err := registry.ImageWithDigest(imageURL)
	if err != nil {
		log.Fatalf("Error while calling ImageWithDigest for '%s': %v", imageURL, err)
	}

	log.Debugf("Replaced imageURL '%s' with '%s'", imageURL, digest)
	return digest
}

// ImageAgeGeneration returns the age of an image
func (d *data) ImageAgeGeneration(imageURL, generation string) float64 {
	age, err := registry.ImageAge(imageURL)
	if err != nil {
		log.Fatalf("Error while calling ImageAge for '%s': %v", imageURL, err)
	}
	b, err := time.ParseDuration(generation)
	if err != nil {
		log.Fatalf("Error while ParseDuration '%s': %v", generation, err)
	}
	return math.Floor(age.Seconds() / b.Seconds())
}

// HasFile returns whether a file exist in the local context or not
func (d *data) HasFile(filePath string) bool {
	_, err := os.Stat(path.Join(d.buildData.localContext, filePath))
	return err == nil
}

// Concat concats an array of string together
func (d *data) Concat(args ...string) string {
	return strings.Join(args, "")
}

// GitCommitShort returns the git commit of the local context
func (d *data) GitCommitShort() string {
	out := bytes.Buffer{}
	err := d.exec.NewCommand("git", "rev-parse", "HEAD").WithDir(d.buildData.localContext).WithCombinedOutput(&out).Run()
	if err != nil {
		log.Fatal(out.String())
	}
	return utils.Chomp(out.String())
}

// MandatoryParameter returns a parameter from ImageSpec or fails
func (d *data) MandatoryParameter(parameterName string) interface{} {
	value, ok := d.buildData.build.ImageSpec[parameterName]
	if !ok {
		log.Fatalf("Could not find mandatory parameter in: %v", d.buildData.build.ImageSpec)
	}
	return value
}

// ParameterWithOptionalDefault returns a parameter from ImageSpec or a default value
func (d *data) ParameterWithOptionalDefault(args ...string) interface{} {
	if len(args[0]) > 0 && d.buildData.build.ImageSpec[args[0]] != nil {
		return d.buildData.build.ImageSpec[args[0]]
	} else if len(args) > 1 {
		return args[1]
	}
	return ""
}

// File read the content of a file from the local context
func (d *data) File(filePath string) interface{} {
	buildData, err := ioutil.ReadFile(path.Join(d.buildData.localContext, filePath))
	if err != nil {
		panic(err)
	}
	return utils.Chomp(string(buildData))
}
