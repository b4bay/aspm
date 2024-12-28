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
		artefactPath   string
		artefactId     string
		artefactName   string
		artefactAuthor string
		artefactInfo   os.FileInfo
		reportsPath    []string
	)

	var err error
	var collectPayload shared.CollectMessageBody

	fs := flag.NewFlagSet(string(shared.CliModeCollect), flag.ExitOnError)
	fs.Parse(args)

	unnamed := fs.Args()
	if len(unnamed) < 2 {
		fmt.Println("Error: at least artefact and one report required")
		Exit(1)
	} else {
		artefactPath = unnamed[0]
		reportsPath = unnamed[1:]
	}

	// Processing artefact
	artefactInfo, err = os.Stat(artefactPath)
	if os.IsNotExist(err) {
		fmt.Printf("Error: Product not found '%s'", artefactPath)
		Exit(1)
	}

	if artefactInfo.IsDir() {
		artefactType = shared.ArtefactTypeGit
	} else {
		artefactType = shared.ArtefactTypeBin
	}

	if artefactType == shared.ArtefactTypeGit {
		artefactId, err = cli.IdGit(artefactPath)
		if err != nil {
			fmt.Printf("Error: Invalid artefact (id) '%s': %v\n", artefactPath, err)
			Exit(1)
		}
		artefactName, err = cli.NameGit(artefactPath)
		if err != nil {
			fmt.Printf("Error: Invalid artefact (name) '%s': %v\n", artefactPath, err)
			Exit(1)
		}
		artefactAuthor = cli.GetAuthorFromGit()
	}

	if artefactType == shared.ArtefactTypeBin {
		artefactId, err = cli.IdBin(artefactPath)
		if err != nil {
			fmt.Printf("Error: Invalid artefact '%s': %v\n", artefactPath, err)
			Exit(1)
		}
	}

	fmt.Printf("Running in 'collect' mode: artefact=%s, reports=%v\n", artefactPath, reportsPath)

	collectPayload.Artefact = shared.ProductMessage{
		Id:     artefactId,
		Type:   artefactType,
		Name:   artefactName,
		Author: artefactAuthor,
	}
	collectPayload.Environment = cli.GetEnvironment()
	collectPayload.Reports = cli.GetReports(reportsPath)

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
	var originPayload shared.OriginMessageBody
	var (
		productPath         string
		productId           string
		productName         string
		productAuthor       string
		productInfo         os.FileInfo
		productArtefactType shared.ArtefactType
		origins             []shared.ProductMessage
		originsPath         []string
		originPath          string
		originId            string
		originName          string
		originAuthor        string
		originInfo          os.FileInfo
		originArtefactType  shared.ArtefactType
	)

	fs := flag.NewFlagSet(string(shared.CliModeOrigin), flag.ExitOnError)
	method := fs.String("method", string(shared.ProductionMethodDefault), "Method (compile or pack)")
	fs.Parse(args)

	unnamed := fs.Args()
	if len(unnamed) < 2 {
		fmt.Println("Error: at least artefact and one origin required")
		Exit(1)
	} else {
		productPath = unnamed[0]
		originsPath = unnamed[1:]
	}

	if shared.IsValidProductionMethod(shared.ProductionMethod(*method)) {
		productionMethod = shared.ProductionMethod(*method)
	} else {
		fmt.Printf("Error: Invalid method '%s'", *method)
		Exit(1)
	}

	fmt.Printf("Running in 'origin' mode: method=%s, product=%s, origins=%v\n", *method, productPath, originsPath)

	// Processing Product
	productInfo, err = os.Stat(productPath)
	if os.IsNotExist(err) {
		fmt.Printf("Error: Product not found '%s'", productPath)
		Exit(1)
	}

	if productInfo.IsDir() {
		productArtefactType = shared.ArtefactTypeGit
	} else {
		productArtefactType = shared.ArtefactTypeBin
	}

	if productArtefactType == shared.ArtefactTypeGit {
		productId, err = cli.IdGit(productPath)
		if err != nil {
			fmt.Printf("Error: Invalid product (id) '%s': %v\n", productPath, err)
			Exit(1)
		}
		productName, err = cli.NameGit(productPath)
		if err != nil {
			fmt.Printf("Error: Invalid product (name) '%s': %v\n", productPath, err)
			Exit(1)
		}
		productAuthor = cli.GetAuthorFromGit()
	}

	if productArtefactType == shared.ArtefactTypeBin {
		productId, err = cli.IdBin(productPath)
		if err != nil {
			fmt.Printf("Error: Invalid product (id) '%s': %v\n", productPath, err)
			Exit(1)
		}
		productName, err = cli.NameBin(productPath)
		if err != nil {
			fmt.Printf("Error: Invalid product (name) '%s': %v\n", productPath, err)
			Exit(1)
		}
	}

	// Processing Origins
	for _, originPath = range originsPath {
		originInfo, err = os.Stat(originPath)
		if os.IsNotExist(err) {
			fmt.Printf("Error: Origin not found '%s'", originPath)
			Exit(1)
		}

		if originInfo.IsDir() {
			originArtefactType = shared.ArtefactTypeGit
		} else {
			originArtefactType = shared.ArtefactTypeBin
		}

		if originArtefactType == shared.ArtefactTypeGit {
			originId, err = cli.IdGit(originPath)
			if err != nil {
				fmt.Printf("Error: Invalid origin (id) '%s': %v\n", originPath, err)
				Exit(1)
			}
			originName, err = cli.NameGit(originPath)
			if err != nil {
				fmt.Printf("Error: Invalid origin (name) '%s': %v\n", originPath, err)
				Exit(1)
			}
			originAuthor = cli.GetAuthorFromGit()
		}

		if originArtefactType == shared.ArtefactTypeBin {
			originId, err = cli.IdBin(originPath)
			if err != nil {
				fmt.Printf("Error: Invalid origin (id) '%s': %v\n", originPath, err)
				Exit(1)
			}

			originName, err = cli.NameBin(originPath)
			if err != nil {
				fmt.Printf("Error: Invalid origin (name) '%s': %v\n", originPath, err)
				Exit(1)
			}

		}

		origins = append(origins, shared.ProductMessage{
			Id:     originId,
			Name:   originName,
			Type:   originArtefactType,
			Author: originAuthor,
		})
	}

	originPayload.Product.Id = productId
	originPayload.Product.Name = productName
	originPayload.Product.Type = productArtefactType
	originPayload.Product.Author = productAuthor
	originPayload.Origins = origins
	originPayload.ProductionMethod = productionMethod
	originPayload.Environment = cli.GetEnvironment()

	aspmClient.Post("/"+string(shared.CliModeOrigin), originPayload)

}
