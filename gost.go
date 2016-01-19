package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"code.google.com/p/goauth2/oauth"
	"github.com/atotto/clipboard"
	"github.com/docopt/docopt.go"
	"github.com/google/go-github/github"
)

var (
	version = "Gost 1.1.1"
	usage   = `Gost - A simple command line utility for easily creating Gists for Github

        Usage:
         gost [--file=<file>] [--clip] [--name=<name>] [--description=<description>] [--token=<token>] [--public] [--paste]
         gost (--help | --version)

        Options:
          -t --token=<token>             Optional Github API authentication token. If excluded, your Gist will be created anonymously.
          -f --file=<file>               Create a Gist from file.
          -n --name=<name>               Optional name for your new Gist.
          -d --description=<description> Optional description for your new Gist.
          -c --clip                      Create a Gist from the contents of your clipboard.
          -p --public                    Make this Gist public [default: false].
		  -P --paste                     Will paste your latest gist to stdout and local clipboard.
          -h --help                      Will display this help screen.
          -v --version                   Displays the current version of Gost.`
)

func main() {
	arguments, err := docopt.Parse(usage, nil, true, version, false)

	if err != nil {
		fmt.Println("Could not properly execute command; exiting ...")
		os.Exit(1)
	}

	file := arguments["--file"]
	content := ""

	fi, err := os.Stdin.Stat()
	if err != nil {
		fmt.Println("Cannot read from Stdin;", err)
		os.Exit(1)
	}

	if fi.Mode()&os.ModeNamedPipe != 0 {
		stdin, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Println("Cannot read from Stdin;", err)
		}

		content = string(stdin)
		if len(strings.TrimSpace(content)) == 0 {
			fmt.Println("Stdin is empty; exiting ...")
			os.Exit(1)
		}
	} else if file == nil {
		if arguments["--clip"] == false {
			fmt.Println("Please specify a valid file with -f or --file, or add something to your clipboard.")
			os.Exit(1)
		}

		content, err = clipboard.ReadAll()

		if err != nil {
			fmt.Println("Error reading clipboard; exiting ...")
			os.Exit(1)
		}

		if len(strings.TrimSpace(content)) == 0 {
			fmt.Println("Your clipboard is empty; exiting ...")
			os.Exit(1)
		}
	} else {
		bytes, err := ioutil.ReadFile(file.(string))
		if err != nil {
			fmt.Println("Invalid file specified;", err)
			os.Exit(1)
		}
		content = string(bytes)
	}

	name := arguments["--name"]
	if name == nil {
		if arguments["--file"] != nil {
			name = path.Base(file.(string))
		} else {
			name = "gistfile"
		}
	}

	token := arguments["--token"]
	if token == nil {
		token = os.Getenv("GOST")
	}

	client := github.NewClient(nil)
	if len(strings.TrimSpace(token.(string))) > 0 {
		t := &oauth.Transport{
			Token: &oauth.Token{AccessToken: token.(string)},
		}

		client = github.NewClient(t.Client())
	}

	description := arguments["--description"]
	if description == nil {
		description = ""
	}

	public := arguments["--public"].(bool)

	desc := description.(string)

	input := &github.Gist{
		Description: &desc,
		Public:      &public,
		Files: map[github.GistFilename]github.GistFile{
			github.GistFilename(name.(string)): github.GistFile{Content: &content},
		},
	}

	fmt.Println("Gosting Gist ... ")

	gist, _, err := client.Gists.Create(input)
	if err != nil {
		fmt.Println("Unable to create gist:", err)
		os.Exit(1)
	}

	fmt.Println("Done!")
	fmt.Println("Gist URL:", string(*gist.HTMLURL))
	os.Exit(0)
}
