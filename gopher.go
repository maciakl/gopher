package main

import (
    "os"
    "fmt"
    "flag"
    "bufio"
    "runtime"
    "strings"
    "os/exec"
    "path/filepath"

    "github.com/fatih/color"
)

const version = "0.1.5"

const templateUrl = "https://gist.githubusercontent.com/maciakl/b5877bcb8b1ad21e2e798d3da3bff13b/raw/3fb1c32e3766bf2cf3926ee72225518e827a1228/hello.go"

func main() {

    var name string
    flag.StringVar(&name, "init", "", "bootstrap a new project with a given name")

    var wrap bool
    flag.BoolVar(&wrap, "wrap", false, "build the project and zip it (windows only for now)")

    var scoop bool
    flag.BoolVar(&scoop, "scoop", false, "generate a scoop manifest file for the project")
    
    var ver bool
    flag.BoolVar(&ver, "version", false, "display version number and exit")
    flag.Parse()

    // show version and exit
    if ver {
        fmt.Println(filepath.Base(os.Args[0]), "version", version)
        os.Exit(0)
    }
    
    // bootstrap a new project
    if name != "" && !wrap && !scoop {
        createProject(name)
    }

    // build the project and zip it
    if wrap && name == "" && !scoop {
        buildProject()
    }


    // generate a scoop manifest file
    if scoop && name == "" && !wrap {
        generateScoopFile()
    }


    if name == "" && !wrap && !scoop {
        banner()
        color.Red("❌  No arguments provided. Use -init, -wrap or -scoop.")
    }
    
}

func createProject(name string) {

    banner()

    errors := 0

    // create a new directory
    color.Cyan("Creating project "+ name + "...")
    os.Mkdir(name, 0755)
    os.Chdir(name)


    // run the go mod init command
    color.Cyan("Running go mod init "+ name + "...")
    cmd := exec.Command("go", "mod", "init", name)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    e := cmd.Run()

    if e != nil { errors++ }

    // create .gitignore file 
    color.Cyan("Creating .gitignore file...")
    gfile, err := os.Create(".gitignore")
    if err != nil {
        color.Red("Error creating .gitignore", err)
        os.Exit(1)
    }
    gfile.WriteString(name+"\n")
    gfile.WriteString(name+"*.exe\n")
    gfile.WriteString(name+".zip\n")
    gfile.WriteString(name+".tgz\n")
    gfile.WriteString(name+"_*.zip\n")
    gfile.WriteString(name+"_*.tgz\n")
    gfile.Close()

    // create README.md file
    color.Cyan("Creating README.md file...")
    rfile, err := os.Create("README.md")
    if err != nil {
        fmt.Println("Error creating README.md file:", err)
        os.Exit(1)
    }
    rfile.WriteString("# " + name + "\n")
    rfile.Close()


    // download [name].go file fom the gist template url
    color.Cyan("Creating "+name+".go file")
    cmd = exec.Command("curl", "-o", name+".go", templateUrl)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    e = cmd.Run()

    if e != nil { errors++ }

    // run the git init command with -b main
    color.Cyan("Running git init -b main...")
    cmd = exec.Command("git", "init", "-b", "main")
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    e = cmd.Run()

    if e != nil { errors++ }

    // print the success message
    if errors == 0 {
        color.Green("✔  Project "+ name + " created successfully.")
    } else {
        color.Green("⚠  Project "+ name + " created with some errors.")
    }

}

func buildProject() {

    banner()

    errors := 0

    // get the module name from go.mod file
    color.Cyan("Getting module name from go.mod file...")
    name := getModuleName()


    // change to the current directory (pwd)
    dir, err := os.Getwd()
    if err != nil {
        color.Red("Error getting the current directory", err)
        os.Exit(1)
    }
    os.Chdir(dir)

    // run the go build 
    color.Cyan("Running go build...")
    cmd := exec.Command("go", "build")
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    e := cmd.Run()

    if e != nil { errors++ }

    // detect the operating system
    color.Cyan("Detecting the operating system...")
    current_os := runtime.GOOS
    color.Cyan("Operating system is "+current_os)


    var suffix string
    if current_os == "windows" { suffix = "win" }
    if current_os == "darwin" { suffix = "mac" }
    if current_os == "linux" { suffix = "lin" }

    ext := ""
    if current_os == "windows" { ext = ".exe" }

    // create a zip file with the windows executable
    color.Cyan("Creating "+name+"_"+suffix+".zip file...")
    cmd = exec.Command("zip", "-r", name+"_"+suffix+".zip", name+ext)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    e = cmd.Run()

    if e != nil { errors++ }
    
    // print the success message
    if errors == 0 {
        color.Green("✔  Project "+ name + " wrapped successfully.")
    } else {
        color.Green("⚠  Project "+ name + " wrapped with some errors.")
    }

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
    version = getVersion(name+".go")

    color.Cyan("Adding generic description, you can edit it later...")
    description = "A new scoop package"

    color.Cyan("Creating the homepage url...")
    homepage = "https://github.com/"+username+"/"+name

    color.Cyan("Creating the download url...")
    url = "https://github.com/"+username+"/"+name+"/releases/download/v"+version+"/"+name+"_win.zip"


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
    color.Cyan("Creating "+name+".json file")
    mfile, err := os.Create(name+".json")
    if err != nil {
        color.Red("Error creating "+name+".json file", err)
        os.Exit(1)
    }
    defer mfile.Close()

    mfile.WriteString(manifest)
    color.Green("✔  Scoop manifest file "+name+".json created successfully.")

}


// searches the file name.go for a constant named version and returns its value
func getVersion(filename string) string {
    // open the file
    file, err := os.Open(filename)
    if err != nil {
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

