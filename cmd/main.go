package main

import (
	"fmt"

	"github.com/AndrewDonelson/Cosmos.WASM.SDK/internal/common"
	"github.com/AndrewDonelson/Cosmos.WASM.SDK/internal/models"
)

func main() {
	yml2sdk := models.NewYaml2SDK(common.YAML_DIR + common.YAML_FILE)
	if yml2sdk == nil {
		fmt.Println("Error creating Yaml2SDK")
		return
	}
}
