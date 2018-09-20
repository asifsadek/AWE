package cwl

import (
	"fmt"
	"reflect"

	"github.com/davecgh/go-spew/spew"
	"github.com/mitchellh/mapstructure"
)

// https://www.commonwl.org/v1.0/Workflow.html#WorkflowOutputParameter
type WorkflowOutputParameter struct {
	OutputParameter `yaml:",inline" json:",inline" bson:",inline" mapstructure:",squash"` // provides Id, Label, SecondaryFiles, Format, Streamable, OutputBinding, Type
	Doc             string                                                                `yaml:"doc,omitempty" bson:"doc,omitempty" json:"doc,omitempty"`
	OutputSource    interface{}                                                           `yaml:"outputSource,omitempty" bson:"outputSource,omitempty" json:"outputSource,omitempty"` //string or []string
	LinkMerge       LinkMergeMethod                                                       `yaml:"linkMerge,omitempty" bson:"linkMerge,omitempty" json:"linkMerge,omitempty"`
}

func NewWorkflowOutputParameter(original interface{}, schemata []CWLType_Type, context *WorkflowContext) (wop *WorkflowOutputParameter, err error) {
	var output_parameter WorkflowOutputParameter

	original, err = MakeStringMap(original, context)
	if err != nil {
		return
	}

	var op *OutputParameter
	op, err = NewOutputParameterFromInterface(original, schemata, "Output", context)
	if err != nil {
		err = fmt.Errorf("(NewWorkflowOutputParameter) NewOutputParameterFromInterface returns %s", err.Error())
		return
	}

	switch original.(type) {

	case map[string]interface{}:
		original_map, ok := original.(map[string]interface{})
		if !ok {
			err = fmt.Errorf("(NewWorkflowOutputParameter) type switch error")
			return
		}

		outputSource_if, ok := original_map["outputSource"]
		if ok {

			switch outputSource_if.(type) {
			case string:
				original_map["outputSource"] = outputSource_if.(string)

			case []string:
				original_map["outputSource"] = outputSource_if.([]string)
			case []interface{}:
				outputSource_if_array := outputSource_if.([]interface{})

				outputSource_string_array := []string{}
				for _, elem := range outputSource_if_array {
					elem_str, ok := elem.(string)
					if !ok {
						err = fmt.Errorf("(NewWorkflowOutputParameter) not a string ?!")
						return
					}
					outputSource_string_array = append(outputSource_string_array, elem_str)
				}

				original_map["outputSource"] = outputSource_string_array

			default:

				outputSource_str, ok := outputSource_if.(string)
				if ok {
					original_map["outputSource"] = []string{outputSource_str}
				} else {

					spew.Dump(outputSource_if)
					err = fmt.Errorf("(NewWorkflowOutputParameter) type of outputSource_if unknown: %s", reflect.TypeOf(outputSource_if))
					return
				}
			}

		}

		// wop_type, has_type := original_map["type"]
		// if has_type {

		// 	wop_type_array, xerr := NewWorkflowOutputParameterTypeArray(wop_type, schemata)
		// 	if xerr != nil {
		// 		err = fmt.Errorf("from NewWorkflowOutputParameterTypeArray: %s", xerr.Error())
		// 		return
		// 	}
		// 	//fmt.Println("type of wop_type_array")
		// 	//fmt.Println(reflect.TypeOf(wop_type_array))
		// 	//fmt.Println("original:")
		// 	//spew.Dump(original)
		// 	//fmt.Println("wop_type_array:")
		// 	//spew.Dump(wop_type_array)
		// 	original_map["type"] = wop_type_array

		// }

		err = mapstructure.Decode(original, &output_parameter)
		if err != nil {
			err = fmt.Errorf("(NewWorkflowOutputParameter) decode error: %s", err.Error())
			return
		}
		wop = &output_parameter

		wop.OutputParameter = *op

	default:
		err = fmt.Errorf("(NewWorkflowOutputParameter) type unknown, %s", reflect.TypeOf(original))
		return

	}

	return
}

// WorkflowOutputParameter
func NewWorkflowOutputParameterArray(original interface{}, schemata []CWLType_Type, context *WorkflowContext) (new_array_ptr *[]WorkflowOutputParameter, err error) {

	new_array := []WorkflowOutputParameter{}
	switch original.(type) {
	case map[interface{}]interface{}:
		for k, v := range original.(map[interface{}]interface{}) {
			//fmt.Printf("A")

			output_parameter, xerr := NewWorkflowOutputParameter(v, schemata, context)
			if xerr != nil {
				err = xerr
				return
			}
			output_parameter.Id = k.(string)
			//fmt.Printf("C")
			new_array = append(new_array, *output_parameter)
			//fmt.Printf("D")

		}
		new_array_ptr = &new_array
		return
	case []interface{}:

		for _, v := range original.([]interface{}) {
			//fmt.Printf("A")

			output_parameter, xerr := NewWorkflowOutputParameter(v, schemata, context)
			if xerr != nil {
				err = xerr
				return
			}
			//output_parameter.Id = k.(string)
			//fmt.Printf("C")
			new_array = append(new_array, *output_parameter)
			//fmt.Printf("D")

		}
		new_array_ptr = &new_array
		return

	default:
		spew.Dump(new_array)
		err = fmt.Errorf("(NewWorkflowOutputParameterArray) type %s unknown", reflect.TypeOf(original))
	}
	//spew.Dump(new_array)
	return
}
