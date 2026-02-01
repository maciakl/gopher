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

const version = "0.7.11"

func main() {
	
	_ , err := run()

	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}

func run() (string, error) {

	var err error = nil

	// get number of arguments
	if len(os.Args) == 1 {
		banner()
		color.Red("❌  Missing subcommand.")
		printUsage()
		return "missing subcommand", fmt.Errorf("missing subcommand")
	}

	// get the first argument
	arg := os.Args[1]

	switch arg {
	case "version", "-version", "--version", "-v":
		banner()
		return "banner", nil

	case "help", "-help", "--help", "-h":
		banner()
		printUsage()
		return "usage", nil

	// bootstrap a new project
	case "init":
		banner()

		if len(os.Args) < 3 {
			color.Red("❌  Missing argument for init subcommand.")
			printUsage()
			return "missing argument for init", fmt.Errorf("missing argument for init")
		}

		err = createProject(os.Args[2])

	// create a Makefile for the project
	case "make":
		banner()
		err = createMakefile()

	// create a Justfile for the Project
	case "just":
		banner()
		err = createJustfile()

	// build the project and zip it
	case "release":
		banner()
		err = release()

	case "info":
		banner()
		_, err = info(false)

	// generate a scoop manifest file
	case "scoop":
		banner()
		err = generateScoopFile()

	case "install":
		banner()
		err = installProject()

    // bump version number
    case "bump":
        banner()

        if len(os.Args) < 3 {
            color.Red("❌  Missing argument for bump subcommand. Use minor, major, or patch.")
            printUsage()
            return "missing argument for bump", fmt.Errorf("missing argument for bump")
        }

        err = versionBump(os.Args[2])

	// print usage and exit
	default:
		banner()
		color.Red("❌  Unknown subcommand.")
		printUsage()
		return "unknown subcommand", fmt.Errorf("unknown subcommand")
	}

	if err != nil {
		return "subcommand error", err
	}

	return "program terminated normally", nil
}

func printUsage() {

	fmt.Println("\nUsage: gopher [subcommand] <arguments>")
	fmt.Println("\nSubcommands:")
	fmt.Println("")
	fmt.Println("  init <string>")
	fmt.Println("        bootstrap a new project with where the <string> is the project name")
	fmt.Println("        in the format username/projectname or a full github uri like github.com/username/projectname")
	fmt.Println("")
	fmt.Println("  info")
	fmt.Println("        print project information known to gopher")
	fmt.Println("")
	fmt.Println("  make")
	fmt.Println("        create a simple Makefile for the project")
	fmt.Println("")
	fmt.Println("  just")
	fmt.Println("        create a simple Justfile for the project")
	fmt.Println("")
	fmt.Println("  release")
	fmt.Println("        build and release the project using goreleaser")
	fmt.Println("")
	fmt.Println("  scoop")
	fmt.Println("        generate a Scoop manifest file for the project")
	fmt.Println("")
	fmt.Println("  install")
	fmt.Println("        install the project binary in the user's private bin directory")
	fmt.Println("        typically ~/bin")
	fmt.Println("")
    fmt.Println("  bump <string>")
    fmt.Println("        bump the version number in the main file")
    fmt.Println("        the <string> can be major, minor, or build / patch")
	fmt.Println("")
	fmt.Println("  version")
	fmt.Println("        display version number of the gopher tool and exit")
	fmt.Println("")
	fmt.Println("  help")
	fmt.Println("        display this help message and exit")
}

func check() error {
	// check if go is installed
	_, err := exec.LookPath("go")
	if err != nil {
		fmt.Print("💥 ")
		color.Red("Go is not installed. Please install Go and try again.")
		return err
	}

	// check if git is installed
	_, err = exec.LookPath("git")
	if err != nil {
		fmt.Print("💥 ")
		color.Red("Git is not installed. Please install Git and try again.")
		return err
	}

	// check if goreleaser is installed
	_, err = exec.LookPath("goreleaser")
	if err != nil {
		fmt.Print("💥 ")
		color.Red("Goreleaser is not installed. Please install Goreleaser and try again.")
		return err
	}
	
	return nil
}

