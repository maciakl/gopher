package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/fatih/color"
	cp "github.com/otiai10/copy"
)

const version = "0.4.3"

const templateUrl = "https://gist.githubusercontent.com/maciakl/b5877bcb8b1ad21e2e798d3da3bff13b/raw/e5d5196713bb404236eb28a5bf9b0183b648a4a9/hello.go"

func main() {

	// get number of arguments
	if len(os.Args) == 1 {
		banner()
		color.Red("❌  Missing subcommand.")
		printUsage()
		os.Exit(1)
	}

	// get the first argument
	arg := os.Args[1]

	switch arg {
	case "version", "-version", "--version", "-v":
		banner()
		os.Exit(0)

	case "help", "-help", "--help", "-h":
		banner()
		printUsage()
		os.Exit(0)

	// bootstrap a new project
	case "init":
		banner()

		if len(os.Args) < 3 {
			color.Red("❌  Missing argument for init subcommand.")
			printUsage()
			os.Exit(1)
		}

		createProject(os.Args[2])

	// create a Makefile for the project
	case "make":
		banner()
		createMakefile()

	// create a Justfile for the Project
	case "just":
		banner()
		createJustfile()

	// build the project and zip it
	case "release", "wrap":
		banner()
		release()

	// generate a scoop manifest file
	case "scoop":
		banner()
		generateScoopFile()

	case "install":
		banner()
		installProject()

	// print usage and exit
	default:
		banner()
		color.Red("❌  Unknown subcommand.")
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {

	fmt.Println("\nUsage: gopher [subcommand] <arguments>")
	fmt.Println("\nSubcommands:")
	fmt.Println("  init <string>")
	fmt.Println("        bootstrap a new project with a given <string> name or url")
	fmt.Println("  make")
	fmt.Println("        create a simple Makefile for the project")
	fmt.Println("  just")
	fmt.Println("        create a simple Justfile for the project")
	fmt.Println("  release")
	fmt.Println("        build the project for windows, linux and mac, then and zip the binaries")
	fmt.Println("  scoop")
	fmt.Println("        generate a Scoop.sh manifest file for the project")
	fmt.Println("  install")
	fmt.Println("        install the project binary in the user's private bin directory")
	fmt.Println("  version")
	fmt.Println("        display version number and exit")
	fmt.Println("  help")
	fmt.Println("        display this help message and exit")
}

// This function creates a new project with a given name.
func createProject(uri string) {

	errors := 0
	var name string

	// check if we got a name or a uri
	if strings.Contains(uri, "/") {
		color.Cyan("Detected a github uri, extracting the name...")
		name = getName(uri)
	} else {
		name = uri
	}

	// create a new directory
	color.Cyan("Creating project " + name + "...")
	os.Mkdir(name, 0755)
	os.Chdir(name)

	// run the go mod init command
	color.Cyan("Running go mod init " + uri + "...")
	cmd := exec.Command("go", "mod", "init", uri)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	e := cmd.Run()

	if e != nil {
		fmt.Print("💥 ")
		color.Red("Error running go mod command")
		errors++
	}

	// create .gitignore file
	color.Cyan("Creating .gitignore file...")
	gfile, err := os.Create(".gitignore")
	if err != nil {
		fmt.Print("💥 ")
		color.Red("Error creating .gitignore", err)
		os.Exit(1)
	}
	gfile.WriteString(name + "\n")
	gfile.WriteString(name + "*.exe\n")
	gfile.WriteString(name + ".zip\n")
	gfile.WriteString(name + ".tgz\n")
	gfile.WriteString(name + "_*.zip\n")
	gfile.WriteString(name + "_*.tgz\n")
	gfile.Close()

	// create README.md file
	color.Cyan("Creating README.md file...")
	rfile, err := os.Create("README.md")
	if err != nil {
		fmt.Print("💥 ")
		color.Red("Error creating README.md file")
		color.Red(err.Error())
		os.Exit(1)
	}
	rfile.WriteString("# " + name + "\n")
	rfile.Close()

	// download [name].go file fom the gist template url
	color.Cyan("Creating " + name + ".go file")
	cmd = exec.Command("curl", "-o", name+".go", templateUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	e = cmd.Run()

	if e != nil {
		fmt.Print("💥 ")
		color.Red(e.Error())
		errors++
	}

	// run the git init command with -b main
	color.Cyan("Running git init -b main...")
	cmd = exec.Command("git", "init", "-b", "main")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	e = cmd.Run()

	if e != nil {
		fmt.Print("💥 ")
		color.Red(e.Error())
		errors++
	}

	// print the success message
	if errors == 0 {
		color.Green("✔  Project " + name + " created successfully.")
	} else {
		color.Green("⚠  Project " + name + " created with some errors.")
	}

}

func release() {
	color.Cyan("Building the project using gopher defaults...")
	color.Cyan("This will create 3 zip files with the executables for windows, mac and linux.")
	buildProject()
}

// This is gopher's internal function to build the project and zip it
// by default it builds for windows, mac and linux and generates zip files
func buildProject() {

	errors := 0

	// get the module name from go.mod file
	color.Cyan("Getting module name from go.mod file...")
	name := getModuleName()

	// change to the current directory (pwd)
	dir, err := os.Getwd()
	if err != nil {
		fmt.Print("💥 ")
		color.Red("Error getting the current directory")
		color.Red(err.Error())
		os.Exit(1)
	}
	os.Chdir(dir)

	// detect the operating system
	color.Cyan("Detecting the operating system...")
	current_os := runtime.GOOS
	color.Cyan("Operating system is " + current_os)

	// run the go build command for the common operating systems
	color.Cyan("Attempting to cross-compile the project for windows, mac and linux...")
	errors += buildAndZip("windows", name)
	errors += buildAndZip("linux", name)
	errors += buildAndZip("darwin", name)

	// print the success message
	if errors == 0 {
		color.Green("✔  Project " + name + " wrapped successfully.")
	} else {
		color.Green("⚠  Project " + name + " wrapped with some errors.")
	}

}

// build and zip the project, return the number of errors encountered
func buildAndZip(current_os string, name string) int {

	errors := 0

	// run the go build command
	color.Cyan("Running go build for " + current_os + "...")
	cmd := exec.Command("go", "build")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOOS="+current_os)
	e := cmd.Run()

	if e != nil {
		fmt.Print("💥 ")
		color.Red(e.Error())
		errors++
	}

	var suffix string
	if current_os == "windows" {
		suffix = "win"
	}
	if current_os == "darwin" {
		suffix = "mac"
	}
	if current_os == "linux" {
		suffix = "lin"
	}

	// windows executables have .exe extension, others don't
	ext := ""
	if current_os == "windows" {
		ext = ".exe"
	}

	// create a zip file with the executable
	// requires the zip command to be installed
	// on windows it's not available by default
	// you can install it with scoop install zip
	color.Cyan("Creating " + name + "_" + suffix + ".zip file...")
	cmd = exec.Command("zip", "-r", name+"_"+suffix+".zip", name+ext)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	e = cmd.Run()

	if e != nil {
		fmt.Print("💥 ")
		color.Red(e.Error())
		errors++
	}

	return errors

}

// generate a scoop manifest file
func generateScoopFile() {

	color.Cyan("Generating scoop manifest file...")
	// declare multiple string variables
	var name, username, version, description, homepage, url string

	color.Cyan("Getting module string from go.mod file...")
	uri := getModule()

	// check if the module string is a uri
	if strings.Contains(uri, "/") {
		name = getName(uri)
		username = getUsername(uri)
	} else {
		name = uri
		// ask user for github username since it's not in the module string
		color.Yellow("⭐ Enter your github username and press [ENTER]:")
		fmt.Scanln(&username)
	}

	color.Cyan("Getting version from gopher.go file...")
	version = getVersion(name + ".go")

	color.Cyan("Adding generic description, you can edit it later...")
	description = "A new scoop package"

	color.Cyan("Creating the homepage url...")
	homepage = "https://github.com/" + username + "/" + name

	color.Cyan("Creating the download url...")
	url = "https://github.com/" + username + "/" + name + "/releases/download/v" + version + "/" + name + "_win.zip"

	color.Cyan("Creating the scoop manifest...")

	// create the scoop manifest file
	manifest := fmt.Sprintf(`{
    "version": "%s",
    "description": "%s",
    "homepage": "%s",
    "checkver": "github",
    "url": "%s",
    "bin": "%s",
    "license": "freeware"
}`, version, description, homepage, url, name+".exe")

	// write the manifest to the file
	color.Cyan("Creating " + name + ".json file")
	mfile, err := os.Create(name + ".json")
	if err != nil {
		color.Red("Error creating " + name + ".json file")
		color.Red(err.Error())
		os.Exit(1)
	}
	defer mfile.Close()

	mfile.WriteString(manifest)
	color.Green("✔  Scoop manifest file " + name + ".json created successfully.")

}

// searches the file name.go for a constant named version and returns its value
func getVersion(filename string) string {
	// open the file
	file, err := os.Open(filename)
	if err != nil {
		fmt.Print("💥 ")
		color.Red("Error opening file")
		color.Red(err.Error())
		os.Exit(1)
	}
	defer file.Close()

	// create a scanner
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "const version") {
			quoted_version := strings.Split(line, "=")[1]
			return strings.Trim(quoted_version, " \"")
		}
	}
	return ""
}

// searches go.mod file for the module name and returns it as string
func getModuleName() string {
	// open the file
	file, err := os.Open("go.mod")
	if err != nil {
		fmt.Print("💥 ")
		color.Red("Error opening go.mod file")
		color.Red(err.Error())
		os.Exit(1)
	}
	defer file.Close()

	// create a scanner
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "module") {

			url := strings.Split(line, " ")[1]

			// if url contains / then it's a github url an we need to extract the last part
			if strings.Contains(url, "/") {
				return getName(url)
			} else {
				return url
			}
		}
	}
	return ""
}

