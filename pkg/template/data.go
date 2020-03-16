package template

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/maxlaverse/image-builder/pkg/config"
	"github.com/maxlaverse/image-builder/pkg/executor"
	"github.com/maxlaverse/image-builder/pkg/registry"
	"github.com/maxlaverse/image-builder/pkg/utils"
	log "github.com/sirupsen/logrus"
)

type StageResolver func(string) (string, error)

// data represents the buildData provided to the Go templating
// engine when rendering Dockerfiles
type data struct {
	buildConf      config.BuildConfiguration
	currentContext string
	deps           map[string]struct{}
	exec           executor.Executor
	resolver       StageResolver
}

// newTemplateData returns a new instance of Data
func newTemplateData(buildConf config.BuildConfiguration, currentContext string, resolver StageResolver, exec executor.Executor) data {
	return data{
		buildConf:      buildConf,
		currentContext: currentContext,
		resolver:       resolver,
		exec:           exec,
		deps:           map[string]struct{}{},
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
	d.deps[stageName] = struct{}{}
	imageURL, err := d.resolver(stageName)
	if err != nil {
		err = fmt.Errorf("cannot replace BuilderStage('%s'): %w", stageName, err)
		log.Error(err)
		panic(err)
	}
	if len(imageURL) == 0 {
		return fmt.Sprintf("{{ BuilderStage \"%s\"}}", stageName)
	}
	log.Debugf("Replacing BuilderStage('%s') with '%s'", stageName, imageURL)
	return imageURL
}

// ExternalImage returns an imageURL referenced by its sha256
func (d *data) ExternalImage(imageURL string) string {
	digest, err := registry.ImageWithDigest(imageURL)
	if err != nil {
		log.Fatalf("Error when calling ImageWithDigest('%s'): %w", imageURL, err)
	}

	log.Debugf("Replacing ExternalImage('%s') with '%s'", imageURL, digest)
	return digest
}

// ImageAgeGeneration returns the age of an image
func (d *data) ImageAgeGeneration(imageURL, generation string) float64 {
	// TODO: Cache result of this call
	age, err := registry.ImageAge(imageURL)
	if err != nil {
		log.Fatalf("Error when calling ImageAge('%s'): %w", imageURL, err)
	}
	b, err := time.ParseDuration(generation)
	if err != nil {
		log.Fatalf("Error when calling ParseDuration('%s'): %w", generation, err)
	}
	return math.Floor(age.Seconds() / b.Seconds())
}

// HasFile returns whether a file exist in the local context or not
func (d *data) HasFile(filePath string) bool {
	_, err := os.Stat(path.Join(d.currentContext, filePath))
	return err == nil
}

// Concat concats an array of string together
func (d *data) Concat(args ...string) string {
	return strings.Join(args, "")
}

// GitCommitShort returns the git commit of the local context
func (d *data) GitCommitShort() string {
	out := bytes.Buffer{}
	err := d.exec.NewCommand("git", "rev-parse", "HEAD").WithDir(d.currentContext).WithCombinedOutput(&out).Run()
	if err != nil {
		log.Fatal(out.String())
	}
	return utils.Chomp(out.String())
}

// MandatoryParameter returns a parameter from ImageSpec or fails
func (d *data) MandatoryParameter(parameterName string) interface{} {
	value, ok := d.buildConf.ImageSpec[parameterName]
	if !ok {
		log.Fatalf("Could not find mandatory parameter in: %v", d.buildConf.ImageSpec)
	}
	return value
}

// ParameterWithOptionalDefault returns a parameter from ImageSpec or a default value
func (d *data) ParameterWithOptionalDefault(args ...string) interface{} {
	if len(args[0]) > 0 && d.buildConf.ImageSpec[args[0]] != nil {
		return d.buildConf.ImageSpec[args[0]]
	} else if len(args) > 1 {
		return args[1]
	}
	return ""
}

// File read the content of a file from the local context
func (d *data) File(filePath string) interface{} {
	buildData, err := ioutil.ReadFile(path.Join(d.currentContext, filePath))
	if err != nil {
		panic(err)
	}
	return utils.Chomp(string(buildData))
}
