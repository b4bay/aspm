package shared

type CliMode string

const (
	CliModeCollect CliMode = "collect"
	CliModeGW      CliMode = "gw"
	CliModeOrigin  CliMode = "origin"
	CliModeDefault         = CliModeCollect
)

var AllowedCliModes = []CliMode{CliModeCollect, CliModeGW, CliModeOrigin}

func IsValidCliMode(cliMode CliMode) bool {
	for _, a := range AllowedCliModes {
		if a == cliMode {
			return true
		}
	}
	return false
}

type ArtefactType string

const (
	ArtefactTypeGit     ArtefactType = "git"
	ArtefactTypeBin     ArtefactType = "bin"
	ArtefactTypeDefault ArtefactType = ArtefactTypeGit
)

var AllowedArtefactTypes = []ArtefactType{ArtefactTypeGit, ArtefactTypeBin}

func IsValidArtefactType(artefactType ArtefactType) bool {
	for _, a := range AllowedArtefactTypes {
		if a == artefactType {
			return true
		}
	}
	return false
}

type ProductionMethod string

const (
	ProductionMethodCompile ProductionMethod = "compile"
	ProductionMethodPack    ProductionMethod = "pack"
	ProductionMethodDefault ProductionMethod = ProductionMethodCompile
)

var AllowedProductionMethods = []ProductionMethod{ProductionMethodCompile, ProductionMethodPack}

func IsValidProductionMethod(productionMethod ProductionMethod) bool {
	for _, a := range AllowedProductionMethods {
		if a == productionMethod {
			return true
		}
	}
	return false
}

type ProductMessage struct {
	Id     string       `json:"id"`
	Name   string       `json:"name"`
	Type   ArtefactType `json:"type"`
	Author string       `json:"author"`
}

type OriginMessageBody struct {
	Environment      map[string]string `json:"environment"`
	Product          ProductMessage    `json:"product"`
	Origins          []ProductMessage  `json:"origins"`
	ProductionMethod ProductionMethod  `json:"production_method"`
}

type CollectMessageBody struct {
	Environment map[string]string `json:"environment"`
	Artefact    ProductMessage    `json:"artefact"`
	Reports     map[string]string `json:"reports"`
}