// find a text line inside a file that matches the pattern
func findInFile(filename string, pattern string) (string, error) {

	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, pattern) {
			return line, nil
		}
	}
	return "", nil
}

// perform an in-place find and replace in a text file
func replaceInFile(filename string, find string, replace string) error {

	file, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	file = bytes.Replace(file, []byte(find), []byte(replace), 1)
	err = os.WriteFile(filename, file, 0644)
	if err != nil {
		return err
	}
	return nil
}

// This function creates a new project with a given name.
func createProject(uri string) error {

	err := check()
	if err != nil {	return err }

	errors := 0
	var name, username string

	// check if we got a name or a uri
	if strings.Contains(uri, "/") {
		color.Cyan("Detected a github uri, extracting the name...")
		name = getName(uri)
		username = getUsername(uri)
		if !strings.Contains(uri, "github.com") {
			uri = "github.com/" + uri
		}
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
	defer gfile.Close()

	gfile.WriteString(".env\n")
	gfile.WriteString(name + "\n")
	gfile.WriteString(name + "*.exe\n")
	gfile.WriteString(name + ".zip\n")
	gfile.WriteString(name + ".tgz\n")
	gfile.WriteString(name + "_*.zip\n")
	gfile.WriteString(name + "_*.tgz\n")

	color.Blue("🆗 .gitignore file created.")

	// create README.md file
	color.Cyan("Creating README.md file...")
	rfile, err := os.Create("README.md")
	if err != nil {
		fmt.Print("💥 ")
		color.Red("Error creating README.md file")
		color.Red(err.Error())
		return err
	}
	defer rfile.Close()
	rfile.WriteString("# " + name + "\n")

	color.Blue("🆗 README file created.")

	// create a simple main.go file
	createMainFile()

	// create a simple test file
	createTestFile()

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

	// add github as origin
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

	// run goreleaser init
	color.Cyan("Running goreleaser init...")
	cmd = exec.Command("goreleaser", "init")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	e = cmd.Run()
	if e != nil {
		fmt.Print("💥 ")
		color.Red(e.Error())
		errors++
	}
	color.Blue("🆗 goreleaser configuration created.")
	color.White("💬  You can edit the .goreleaser.yml file to customize the release process.")

	yml_err := 0


	// modify the goreleaser.yaml and replace {{ .ProjectName }}_ with {{ .ProjectName }}_{{ .Version }}_

	color.Cyan("Modifying the .goreleaser.yaml file to include version in the archive names...")

	err = replaceInFile(".goreleaser.yaml", "{{ .ProjectName }}_", "{{ .ProjectName }}_{{ .Version }}_")

	if err != nil {
		fmt.Print("💥 ")
		color.Red("Error modifying .goreleaser.yaml file")
		color.Red(err.Error())
		errors++
		yml_err++
	}
	
	if yml_err == 0 {
		color.Blue("🆗 .goreleaser.yaml file modified successfully.")
	}

	// print the success message
	if errors == 0 {
		color.Green("✔  Project " + name + " created successfully.")
		return nil
	} else {
		color.Green("⚠  Project " + name + " created with some errors.")
		return fmt.Errorf("project creation completed with %d errors", errors)
	}
}

// get the main file name
func getMainFileName() (string, error) {
	name := "main"
	var e error
	// check if main.go exists, if not, get the module name and check if that file exists
	if _, err := os.Stat("main.go"); os.IsNotExist(err) {
		name, e = getModuleName()
		if e != nil { return "", e }	
		if _, err := os.Stat(name + ".go"); os.IsNotExist(err) {
			fmt.Print("💥 ")
			color.Red("Could not find main.go or " + name + ".go file in the current directory")
			return "", fmt.Errorf("Could not find main.go or %s.go file in the current directory", name)
		}
	}
	return name	, nil
}

type Info struct {
	name			string
	project			string
	version			string
	git_tag 		string
	git_tag_commit	string
	git_head		string
	git_branch		string
	git_state		string
	gh_username		string	
	gh_uri			string
	gh_origin		string
}

// dislpay info about the project
func info(silent bool) (Info, error) {

	var err error
	info := Info{}

	info.name, err = getMainFileName()
	if err != nil { return Info{}, err }

	info.project, err = getModuleName()
	if err != nil { return Info{}, err }


	info.version, err = getVersion(info.name + ".go")
	if err != nil { return Info{},err }

	info.gh_uri, err = getModule()
	if err != nil { return Info{}, err }

	info.gh_username = getUsername(info.gh_uri)
	info.gh_origin, err = getGitOrigin()
	if err != nil { return Info{}, err }

	info.git_tag = getGitTag()
	info.git_tag_commit = getGitCommit(info.git_tag)
	info.git_head = getGitCommit("HEAD")

	if getGitClean() {
		info.git_state = "✔️ " + color.BlueString("clean")
	} else {
		info.git_state = "❌ " + color.RedString("dirty")
	}

	// get git branch name
	info.git_branch, err = getGitBranch()
	if err != nil { return Info{}, err }

	var branch string
	if info.git_branch == "main" || info.git_branch == "master" {
		branch = color.GreenString(info.git_branch)
	} else {
		branch = color.YellowString(info.git_branch)
	}

	if !silent {

		fmt.Println()
		color.White("📝 Project information:")
		color.White("  Project Name:\t" + info.project)
		color.White("  Version:\t" + info.version)
		color.White("  Git tag: \t" + info.git_tag + " (" + info.git_tag_commit + ")")
		color.White("  Git HEAD: \t" + info.git_head)
		color.White("  Git branch:\t" + branch)
		color.White("  Git State: \t" + info.git_state)
		color.White("  Github user: \t" + info.gh_username)
		color.White("  Github URI: \t" + info.gh_uri)
		color.White("  Github repo: \t" + info.gh_origin)
		fmt.Println()


		color.White("📃 Recent git commits:")

		cmd := exec.Command("git", "--no-pager", "log", "--oneline", "--graph", "--decorate", "-10")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		e := cmd.Run()
		if e != nil {
			// print warning	
			color.Yellow("⚠  Failed to fetch git log")
		}

		fmt.Println()
	}
	return info, nil
}

func getGitBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		fmt.Print("💥 ")
		color.Red("Error getting git branch")
		color.Red(err.Error())
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// check if git repo is clean
func getGitClean() bool {

	cmd := exec.Command("git", "diff", "--quiet")
	e := cmd.Run()
	if e != nil {
		return false
	}
	return true
}



// get github origin from git
func getGitOrigin() (string, error) {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	output, err := cmd.Output()
	if err != nil {
		fmt.Print("💥 ")
		color.Red("Error getting git origin")
		color.Red(err.Error())
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// get latest git tag from git 
func getGitTag() string {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// get short commit id from git given a tag
func getGitCommit(tag string) string {
	cmd := exec.Command("git", "rev-parse", "--short", tag)
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// release the project using goreleaser
func release() error {

	err := check()
	if err != nil { return err }

	color.Cyan("Releasing the project ...")
	color.Cyan("This will build the project for multiple platforms and create a new github release.")
	color.White("💬  Make sure you edit the .goreleaser.yml file to set up how the project should get released.")

	name, en := getMainFileName()
	if en != nil { return en }

	version, ev := getVersion(name + ".go")
	if ev != nil { return ev }

	// add a tag for the current version
	color.Cyan("Tagging the current version v" + version + "...")
	cmd := exec.Command("git", "tag", "v"+version)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	e := cmd.Run()
	if e != nil {
		fmt.Print("💥 ")
		color.Red(e.Error())
		return e
	}
	color.Blue("🆗 Git tag added successfully.")

	// run goreleaser release command
	color.Cyan("Running goreleaser release...")

	cmd = exec.Command("goreleaser", "release", "--clean")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	e = cmd.Run()

	if e != nil {
		fmt.Print("💥 ")
		color.Red(e.Error())

		// delete the tag we just created
		color.Cyan("Deleting the git tag v" + version + "...")
		cmd = exec.Command("git", "tag", "-d", "v"+version)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		e = cmd.Run()
		if e != nil {
			fmt.Print("💥 ")
			color.Red(e.Error())
			color.Yellow("⚠  Failed to remove tag, run git tag v" + version + " manually before re-running this command")
			return e
		}
	}

	color.Blue("🆗 goreleaser ran successfully.")
	color.Green("✔  Project released successfully. Check your github page for the new release")
	return nil
}


// generate a scoop manifest file
func generateScoopFile() error {

	color.Cyan("Generating scoop manifest file...")

	// check if the dist/ folder exists in the project directory and if not exit
	if _, err := os.Stat("dist"); os.IsNotExist(err) {
		// warn
		color.Yellow("⚠  dist/ folder does not exist in the project directory.")
		color.White("💬  Make sure you have built the project using gopher release")
		return err
	}

	
	var name, username, version, description, homepage, url string

	color.Cyan("Getting module string from go.mod file...")
	uri, em := getModule()
	if em != nil { return em }

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
	mainfile, en := getMainFileName()
	if en != nil { return en }

	var ev error
	version, ev = getVersion(mainfile + ".go")
	if ev != nil { return ev }

    color.Blue("🆗 Got the project version: " + version)

	color.Cyan("Adding generic description, you can edit it later...")
	description = "A new scoop package"

	color.Cyan("Creating the homepage url...")
	homepage = "https://github.com/" + username + "/" + name

    color.Blue("🆗 Homepage url: " + homepage)

	color.Cyan("Creating the download url...")
	url = "https://github.com/" + username + "/" + name + "/releases/download/v" + version + "/" + name + "_" + version + "_Windows_x86_64.zip"

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

	scoopfile_path := "dist" + string(os.PathSeparator) + name + ".json"

	// write the manifest to the file
	color.Cyan("Creating " + scoopfile_path)
	mfile, err := os.Create(scoopfile_path)
	if err != nil {
		color.Red("Error creating " + scoopfile_path)
		color.Red(err.Error())
		return err
	}
	defer mfile.Close()

    color.Blue("🆗 file created.")

	mfile.WriteString(manifest)
    color.Blue("🆗 scoop file has been written to disk.")

	errors := 0

	color.Cyan("Checking for for existing windows binary release in dist/ folder...")

	// use correct path operator
	checksum_file := "dist" + string(os.PathSeparator) + name + "_" + version + "_checksums.txt"
	zip_file :=  name + "_" + version + "_Windows_x86_64.zip"
	hash := ""

	line, err := findInFile(checksum_file, name+"_"+version+"_Windows_x86_64.zip")

	if err != nil {
		color.Yellow("⚠  Could not find the checksum file: " + checksum_file)
		color.White("💬  Make sure you have built the windows binary using gopher release command.")
		errors++
	} else {
		if line == "" {
			color.Yellow("⚠  Could not find a checksum for " + zip_file + " in the checksum file.")
			color.White("💬  Make sure you have built the windows binary using gopher release command.")
			errors++
		} else {
			hash = strings.Split(line, " ")[0]
			color.Blue("🆗 Found a checksum for " + zip_file +" -> " + hash)
		}
	}

	if hash != "" {
		color.Cyan("Adding the sha256 checksum to the manifest file...")

		bin_line := fmt.Sprintf(`    "url": "%s",`, url)
		new_line := fmt.Sprintf(`    "url": "%s", "hash": "%s",`, url, hash)

		err = replaceInFile(scoopfile_path, bin_line, new_line)
		if err != nil {
			// print a warning
			color.Yellow("⚠  Could not add the hash to the manifest file.")
			color.White("💬  You can add it manually later.")
			errors++
		} else {
			color.Blue("🆗 Manifest file updated successfully.")
		}

	}

	if errors == 0 {
		color.Green("✔  Scoop manifest file " + name + ".json created successfully.")
		return nil
	} else {
		color.Green("⚠  Scoop manifest file " + name + ".json created with some warnings.")
		return fmt.Errorf("scoop manifest creation completed with %d warnings", errors)
	}

}

// searches the file name.go for a constant named version and returns its value
func getVersion(filename string) (string, error) {

	line, err := findInFile(filename, "const version")

	if err != nil {
		fmt.Print("💥 ")
		color.Red("Error opening file " + filename)
		color.Red(err.Error())
		return "", err
	}

	quoted_version := strings.Split(line, "=")[1]
	return strings.Trim(quoted_version, " \""), nil
}

// searches go.mod file for the module name and returns it as string
func getModuleName() (string, error) {

	line, err := findInFile("go.mod", "module")

	if err != nil {
		fmt.Print("💥 ")
		color.Red("Error opening go.mod file")
		color.Red(err.Error())
		return "", err
	}

	url := strings.Split(line, " ")[1]

	// if the url contains / then it's a github url an we need to extract the last part
	if strings.Contains(url, "/") {
		return getName(url), nil
	} else {
		return url, nil
	}
}

// search the go.mod file for the module string and return it
func getModule() (string, error) {

	line, err := findInFile("go.mod", "module")

	if err != nil {
		fmt.Print("💥 ")
		color.Red("Error opening go.mod file")
		color.Red(err.Error())
		return "", err
	}

	return strings.Split(line, " ")[1], nil
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
func installProject() error {

	// get project name from go.mod file
	name, em := getModuleName()
	if em != nil { return em }

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
			return err
		}

		// copy the binary to the bin directory
		color.Cyan("Copying the binary to the bin directory...")

		// on windows cp is a built in shell feature so we will use the copy library instead
		err := cp.Copy(name+".exe", installpath+"\\"+name+".exe")
		if err != nil {
			fmt.Print("💥 ")
			color.Red(err.Error())
			return err
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
			return err
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
			return e
		}

	}

    color.Blue("🆗 copy successful.")

    color.White("💬 Make sure " + installpath + " is in your PATH")
	color.Green("✔  " + name + " installed successfully into " + installpath)
	return nil
}

func createMakefile() error {

	color.Cyan("Creating Makefile...")

	color.Cyan("Getting module name from go.mod file...")
	name, em := getModuleName()
	if em != nil { return em }

	color.Cyan("Generating the Makefile content...")
	content := fmt.Sprintf(`BINARY_NAME=%s

.PHONY: build
build: tidy
	go build

.PHONY: clean
clean:
	go clean
	rm -rf dist
    

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
		return err
	}
	defer mfile.Close()

	color.Cyan("Writing the Makefile content to disk...")
	mfile.WriteString(content)
	color.Green("✔  Makefile created successfully.")
	return nil
}

func createJustfile() error {

	color.Cyan("Creating Justfile...")

	color.Cyan("Getting module name from go.mod file...")
	name, em := getModuleName()
	if em != nil { return em }

	color.Cyan("Generating the Justfile content...")
	content := fmt.Sprintf(`BINARY_NAME := "%s"

build: tidy
	go build

clean:
	fo clean
	rm -rf dist
    

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
		return err
	}
	defer jfile.Close()

	color.Cyan("Writing the Justfile content to disk...")
	jfile.WriteString(content)
	color.Green("✔  Justfile created successfully.")
	return nil
}

func createMainFile() error {

	color.Cyan("Getting module name from go.mod file contents...")
	name := "main"

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
		return err
	}
	defer gfile.Close()

	color.Cyan("Writing the " + name + ".go file content to disk...")
	gfile.WriteString(content)
	color.Blue("🆗 " + name + ".go file created successfully.")
	return nil
}

func createTestFile() error {

	color.Cyan("Getting module name from go.mod file contents...")
	name, em := getMainFileName()

	if em != nil { return em }
	filename := name + "_test.go"

	color.Cyan("Generating the " + filename + " file...")

	content := `package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var (
	binName  = "test"
	cmdPath  string
	exitCode int
)

func TestMain(m *testing.M) {
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}

	build := exec.Command("go", "build", "-o", binName)
	if err := build.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "cannot build %s: %s", binName, err)
		os.Exit(1)
	}

	var err error
	cmdPath, err = filepath.Abs(binName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot get absolute path to %s: %s", binName, err)
		os.Exit(1)
	}

	exitCode = m.Run()

	os.Remove(binName)
	os.Exit(exitCode)
}

func TestNoArgs(t *testing.T) {
	cmd := exec.Command(cmdPath)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run()

	expected := "Usage:"
	if !strings.Contains(out.String(), expected) {
		t.Errorf("expected to contain %q, got %q", expected, out.String())
	}
}

func TestCorrectFlags(t *testing.T) {
	testCases := []struct {
		args     []string
		expected string
	}{
		{[]string{"-v"}, "version"},
		{[]string{"--version"}, "version"},
		{[]string{"-h"}, "Usage:"},
		{[]string{"--help"}, "Usage:"},
	}

	for _, tc := range testCases {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			cmd := exec.Command(cmdPath, tc.args...)
			var out bytes.Buffer
			cmd.Stdout = &out
			cmd.Run()

			if !strings.Contains(out.String(), tc.expected) {
				t.Errorf("expected to contain %q, got %q", tc.expected, out.String())
			}
		})
	}
}

func TestWrongFlag(t *testing.T) {
	cmd := exec.Command(cmdPath, "-wrong")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run()

	expected := "Usage:"
	if !strings.Contains(out.String(), expected) {
		t.Errorf("expected to contain %q, got %q", expected, out.String())
	}
}`

	color.Cyan("Creating the " + filename + " file on disk...")
	gfile, err := os.Create(filename)
	if err != nil {
		fmt.Print("💥 ")
		color.Red("Error creating " + filename + " file")
		color.Red(err.Error())
		return err
	}
	defer gfile.Close()

	color.Cyan("Writing the " + filename + " file content to disk...")
	gfile.WriteString(content)
	color.Blue("🆗 " + filename + " file created successfully.")
	return nil
}


func incString(s string) string {
    n, _ := strconv.Atoi(s)
    return strconv.Itoa(n + 1)
}


// bump the version number in the version constant of the main file
func versionBump(what string) error {

    // check for the existence of go.mod
    if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
        fmt.Print("💥 ")
        color.Red("Error: go.mod file not found in the current directory.")
        return fmt.Errorf("go.mod file not found")
    }

    if what != "major" && what != "minor" && what != "build" && what != "patch"{
        fmt.Print("💥 ")
        color.Red("Invalid argument for bump subcommand. Use major, minor, or patch.")
        printUsage()
        return fmt.Errorf("invalid argument for bump subcommand")
    }

    color.Cyan("Determining the name of the main file...")
	name, em := getMainFileName()
	if em != nil { return em }
	
    // get the current version
    color.Cyan("Getting current version from " + name + ".go file...")
    version, ev := getVersion(name + ".go")
	if ev != nil { return ev }

    color.Blue("🆗 Current version is " + version)
    color.Cyan("Bumping the version number...")

    // split the version into parts
    parts := strings.Split(version, ".")
    major := parts[0]
    minor := parts[1]
    build := parts[2]

    // increment the build number
	switch what {

	case "build", "patch":
		build = incString(build)
	case "minor":
		minor = incString(minor)
		build = "0"
	case "major":
		major = incString(major)
		minor = "0"
		build = "0"
	}

    // create the new version string
    new_version := major + "." + minor + "." + build

    color.Blue("🆗 New version is " + new_version)

    color.Cyan("Replacing the version number in " + name + ".go file...")

    find := "const version = \"" + version + "\""
    replace := "const version = \"" + new_version + "\""

	err := replaceInFile(name+".go", find, replace)

    if err != nil {
        fmt.Print("💥 ")
        color.Red("Error modifying the source file")
        color.Red(err.Error())
        return err
    }

    color.Blue("🆗 Version number replaced successfully.")
    color.Green("✔  Version bumped to " + new_version)

	return nil
}
