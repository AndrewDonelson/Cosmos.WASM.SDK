package models

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

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

// generateMain() generates the main() function for the SDK. This function uses a channel to block the main thread and
// SDK from exiting. This is a hack to prevent the program from exiting and allow the global functions to be called from
// the JS code.
func (y *Yaml2SDK) generateMain() string {
	fnCode := "\nfunc main() {\n"
	fnCode += y.getJSExports()
	fnCode += "\n\tfmt.Println(SDKVersion(\"COSM-SDK\"), \"ready\")\n"
	fnCode += "\tc := make(chan bool)\n"
	fnCode += "\t<-c\n"
	fnCode += "\tfmt.Println(SDKVersion(\"COSM-SDK\"), \"exited\")\n"
	fnCode += "}\n"
	return fnCode
}

// getJSFunctions() returns the code for the JS functions that will be called from the JS code.
// Example:
//
//	func jsCosmosStakingV1Beta1ValidatorDelegations(this js.Value, args []js.Value) interface{} {
//		return callCosmosStakingV1Beta1ValidatorDelegations(args[0].String(), args[1].String(), args[2].String(), args[3].String(), args[4].Bool(), args[5].Bool())
//	}
func (y *Yaml2SDK) getJSFunctions() string {
	/*** [Begin jsFunctions] ***/
	fnCode := "\n"
	fnCode += "func jsSKDVersion(this js.Value, args []js.Value) interface{} {\n"
	fnCode += "\treturn SDKVersion(\"GWF-SDK\")\n"
	fnCode += "}\n"

	return fnCode

}

// getJSExports() exports the code for the JS that will be called from the JS code.
func (y *Yaml2SDK) getJSExports() string {
	fnCode := "\n"
	fnCode += "\tjs.Global().Set(\"SDKVersion\", js.FuncOf(jsSKDVersion))\n"
	//TODO: Loop through and add the rest of the functions here
	// for _, e := range y.source {
	// 	fnCode += y.handleFunction(&e)
	// }
	return fnCode
}

// getImports() returns the code for the imports that will be used in the SDK.
func (y *Yaml2SDK) getImports() string {
	imports := `import (
	"fmt"
	"reflect"
	"io"
	"encoding/json"
	"net/http"	
	"syscall/js"
)
`
	return imports
}

// parseYaml() parses the YAML file and creates the Endpoints struct.
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
									prop.Type = y.translateType(pType)
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

// getpath() returns the operationId and description for the path.
func (y *Yaml2SDK) getPath(data interface{}) (opID, desc string, err error) {
	opID = data.(MapIntInt)["operationId"].(string)
	desc = data.(MapIntInt)["summary"].(string)
	return
}

// getTags() returns the tags for the path.
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

// getParameters() returns the parameters for the path.
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
					//param.Type = pType
					param.Type = y.translateType(pType)
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

// getproperty() returns the property for the path.
func (y *Yaml2SDK) parseProperty(data MapIntInt) *YAMLProperty {
	prop := &YAMLProperty{}
	//prop.Type = data["type"].(string)

	if pType, exists := data["type"].(string); exists {
		prop.Type = y.translateType(pType)
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

// getproperties() returns the properties for the path.
func (y *Yaml2SDK) getProperties(data MapIntInt) *[]YAMLProperty {
	properties := make([]YAMLProperty, 0)
	for _, v := range data {
		properties = append(properties, *y.parseProperty(v.(MapIntInt)))
	}

	return &properties
}

// handleCodeComment() handles the code comment for the path.
func (y *Yaml2SDK) handleCodeComment(s string) string {
	result := ""
	cs := common.SplitStrBySize(s, common.DEFAULT_LINE_LENGTH)
	for _, c := range cs {
		result += fmt.Sprintf("// %s\n", c)
	}

	return result
}

// generateHeader() generates the header for the generated SDK golang file
func (y *Yaml2SDK) generateHeader() string {
	codeGenMsg := "Code generated by wasm-sdk. DO NOT EDIT.\n"
	fnCode := "// go:build js && wasm\n"
	fnCode += fmt.Sprintf("// %s\n", codeGenMsg)
	fnCode += "package main\n\n"
	fnCode += y.getImports()
	return fnCode
}

func (y *Yaml2SDK) translateType(t string) string {
	switch t {
	case "":
		return "string"
	case "integer":
		return "int64"
	case "body":
		return "string"
	case "boolean":
		return "bool"
	case "array":
		return "[]string"
	case "object":
		return "map[string]interface{}"
	default:
		return t
	}
}

// handleFunction() handles the function for the path.
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
			p.Type = y.translateType(p.Type)
			fnParams = append(fnParams, fmt.Sprintf("%s %s", p.Name, p.Type))
			callFunction += fmt.Sprintf("//\t%s: %s\n", p.Name, y.handleCodeComment(p.Description))
		}
		params = strings.Join(fnParams, ", ")
	} else {
		callFunction += "// Parameters: none\n"
	}
	callFunction += fmt.Sprintf("func %s(%s) {\n", fnName, params)
	//callFunction += fmt.Sprintf("\t// TODO: Implement %s", e.OperationID)
	callFunction += fmt.Sprintf(`	result, err := consumeAPI("%s", "%s")`, common.DEFAULT_API_HOST, e.Path)
	callFunction += "\n\t if err != nil {\n"
	callFunction += "\t\t fmt.Printf(\"Error: %s\", err)\n"
	callFunction += "\t }\n"
	callFunction += "\t fmt.Printf(\"Result: %v\", result)\n"
	callFunction += "\n}\n"

	// add function name to map to be used in the interface declaration
	y.intMethods = append(y.intMethods, fmt.Sprintf("%s(%s)", fnName, params))
	return callFunction
}

