package cwl

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/MG-RAST/AWE/lib/logger"
	"github.com/davecgh/go-spew/spew"
)

type WorkflowContext struct {
	CWL_document `yaml:",inline" json:",inline" bson:",inline" mapstructure:",squash"` // fields: CwlVersion, Base, Graph, Namespaces, Schemas (all interface-based !)
	Path         string
	//Namespaces   map[string]string
	//CWLVersion
	//CwlVersion CWLVersion    `yaml:"cwl_version"  json:"cwl_version" bson:"cwl_version" mapstructure:"cwl_version"`
	//CWL_graph  []interface{} `yaml:"cwl_graph"  json:"cwl_graph" bson:"cwl_graph" mapstructure:"cwl_graph"`
	// old ParsingContext
	If_objects map[string]interface{} `yaml:"-"  json:"-" bson:"-" mapstructure:"-"`
	Objects    map[string]CWL_object  `yaml:"-"  json:"-" bson:"-" mapstructure:"-"` // stores all objects (replaces All ???)

	Workflows          map[string]*Workflow          `yaml:"-"  json:"-" bson:"-" mapstructure:"-"`
	WorkflowStepInputs map[string]*WorkflowStepInput `yaml:"-"  json:"-" bson:"-" mapstructure:"-"`
	CommandLineTools   map[string]*CommandLineTool   `yaml:"-"  json:"-" bson:"-" mapstructure:"-"`
	ExpressionTools    map[string]*ExpressionTool    `yaml:"-"  json:"-" bson:"-" mapstructure:"-"`
	Files              map[string]*File              `yaml:"-"  json:"-" bson:"-" mapstructure:"-"`
	Strings            map[string]*String            `yaml:"-"  json:"-" bson:"-" mapstructure:"-"`
	Ints               map[string]*Int               `yaml:"-"  json:"-" bson:"-" mapstructure:"-"`
	Booleans           map[string]*Boolean           `yaml:"-"  json:"-" bson:"-" mapstructure:"-"`
	All                map[string]CWL_object         `yaml:"-"  json:"-" bson:"-" mapstructure:"-"` // everything goes in here
	//Job_input          *Job_document
	Job_input_map *JobDocMap `yaml:"-"  json:"-" bson:"-" mapstructure:"-"`

	Schemata    map[string]CWLType_Type `yaml:"-"  json:"-" bson:"-" mapstructure:"-"`
	Initialized bool                    `yaml:"-"  json:"-" bson:"-" mapstructure:"-"`
}

func NewWorkflowContext() (context *WorkflowContext) {
	context = &WorkflowContext{}

	return
}

