package main

import (
	"bufio"
	"fmt"
	"os"
    "bytes"
	"os/exec"
	"runtime"
	"strings"
    "strconv"

	"github.com/fatih/color"
	cp "github.com/otiai10/copy"
)

const version = "0.6.1"

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

    // bump version number
    case "bump":
        banner()

        if len(os.Args) < 3 {
            color.Red("❌  Missing argument for bump subcommand. Use minor, major, or build.")
            printUsage()
            os.Exit(1)
        }

        versionBump(os.Args[2])

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
    fmt.Println("  bump <string>")
    fmt.Println("        bump the major, minor, or build version number in the main file")
	fmt.Println("  version")
	fmt.Println("        display version number and exit")
	fmt.Println("  help")
	fmt.Println("        display this help message and exit")
}

// This function creates a new project with a given name.
func createProject(uri string) {

	errors := 0
	var name, username string

	// check if we got a name or a uri
	if strings.Contains(uri, "/") {
		color.Cyan("Detected a github uri, extracting the name...")
		name = getName(uri)
		username = getUsername(uri)
	} else {
		color.Yellow("⚠  project name is not a github URI")
		color.Cyan("Checking if GOPHER_USERNAME environment variable is set...")

		// get the value of GOPHER_USERNAME environment variable
		gh_username := os.Getenv("GOPHER_USERNAME")

		if gh_username == "" {
			color.Yellow("⚠  GOPHER_USERNAME environment variable is not set.")
			color.Red("🛑 STOP: INPUT REQUIRED")
			fmt.Print("❓Enter your github username and press [ENTER]: ")
			fmt.Scanln(&gh_username)
			color.White("💬 Don't want to be asked again? Use a github uri when initializing the project.")
			color.White("💬 Example: gopher init github.com/username/project")
			color.White("💬 Alternatively set the GOPHER_USERNAME environment variable.")
		}

		color.Blue("🆗 Got your github username: " + gh_username)
		name = uri
		username = gh_username
		uri = "github.com/" + gh_username + "/" + name
	}

	gh_origin := "git@github.com:" + username + "/" + name + ".git"

	fmt.Println()
	color.White("📝 Project information:")
	color.White("  Project Name:\t" + name)
	color.White("  Github user: \t" + username)
	color.White("  Github URI: \t" + uri)
	color.White("  Github repo: \t" + gh_origin)
	fmt.Println()

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

	color.Blue("🆗 go module initiated.")

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

	color.Blue("🆗 .gitignore file created.")

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

	color.Blue("🆗 README file created.")

	// create a simple main.go file
	createMainFile()

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

	color.Blue("🆗 git repository initiated.")

	// add github as irigin
	color.Cyan("Running git remote add origin...")
	cmd = exec.Command("git", "remote", "add", "origin", gh_origin)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	e = cmd.Run()

	if e != nil {
		fmt.Print("💥 ")
		color.Red(e.Error())
		errors++
	}

	color.Blue("🆗 new origin repository added.")
	color.White("💬  You can run git push -u origin main to push your project to github.")

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

	werrors, lerrors, merrors := 0, 0, 0

	errors += buildAndZip("windows", name)
	if werrors == 0 {
		color.Blue("🆗 Windows build successful.")
	}

	errors += buildAndZip("linux", name)
	if lerrors == 0 {
		color.Blue("🆗 Linux build successful.")
	}

	errors += buildAndZip("darwin", name)
	if merrors == 0 {
		color.Blue("🆗 Mac build successful.")
	}

	errors += werrors + lerrors + merrors

	// print the success message
	if errors == 0 {
		color.Green("✔  Project " + name + " released successfully.")
	} else {
		color.Green("⚠  Project " + name + " released with some errors.")
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
        // check if the username is in an environment variable
        color.Cyan("Checking if GOPHER_USERNAME environment variable is set...")
        username = os.Getenv("GOPHER_USERNAME")

        if username == "" {
            color.Yellow("⚠  GOPHER_USERNAME environment variable is not set.")
            // ask user for github username since it's not in the module string
            color.Red("🛑 STOP: INPUT REQUIRED")
            fmt.Print("❓ Enter your github username and press [ENTER]: ")
            fmt.Scanln(&username)
        }
	}

    color.Blue("🆗 Got the project name: " + name)
    color.Blue("🆗 Got your github username: " + username)

	color.Cyan("Getting version from gopher.go file...")
	version = getVersion(name + ".go")

    color.Blue("🆗 Got the project version: " + version)

	color.Cyan("Adding generic description, you can edit it later...")
	description = "A new scoop package"

	color.Cyan("Creating the homepage url...")
	homepage = "https://github.com/" + username + "/" + name

    color.Blue("🆗 Homepage url: " + homepage)

	color.Cyan("Creating the download url...")
	url = "https://github.com/" + username + "/" + name + "/releases/download/v" + version + "/" + name + "_win.zip"

    color.Blue("🆗 Download url: " + url)

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

    color.Blue("🆗 Manifest created successfully.")

	// write the manifest to the file
	color.Cyan("Creating " + name + ".json file")
	mfile, err := os.Create(name + ".json")
	if err != nil {
		color.Red("Error creating " + name + ".json file")
		color.Red(err.Error())
		os.Exit(1)
	}
	defer mfile.Close()

    color.Blue("🆗 file created successfully.")

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

    // check if environment variable GOPHER_ISTALLPATH is set
    color.Cyan("Checking if GOPHER_INSTALLPATH environment variable is set...")
    installpath := os.Getenv("GOPHER_INSTALLPATH")

    if installpath == "" {
        color.Yellow("⚠  GOPHER_INSTALLPATH environment variable is not set.")
        color.White("💬 You can set it to the directory where you want gopher to install all the binaries.")
        color.White("💬 Gopher will use ~/bin or %USERPROFILE%\\bin if GOPHER_INSTALLPATH is not set.")
    }

	color.Cyan("Checking the os...")

	//check if we are on windows
	if runtime.GOOS == "windows" {

		color.Cyan("We are on Windows...")

        if installpath == "" {
            // get the user's home directory
            color.Cyan("Getting the user's home directory...")
            home := os.Getenv("USERPROFILE")
            installpath = home + "\\bin"
        }
        color.Blue("🆗 Attempting to install to: " + installpath)

		// check if the bin directory exists and bail out if it does not
		color.Cyan("Checking if the install directory exists...")
		if _, err := os.Stat(installpath); os.IsNotExist(err) {
			fmt.Print("💥 ")
			color.Red("The " + installpath + " directory does not exist. Please create it and add it to your path first.")
			os.Exit(1)
		}

		// copy the binary to the bin directory
		color.Cyan("Copying the binary to the bin directory...")

		// on windows cp is a built in shell feature so we will use the copy library instead
		err := cp.Copy(name+".exe", installpath+"\\"+name+".exe")
		if err != nil {
			fmt.Print("💥 ")
			color.Red(err.Error())
			os.Exit(1)
		}

	} else {

		color.Cyan("We are on Linux or Mac...")

        if installpath == "" {
            // get the user's home directory
            color.Cyan("Getting the user's home directory...")
            home := os.Getenv("HOME")
            installpath = home + "/bin"
        }

        color.Blue("🆗 Attempting to install to: " + installpath)

		// check if the bin directory exists and bail out if it does not
		color.Cyan("Checking if the bin directory exists...")
		if _, err := os.Stat(installpath); os.IsNotExist(err) {
			fmt.Print("💥 ")
			color.Red("The " + installpath + " directory does not exist. Please create it and add it to your path first.")
			os.Exit(1)
		}

		// copy the binary to the bin directory
		color.Cyan("Copying the binary to the bin directory...")
		cmd := exec.Command("cp", name, installpath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		e := cmd.Run()

		if e != nil {
			fmt.Print("💥 ")
			color.Red(e.Error())
			os.Exit(1)
		}

	}

    color.Blue("🆗 copy successful.")

    color.White("💬 Make sure " + installpath + " is in your PATH")
	color.Green("✔  " + name + " installed successfully into " + installpath)
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

func createMainFile() {

	color.Cyan("Getting module name from go.mod file contents...")
	name := getModuleName()

	color.Cyan("Generating the " + name + ".go file...")
	content := `package main    

import (
"os"
"fmt"
"path/filepath"
)

const version = "0.1.0"

func main() {

    if len(os.Args) > 1 {
        switch os.Args[1] {
        case "-v", "--version":
            Version()
        case "-h", "--help":
            Usage()
        default:
            Usage()
        } 
    } else {
        Usage()
    }
}

func Version() {
    fmt.Println(filepath.Base(os.Args[0]), "version", version)
    os.Exit(0)
}

func Usage() {
    fmt.Println("Usage:", filepath.Base(os.Args[0]), "[options]")
    fmt.Println("Options:")
    fmt.Println("  -v, --version    Print version information and exit")
    fmt.Println("  -h, --help       Print this message and exit")
    os.Exit(0)
}`

	color.Cyan("Creating the " + name + ".go file on disk...")
	gfile, err := os.Create(name + ".go")
	if err != nil {
		fmt.Print("💥 ")
		color.Red("Error creating " + name + ".go file")
		color.Red(err.Error())
		os.Exit(1)
	}

	color.Cyan("Writing the " + name + ".go file content to disk...")
	gfile.WriteString(content)
	color.Blue("🆗 " + name + ".go file created successfully.")
}

func incString(s string) string {
    n, _ := strconv.Atoi(s)
    return strconv.Itoa(n + 1)
}


// bump the version number in the version constant of the main file
func versionBump(what string) {

    if what != "major" && what != "minor" && what != "build" {
        fmt.Print("💥 ")
        color.Red("Invalid argument for bump subcommand.")
        printUsage()
        os.Exit(1)
    }

    // get the module name from go.mod file
    color.Cyan("Getting module name from go.mod file...")
    name := getModuleName()

    // get the current version
    color.Cyan("Getting current version from " + name + ".go file...")
    version := getVersion(name + ".go")

    color.Blue("🆗 Current version is " + version)
    color.Cyan("Bumping the version number...")

    // split the version into parts
    parts := strings.Split(version, ".")
    major := parts[0]
    minor := parts[1]
    build := parts[2]

    // increment the build number
    if what == "build" {
        build = incString(build)
    } else if what == "minor" {
        minor = incString(minor)
        build = "0"
    } else if what == "major" {
        major = incString(major)
        minor = "0"
        build = "0"
    }

    // create the new version string
    new_version := major + "." + minor + "." + build

    color.Blue("🆗 New version is " + new_version)

    // read the file into memory
    file, err := os.ReadFile(name + ".go")
    if err != nil {
        fmt.Print("💥 ")
        color.Red("Error reading file")
        color.Red(err.Error())
        os.Exit(1)
    }

    color.Cyan("Replacing the version number in " + name + ".go file...")

    find := "const version = \"" + version + "\""
    replace := "const version = \"" + new_version + "\""

    file = bytes.Replace(file, []byte(find), []byte(replace), -1)

    // write the file back to disk
    err = os.WriteFile(name+".go", file, 0644)
    if err != nil {
        fmt.Print("💥 ")
        color.Red("Error writing file")
        color.Red(err.Error())
        os.Exit(1)
    }

    color.Blue("🆗 Version number replaced successfully.")
    color.Green("✔  Version bumped to " + new_version)

}
