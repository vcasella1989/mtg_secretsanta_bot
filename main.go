package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	Token        = "your token"
	participants = []*discordgo.User{}
)

type Participant struct {
	Name          string
	Username      string
	ID            string
	PreviousSanta string
}

func main() {

	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("Error creating Discord session:", err)
		return
	}

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening connection:", err)
		return
	}

	fmt.Println("Bot is running. Press CTRL+C to exit.")
	select {}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	switch m.Content {
	case "!joinsecretsanta":
		for _, user := range participants {
			if user.ID == m.Author.ID {
				s.ChannelMessageSend(m.ChannelID, "You're already participating!")
				return
			}
		}
		participants = append(participants, m.Author)
		s.ChannelMessageSend(m.ChannelID, m.Author.Username+" has been added to Secret Santa!")
		WriteToFile("participants.txt", m.Author.Username+","+m.Author.ID+"\n")

	case "!startsecretsanta":
		//Get current participants
		var homies []Participant

		file, err := os.Open("participants.txt")
		if err != nil {
			log.Fatalf("failed to open file: %s", err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {

			line := scanner.Text()
			parts := strings.Split(string(line), ",")
			fmt.Println(parts)
			participant := Participant{
				Name:          parts[0],
				ID:            parts[1],
				Username:      parts[2],
				PreviousSanta: parts[3],
			}
			homies = append(homies, participant)
		}

		assignments, err := assignSecretSanta(homies)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		for p, santa := range assignments {
			SendMDM(s, p, santa)
		}

	case "!hohoho":
		s.ChannelMessageSend(m.ChannelID, "HO HO HO to you "+m.Author.Username+"!")
	case "!privatehohoho":
		private_channel, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			fmt.Println("Failed to create DM for "+m.Author.Username+":", err)
			return
		}
		s.ChannelMessageSend(private_channel.ID, "HO HO HO to you "+m.Author.Username+"!")
	}

}

func SendMDM(s *discordgo.Session, p Participant, santa Participant) {
	privateChannel, err := s.UserChannelCreate(p.ID)
	if err != nil {
		fmt.Println("Failed to create DM for "+p.Username+":", err)
		return
	}
	fmt.Println("sending message to " + p.Username)

	s.ChannelMessageSend(privateChannel.ID, fmt.Sprintf("HO HO HO %s! You are the Secret Santa for %s!", p.Name, santa.Name))
}

func WriteToFile(filename string, data string) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}
	defer file.Close()

	// Append content to the file
	_, err = file.WriteString(data + "\n")
	if err != nil {
		log.Fatalf("failed writing to file: %s", err)
	}

	fmt.Println("Appended successfully!")
}

func assignSecretSanta(people []Participant) (map[Participant]Participant, error) {
	rand.Seed(time.Now().UnixNano())

	candidates := make([]Participant, len(people))
	copy(candidates, people)

	assignments := make(map[Participant]Participant)

	for _, p := range people {
		possible := make([]Participant, 0)

		for _, c := range candidates {
			if c.ID != p.ID && c.Name != p.PreviousSanta {
				possible = append(possible, c)
			}
		}

		if len(possible) == 0 {
			return nil, errors.New("couldn't find a suitable Secret Santa for everyone")
		}

		chosenIndex := rand.Intn(len(possible))
		chosen := possible[chosenIndex]

		assignments[p] = chosen
		candidates = remove(candidates, chosen)
	}

	return assignments, nil
}

func remove(people []Participant, person Participant) []Participant {
	for i, p := range people {
		if p.ID == person.ID {
			return append(people[:i], people[i+1:]...)
		}
	}
	return people
}