// generateSDKFunctions() generates the functions for the generated SDK golang file
func (y *Yaml2SDK) generateSDKFunctions() string {
	var fnCode string
	for _, e := range y.source {
		fnCode += y.handleFunction(&e)
	}
	return fnCode
}

// generateSDKInterface() generates the interface for the generated SDK golang file
func (y *Yaml2SDK) generateSDKInterface() string {
	sdkInt := "// SDKInterface is the interface for the SDK\n"
	sdkInt += "type SDKInterface interface {\n"
	for _, v := range y.intMethods {
		sdkInt += fmt.Sprintf("\t%s\n", v)
	}
	sdkInt += "}\n"

	return sdkInt
}

// generateSDK() generates the SDK golang file
func (y *Yaml2SDK) generateSDK() {
	s := spinner.StartNew("Generating SDK...")
	sdkHeader := y.generateHeader()
	sdkFunctions := y.generateSDKFunctions()
	sdkInterface := y.generateSDKInterface()
	sdkFile := sdkHeader +
		sdkInterface +
		sdkFunctions +
		y.getJSFunctions() +
		y.generateCommonCode() +
		y.generateMain()
	s.Stop()
	y.saveGoFile(common.GOLANG_FILE, sdkFile)
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

func (y *Yaml2SDK) generateCommonCode() string {
	fncode := `
const (
	DEFAULT_API_HOST    = "http://localhost:1317"
)

// These fields are populated by by go build flags (see ./Makefile && docker/build.sh)
var (
	Version   string
	BuildDate string
)

// consumeAPI consumes and API endpoint and returns the response as a map
// Scope: Private
// Parameters:
//
//	url: The URL of the API
//	endpoint: The endpoint to consume
//
// Returns:
//
//	map[string]interface{}: The response as a map
//	error: Any error that occurred
func consumeAPI(url string, endpoint string) (map[string]interface{}, error) {
	if url == "" {
		url = DEFAULT_API_HOST
	}
	fmt.Println("Request:", url, endpoint)

	epURL := url + endpoint

	// Make an HTTP GET request to the API endpoint
	response, err := http.Get(epURL)
	if err != nil {
		return nil, fmt.Errorf("error making HTTP GET request: %v", err)
	}
	defer response.Body.Close()

	// Check that the HTTP status is 200 OK
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status: %d", response.StatusCode)
	}

	// Check that the content type is JSON
	contentType := response.Header.Get("Content-Type")
	if contentType != "application/json" {
		return nil, fmt.Errorf("unexpected content type: %s", contentType)
	}

	// Read the response body and unmarshal it into a map
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}
	var data map[string]interface{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	fmt.Println("Response:", PrettyPrint(data))
	return data, nil
}

// PrettyPrint is used to display any type nicely in the log output
func PrettyPrint(v interface{}) string {

	name := GetType(v)
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return ""
	}

	return fmt.Sprintf("Dump of [%s]:\n%s\n", name, string(b))
}

// GetType will return the name of the provided interface using reflection
func GetType(i interface{}) string {
	t := reflect.TypeOf(i)
	if t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	}

	return t.Name()
}

/* SDKVersion() is a function that returns the version of the SDK.
	* This function is called from the JS code.
	*/
func SDKVersion(name string) string {
	return fmt.Sprintf("%s v%s (%s)", name, Version, BuildDate)
}	
	`
	return fncode
}
