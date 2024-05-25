// This script generates new module in internal/modules/. It can be invoked by running go run main.go <moduleName> <moduleInitial>
package gen

import (
	_ "embed"
	"fmt"
	"golang-api-starter/internal/helper/utils"
	"log"
	"os"
	"os/exec"
	"slices"
	"strings"
	"sync"
	"text/template"

	// "golang.org/x/text/cases"
	// "golang.org/x/text/language"
	"github.com/iancoleman/strcase"
)

type entity struct {
	ModuleName string
	StructName string // just Uppercase 1st char
	Plural     string // the plural of the new module
	Initial    *string
	RouteName  *string
	TableName  *string
	Path       string
}

var wg sync.WaitGroup

var basepath = utils.RootDir(2)

func GenerateNewModule() {
	if len(os.Args) <= 2 {
		fmt.Println("error: missing arg[2] arg[3], try go run gen.go userDocument ud")
		return
	}
	if len(os.Args) == 3 {
		fmt.Println("error: missing new module name")
		fmt.Println("try: go run gen.go <module-name-in-singular-lower-case e.g: userDocument> <initial e.g: u (for ud)>")
		return
	}

	newModule := getNewModuleStruct(os.Args[2])

	fmt.Printf("newModule %+v,\n initial: %s,\n route: %s,\n tableName: %s\n", newModule, *newModule.Initial, *newModule.RouteName, *newModule.TableName)

	/* create directory */
	if err := os.Mkdir(newModule.Path, 0755); err != nil {
		fmt.Println("err:", err)
	}
	fmt.Printf("created %s\n\n", newModule.Path)
	// wg.Add(7) // number of go routines
	/* generate all related files route, controller, service, repo, interface, model, migration */
	// modelsDirectory := "internal/database/models"
	newModule.createFile("type.go", typeTemplate)
	newModule.createFile("route.go", routeTemplate)
	newModule.createFile("controller.go", controllerTemplate)
	newModule.createFile("service.go", serviceTemplate)
	newModule.createFile("repository.go", repositoryTemplate)
	newModule.generateMigration()
	// wg.Wait()

	/* re-generate server.go */
	reGenerateServerFile()
}

func getNewModuleStruct(inputName string) *entity {
	var (
		// inputName string = os.Args[2]
		// structName string = fmt.Sprintf("%s", cases.Title(language.English, cases.Compact).String(inputName))
		structName string = strcase.ToCamel(inputName)
		plural     string = Pluralfy(structName)
		routeName  string = strcase.ToKebab(plural)
		tableName  string = strings.ToLower(Pluralfy(strcase.ToSnake(inputName)))
		initial    string = os.Args[3]
	)

	newDirectory := fmt.Sprintf("internal/modules/%s", inputName)
	newModule := &entity{inputName, structName, plural, &initial, &routeName, &tableName, newDirectory}
	return newModule
}

func Pluralfy(word string) (plural string) {
	if word[len(word)-1:] == "y" { // handle the word ends with y --> ies
		plural = word[0:len(word)-1] + "ies"
	} else if word[len(word)-1:] == "s" {
		plural = word + "es" // handle the word ends with s --> es
	} else {
		plural = word + "s"
	}
	return plural
}

func (e *entity) createFile(fileName, templateFile string) {
	filePath := e.Path
	if len(fileName) > 0 {
		filePath = fmt.Sprintf("%s/%s", e.Path, fileName)
	}
	file, err := os.Create(filePath)
	if err != nil {
		log.Printf("create %s failed: %s\n", filePath, err)
		return
	}

	t := template.Must(template.New(filePath).Parse(templateFile))
	t.Execute(file, e)
}

func (e *entity) generateType() {
	// wg.Done()
}

// func generateRoute(newModule entity, dirPath string) {
// 	filePath := fmt.Sprintf("%s/route.go", dirPath)
// 	createFile(newModule, filePath, "route", routeTemplate)
// 	wg.Done()
// }

// func generateController(newModule entity, dirPath string) {
// 	filePath := fmt.Sprintf("%s/controller.go", dirPath)
// 	createFile(newModule, filePath, "controller", controllerTemplate)
// 	wg.Done()
// }

// func generateService(newModule entity, dirPath string) {
// 	filePath := fmt.Sprintf("%s/service.go", dirPath)
// 	createFile(newModule, filePath, "service", serviceTemplate)
// 	wg.Done()
// }

// func generateRepository(newModule entity, dirPath string) {
// 	filePath := fmt.Sprintf("%s/repository.go", dirPath)
// 	createFile(newModule, filePath, "repository", repositoryTemplate)
// 	wg.Done()
// }

// func generateModel(newModule entity, dirPath string) {
// 	filePath := fmt.Sprintf("%s/%s.go", dirPath, newModule.ModuleName)
// 	createFile(newModule, filePath, "model", modelTemplate)
// 	fmt.Printf("created %s/%s.go, \nplease add the fields(columns) in this file\n\n", dirPath, newModule.ModuleName)
// 	wg.Done()
// }

