package main

import (
	"flag"
	"fmt"
	"github.com/b4bay/aspm/internal/cli"
	"github.com/b4bay/aspm/internal/shared"
	"os"
)

var Exit = os.Exit
var server = os.Getenv("ASPM_SERVER_URL")
var key = os.Getenv("ASPM_SERVER_KEY")
var aspmClient cli.ASPMClientInterface = cli.NewASPMClient(server, key)

var (
	DefaultArtefact, _ = os.Getwd()
	DefaultScope, _    = os.Getwd()
)

func main() {
	var mode shared.CliMode

	// Check if mode is explicitly set as the first argument
	if len(os.Args) > 1 && os.Args[1][0] != '-' {
		mode = shared.CliMode(os.Args[1]) // Set mode to the first argument
		if !shared.IsValidCliMode(mode) {
			fmt.Printf("Error: Unknown mode '%s'. Supported modes are %v.\n", mode, shared.AllowedCliModes)
			Exit(1)
		} else { // Shift arguments for flag parsing
			os.Args = append([]string{os.Args[0]}, os.Args[2:]...)
		}
	} else {
		mode = shared.CliModeDefault
	}

	args := os.Args[1:]

	switch mode {
	case shared.CliModeCollect:
		handleCollectMode(args)
	case shared.CliModeGW:
		handleGWMode(args)
	case shared.CliModeOrigin:
		handleOriginMode(args)
	default:
		fmt.Printf("Error: Unknown mode '%s'. Supported modes are 'origin', 'collect' and 'gw'.\n", mode)
		Exit(1)
	}
}

func handleCollectMode(args []string) {
	var artefactType shared.ArtefactType
	var (
		artefact   string
		artefactId string
		reports    []string
	)

	var err error
	var collectPayload shared.CollectMessageBody

	fs := flag.NewFlagSet(string(shared.CliModeCollect), flag.ExitOnError)
	typ := fs.String("type", string(shared.ArtefactTypeDefault), "Artefact type")
	fs.Parse(args)

	unnamed := fs.Args()
	if len(unnamed) < 2 {
		fmt.Println("Error: at least artefact and one report required")
		Exit(1)
	} else {
		artefact = unnamed[0]
		reports = unnamed[1:]
	}

	if shared.IsValidArtefactType(shared.ArtefactType(*typ)) {
		artefactType = shared.ArtefactType(*typ)
	} else {
		fmt.Printf("Error: Invalid artefact type '%s'", *typ)
		Exit(1)
	}

	if artefactType == shared.ArtefactTypeGit {
		artefactId, err = cli.IdGit(artefact)
		if err != nil {
			fmt.Printf("Error: Invalid artefact '%s': %v\n", artefact, err)
			Exit(1)
		}
	} else if artefactType == shared.ArtefactTypeBin {
		artefactId, err = cli.IdBin(artefact)
		if err != nil {
			fmt.Printf("Error: Invalid artefact '%s': %v\n", artefact, err)
			Exit(1)
		}
	}

	fmt.Printf("Running in 'collect' mode: type=%s, artefact=%s, reports=%v\n", *typ, artefact, reports)

	collectPayload.ArtefactId = artefactId
	collectPayload.Environment = cli.GetEnvironment()
	collectPayload.Reports = cli.GetReports(reports)

	aspmClient.Post("/"+string(shared.CliModeCollect), collectPayload)
}

func handleGWMode(args []string) {
	fs := flag.NewFlagSet(string(shared.CliModeGW), flag.ExitOnError)
	typ := fs.String("type", string(shared.ArtefactTypeDefault), "Type value")
	scope := fs.String("scope", DefaultScope, "Scope path")
	fs.Parse(args)

	if *typ != "" && !shared.IsValidArtefactType(shared.ArtefactType(*typ)) {
		fmt.Printf("Error: Invalid type '%s'\n", *typ)
		Exit(1)
	}

	unnamed := fs.Args()
	var artefact = DefaultArtefact
	if len(unnamed) > 1 {
		fmt.Println("Error: only one artefact allowed")
		Exit(1)
	}

	if len(unnamed) == 1 {
		artefact = unnamed[0]
	}

	fmt.Printf("Running in 'gw' mode: type=%s, scope=%s, artefact=%s\n", *typ, *scope, artefact)

}

