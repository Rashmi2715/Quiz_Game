package main

import (
	"encoding/csv"
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
	Answer   string   `json:"answer"`
}

var quizQuestions []Quiz
var currentQuestionIndex int
var score int = 0 // Variable to track the score

func main() {
	// Read quiz questions from CSV file
	quizQuestions, err := readCSV("quiz_questions.csv")
	if err != nil {
		fmt.Println("Error reading CSV file:", err)
		return
	}

	// Set initial question index and score
	currentQuestionIndex = 0

	// Handle requests to fetch quiz questions
	http.HandleFunc("/quiz", func(w http.ResponseWriter, r *http.Request) {
		// Check if there are more questions
		if currentQuestionIndex >= len(quizQuestions) {
			currentQuestionIndex = 0 // Reset the quiz for next run
		}

		// Convert current quiz question to JSON format
		quizJSON, err := json.Marshal(quizQuestions[currentQuestionIndex])
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Set response headers
		w.Header().Set("Content-Type", "application/json")

		// Write JSON response with quiz question
		w.Write(quizJSON)
	})

	// Handle requests to validate user answers
	http.HandleFunc("/validate", func(w http.ResponseWriter, r *http.Request) {
		// Parse user's answer index from query parameter
		userAnswerIndex := r.URL.Query().Get("answerIndex")

		// Convert answer index to integer
		answerIndex := 0
		fmt.Sscanf(userAnswerIndex, "%d", &answerIndex)

		// Retrieve correct answer from quizQuestions array
		correctAnswer := quizQuestions[currentQuestionIndex].Answer

		// Compare user's answer with correct answer
		var result string
		if quizQuestions[currentQuestionIndex].Options[answerIndex-1] == correctAnswer {
			score++
			result = "Correct!"
		} else {
			result = "Wrong!"

		}
		// Increment current question index for the next question
		currentQuestionIndex++

		// Check if quiz is completed
		if currentQuestionIndex >= len(quizQuestions) {
			currentQuestionIndex = 0 // Reset the quiz for next run
		}

		response := struct {
			Result        string `json:"result"`
			CorrectAnswer string `json:"correctAnswer"`
		}{
			Result:        result,
			CorrectAnswer: correctAnswer,
		}

		// Convert response to JSON format
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Set response headers
		w.Header().Set("Content-Type", "application/json")

		// Write JSON response with result and correct answer
		w.Write(jsonResponse)

		// Flush the response to ensure it is sent before closing the connection
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	})

	http.HandleFunc("/score", func(w http.ResponseWriter, r *http.Request) {
		// Send the score as part of the response
		fmt.Fprintf(w, "Your score: %d out of %d\n", score, len(quizQuestions))
		score = 0 // Reset the score for the next quiz
	})

	// Start the server
	fmt.Println("Server listening on port 8080...")
	http.ListenAndServe(":8080", nil)
}

// Function to read quiz questions from CSV file
func readCSV(filename string) ([]Quiz, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	if _, err := reader.Read(); err != nil {
		return nil, err
	}

	lines, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var quizQuestions []Quiz
	for _, line := range lines {
		quiz := Quiz{
			Question: line[0],
			Options:  splitOptions(line[1]),
			Answer:   line[2],
		}
		quizQuestions = append(quizQuestions, quiz)
	}

	return quizQuestions, nil
}

// Function to split options string into slice of options
func splitOptions(optionsStr string) []string {
	options := strings.Split(optionsStr, ",")
	for i, opt := range options {
		options[i] = strings.TrimSpace(opt)
	}
	return options
}
