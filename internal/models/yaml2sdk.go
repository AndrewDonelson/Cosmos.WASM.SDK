package models

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/AndrewDonelson/Cosmos.WASM.SDK/internal/common"
	"github.com/janeczku/go-spinner"
	"gopkg.in/yaml.v2"
)

type Yaml2SDK struct {
	file       string
	yaml       MapIntInt
	source     map[string]Endpoints
	intMethods []string
}

func NewYaml2SDK(file string) *Yaml2SDK {
	obj := &Yaml2SDK{
		file: file,
	}
	obj.source = make(map[string]Endpoints)

	err := obj.loadYaml(file)
	if err != nil {
		fmt.Println("Error loading YAML file: ", err)
		return nil
	}

	obj.parseYaml()

	obj.generateSDK()

	return obj
}

func (y *Yaml2SDK) loadYaml(file string) error {
	y.yaml = make(MapIntInt)

	yfile, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(yfile, &y.yaml)
	if err != nil {
		return err
	}

	fmt.Println("Loaded YAML file: ", file)
	return err
}

func (y *Yaml2SDK) parseYaml() error {

	for rk, rv := range y.yaml {
		if rk == "paths" {
			paths := rv.(MapIntInt)
			for pk, pv := range paths {
				ep := Endpoints{}
				ep.Path = pk.(string)
				ep.Tags, _ = y.getTags(ep.Path)
				for dk, dv := range pv.(MapIntInt) {
					ep.Action = strings.ToUpper(dk.(string))
					ep.OperationID, ep.Description, _ = y.getPath(dv)
					ep.Parameters = *y.getParameters(dv.(MapIntInt))
				}

				y.source[ep.OperationID] = ep
			}
		} else if rk == "definitions" {
			defs := rv.(MapIntInt)
			for _, dv := range defs {
				if dValue, exists := dv.(MapIntInt); exists {
					def := YAMLDefinition{}
					def.Type = dValue["type"].(string)
					if dDesc, exists := dv.(MapIntInt)["description"].(string); exists {
						def.Description = dDesc
					}
					def.Properties = make([]YAMLProperty, 0)
					for pk, pv := range dValue {
						if pk == "properties" {
							for pk, pv := range pv.(MapIntInt) {
								prop := YAMLProperty{}
								prop.Name = pk.(string)

								if pType, exists := pv.(MapIntInt)["type"].(string); exists {
									prop.Type = pType
								}

								if pDesc, exists := pv.(MapIntInt)["description"].(string); exists {
									prop.Description = pDesc
								}
								def.Properties = append(def.Properties, prop)
							}
						}
					}
				}
			}

		}
	}

	return nil
}

func (y *Yaml2SDK) getPath(data interface{}) (opID, desc string, err error) {
	opID = data.(MapIntInt)["operationId"].(string)
	desc = data.(MapIntInt)["summary"].(string)
	return
}

func (y *Yaml2SDK) getTags(path string) (tags []string, err error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("path is empty")
	}

	if path[0] != '/' {
		return nil, fmt.Errorf("path does not start with /")
	}

	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		return nil, fmt.Errorf("path does not have enough parts")
	}

	return parts[1 : len(parts)-2], nil
}

func (y *Yaml2SDK) getParameters(data MapIntInt) *[]YAMLParameter {
	parameters := make([]YAMLParameter, 0)
	for k, v := range data {
		if k == "parameters" {
			param := &YAMLParameter{}
			for _, pv := range v.([]interface{}) {

				if pName, exists := pv.(MapIntInt)["name"].(string); exists {
					param.Name = pName
				}

				if pDesc, exists := pv.(MapIntInt)["description"].(string); exists {
					param.Description = pDesc
				}

				if pReq, exists := pv.(MapIntInt)["required"].(bool); exists {
					param.Required = pReq
				}

				if pType, exists := pv.(MapIntInt)["type"].(string); exists {
					param.Type = pType
				}

				if pFmt, exists := pv.(MapIntInt)["format"].(string); exists {
					param.Format = pFmt
				}

				if pIn, exists := pv.(MapIntInt)["in"].(string); exists {
					param.In = pIn
				}
				parameters = append(parameters, *param)
			}
		}
	}

	return &parameters
}

