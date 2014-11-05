package main

import "ayu"
import "bufio"
import "bytes"
import "os/exec"
import "errors"
import "flag"
import "fmt"
import "io"
import "io/ioutil"
import "log"
import "encoding/json"
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

// Fetched from server
var game_state ayu.State

// Based on --player argument
var player_cmd exec.Cmd
var player_in io.WriteCloser
var player_out, player_err io.ReadCloser

func readStrings(input io.ReadCloser, delimiter byte, output chan<- string) {
	reader := bufio.NewReader(input)
	for {
		if line, err := reader.ReadString(delimiter); line != "" {
			output <- line
		} else if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal("Error reading from player!", err)
			break
		}
	}
	fmt.Println("Exiting")
	input.Close()
}

func relativePathToUrl(rel_path string) url.URL {
	return url.URL{
		game_url.Scheme, game_url.Opaque, game_url.User, game_url.Host,
		path.Join(path.Dir(game_url.Path), rel_path), "", ""}
}

func pollGame(version int) error {
	params := url.Values{}
	params.Set("game", *game_id)
	params.Set("version", fmt.Sprintf("%d", version))
	poll_url := relativePathToUrl("poll")
	poll_url.RawQuery = params.Encode()
poll:
	if response, err := http.Get(poll_url.String()); err != nil {
		return err
	} else {
		defer response.Body.Close()
		if response.StatusCode == 200 /* OK */ {
			if body, err := ioutil.ReadAll(response.Body); err != nil {
				return err
			} else {
				var state struct {
					NextPlayer int
					Size       int
					Fields     ayu.Fields
					History    ayu.History
				}
				if err := json.Unmarshal(body, &state); err != nil {
					return err
				}
				if state.Size != ayu.S {
					return fmt.Errorf(
						"Unexpected board size: %d (expected: %d)",
						state.Size, ayu.S)
				}
				if len(state.History) != version {
					return fmt.Errorf(
						"Unexpected number of moves: %d (expected: %d)",
						len(state.History), version)
				}
				game_state = ayu.State{state.Fields, state.History}
				return nil
			}
		} else if response.StatusCode == 204 /* No Content */ {
			// TODO: rate limit polling!
			goto poll
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

func postLastMove() error {
	key := ""
	if white_key != nil {
		key = *white_key
	} else if black_key != nil {
		key = *black_key
	}
	version := len(game_state.History) - 1
	update := map[string]interface{}{
		"game":    game_id,
		"version": version,
		"key":     key,
		"move":    game_state.History[version],
	}
	update_url := relativePathToUrl("update")
	if update_bytes, err := json.Marshal(update); err != nil {
		return err
	} else if response, err := http.Post(update_url.String(), "application/json",
			bytes.NewReader(update_bytes)); err != nil {
		return err
	} else if response.StatusCode != 200 {
		return fmt.Errorf("Unexpected response status: %d %s",
			response.StatusCode, response.Status)
	}
	return nil
}

func main() {
	flag.Parse()
	if err := parseGameUrl(*url_arg); err != nil {
		fmt.Println("Could not parse game URL!", err)
		return
	}
	if err := runPlayerCommand(*player_arg); err != nil {
		fmt.Println("Could not execute player command! ", err)
		return
	}
	game_state.Create()
	go io.Copy(os.Stderr, player_err)
	lines := make(chan string)
	go readStrings(player_out, '\n', lines)
	for !game_state.Over() {
		if (game_state.Next() == 0) == (white_key != nil) {
			// Player's turn
			if len(game_state.History) == 0 {
				player_in.Write([]byte("Start\n"))
			}
			move_str := strings.TrimSpace(<-lines)
			if move, ok := ayu.ParseMove(move_str); !ok {
				fmt.Println("Could not parse move:", move_str)
				break
			} else if !game_state.Execute(move) {
				fmt.Println("Player made invalid move:", move_str)
				break
			} else if err := postLastMove(); err != nil {
				fmt.Printf("Failed to post move '%s': %s\n", move, err)
				break
			} else {
				fmt.Println(">", move)
			}
		} else {
			// Opponent's turn
			if err := pollGame(len(game_state.History) + 1); err != nil {
				fmt.Println("Could not poll game state!", err)
				break
			}
			last_move := game_state.History[len(game_state.History)-1]
			fmt.Println("<", last_move)
			player_in.Write([]byte(last_move.String() + "\n"))
		}
	}
	fmt.Println("Game over.")
	player_in.Write([]byte("Quit\n"))
	player_cmd.Wait()
}
