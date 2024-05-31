package main

import (
    "os"
    "fmt"
    "flag"
    "os/exec"
    "path/filepath"

    "github.com/fatih/color"
)

const version = "0.1.0"
const templateUrl = "https://gist.githubusercontent.com/maciakl/b5877bcb8b1ad21e2e798d3da3bff13b/raw/3fb1c32e3766bf2cf3926ee72225518e827a1228/hello.go"

func main() {

    var name string
    flag.StringVar(&name, "init", "", "bootstrap a new project with a given name")

    var wrap string
    flag.StringVar(&wrap, "wrap", "", "build the project and zip it (windows only for now)")

    var ver bool
    flag.BoolVar(&ver, "version", false, "display version number and exit")
    flag.Parse()

    // show version and exit
    if ver {
        fmt.Println(filepath.Base(os.Args[0]), "version", version)
        os.Exit(0)
    }
    
    // bootstrap a new project
    if name != "" {
        createProject(name)
    }


    // build the project and zip it
    if wrap != "" {
        buildProject(wrap)
    }
    
}

func createProject(name string) {

    color.Magenta("Gopher v" + version + "\n")

    // create a new directory
    color.Cyan("Creating project "+ name)
    os.Mkdir(name, 0755)
    os.Chdir(name)


    // run the go mod init command
    color.Cyan("Running go mod init "+ name)
    cmd := exec.Command("go", "mod", "init", name)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Run()

    // create .gitignore file 
    color.Cyan("Creating .gitignore file")
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
    color.Cyan("Creating README.md file")
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
    cmd.Run()

    // run the git init command with -b main
    color.Cyan("Running git init -b main")
    cmd = exec.Command("git", "init", "-b", "main")
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Run()

    // print the success message
    color.Green("✔  Project "+ name + " created successfully.")

}

func buildProject(name string) {

    color.Magenta("Gopher v" + version + "\n")

    // change to the current directory (pwd)
    dir, err := os.Getwd()
    if err != nil {
        color.Red("Error getting the current directory", err)
        os.Exit(1)
    }
    os.Chdir(dir)

    

    // run the go build 
    color.Cyan("Running go build for windows")
    cmd := exec.Command("go", "build")
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Run()

    // create a zip file with the windows executable
    color.Cyan("Creating "+name+"_win.zip file")
    cmd = exec.Command("zip", "-r", name+"_win.zip", name+".exe")
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Run()
    
    // print the success message
    color.Green("✔  Project "+ name + " wrapped successfully.")

}
