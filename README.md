# 🐿 gopher

A minimal go project bootstrapping tool.

    Usage: gopher [subcommand] <arguments>

    Subcommands:
      init <string>
            bootstrap a new project with a given <string> name or url
      make
            build the project using a Makefile, falle back on wrap
      wrap
            build the project for windows, linux and mac, then and zip the binaries
      scoop
            generate a Scoop.sh manifest file for the project
      install
            install the project binary in the user's private bin directory
      version
            display version number and exit
      help
            display this help message and exit

## Using the tool

Currently gopher supports 4 actions.

- Bootstraping a project `init`
- Building a project `make` or `wrap`
- Installing a project `install`
- Creating a `scoop` manifest `scoop`

### Create a new project

To create a new go project run:

    gopher init <string>

The `<string>` can be a project name or a git repository uri (without https:// part).

Goper will extract the project `name` from the uri and use it to create the project. Then it will:

- create a folder `name`
- inside it will:
  - run `go mod init uri`
  - create `.gitignore` file
  - create `README.md` file
  - create `name.go` with simple hello world code
  - run `git init -b main`
 
### Compile the project and create a zip files for distribution

In most cases you should use the following command while in the project directory:

    gopher make

If you do:

- Gopher will check if a `Makefile` exists in the project directory
- If it does, it will run the make command and execute the default build
- If there is no `Makefile`, it will use the gopher defaults instead

To build the project using the gopher defaults and skip Makefile even if one exists run:

    gopher wrap

This must be run in the project directory. It will:

- cross compile the project for windows, mac and linux
- generate zip files for each os named `name_win.zip`, `name_mac.zip` and `name_lin.zip` respectively

Note that if a `Makefile` is present, the `make` subcommand does not generate any zip files and assumes any kind of packaging will be handled by your project Makefile.

### Generate a Scoop Manifest

To create a Scoop manifest (see [scoop.sh](https://scoop.sh)) for the project run:

    gopher scoop

This will generate `name.json` file that you can add to your scoop bucket using the data in your `go.mod` file. If the module name is not a valid uri, gopher will ask you to provide your Github username
and then use that to create the appropriate URL's for the manifest.

Don't forget to edit the description and verify all the details are correct before uploading the file.

### Installing

To install the project on your system run

    gopher install

This will rebuild the project using `go build` and then copy the executable to your private user directory. This is `~/bin/` on mac/linux or `%USERPROFILE%\bin\` on windows. If such directory does not exist, gopher will bail out with an error.

⚠️ Note: you must create the `bin` directory and add it to your `PATH` manually, gopher won't do that for you

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
