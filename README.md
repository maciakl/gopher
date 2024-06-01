# gopher

A minimal go project bootstrapping tool

    Usage:
      -init string
            bootstrap a new project with a given name
      -version
            display version number and exit
      -wrap
            build the project and zip it (windows only for now)
     -scoop
            create a scoop manifest file for the project

## Using the tool

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
 
To build the project and generate zip file with the executable run:

    gopher -wrap

This must be run in the project directory. It will:

- run `go build`
- zip up `name.exe` and create `name_win.zip`

Currently, wrapping is only supported on windows.

To create a scoop manifest file for the project run:

    gopher -scoop

This will generate `name.json` file that you can add to your scoop bucket. Don't forget to edit the description and verify all the details are correct before uploading the file.
 

## Installing

 On Windows, this tool is distributed via `scoop` (see [scoop.sh](https://scoop.sh)).

 First, you need to add my bucket:

    scoop bucket add maciak https://github.com/maciakl/bucket
    scoop update

 Next simply run:
 
    scoop install gopher

If you don't want to use `scoop` you can simply download the executable from the release page and extract it somewhere in your path.
