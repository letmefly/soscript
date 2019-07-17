package main

import (
	"fmt"
	"github.com/urfave/cli"
	"log"
	"os"
)

func compileOne(varDefFilePath string, varConfigFilePath string, sourceFilePath string, outputFilePath string) {
	varDefFile, err := os.Open(varDefFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer varDefFile.Close()
	varConfigFile, err := os.Open(varConfigFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer varConfigFile.Close()
	sourceFile, err := os.Open(sourceFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer sourceFile.Close()
	//outputFile, err := os.Open(outputFilePath)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer outputFile.Close()
	varLexer := newLexer("ss", varDefFile)
	configLexer := newLexer("ss", varConfigFile)
	sourceLexer := newLexer("not_ss", sourceFile)
	parser := newParser(varLexer, configLexer)
	parser.parseSourceCode(sourceLexer)
	generator := newServerGen(outputFilePath, parser)
	generator.gen()
}

func compileDir(varDefFilePath string, varConfigFilePath string, sourceDir string, outputDir string) {

}

func main() {
	app := cli.NewApp()
	app.Commands = []cli.Command{
		{
			Name:    "compile",
			Aliases: []string{"compile"},
			Usage:   "Compile Source File",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "variable, v",
					Usage: "Load Variable Definition File",
				},
				cli.StringFlag{
					Name: "config, c",
					Usage: "Load Variable Config File",
				},
				cli.StringFlag{
					Name: "source, s",
					Usage: "Load Source File",
				},
				cli.StringFlag{
					Name: "output, o",
					Usage: "Store Compile Output File",
				},
			},
			Action:  func(c *cli.Context) error {
				fmt.Println("haha")
				fmt.Println(c.String("v"))
				fmt.Println(c.String("c"))
				fmt.Println(c.String("s"))
				fmt.Println(c.String("o"))
				varDefFilePath := c.String("v")
				varConfigFilePath := c.String("c")
				sourceFilePath := c.String("s")
				outputFilePath := c.String("o")
				compileOne(varDefFilePath, varConfigFilePath, sourceFilePath, outputFilePath)
				return nil
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}



	//srcRoot := "./"
	//dirs, _ := ioutil.ReadDir(srcRoot + "pb")
	//for _, dir := range dirs {
	//	if dir.IsDir() {
	//		files, _ := ioutil.ReadDir(srcRoot + "pb/" + dir.Name())
	//		for _, file := range files {
	//			if strings.HasSuffix(file.Name(), ".proto") {
	//				//log.Println(file.Name())
	//				strs := strings.Split(file.Name(), "_")
	//				serverName := strs[0]
	//				filePath := srcRoot + "pb/" + dir.Name() + "/" + file.Name()
	//				log.Println(filePath)
	//				protoFile, err := os.Open(filePath)
	//				if err != nil {
	//					log.Fatal(err)
	//				}
	//				lexer := newLexer(protoFile)
	//				defer protoFile.Close()
	//
	//				parser := newParser(lexer)
	//				parser.parse()
	//				genType := "rpc"
	//				genFile := ""
	//				if strings.Contains(file.Name(), "_msg") {
	//					genType = "msg"
	//					genFile = srcRoot + "pb/" + dir.Name() + "/" + serverName + "_msg.gen.go"
	//				} else if strings.Contains(file.Name(), "_rpc") {
	//					genType = "rpc"
	//					genFile = srcRoot + "pb/" + dir.Name() + "/" + serverName + "_rpc.gen.go"
	//				}
	//
	//				log.Println("genFile:", genFile)
	//				generator := newServerGen(genType, genFile, serverName, parser)
	//				generator.gen()
	//			}
	//		}
	//	}
	//}
}