func (ent *entity) generateMigration() {
	ent.Plural = strings.ToLower(ent.Plural)
	whichOutput, _ := exec.Command("which", "migrate").Output()
	// fmt.Println("whichOutput",string(whichOutput))
	migrateBinPath := strings.Fields(string(whichOutput))
	if len(migrateBinPath) == 0 {
		log.Fatal("migrate command not found")
	}
	// fmt.Println("migrateBinPath?",migrateBinPath[1])

	migrationName := fmt.Sprintf("create_%s", *ent.TableName)
	// fmt.Println("up", migrationOut[0])
	// fmt.Println("down", migrationOut[1])

	dbEngines := map[string]map[string]string{
		"postgres": {
			"up":   migratePgUpTemplate,
			"down": migratePgDownTemplate,
			"ext":  "sql",
		},
		"mariadb": {
			"up":   migrateMariaUpTemplate,
			"down": migrateMariaDownTemplate,
			"ext":  "sql",
		},
		"sqlite": {
			"up":   migrateSqliteUpTemplate,
			"down": migrateSqliteDownTemplate,
			"ext":  "sql",
		},
		"mongodb": {
			"up":   migrateMongoUpTemplate,
			"down": migrateMongoDownTemplate,
			"ext":  "json",
		},
	}
	for dbEngine, migrations := range dbEngines {
		argstr := []string{"create", "-ext", migrations["ext"], "-dir", fmt.Sprintf("%s/migrations/%s", basepath, dbEngine), "-seq", migrationName}
		out, err := exec.Command(migrateBinPath[0], argstr...).CombinedOutput()
		migrationOut := strings.Split(strings.ReplaceAll(string(out), "\r\n", "\n"), "\n")

		/* create migrateion-up.sql */
		ent.Path = migrationOut[0]
		ent.createFile("", migrations["up"])
		fmt.Printf("created %s\n", migrationOut[0])

		/* create migrateion-down.sql */
		ent.Path = migrationOut[1]
		ent.createFile("", migrations["down"])
		fmt.Printf("created %s\n\n", migrationOut[1])

		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}

	fmt.Printf("DB migration files for %s created in ./migrations, \nplease go to add the SQL statements in up+down files, and then run: make migrate-up \n\n", ent.ModuleName)
	// wg.Done()
}

func reGenerateServerFile() {
	var allModules []entity
	// moduleDirs, err := ioutil.ReadDir(fmt.Sprintf("%s/internal/modules/", basepath))
	moduleDirs, err := os.ReadDir(fmt.Sprintf("%s/internal/modules/", basepath))
	if err != nil {
		log.Fatal("failed to open module ", err)
	}

	namesSkipForRegenerate := []string{"sample", "user", "group", "groupUser"}
	for _, dir := range moduleDirs {
		if slices.Contains(namesSkipForRegenerate, dir.Name()) || !dir.IsDir() {
			continue
		}
		// structName := fmt.Sprintf("%s", cases.Title(language.English, cases.Compact).String(dir.Name()))
		// structName := strcase.ToCamel(dir.Name())
		// routeName := strcase.ToKebab(dir.Name())
		// tableName := strings.ToLower(Pluralfy(routeName))
		module := getNewModuleStruct(dir.Name())
		module.Initial = nil

		// module := &entity{ModuleName: dir.Name(), StructName: structName, Plural: strings.ToLower(Pluralfy(structName)), RouteName: &routeName, TableName:&tableName }
		allModules = append(allModules, *module)
	}

	// fmt.Println(allModules)
	filePath := fmt.Sprintf("%s/cmd/server/main.go", basepath)
	tmplData := map[string][]entity{"Modules": allModules}

	file, err := os.Create(filePath)
	if err != nil {
		log.Println("re-generate server.go failed: ", err)
		return
	}

	t := template.Must(template.New("server.go").Parse(serverTemplate))
	t.Execute(file, tmplData)
}

//go:embed skel/route.tmpl
var routeTemplate string

//go:embed skel/controller.tmpl
var controllerTemplate string

//go:embed skel/service.tmpl
var serviceTemplate string

//go:embed skel/repository.tmpl
var repositoryTemplate string

//go:embed skel/type.tmpl
var typeTemplate string

//go:embed skel/migrate-pg-up.tmpl
var migratePgUpTemplate string

//go:embed skel/migrate-pg-down.tmpl
var migratePgDownTemplate string

//go:embed skel/migrate-mariadb-up.tmpl
var migrateMariaUpTemplate string

//go:embed skel/migrate-mariadb-down.tmpl
var migrateMariaDownTemplate string

//go:embed skel/migrate-sqlite-up.tmpl
var migrateSqliteUpTemplate string

//go:embed skel/migrate-sqlite-down.tmpl
var migrateSqliteDownTemplate string

//go:embed skel/migrate-mongo-up.tmpl
var migrateMongoUpTemplate string

//go:embed skel/migrate-mongo-down.tmpl
var migrateMongoDownTemplate string

//go:embed skel/server.tmpl
var serverTemplate string
