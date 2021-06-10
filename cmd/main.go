package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/curd-boy/sql-gen/mysql"
	"github.com/spf13/cobra"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

var sqlFP string
var sqlF string
var sqlS string
var tplP string
var fileOut string

func main() {
	rootCMD := &cobra.Command{
		Use: "sgen",
		Long: `生成go function
[-d] 解析指定目录下所有.sql文件 eg. "./"
[-f] 解析指定.sql文件中 eg. "./xx.sql"
[-s] 解析指定sql语句 eg. "select * from users where id = ?;"
[-t] 指定模板 eg. "./xx.tpl"
[-o] 输出到指定目录 默认./`,
		Args: nil,
		Run: func(cmd *cobra.Command, args []string) {
			if sqlFP != "" {
				parseFP(sqlFP)
				return
			}
			if sqlFP != "" {
				parseF(sqlF)
				return
			}
			if sqlS != "" {
				parseS(sqlS)
				return
			}
		},
	}
	rootCMD.Flags().StringVar(&sqlFP, "d", "", "director paths")
	rootCMD.Flags().StringVar(&sqlF, "f", "", "file path")
	rootCMD.Flags().StringVar(&sqlS, "s", "", "sql string")
	rootCMD.Flags().StringVar(&fileOut, "o", "./", "output file path")
	rootCMD.Flags().StringVar(&tplP, "t", "", "template file")

	rootCMD.AddCommand(&cobra.Command{Use: "version", Run: func(cmd *cobra.Command, args []string) {
		log.Println("0.0.1")
	}})
	rootCMD.Execute()
}
func getPackageName() string {
	packName := "main"
	pp, err := filepath.Abs(fileOut)
	if err != nil {
		log.Println(err.Error())
	}
	packName = filepath.Base(pp)
	if strings.Contains(packName, ".") {
		return "main"
	}
	packName = strings.ReplaceAll(packName, "/", "_")
	packName = strings.ReplaceAll(packName, " ", "_")
	return packName
}

func parseFP(fp string) {
	rds, err := mysql.ParseSqlPath(fp)
	if err != nil {
		log.Println(err)
		return
	}
	err = mysql.Parse(rds, getPackageName())
	if err != nil {
		log.Println(err)
		return
	}
}
func parseF(fp string) {
	f, err := os.ReadFile(fp)
	if err != nil {
		log.Println(err.Error())
		return
	}
	err = mysql.Parse(bytes.NewReader(f), getPackageName())
	if err != nil {
		fmt.Println(err)
	}
}
func parseS(s string) {
	err := mysql.Parse(strings.NewReader(s), getPackageName())
	if err != nil {
		fmt.Println(err)
	}
}
