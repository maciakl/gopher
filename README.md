# gopher

A minimal go project bootstrapping tool

    Usage:
      -init string
            bootstrap a new project with a given name
      -version
            display version number and exit
      -wrap string
            build the project and zip it (windows only for now)

## Installing

 On Windows, this tool is distributed via `scoop` (see [scoop.sh](https://scoop.sh)).

 First, you need to add my bucket:

    scoop bucket add maciak https://github.com/maciakl/bucket
    scoop update

 Next simply run:
 
    scoop install gopher

If you don't want to use `scoop` you can simply download the executable from the release page and extract it somewhere in your path.
