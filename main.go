package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"

	"github.com/gorilla/mux"
)

// types
type Survey struct {
	Id        bson.ObjectID `json:"id" bson:"_id"`
	Token     string        `json:"token" bson:"token"`
	CreatedAt time.Time     `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time     `json:"updated_at" bson:"updated_at"`
	Title     string        `json:"title" bson:"title"`
	Questions []Question    `json:"questions,omitempty" bson:"questions"`
}

// extra: for displaying a list of surveys as a entry point to lookup existing survey on frontend
type SurveysList struct {
	Token string `json:"token" bson:"token"`
	Title string `json:"title" bson:"title"`
}

type Question struct {
	Id            bson.ObjectID `json:"id" bson:"_id"`
	QuestionTitle string        `json:"question_title" bson:"question_title"`
	QuestionType  string        `json:"question_type" bson:"question_type"`
	Answers       []string      `json:"answers,omitempty" bson:"answers"`
}

type Response struct {
	Id           bson.ObjectID `json:"id" bson:"_id"`
	UserId       bson.ObjectID `json:"user_id" bson:"user_id"`
	CreatedAt    time.Time     `json:"created_at" bson:"created_at"`
	SurveyId     bson.ObjectID `json:"survey_id" bson:"survey_id"`
	QuestionId   bson.ObjectID `json:"question_id" bson:"question_id"`
	ResponseText string        `json:"response_text" bson:"response_text"`
}

type ResponseInput struct {
	QuestionId   bson.ObjectID `json:"question_id" bson:"question_id"`
	ResponseText string        `json:"response_text" bson:"response_text"`
}

// global variable
var client *mongo.Client
var surveysCollection *mongo.Collection
var responsesCollection *mongo.Collection

// initial database
func initDB() {
	err := godotenv.Load() // load .env
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	uri := os.Getenv("MONGODB_URI") // get env from .env
	docs := "www.mongodb.com/docs/drivers/go/current/"
	if uri == "" { // check if mongodb uri is missing...
		log.Fatal("Set your 'MONGODB_URI' environment variable. " +
			"See: " + docs +
			"usage-examples/#environment-variable")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().
		ApplyURI(uri))
	_ = client.Ping(ctx, readpref.Primary())

	if err != nil {
		panic(err)
	}

	db := client.Database("OSP_backend")

	surveysCollection = db.Collection("surveys")
	responsesCollection = db.Collection("responses")

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "token", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err = surveysCollection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		log.Fatal(err)
	}

}

// generate token
func genToken() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 5)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func isSurveyIdExist(w http.ResponseWriter, id bson.ObjectID) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var test Survey
	err := surveysCollection.FindOne(ctx, bson.M{"_id": id}).Decode(test)

	if err == mongo.ErrNoDocuments {
		http.Error(w, "the survey does not exist, please provide correct survey id", http.StatusBadRequest)
		return false
	}

	return true
}

func validateQuestionTypes(w http.ResponseWriter, t string, a []string) bool {
	switch t {
	case "Multiple Choice":
		if len(a) < 2 {
			http.Error(w, "Failed to create survey, MC Question should have more than 1 answer", http.StatusBadRequest)
			return false
		}
	case "Likert Scale":
		if len(a) < 3 {
			http.Error(w, "Failed to create survey, Likert Scale Question should have more than 2 answers", http.StatusBadRequest)
			return false
		}
	}
	return true
}

// extra: get all existing surveys token for displaying a list of surveys
func getAllSurveysList(w http.ResponseWriter, r *http.Request) {
	fmt.Println("get surveys list")
	page := r.URL.Query().Get("page")
	limit := r.URL.Query().Get("limit")
	p, err := strconv.ParseInt(page, 10, 64)
	l, err := strconv.ParseInt(limit, 10, 64)
	if p < 1 || l < 1 || p > l {
		//return all surveys if you enter unexpected string or number
		p = 0
		l = 0
	}
	skip := p*l - l
	fOpt := options.Find().SetSkip(skip).SetLimit(l)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := surveysCollection.Find(ctx, bson.D{{}}, fOpt)
	if err != nil {
		panic(err)
	}
	var surveysList []SurveysList
	if err = cursor.All(ctx, &surveysList); err != nil {
		log.Panic(err)
	}
	defer cursor.Close(ctx)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(surveysList)
}

// create survey
func createSurvey(w http.ResponseWriter, r *http.Request) {
	fmt.Println("create survey")
	var survey Survey
	var validate bool = true
	_ = json.NewDecoder(r.Body).Decode(&survey)
	if survey.Title == "" {
		http.Error(w, "Title is required, please make sure the title field is filled", http.StatusBadRequest)
		return
	}
	survey.Id = bson.NewObjectID()
	survey.Token = genToken()
	survey.CreatedAt = time.Now()
	survey.UpdatedAt = survey.CreatedAt

	for i := range survey.Questions {
		if !validateQuestionTypes(w, survey.Questions[i].QuestionType, survey.Questions[i].Answers) {
			validate = false
		}
		survey.Questions[i].Id = bson.NewObjectID()
	}

	if !validate {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := surveysCollection.InsertOne(ctx, survey)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(survey)
}

// update survey by id
func updateSurvey(w http.ResponseWriter, r *http.Request) {
	fmt.Println("edit survey")
	queries := mux.Vars(r)
	id, err := bson.ObjectIDFromHex(queries["survey_id"])
	if err != nil {
		http.Error(w, "Invalid Survey Id", http.StatusBadRequest)
		return
	}
	if !isSurveyIdExist(w, id) {
		return
	}

	var input Survey
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		panic(err)
	}

	updatedSurvey := bson.M{}

	if input.Title != "" {
		updatedSurvey["title"] = input.Title
	}

	if len(input.Questions) > 0 {
		for i := range input.Questions {
			if input.Questions[i].QuestionTitle == "" || input.Questions[i].QuestionType == "" {
				http.Error(w, "Invalid Question without title or type", http.StatusBadRequest)
				return
			}
			if input.Questions[i].Id.IsZero() {
				input.Questions[i].Id = bson.NewObjectID()
			}
		}
		updatedSurvey["questions"] = input.Questions
	}

	if len(updatedSurvey) == 0 {
		http.Error(w, "No updates", http.StatusBadRequest)
		return
	}

	updatedSurvey["updated_at"] = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := surveysCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": updatedSurvey})
	if err != nil {
		if res.MatchedCount == 0 {
			http.Error(w, "No survey found", http.StatusNotFound)
			return
		}
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "survey updated"})
}

// delete survey by id
func deleteSurvey(w http.ResponseWriter, r *http.Request) {
	fmt.Println("delete survey")
	queries := mux.Vars(r)
	id, err := bson.ObjectIDFromHex(queries["survey_id"])
	if err != nil {
		http.Error(w, "Invalid Survey Id", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := surveysCollection.DeleteOne(ctx, bson.M{"_id": id})

	if res.DeletedCount == 0 {
		http.Error(w, "Failed to delete survey, survey might have already removed", http.StatusInternalServerError)
		return
	}
	if err != nil {
		panic(err)
	}

	_, err = responsesCollection.DeleteMany(ctx, bson.M{"survey_id": id})
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "survey deleted"})
}

// get survey by token
func getSurveyByToken(w http.ResponseWriter, r *http.Request) {
	fmt.Println("get survey by token")
	queries := mux.Vars(r)
	token := queries["token"]

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var survey Survey
	err := surveysCollection.FindOne(ctx, bson.M{"token": token}).Decode(&survey)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("No survey found")
		} else {
			panic(err)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(survey)
}

// submit response
func submitResponse(w http.ResponseWriter, r *http.Request) {
	fmt.Println("submit response")
	queries := mux.Vars(r)
	id, err := bson.ObjectIDFromHex(queries["survey_id"])
	if err != nil {
		http.Error(w, "Invalid Survey Id", http.StatusBadRequest)
		return
	}

	if !isSurveyIdExist(w, id) {
		return
	}
	var responseInputs []ResponseInput
	err = json.NewDecoder(r.Body).Decode(&responseInputs)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userId := bson.NewObjectID()

	for _, input := range responseInputs {
		if input.QuestionId.IsZero() || input.ResponseText == "" {
			http.Error(w, "Invalid input from submission", http.StatusBadRequest)
			return
		}
		var response Response
		response.Id = bson.NewObjectID()
		response.UserId = userId
		response.CreatedAt = time.Now()
		response.SurveyId = id
		response.QuestionId = input.QuestionId
		response.ResponseText = input.ResponseText

		_, err := responsesCollection.InsertOne(ctx, response)
		if err != nil {
			http.Error(w, "Failed to submit response", http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responseInputs)
}

// get responses by survey id
func getResponses(w http.ResponseWriter, r *http.Request) {
	fmt.Println("get all responses")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := responsesCollection.Find(ctx, bson.D{{}})
	if err != nil {
		panic(err)
	}
	defer cursor.Close(ctx)
	var responsesList []Response
	if err = cursor.All(ctx, &responsesList); err != nil {
		log.Panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responsesList)
}

// get responses by survey id
func getResponsesById(w http.ResponseWriter, r *http.Request) {
	fmt.Println("get responses by survey id")

	queries := mux.Vars(r)
	id, err := bson.ObjectIDFromHex(queries["survey_id"])
	if err != nil {
		http.Error(w, "Invalid Survey Id", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := responsesCollection.Find(ctx, bson.D{{"survey_id", id}})
	if err != nil {
		panic(err)
	}
	defer cursor.Close(ctx)
	var responsesList []Response
	if err = cursor.All(ctx, &responsesList); err != nil {
		log.Panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responsesList)
}

func main() {
	initDB()
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	r := mux.NewRouter()
	r.HandleFunc("/surveys", getAllSurveysList).Methods("GET")              //list out all created survey by page, default 10 item in 1 page
	r.HandleFunc("/surveys", createSurvey).Methods("POST")                  //create survey
	r.HandleFunc("/surveys/{survey_id}", updateSurvey).Methods("PUT")       //update survey
	r.HandleFunc("/surveys/{survey_id}", deleteSurvey).Methods("DELETE")    //delete survey
	r.HandleFunc("/surveys/token/{token}", getSurveyByToken).Methods("GET") //get survey by token
	r.HandleFunc("/responses/{survey_id}", submitResponse).Methods("POST")  //submit response with survey id
	r.HandleFunc("/responses", getResponses).Methods("GET")                 //get all responses
	r.HandleFunc("/responses/{survey_id}", getResponsesById).Methods("GET") //get response by survey id

	fmt.Println("Server is running on http://localhost:5050")
	log.Fatal(http.ListenAndServe(":5050", r))
}