func (y *Yaml2SDK) parseProperty(data MapIntInt) *YAMLProperty {
	prop := &YAMLProperty{}
	//prop.Type = data["type"].(string)

	if pType, exists := data["type"].(string); exists {
		prop.Type = pType
	}

	if pDesc, exists := data["description"].(string); exists {
		prop.Description = pDesc
	}

	// if pProp, exists := data["properties"].(MapIntInt); exists {
	// 	prop.Properties = *y.getProperties(pProp)
	// }

	// for k, v := range data {
	// 	fmt.Printf("Found property: [%s] = %+v\n", k, v)
	// 	if k == "type" {
	// 	if pName, exists := v.(MapIntInt)["name"].(string); exists {
	// 		prop.Name = pName
	// 	}

	// 	if pDesc, exists := v.(MapIntInt)["description"].(string); exists {
	// 		prop.Description = pDesc
	// 	}

	// 	if pType, exists := v.(MapIntInt)["type"].(string); exists {
	// 		prop.Type = pType
	// 	}

	// 	if pFmt, exists := v.(MapIntInt)["format"].(string); exists {
	// 		prop.Format = pFmt
	// 	}

	// 	if pDefault, exists := v.(MapIntInt)["default"].(string); exists {
	// 		prop.Default = pDefault
	// 	}
	//}

	fmt.Printf("\t-Found property: %s: %s\n", prop.Name, prop.Type)
	return prop
}

func (y *Yaml2SDK) getProperties(data MapIntInt) *[]YAMLProperty {
	properties := make([]YAMLProperty, 0)
	for _, v := range data {
		properties = append(properties, *y.parseProperty(v.(MapIntInt)))
	}

	return &properties
}

func (y *Yaml2SDK) handleCodeComment(s string) string {
	result := ""
	cs := common.SplitStrBySize(s, common.DEFAULT_LINE_LENGTH)
	for _, c := range cs {
		result += fmt.Sprintf("// %s\n", c)
	}

	return result
}

func (y *Yaml2SDK) generateHeader() string {
	codeGenMsg := "Code generated by wasm-sdk. DO NOT EDIT.\n"
	fnCode := fmt.Sprintf("// %s\n", codeGenMsg)
	fnCode += "package main\n\n"
	return fnCode
}

func (y *Yaml2SDK) handleFunction(e *Endpoints) string {
	var callFunction string
	var fnParams []string
	var params = ""

	fnName := fmt.Sprintf("call%s", e.OperationID)
	callFunction = fmt.Sprintf("\n// %s\n", fnName)
	callFunction += y.handleCodeComment(e.Description)
	callFunction += fmt.Sprintf("// Method: %s\n", e.Action)
	callFunction += fmt.Sprintf("// Path: %s\n", e.Path)
	if len(e.Parameters) > 0 {
		callFunction += "// Parameters:\n"
		for _, p := range e.Parameters {
			p.Name = strings.ReplaceAll(p.Name, ".", "_")
			if p.Type == "integer" {
				p.Type = "int64"
			} else if p.Type == "boolean" {
				p.Type = "bool"
			} else if p.Type == "array" {
				p.Type = "[]string"
			}
			fnParams = append(fnParams, fmt.Sprintf("%s %s", p.Name, p.Type))
			callFunction += fmt.Sprintf("//\t%s: %s\n", p.Name, y.handleCodeComment(p.Description))
		}
		params = strings.Join(fnParams, ", ")
	} else {
		callFunction += "// Parameters: none\n"
	}
	callFunction += fmt.Sprintf("func (sdk *SDKInterface) %s(%s) {\n", fnName, params)
	callFunction += fmt.Sprintf("\t// TODO: Implement %s", e.OperationID)
	callFunction += "\n}\n"

	// add function name to map to be used in the interface declaration
	y.intMethods = append(y.intMethods, fmt.Sprintf("%s(%s)", fnName, params))
	return callFunction
}

func (y *Yaml2SDK) generateSDKFunctions() string {
	var fnCode string
	for _, e := range y.source {
		fnCode += y.handleFunction(&e)
	}
	return fnCode
}

func (y *Yaml2SDK) generateSDKInterface() string {
	sdkInt := "// SDKInterface is the interface for the SDK\n"
	sdkInt += "type SDKInterface interface {\n"
	for _, v := range y.intMethods {
		sdkInt += fmt.Sprintf("\t%s\n", v)
	}
	sdkInt += "}\n"

	return sdkInt
}

func (y *Yaml2SDK) generateSDK() {
	time.Sleep(3 * time.Second)
	s := spinner.StartNew("Generating SDK...")
	sdkHeader := y.generateHeader()
	sdkFunctions := y.generateSDKFunctions()
	sdkInterface := y.generateSDKInterface()
	sdkFile := sdkHeader + sdkInterface + sdkFunctions
	s.Stop()
	y.saveGoFile("../sdk.go", sdkFile)
}

// saveGoFile writes the generated golang sdk file to disk
func (y *Yaml2SDK) saveGoFile(fname string, code string) {
	f, err := os.Create(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	n4, err := w.WriteString(code)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created file [%s] - %d bytes\n", fname, n4)

	w.Flush()

}
