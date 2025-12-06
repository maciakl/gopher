# 🐿 gopher

A minimal, oppinionated go project bootstrapping tool.

Gopher will bootstrap a project with sensible defaults, and provide additional project management tooling for writing simple command line appssuch as [weather](https://github.com/maciakl/weather), [jjmp](https://github.com/maciakl/jjmp), [myip](https://github.com/maciakl/myip) or itself. It is not intended to be used for web based projects, GUI applications, services or large projects.

    Usage: gopher [subcommand] <arguments>

    Subcommands:
      init <string>
            bootstrap a new project with a given <string> in the format
            username/project or github.com/username/project
      make
            print project information known to gopher
      make
            create a simple Makefile for the project
      just
            create a simple Justfile for the project  
      release
            build the project using goreleaser
      scoop
            generate a Scoop manifest file for the project
      install
            install the project binary in the user's private bin directory
            typically ~/bin or %USERPROFILE%\bin
      bump <string>
            bump the project version number in the main file; the <string> can
            be one of: major, minor, build | patch
      version
            display version number of this tool and exit
      help
            display this help message and exit

## Using the tool

Currently gopher supports 7 actions.

- Bootstraping a project: `init`
- Generating build files using: `make` and `just`
- Building and packaging a project: `release`
- Installing a project: `install`
- Creating a [Scoop.sh](https://scoop.sh) manifest: `scoop`
- Bumping the version number in your main file to the next one: `bump`

### Create a new project

To create a new go project run:

    gopher init <string>

The `<string>` can be a project name or a git repository uri (you can leave out https://github.com/ part if it's a github repository).

If you provide a valid uri, gopher will extract following information from it: 

- the `project_name` from the uri and use it to create the project.
- your `github_username`
- the github `repository_address` (ssh format)

If all you provided was a project name gopher will try the following:

1. Check if `GOPHER_USERNAME` variable is set, and if so use it's contents as your `github_username`
2. If the variable is not set, gopher will stop and ask you to type in your `github_username`
3. Construct the `uri` and `repo_address` from the above

Once it knows all the relevant information it will do the following:

- create a folder `project_name`
- inside it will:
  - run `go mod init uri`
  - create `.gitignore` file
  - create `README.md` file
  - create `name.go` with simple hello world code
  - run `git init -b main`
  - run `git remote add origin repo_address`
 
So, for example, if you run:

    gopher init maciakl/test

Gopher will generate the following folder structure:

    test/
     |
     +--- .gitignore
     |
     +--- go.mod
     |
     +--- README.md
     |
     +--- test.go
     |
     +--- .goreleser.yml

### Generating Build Files

You can use the `gopher` tool to create simple build files for your project. To create a simple `Makefile` run:

    gopher make

This will create a simple `Makefile` with couple of targets such as `build`, `run`, `clean`, `tidy` and `test`.

You can also create an equivalent build file for [Just](https://github.com/casey/just) task runner by running:

    gopher just

These build files are intended to be scaffolding that you are encouraged to customize to fit your project.

### Releasing a project

To build the project and create a set of zip files for different distribution platforms run:

    gopher release

This must be run in the project directory. It will:

- add a git tag for the current version of your project (extracted from the version const in your main file)
- cross compile the project for windows, mac and linux
- generate zip files for each os named tagged with the version number and platform
- run `goreleaser release --clean` to create a github release with the generated zip files attached

You can control the behavior of the release process by editing the `.goreleser.yml` file created by `gopher init` command. See [goreleaser documentation](https://goreleaser.com/) for more details.

⚠️ Note: you must set up a github token and make it available for `goreleaser`. 

Create your own github token [here](https://github.com/settings/tokens/new?scopes=repo,write:packages).

The recommended way to expose this to `goreleaser` to add the following lines to `.goreleaser.yml` file:

```yaml
env_files:
  github_token: .env
```
Then create a `.env` file in the project directory with your Github token. Gopher automatically puts that file into your project's `.gitignore` for you so you don't accidentally commit it with your code. 

Alternatively, if you give the token permission to all your repos, you can set the `GITHUB_TOKEN` environment variable on your system and have it available for all your projects.


### Generate a Scoop Manifest

To create a Scoop manifest (see [scoop.sh](https://scoop.sh)) for the project run:

    gopher scoop

This will generate `name.json` file that you can add to your scoop bucket using the data in your `go.mod` file. If the module name is not a valid uri, gopher will ask you to provide your Github username
and then use that to create the appropriate URL's for the manifest.

Don't forget to edit the description and verify all the details are correct before uploading the file.

For example, if you run the following commands:

    gopher init github.com/maciakl/test
    gopher scoop

This will yield the following `test.json` file in the project directory:

```json
    {
        "version": "0.1.0",
        "description": "A new scoop package",
        "homepage": "https://github.com/maciakl/test",
        "checkver": "github",
        "url": "https://github.com/maciakl/test/releases/download/v0.1.0/test_0.1.0_Winods_x86_64.zip",
        "bin": "test.exe",
        "license": "freeware"
    }
```

### Installing a project binary

To install the project on your system run

    gopher install

This will rebuild the project using `go build` and then copy the executable to a directory defined in your `GOPHER_INSTALLPATH` environment variable. If this variable is not set, gopher will attempt to use `~/bin/` on mac/linux or `%USERPROFILE%\bin\` on windows. 

If such directory does not exist, gopher will bail out with an error.

⚠️ Note: you must create the directory and add it to your `PATH` manually, gopher won't do that for you

### Bumping

The `bump` subcommand will search your code for a line that looks something like this:

    const version = "1.2.3"

It will parse out the current version number, and increment and/or update the appropriate digits.

The different digits are called: `MAJOR.MINOR.PATCH`.

You can use the following subcommands:

| Command | Current Version | New Version |
| --- | --- | --- |
| `gopher bump patch` | `A.B.C` | `A.B.C+1` |
| `gopher bump minor` | `A.B.C` | `A.B+1.0` |
| `gopher bump major` | `A.B.C` | `A+1.0.0` |

Currently `gopher bump build` will also work and increment the `patch` number.

⚠️ Note: just in case, commit your changes before using this command as your file is edited in place, so if something goes horribly wrong, you might lose work.

## Configuration (optional)

You can configure gopher by setting appropriate envionment variables.

Following settings are currently supported:

| Environment Variable | Description |
| --- | --- |
| `GOPHER_USERNAME` | Your GitHub username. Setting this variable will prevent gopher from asking you to type it in. |
| `GOPHER_INSTALLPATH` | Default binary install location. If this variable is not set, gopher will try to use `~/bin` (or `%USERPROFILE%\bin` on Windows). |

Use your preferred method for setting environment variables appropriate for your OS. Here are some examples:

Powershell on Windows, put this is `$PROFILE`:

    $env:GOPHER_USERNAME="maciakl"
    $env:GOPHER_INSTALLPATH="c:\Tools"

Zsh on Linux, put this in your `.zshrc`:

    export GOPHER_USERNAME=maciakl
    export GOPHER_INSTALLPATH=/opt/gopher/


## Examples

Sample screenshot of using `gopher` to create a go project, generate a scoop manifest and compiling release files for windows, linux and mac (all done in Powershell on Windows):

![scr2](https://github.com/user-attachments/assets/916d93d6-6e1a-480e-beeb-c3c631903c46)

## Installing

Install via go:
 
    go install github.com/maciakl/gopher@latest

On Windows, this tool is distributed via `scoop` (see [scoop.sh](https://scoop.sh)).

First, you need to add my bucket:

    scoop bucket add maciak https://github.com/maciakl/bucket
    scoop update

Next simply run:
 
    scoop install gopher

If you don't want to use `scoop` you can simply download the executable from the release page and extract it somewhere in your path.