// search for #main and create objects recursively
func (context *WorkflowContext) Init(entrypoint string) (err error) {

	logger.Debug(3, "(WorkflowContext/Init) start")
	if context.Initialized == true {
		err = fmt.Errorf("(WorkflowContext/Init) already initialized")
		return
	}

	if context.If_objects == nil {
		context.If_objects = make(map[string]interface{})
	}

	if context.Objects == nil {
		context.Objects = make(map[string]CWL_object)
	}

	if context.Workflows == nil {
		context.Workflows = make(map[string]*Workflow)
	}

	if context.WorkflowStepInputs == nil {
		context.WorkflowStepInputs = make(map[string]*WorkflowStepInput)
	}

	if context.CommandLineTools == nil {
		context.CommandLineTools = make(map[string]*CommandLineTool)
	}

	if context.ExpressionTools == nil {
		context.ExpressionTools = make(map[string]*ExpressionTool)
	}
	if context.Files == nil {
		context.Files = make(map[string]*File)
	}

	if context.Strings == nil {
		context.Strings = make(map[string]*String)
	}

	if context.Ints == nil {
		context.Ints = make(map[string]*Int)
	}

	if context.Booleans == nil {
		context.Booleans = make(map[string]*Boolean)
	}

	if context.All == nil {
		context.All = make(map[string]CWL_object)
	}

	if context.Schemata == nil {
		context.Schemata = make(map[string]CWLType_Type)
	}

	if context.CwlVersion == "" {
		err = fmt.Errorf("(WorkflowContext/Init) context.CwlVersion ==nil")
		return
	}

	//context := &WorkflowContext{}
	//var If_objects map[string]interface{}

	graph := context.CWL_document.Graph

	if len(graph) == 0 {
		err = fmt.Errorf("(WorkflowContext/Init) len(graph) == 0")
		return
	}

	logger.Debug(3, "(WorkflowContext/Init) len(graph): %d", len(graph))

	// put interface objetcs into map: populate context.If_objects
	for i, _ := range graph {

		//fmt.Printf("graph element type: %s\n", reflect.TypeOf(graph[i]))

		if graph[i] == nil {
			err = fmt.Errorf("(WorkflowContext/Init) graph[i] empty array element")
			return
		}

		var id string
		id, err = GetId(graph[i])
		if err != nil {
			fmt.Println("(WorkflowContext/Init) object without id:")
			spew.Dump(graph[i])
			return
		}
		//fmt.Printf("id=\"%s\\n", id)

		context.If_objects[id] = graph[i]

	}

	logger.Debug(3, "(WorkflowContext/Init) len(context.If_objects): %d", len(context.If_objects))

	main_if, has_main := context.If_objects[entrypoint] // "#main" or enrypoint
	if !has_main {
		var keys string
		for key, _ := range context.If_objects {
			keys += "," + key
		}
		err = fmt.Errorf("(WorkflowContext/Init) entrypoint %s not found in graph (found %s)", entrypoint, keys)
		return
	}

	// start with #main
	// recursivly add objects to context
	var object CWL_object
	var schemata_new []CWLType_Type
	object, schemata_new, err = New_CWL_object(main_if, nil, context)
	if err != nil {
		fmt.Printf("(WorkflowContext/Init) main_if")
		spew.Dump(main_if)
		err = fmt.Errorf("(WorkflowContext/Init) A New_CWL_object returned %s", err.Error())
		return
	}
	context.Objects[entrypoint] = object

	err = context.AddSchemata(schemata_new)
	if err != nil {
		err = fmt.Errorf("(WorkflowContext/Init) context.AddSchemata returned %s", err.Error())
		return
	}
	//for i, _ := range schemata_new {
	//	schemata = append(schemata, schemata_new[i])
	//}

	context.CWL_document.Graph = nil
	context.CWL_document.Graph = []interface{}{}
	for key, value := range context.Objects {
		logger.Debug(3, "(WorkflowContext/Init) adding %s to context.CWL_document.Graph", key)
		context.Add(key, value)
		context.CWL_document.Graph = append(context.CWL_document.Graph, value)
	}
	//fmt.Println("(WorkflowContext/Init) context.Objects: ")
	//spew.Dump(context.Objects)

	context.Initialized = true
	return
}

func (c WorkflowContext) Evaluate(raw string) (parsed string) {

	reg := regexp.MustCompile(`\$\([\w.]+\)`) // https://github.com/google/re2/wiki/Syntax

	parsed = raw
	for {

		matches := reg.FindAll([]byte(parsed), -1)
		fmt.Printf("Matches: %d\n", len(matches))
		if len(matches) == 0 {
			return parsed
		}
		for _, match := range matches {
			key := bytes.TrimPrefix(match, []byte("$("))
			key = bytes.TrimSuffix(key, []byte(")"))

			// trimming of inputs. is only a work-around
			key = bytes.TrimPrefix(key, []byte("inputs."))

			value_str := ""
			value, err := c.GetString(string(key))

			if err != nil {
				value_str = "<ERROR_NOT_FOUND:" + string(key) + ">"
			} else {
				value_str = value.String()
			}

			logger.Debug(1, "evaluate %s -> %s\n", key, value_str)
			parsed = strings.Replace(parsed, string(match), value_str, 1)
		}

	}

}

func (c WorkflowContext) AddSchemata(obj []CWLType_Type) (err error) {
	//fmt.Printf("(AddSchemata)\n")

	if c.Schemata == nil {
		c.Schemata = make(map[string]CWLType_Type)
	}

	for i, _ := range obj {
		id := obj[i].GetId()
		if id == "" {
			err = fmt.Errorf("id empty")
			return
		}

		//fmt.Printf("Adding %s\n", id)

		_, ok := c.Schemata[id]
		if ok {
			return
		}

		c.Schemata[id] = obj[i]
	}
	return
}

func (c WorkflowContext) GetSchemata() (obj []CWLType_Type, err error) {
	obj = []CWLType_Type{}
	for _, schema := range c.Schemata {
		obj = append(obj, schema)
	}
	return
}

