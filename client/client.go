package main

import "bufio"
import "os/exec"
import "errors"
import "flag"
import "fmt"
import "io"
import "io/ioutil"
import "net/http"
import "net/url"
import "os"
import "path"
import "strings"

var url_arg = flag.String("url", "", "Game URL with exactly one player key")
var player_arg = flag.String("player", "", "Command to run player program")

// Parsed from --url argument
var game_url url.URL
var game_id, black_key, white_key *string

// Based on --player argument
var player_cmd exec.Cmd
var player_in io.WriteCloser
var player_out, player_err io.ReadCloser

func readStrings(input io.ReadCloser, delimiter byte, output chan<- string) {
	reader := bufio.NewReader(input)
	for {
		if line, err := reader.ReadString(delimiter); err == io.EOF {
			if line != "" {
				output <- line
			}
			break
		} else if err != nil {
			fmt.Println("Error reading from player!")
			break
		} else {
			output <- line
		}
	}
	input.Close()
}

func pollGame(version int) error {
	poll_url := game_url
	poll_url.Path = path.Join(path.Dir(poll_url.Path), "poll")
	poll_url.Fragment = ""
	params := url.Values{}
	params.Set("game", *game_id)
	params.Set("version", fmt.Sprintf("%d", version))
	poll_url.RawQuery = params.Encode()
	fmt.Println(poll_url.String())
	if response, err := http.Get(poll_url.String()); err != nil {
		return err
	} else {
		defer response.Body.Close()
		if response.StatusCode == 200 /* OK */ {
			if body, err := ioutil.ReadAll(response.Body); err != nil {
				return err
			} else {
				fmt.Println(string(body))
				// TODO: parse body, return it
				return nil
			}
		} else if response.StatusCode == 204 /* No Content */ {
			fmt.Println("TODO: poll again!")
			return nil
		} else {
			return errors.New(fmt.Sprintf(
				"Unexpected response status: %s",
				response.Status))
		}
	}
}

func parseGameUrl(url_str string) error {
	if the_url, err := url.Parse(url_str); err != nil {
		return errors.New("Could not parse game URL.")
	} else if fragment_map, err := url.ParseQuery(the_url.Fragment); err != nil {
		return errors.New("Could not parse URL fragment.")
	} else {
		num_keys := 0
		num_games := 0
		for key, values := range fragment_map {
			for _, value := range values {
				switch key {
				case "game":
					game_id = &value
					num_games++
				case "white":
					white_key = &value
					num_keys++
				case "black":
					black_key = &value
					num_keys++
				}
			}
		}
		if num_games != 1 {
			return errors.New("Need exactly one game id.")
		}
		if num_keys != 1 {
			return errors.New("Need exactly one player key.")
		}
		game_url = *the_url
		return nil
	}
}

func runPlayerCommand(command string) error {
	if argv := strings.Fields(*player_arg); len(argv) == 0 {
		return errors.New("No player command given!")
	} else if name, err := exec.LookPath(argv[0]); err != nil {
		return errors.New("Can't find player executable!")
	} else if dir, err := os.Getwd(); err != nil {
		return errors.New("Can't get current working directory!")
	} else {
		player_cmd = exec.Cmd{Path: name, Args: argv, Dir: dir}
		if player_in, err = player_cmd.StdinPipe(); err != nil {
			return errors.New("Could not open player input!")
		} else if player_out, err = player_cmd.StdoutPipe(); err != nil {
			return errors.New("Could not open player output!")
		} else if player_err, err = player_cmd.StderrPipe(); err != nil {
			return errors.New("Could not open player errors!")
		} else if err := player_cmd.Start(); err != nil {
			return errors.New("Could not start player!")
		}
		return nil
	}
}

func main() {
	flag.Parse()

	if err := parseGameUrl(*url_arg); err != nil {
		fmt.Println("Could not parse game URL!", err)
		return
	}
	if err := pollGame(0); err != nil {
		fmt.Println("Could not poll initial game state!", err)
		return
	}
	if err := runPlayerCommand(*player_arg); err != nil {
		fmt.Println("Could not execute player command! ", err)
	}
	go io.Copy(os.Stderr, player_err)
	lines := make(chan string)
	go readStrings(player_out, '\n', lines)

	for line := range lines {
		fmt.Println("Line:", line)
	}

	player_cmd.Wait()

	// TODO!
	//  - one go routine to poll server
	//  - one go routine to read commands, parse them, and tell the server
	fmt.Println(game_url, white_key, black_key, player_in) // used
	// TODO: print stderr from player too
}