func handleOriginMode(args []string) {
	var err error
	var productionMethod shared.ProductionMethod
	var originArtefactType shared.ArtefactType
	var productArtefactType shared.ArtefactType
	var originPayload shared.OriginMessageBody
	var (
		product     string
		productId   string
		productName string
		origins     []string
		originIds   []string
	)

	fs := flag.NewFlagSet(string(shared.CliModeOrigin), flag.ExitOnError)
	method := fs.String("method", string(shared.ProductionMethodDefault), "Method (compile or pack)")
	from := fs.String("from", string(shared.ArtefactTypeDefault), "Origin artefact type")
	to := fs.String("to", string(shared.ArtefactTypeDefault), "Product artefact type")
	fs.Parse(args)

	unnamed := fs.Args()
	if len(unnamed) < 2 {
		fmt.Println("Error: at least artefact and one origin required")
		Exit(1)
	} else {
		product = unnamed[0]
		origins = unnamed[1:]
	}

	if shared.IsValidProductionMethod(shared.ProductionMethod(*method)) {
		productionMethod = shared.ProductionMethod(*method)
	} else {
		fmt.Printf("Error: Invalid method '%s'", *method)
		Exit(1)
	}

	if shared.IsValidArtefactType(shared.ArtefactType(*from)) {
		originArtefactType = shared.ArtefactType(*from)
	} else {
		fmt.Printf("Error: Invalid origin type '%s'", *from)
		Exit(1)
	}

	if shared.IsValidArtefactType(shared.ArtefactType(*to)) {
		productArtefactType = shared.ArtefactType(*to)
	} else {
		fmt.Printf("Error: Invalid product type '%s'", *to)
		Exit(1)
	}

	fmt.Printf("Running in 'origin' mode: method=%s, from=%s, to=%s, artefact=%s, sources=%v\n", *method, *from, *to, product, origins)

	if productArtefactType == shared.ArtefactTypeGit {
		productId, err = cli.IdGit(product)
		if err != nil {
			fmt.Printf("Error: Invalid product (id) '%s': %v\n", product, err)
			Exit(1)
		}
		productName, err = cli.NameGit(product)
		if err != nil {
			fmt.Printf("Error: Invalid product (name) '%s': %v\n", product, err)
			Exit(1)
		}
	} else if productArtefactType == shared.ArtefactTypeBin {
		productId, err = cli.IdBin(product)
		if err != nil {
			fmt.Printf("Error: Invalid product (id) '%s': %v\n", product, err)
			Exit(1)
		}
		productName, err = cli.NameBin(product)
		if err != nil {
			fmt.Printf("Error: Invalid product (name) '%s': %v\n", product, err)
			Exit(1)
		}
	}

	if originArtefactType == shared.ArtefactTypeGit {
		for _, origin := range origins {
			originId, err := cli.IdGit(origin)
			if err != nil {
				fmt.Printf("Error: Invalid origin '%s: %v'\n", origin, err)
				Exit(1)
			} else {
				originIds = append(originIds, originId)
			}
		}
	} else if originArtefactType == shared.ArtefactTypeBin {
		for _, origin := range origins {
			originId, err := cli.IdBin(origin)
			if err != nil {
				fmt.Printf("Error: Invalid origin '%s: %v'\n", origin, err)
				Exit(1)
			} else {
				originIds = append(originIds, originId)
			}
		}
	}

	originPayload.ProductId = productId
	originPayload.ProductName = productName
	originPayload.OriginIds = originIds
	originPayload.ProdMethod = productionMethod
	originPayload.ProductType = productArtefactType
	originPayload.Environment = cli.GetEnvironment()

	aspmClient.Post("/"+string(shared.CliModeOrigin), originPayload)

}