func (c WorkflowContext) AddArray(object_array []Named_CWL_object) (err error) {

	for i, _ := range object_array {
		pair := object_array[i]

		err = c.Add(pair.Id, pair.Value)
		if err != nil {
			return
		}

	}

	return

}

func (c WorkflowContext) Add(id string, obj CWL_object) (err error) {

	if id == "" {
		err = fmt.Errorf("(WorkflowContext/Add) id is empty")
		return
	}

	logger.Debug(3, "(WorkflowContext/Add) Adding object %s to collection (type: %s)", id, reflect.TypeOf(obj))

	_, ok := c.All[id]
	if ok {
		err = fmt.Errorf("(WorkflowContext/Add) Object %s already in collection", id)
		return
	}

	switch obj.(type) {
	case *Workflow:
		c.Workflows[id] = obj.(*Workflow)
	case *WorkflowStepInput:
		c.WorkflowStepInputs[id] = obj.(*WorkflowStepInput)
	case *CommandLineTool:
		c.CommandLineTools[id] = obj.(*CommandLineTool)
	case *ExpressionTool:
		c.ExpressionTools[id] = obj.(*ExpressionTool)
	case *File:
		c.Files[id] = obj.(*File)
	case *String:
		c.Strings[id] = obj.(*String)
	case *Boolean:
		c.Booleans[id] = obj.(*Boolean)
	case *Int:
		obj_int, ok := obj.(*Int)
		if !ok {
			err = fmt.Errorf("could not make Int type assertion")
			return
		}
		c.Ints[id] = obj_int
	default:
		logger.Debug(3, "adding type %s to WorkflowContext.All", reflect.TypeOf(obj))
	}
	c.All[id] = obj

	return
}

func (c WorkflowContext) Get(id string) (obj CWL_object, err error) {
	obj, ok := c.All[id]
	if !ok {
		for k, _ := range c.All {
			logger.Debug(3, "collection: %s", k)
		}
		err = fmt.Errorf("(All) item %s not found in collection", id)
	}
	return
}

func (c WorkflowContext) GetCWLType(id string) (obj CWLType, err error) {
	var ok bool
	obj, ok = c.Files[id]
	if ok {
		return
	}
	obj, ok = c.Strings[id]
	if ok {
		return
	}

	obj, ok = c.Ints[id]
	if ok {
		return
	}
	obj, ok = c.Booleans[id]
	if ok {
		return
	}

	err = fmt.Errorf("(GetType) %s not found", id)
	return

}

func (c WorkflowContext) GetFile(id string) (obj *File, err error) {
	obj, ok := c.Files[id]
	if !ok {
		err = fmt.Errorf("(GetFile) item %s not found in collection", id)
	}
	return
}

func (c WorkflowContext) GetString(id string) (obj *String, err error) {
	obj, ok := c.Strings[id]
	if !ok {
		err = fmt.Errorf("(GetString) item %s not found in collection", id)
	}
	return
}

func (c WorkflowContext) GetInt(id string) (obj *Int, err error) {
	obj, ok := c.Ints[id]
	if !ok {
		err = fmt.Errorf("(GetInt) item %s not found in collection", id)
	}
	return
}

func (c WorkflowContext) GetWorkflowStepInput(id string) (obj *WorkflowStepInput, err error) {
	obj, ok := c.WorkflowStepInputs[id]
	if !ok {
		err = fmt.Errorf("(GetWorkflowStepInput) item %s not found in collection", id)
	}
	return
}

func (c WorkflowContext) GetCommandLineTool(id string) (obj *CommandLineTool, err error) {
	obj, ok := c.CommandLineTools[id]
	if !ok {
		err = fmt.Errorf("(GetCommandLineTool) item %s not found in collection", id)
	}
	return
}

func (c WorkflowContext) GetExpressionTool(id string) (obj *ExpressionTool, err error) {
	obj, ok := c.ExpressionTools[id]
	if !ok {
		err = fmt.Errorf("(GetExpressionTool) item %s not found in collection", id)
	}
	return
}

func (c WorkflowContext) GetWorkflow(id string) (obj *Workflow, err error) {
	obj, ok := c.Workflows[id]
	if !ok {
		err = fmt.Errorf("(GetWorkflow) item %s not found in collection", id)
	}
	return
}