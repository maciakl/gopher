package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"
)

const version = "0.2.1"

const templateUrl = "https://gist.githubusercontent.com/maciakl/b5877bcb8b1ad21e2e798d3da3bff13b/raw/3fb1c32e3766bf2cf3926ee72225518e827a1228/hello.go"

func main() {

	var name string
	flag.StringVar(&name, "init", "", "bootstrap a new project with a given name")

	var wrap bool
	flag.BoolVar(&wrap, "wrap", false, "build the project and zip it (windows only for now)")

	var make bool
	flag.BoolVar(&make, "make", false, "build the project using a Makefile, falle back on wrap")

	var scoop bool
	flag.BoolVar(&scoop, "scoop", false, "generate a scoop manifest file for the project")

	var install bool
	flag.BoolVar(&install, "install", false, "install the project binary in the user's private bin directory")

	var ver bool
	flag.BoolVar(&ver, "version", false, "display version number and exit")
	flag.Parse()

	// show version and exit
	if ver {
		fmt.Println(filepath.Base(os.Args[0]), "version", version)
		os.Exit(0)
	}

	// bootstrap a new project
	if name != "" && !wrap && !scoop && !make && !install {
		createProject(name)
	}

	// build the project and zip it
	if wrap && name == "" && !scoop && !make && !install {
		buildLegacy()
	}

	// build the project using make
	if !wrap && name == "" && !scoop && make && !install {
		build()
	}

	// generate a scoop manifest file
	if scoop && name == "" && !wrap && !make && !install {
		generateScoopFile()
	}

	if !scoop && name == "" && !wrap && !make && install {
        installProject()
	}

	if name == "" && !wrap && !scoop && !make && !install {
		banner()
		color.Red("❌  No arguments provided. Use -init, -make, -wrap or -scoop.")
	}

}

// This function creates a new project with a given name.
func createProject(name string) {

	banner()

	errors := 0

	// create a new directory
	color.Cyan("Creating project " + name + "...")
	os.Mkdir(name, 0755)
	os.Chdir(name)

	// run the go mod init command
	color.Cyan("Running go mod init " + name + "...")
	cmd := exec.Command("go", "mod", "init", name)
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
		color.Red("Error creating README.md file:", err)
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

// The build function decides whether to use make or build it using the internal gopher defaults
func build() {

	banner()
	color.Cyan("Building the project...")
	color.Cyan("Checking if Makefile exists in the project directory...")

	// check if Makefile exists in the project directory
	if _, err := os.Stat("Makefile"); err == nil {
		color.Cyan("Makefile found in the project directory...")
		buildProjectWithMake()
	} else {
		color.Cyan("Makefile not found in the project directory...")
		buildProject()
	}

}

func buildLegacy() {
	banner()
	color.Cyan("Building the project using gopher defaults...")
	buildProject()
}

// This function runs the make command to build the project.
// It assumes make is installed on the system and that Makefile exists in the project directory.
func buildProjectWithMake() {

	errors := 0

	// change to the current directory (pwd)
	dir, err := os.Getwd()
	if err != nil {
		fmt.Print("💥 ")
		color.Red("Error getting the current directory")
        color.Red(err.Error())
		os.Exit(1)
	}
	os.Chdir(dir)

	// run the make command
	color.Cyan("Running make...")
	cmd := exec.Command("make")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	e := cmd.Run()

	if e != nil {
		fmt.Print("💥 ")
		color.Red(e.Error())
		errors++
	}

	// print the success message
	if errors == 0 {
		color.Green("✔  Project built successfully using the project Makefile.")
	} else {
		color.Green("⚠  Project built with some errors.")
	}

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

	banner()
	color.Cyan("Generating scoop manifest file...")
	// declare multiple string variables
	var name, username, version, description, homepage, url string

	// ask user for github username
	color.Yellow("⭐ Enter your github username and press [ENTER]:")
	fmt.Scanln(&username)

	color.Cyan("Getting module name from go.mod file...")
	name = getModuleName()

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
		color.Red("Error creating "+name+".json file", err)
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
		color.Red("Error opening file", err)
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
		color.Red("Error opening go.mod file", err)
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

func banner() {
	color.Cyan("🐿  Gopher v" + version + "\n")
}

// this funtion will istall the project binary in the user's provate bin directory
// this will be ~/bin on linux and mac and %USERPROFILE%\bin on windows
func installProject() {

    banner()

    // get project name from go.mod file
    name := getModuleName()

    color.Cyan("Installing "+name+"...")

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
            color.Red("The "+home+"\\bin"+" directory does not exist. Please create it and add it to your path first.")
            os.Exit(1)
        }

        // copy the binary to the bin directory
        color.Cyan("Copying the binary to the bin directory...")
        cmd := exec.Command("copy", name+".exe", home + "\\bin")
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        e := cmd.Run()

        if e != nil {
            fmt.Print("💥 ")
            color.Red(e.Error())
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
            color.Red("The "+home+"/bin"+" directory does not exist. Please create it and add it to your path first.")
            os.Exit(1)
        }

        // copy the binary to the bin directory
        color.Cyan("Copying the binary to the bin directory...")
        cmd := exec.Command("cp", name, home + "/bin")
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        e := cmd.Run()

        if e != nil {
            fmt.Print("💥 ")
            color.Red(e.Error())
            os.Exit(1)
        }

    }

    color.Green("✔  "+name+" installed successfully.")
}
