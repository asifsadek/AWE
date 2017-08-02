package cwl

import (
	"fmt"
	cwl_types "github.com/MG-RAST/AWE/lib/core/cwl/types"
	"github.com/mitchellh/mapstructure"
)

type CommandOutputBinding struct {
	Glob         []cwl_types.Expression `yaml:"glob" bson:"glob" json:"glob"`
	LoadContents bool                   `yaml:"loadContents" bson:"loadContents" json:"loadContents"`
	OutputEval   cwl_types.Expression   `yaml:"outputEval" bson:"outputEval" json:"outputEval"`
}

func NewCommandOutputBinding(original interface{}) (commandOutputBinding *CommandOutputBinding, err error) {

	switch original.(type) {

	case map[interface{}]interface{}:
		original_map := original.(map[interface{}]interface{})

		glob, ok := original_map["glob"]
		if ok {
			original_map["glob"], err = cwl_types.NewExpressionArray(glob)
			if err != nil {
				return
			}
		}
		outputEval, ok := original_map["outputEval"]
		if ok {
			original_map["outputEval"], err = cwl_types.NewExpression(outputEval)
			if err != nil {
				return
			}
		}
	default:
		err = fmt.Errorf("NewCommandOutputBinding: type unknown")
		return
	}

	commandOutputBinding = &CommandOutputBinding{}
	err = mapstructure.Decode(original, &commandOutputBinding)
	if err != nil {
		err = fmt.Errorf("(NewCommandOutputBinding) %s", err.Error())
		return
	}
	//output_parameter.OutputBinding = outputBinding

	return
}