// search the go.mod file for the module string and return it
func getModule() string {
	// open the file
	file, err := os.Open("go.mod")
	if err != nil {
		fmt.Print("💥 ")
		color.Red("Error opening go.mod file")
		color.Red(err.Error())
		os.Exit(1)
	}
	defer file.Close()

	// create a scanner
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "module") {
			return strings.Split(line, " ")[1]
		}
	}
	return ""
}

// function that takes in a github uri and returns the last part of it
func getName(uri string) string {
	parts := strings.Split(uri, "/")
	return parts[len(parts)-1]
}

// function that takes in a github uri and returns the second to last part of it
func getUsername(uri string) string {
	parts := strings.Split(uri, "/")
	return parts[len(parts)-2]
}

func banner() {
	color.Cyan("🐿  Gopher v" + version + "\n")
}

// this funtion will istall the project binary in the user's provate bin directory
// this will be ~/bin on linux and mac and %USERPROFILE%\bin on windows
func installProject() {

	// get project name from go.mod file
	name := getModuleName()

	color.Cyan("Installing " + name + "...")

	// build it for this system first by running go build
	color.Cyan("Running go build...")
	cmd := exec.Command("go", "build")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	e := cmd.Run()

	if e != nil {
		fmt.Print("💥 ")
		color.Red(e.Error())
	}

	color.Cyan("Checking the os...")

	//check if we are on windows
	if runtime.GOOS == "windows" {

		color.Cyan("We are on Windows...")

		// get the user's home directory
		color.Cyan("Getting the user's home directory...")
		home := os.Getenv("USERPROFILE")

		// check if the bin directory exists and bail out if it does not
		color.Cyan("Checking if the bin directory exists...")
		if _, err := os.Stat(home + "\\bin"); os.IsNotExist(err) {
			fmt.Print("💥 ")
			color.Red("The " + home + "\\bin" + " directory does not exist. Please create it and add it to your path first.")
			os.Exit(1)
		}

		// copy the binary to the bin directory
		color.Cyan("Copying the binary to the bin directory...")

		// on windows cp is a built in shell feature so we will use the copy library instead
		err := cp.Copy(name+".exe", home+"\\bin\\"+name+".exe")
		if err != nil {
			fmt.Print("💥 ")
			color.Red(err.Error())
			os.Exit(1)
		}

	} else {

		color.Cyan("We are on Linux or Mac...")

		// get the user's home directory
		color.Cyan("Getting the user's home directory...")
		home := os.Getenv("HOME")

		// check if the bin directory exists and bail out if it does not
		color.Cyan("Checking if the bin directory exists...")
		if _, err := os.Stat(home + "/bin"); os.IsNotExist(err) {
			fmt.Print("💥 ")
			color.Red("The " + home + "/bin" + " directory does not exist. Please create it and add it to your path first.")
			os.Exit(1)
		}

		// copy the binary to the bin directory
		color.Cyan("Copying the binary to the bin directory...")
		cmd := exec.Command("cp", name, home+"/bin")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		e := cmd.Run()

		if e != nil {
			fmt.Print("💥 ")
			color.Red(e.Error())
			os.Exit(1)
		}

	}

	color.Green("✔  " + name + " installed successfully.")
}

