package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/kamildrazkiewicz/go-stanford-nlp"
	_ "github.com/mattn/go-sqlite3"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

type Cmd struct {
	id   string
	desc string
	cb   func(s *discordgo.Session, m *discordgo.MessageCreate, args []string)
}

var (
	botID    		string
	commands 		[]Cmd
	db       		*sql.DB
	err      		error
	re       		*regexp.Regexp
	tagger   		*pos.Tagger
	token    		string
	model_path	string
	tagger_path string
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.StringVar(&model_path, "model", "$GOPATH/src/hotelcalifornia/x64/stanford-postagger/models/english-left3words-distsim.tagger", "Path to the tagger's model. Default `$GOPATH/src/hotelcalifornia/x64/stanford-postagger/models/english-left3words-distsim.tagger`")
	flag.StringVar(&tagger_path, "tagger", "$GOPATH/src/hotelcalifornia/x64/stanford-postagger/stanford-postagger-3.7.0.jar", "Path to the tagger JAR. Default `$GOPATH/src/hotelcalifornia/x64/stanford-postagger/stanford-postagger-3.7.0.jar`")
	flag.Parse()
	rand.Seed(time.Now().Unix())
	re = regexp.MustCompile(`[^a-zA-Z0-9\s.,?!;:'"\[\]/\\()\-_+@#$%^&*|<>=]`)
}

func getWords(pos string) []string {
	rows, err := db.Query(`SELECT word FROM words WHERE pos=?`, pos)
	if err != nil {
		fmt.Println("error querying databse for pos=", pos, ",", err)
		return nil
	}
	var results []string
	for rows.Next() {
		var word string
		err := rows.Scan(&word)
		if err != nil {
			fmt.Println("error reading query result for pos=", pos, ",", err)
			return nil
		}
		results = append(results, word)
	}
	return results
}

func initCommands() {
	commands = append(commands, Cmd{
		"",
		"",
		func(s *discordgo.Session, m *discordgo.MessageCreate, _ []string) {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Usage: `*<cmd> <args>`\nFor more, see `*help`")
		},
	})
	commands = append(commands, Cmd{
		"ping",
		"ping me!",
		func(s *discordgo.Session, m *discordgo.MessageCreate, _ []string) {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Pong!")
		},
	})
	commands = append(commands, Cmd{
		"test",
		"this is a test",
		func(s *discordgo.Session, m *discordgo.MessageCreate, _ []string) {
			_, _ = s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
				Title: "`THE FOLLOWING IS A TEST OF THE EMERGENCY BROADCAST SYSTEM`",
				Image: &discordgo.MessageEmbedImage{
					URL: "https://i.ytimg.com/vi/WnRrPqgKBS0/hqdefault.jpg",
				},
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "**DOOT DOOT**",
						Value: "doot",
					},
				},
			})
		},
	})
	commands = append(commands, Cmd{
		"ship",
		"Ship two users",
		func(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
			if len(args) < 2 {
				_, _ = s.ChannelMessageSend(m.ChannelID, "I need two names in order to ship :cry:")
				return
			}
			v := 0
			vowel := func(c rune) bool {
				switch c {
				case 'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U':
					return true
				}
				return false
			}
			i := 0
			for x, c := range args[0] {
				if vowel(c) {
					v++
				}
				if v == 2 {
					i = x
					break
				}
			}
			n0 := args[0][0 : i+1]
			v = 0
			f := func(c rune) bool {
				if !vowel(c) {
					v++
				}
				return v < 2
			}
			n1 := strings.TrimLeftFunc(args[1], f)
			_, _ = s.ChannelMessageSend(m.ChannelID, "@"+n0+n1)
		},
	})
	commands = append(commands, Cmd{
		"shitpost",
		"Constructs a random sentence from words that I've collected!",
		func(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
			determiners := getWords("DT")
			if determiners == nil {
				fmt.Println("[DT] something went wrong, you probably know by now")
				return
			}

			adjectives := getWords("JJ")
			if adjectives == nil {
				fmt.Println("[JJ] something went wrong, you probably know by now")
				return
			}
			nouns := getWords("NN")
			if nouns == nil {
				fmt.Println("[NN] something went wrong, you probably know by now")
				return
			}
			nouns = append(nouns, getWords("NNS")...) // who cares about nilchecks anyway
			nouns = append(nouns, getWords("NNP")...)
			nouns = append(nouns, getWords("NNPS")...)
			adverbs := getWords("RB")
			if adverbs == nil {
				fmt.Println("[RB] something went wrong, you probably know by now")
				return
			}
			verbs := getWords("VB")
			if verbs == nil {
				fmt.Println("[VB] something went wrong, you probably know by now")
				return
			}
			verbs = append(verbs, getWords("VBD")...)
			verbs = append(verbs, getWords("VBN")...)
			verbs = append(verbs, getWords("VBP")...)
			_, _ = s.ChannelMessageSend(m.ChannelID,
				determiners[rand.Intn(len(determiners))] + " " + nouns[rand.Intn(len(nouns))] + " " +
					adverbs[rand.Intn(len(adverbs))] + " " + verbs[rand.Intn(len(verbs))] + " " +
					determiners[rand.Intn(len(determiners))] + " " + nouns[rand.Intn(len(nouns))],
			)
		},
	})
	commands = append(commands, Cmd{
		"help",
		"Displays this message",
		func(s *discordgo.Session, m *discordgo.MessageCreate, _ []string) {
			str := ""
			for _, e := range commands {
				if e.id == "" {
					continue
				}
				str += fmt.Sprintf("`%s`\n\t%s\n", e.id, e.desc)
			}
			_, _ = s.ChannelMessageSend(m.ChannelID, str)
		},
	})
}

func main() {
	tagger, err = pos.NewTagger(model_path, tagger_path)
	if err != nil {
		fmt.Println("error initializing tagger,", err)
		return
	}
	db, err = sql.Open("sqlite3", "./words.db")
	if err != nil {
		fmt.Println("error opening database,", err)
		return
	}
	defer db.Close()
	initCommands()
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating session,", err)
		return
	}
	u, err := dg.User("@me")
	if err != nil {
		fmt.Println("error obtaining account details,", err)
	}
	botID = u.ID
	dg.AddHandler(messageCreate)
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	<-make(chan struct{})
	return
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Printf("%20s %20s %20s > %s\n", m.ChannelID, time.Now().Format(time.Stamp), m.Author.Username, m.Content)
	if m.Author.ID == botID {
		return
	}
	if strings.HasPrefix(m.Content, "*") {
		c := strings.Split(strings.TrimPrefix(m.Content, "*"), " ")
		fmt.Printf("%v\n", c)
		for _, e := range commands {
			if e.id == c[0] {
				(e.cb)(s, m, c[1:])
				return
			}
		}
		(commands[0].cb)(s, m, nil)
	} else {
		p := re.ReplaceAllString(m.Content, "")
		var res []*pos.Result
		if p != "" {
			res, err = tagger.Tag(p)
		} else {
			return
		}
		if err != nil {
			fmt.Println("error tagging message content,", err)
			return
		}
		tx, err := db.Begin()
		if err != nil {
			fmt.Println("error beginning db,", err)
			return
		}
		stmt, err := tx.Prepare("insert into words(word, pos) values(?, ?);")
		if err != nil {
			fmt.Println("error preparing statement,", err)
			return
		}
		defer stmt.Close()
		for _, r := range res {
			_, err := stmt.Exec(r.Word, r.TAG)
			if err != nil {
				fmt.Println("error executing stmt,", err)
			}
			fmt.Println(r.Word, r.TAG)
		}
		tx.Commit()
	}
}
