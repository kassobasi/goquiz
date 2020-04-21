package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

type quizItem struct {
	question string
	answer   string
}

func main() {
	f := flag.String("f", "program.csv", "name of the CSV file that contains questions and answers")
	t := flag.Int("t", 30, "timer in seconds")
	shuffle := flag.Bool("s", false, "shuffle questions")

	flag.Parse()

	items, err := getQuizItems(f, shuffle)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	c := context.Background()
	tc, _ := context.WithTimeout(c, time.Duration(*t)*time.Second)

	score := askQuestions(tc, items)
	fmt.Fprintf(os.Stdout, "Your score is %d/%d\n", score, len(items))
}

func askQuestions(c context.Context, items []quizItem) int {
	score := 0
	r := make(chan string)

LOOP:
	for _, item := range items {
		fmt.Fprintf(os.Stdout, "%v? ", item.question)
		go func() {
			var response string
			_, err := fmt.Fscanf(os.Stdin, "%s", &response)
			if err != nil {
				response = ""
			}
			r <- response
		}()
		select {
		case <-c.Done():
			break LOOP
		case response := <-r:
			if response == item.answer {
				score++
			}
		}
	}

	return score
}

func getQuizItems(name *string, shuffle *bool) ([]quizItem, error) {
	f, err := os.Open(*name)
	if err != nil {
		return nil, fmt.Errorf("can't open file: %v %v", *name, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	lines, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("%v is not a valid CSV file: %v", *name, err)
	}

	var items []quizItem
	for i, l := range lines {
		if len(l) != 2 {
			return nil, fmt.Errorf("not a valid question-answer at line %d: %v", i, l)
		}
		items = append(items, quizItem{question: l[0], answer: strings.TrimSpace(l[1])})
	}

	if *shuffle {
		rand.Seed(time.Now().Unix())
		rand.Shuffle(len(items), func(i, j int) {
			items[i], items[j] = items[j], items[i]
		})
	}
	return items, nil
}
