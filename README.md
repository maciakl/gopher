# 🐿 gopher

A minimal go project bootstrapping tool.

    Usage:
      -init string
            bootstrap a new project with a given name
      -version
            display version number and exit
      -wrap
            build the project and zip it
     -scoop
            create a scoop manifest file for the project

## Using the tool

Currently gopher supports 3 actions.

### Create a new project

To create a new go project run:

    gopher -init name

This will:

- create a folder `name`
- inside it will:
  - run `go mod init name`
  - create `.gitignore` file
  - create `README.md` file
  - create `name.go` with simple hello world code
  - run `git init -b main`
 
### Compile the project and create a zip files for distribution

To build the project and generate zip file with the executable run:

    gopher -wrap

This must be run in the project directory. It will:

- cross compile the project for windows, mac and linux
- generate zip files for each os named `name_win.zip`, `name_mac.zip` and `name_lin.zip` respectively


### Generate a Scoop Manifest

To create a Scoop manifest (see [scoop.sh](https://scoop.sh)) for the project run:

    gopher -scoop

This will generate `name.json` file that you can add to your scoop bucket. 

Don't forget to edit the description and verify all the details are correct before uploading the file.

## Examples

Sample screenshot of using `gopher` to create a go project, compile and wrap the executable into a zip file and generate a scoop manifest (all done in Powershell on Windows):

<img width="682" alt="scr" src="https://github.com/maciakl/gopher/assets/189576/8fbf8eea-eff7-41c2-9dec-b4f47ef92ba9">

## Installing

 On Windows, this tool is distributed via `scoop` (see [scoop.sh](https://scoop.sh)).

 First, you need to add my bucket:

    scoop bucket add maciak https://github.com/maciakl/bucket
    scoop update

 Next simply run:
 
    scoop install gopher

If you don't want to use `scoop` you can simply download the executable from the release page and extract it somewhere in your path.