func createMakefile() {

	color.Cyan("Creating Makefile...")

	color.Cyan("Getting module name from go.mod file...")
	name := getModuleName()

	color.Cyan("Generating the Makefile content...")
	content := fmt.Sprintf(`BINARY_NAME=%s

.PHONY: build
build: tidy
	go build

.PHONY: clean
clean:
	go clean
	rm $(BINARY_NAME)_*.zip
    

.PHONY: run
run: build
	./$(BINARY_NAME)

.PHONY: tidy
tidy:
	go mod tidy
	go fmt ./...
	go vet ./...
	go mod verify

.PHONY: test
test: build
	go test`, name)

	color.Cyan("Creating the Makefile file on disk...")

	mfile, err := os.Create("Makefile")
	if err != nil {
		fmt.Print("💥 ")
		color.Red("Error creating Makefile")
		color.Red(err.Error())
		os.Exit(1)
	}
	defer mfile.Close()

	color.Cyan("Writing the Makefile content to disk...")
	mfile.WriteString(content)
	color.Green("✔  Makefile created successfully.")
}

func createJustfile() {

	color.Cyan("Creating Justfile...")

	color.Cyan("Getting module name from go.mod file...")
	name := getModuleName()

	color.Cyan("Generating the Justfile content...")
	content := fmt.Sprintf(`BINARY_NAME := "%s"

build: tidy
	go build

clean:
	go clean
	rm {{BINARY_NAME}}_*.zip
    

run: build
	./{{BINARY_NAME}}

tidy:
	go mod tidy
	go fmt ./...
	go vet ./...
	go mod verify

test: build
	go test`, name)

	color.Cyan("Creating the Justfile file on disk...")
	jfile, err := os.Create("Justfile")
	if err != nil {
		fmt.Print("💥 ")
		color.Red("Error creating Justfile")
		color.Red(err.Error())
		os.Exit(1)
	}
	defer jfile.Close()

	color.Cyan("Writing the Justfile content to disk...")
	jfile.WriteString(content)
	color.Green("✔  Justfile created successfully.")
}
