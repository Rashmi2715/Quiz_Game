package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// Define the Quiz struct to represent quiz questions
type Quiz struct {
	Question string   `json:"question"`
	Options  []string `json:"options"`
}

func main() {
	fmt.Println("Welcome to the Quiz!")
	fmt.Println("Choose a quiz:")
	fmt.Println("1. 5-mark quiz")
	fmt.Println("2. 10-mark quiz")
	fmt.Println("3. Exit")

	reader := bufio.NewReader(os.Stdin)

	var choice string
	for {
		fmt.Print("Enter your choice (1, 2, or 3): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			return
		}

		choice = strings.TrimSpace(input)
		if choice == "1" || choice == "2" || choice == "3" {
			break
		}
		fmt.Println("Invalid choice. Please enter 1, 2, or 3.")
	}

	if choice == "3" {
		fmt.Println("Exiting the quiz.")
		return
	}

	numQuestions := 5

	if choice == "2" {
		numQuestions = 10
	}

	fmt.Println("Starting the quiz...")

	var totalScore int // Variable to accumulate the score

	for i := 1; i <= numQuestions; i++ {
		// Fetch current quiz question from the server
		resp, err := http.Get("http://localhost:8080/quiz")
		if err != nil {
			fmt.Println("Error fetching quiz question:", err)
			return
		}
		defer resp.Body.Close()

		// Read response body
		var currentQuiz Quiz
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(&currentQuiz); err != nil {
			fmt.Println("Error decoding JSON:", err)
			return
		}

		// Display quiz question to the user
		fmt.Printf("Question %d:\n", i)
		fmt.Println(currentQuiz.Question)
		fmt.Println("Options:")
		for i, option := range currentQuiz.Options {
			fmt.Printf("%d. %s\n", i+1, option)
		}

		var userAnswer string
		for {
			fmt.Print("Enter your answer (1, 2, or 3): ")
			answer, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading answer:", err)
				return
			}

			userAnswer = strings.TrimSpace(answer)
			if userAnswer == "1" || userAnswer == "2" || userAnswer == "3" {
				break
			}
			fmt.Println("Invalid answer. Please enter 1, 2, or 3.")
		}

		// Trim whitespace and convert user answer to integer
		userAnswer = strings.TrimSpace(userAnswer)
		answerIndex := 0
		fmt.Sscanf(userAnswer, "%d", &answerIndex)

		// Send user's answer index to the server for validation
		resp, err = http.Post("http://localhost:8080/validate?answerIndex="+fmt.Sprint(answerIndex), "text/plain", nil)
		if err != nil {
			fmt.Println("Error sending answer to server:", err)
			return
		}
		defer resp.Body.Close()
		// Read the server's response line by line
		var response struct {
			Result        string `json:"result"`
			CorrectAnswer string `json:"correctAnswer"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			fmt.Println("Error decoding JSON response:", err)
			return
		}

		// Extract result and correct answer from the response struct
		result := response.Result
		correctAnswer := response.CorrectAnswer

		// Display the result to the user
		fmt.Println("Result:", result)
		if result == "Wrong!" {
			fmt.Println("The correct answer is ", correctAnswer)
		}
		fmt.Println()

		// Accumulate the score
		if result == "Correct!" {
			totalScore++
		}

	}

	// Fetch the score from the server
	scoreResp, err := http.Get("http://localhost:8080/score")
	if err != nil {
		fmt.Println("Error fetching score from server:", err)
		return
	}
	defer scoreResp.Body.Close()

	// Read the score from the server response
	var scoreMsg string
	if _, err := fmt.Fscan(scoreResp.Body, &scoreMsg); err != nil {
		fmt.Println("Error reading score from server:", err)
		return
	}

	// Display the final score
	fmt.Println("Quiz ended.")
	fmt.Println("Your Score:", totalScore)

}
