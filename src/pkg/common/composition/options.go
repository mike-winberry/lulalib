package composition

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mike-winberry/lulalib/src/cmd/common"
	"github.com/mike-winberry/lulalib/src/internal/template"
	"github.com/mike-winberry/lulalib/src/pkg/message"
)

type Option func(*Composer) error

// TODO: add remote option?
func WithModelFromLocalPath(path string) Option {
	return func(c *Composer) error {
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			return fmt.Errorf("input-file: %v does not exist - unable to digest document", path)
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("error getting absolute path: %v", err)
		}
		c.modelDir = filepath.Dir(absPath)

		return nil
	}
}

func WithRenderSettings(renderTypeString string, renderValidations bool) Option {
	return func(c *Composer) error {
		if renderTypeString == "" {
			c.renderTemplate = false
			c.renderValidations = false
			if renderValidations {
				message.Warn("`render` not specified, `render-validations` will be ignored")
			}
			return nil
		}
		c.renderTemplate = true
		c.renderValidations = renderValidations

		// Get the template render type
		renderType, err := template.ParseRenderType(renderTypeString)
		if err != nil {
			message.Warnf("invalid render type, defaulting to non-sensitive: %v", err)
			renderType = template.NONSENSITIVE
		}
		c.renderType = renderType

		return nil
	}
}

func WithTemplateRenderer(renderTypeString string, constants map[string]interface{}, variables []template.VariableConfig, setOpts []string) Option {
	return func(c *Composer) error {
		if renderTypeString == "" {
			c.renderTemplate = false
			if len(setOpts) > 0 {
				message.Warn("`render` not specified, the --set options will be ignored")
			}
			return nil
		}

		// Get overrides from setOpts flag
		overrides, err := common.ParseTemplateOverrides(setOpts)
		if err != nil {
			return fmt.Errorf("error parsing template overrides: %v", err)
		}

		// Handles merging viper config file data + environment variables
		// Throws an error if config keys are invalid for templating
		templateData, err := template.CollectTemplatingData(constants, variables, overrides)
		if err != nil {
			return fmt.Errorf("error collecting templating data: %v", err)
		}

		// need to update the template with the templateString...
		tr := template.NewTemplateRenderer(templateData)

		c.templateRenderer = tr

		return nil
	}
}
