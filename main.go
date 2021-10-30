//RECIPES API
//
//This is a sample recipes API. You can find out more about
//the API at https://github.com/ajaix734/gin_go_api
//
// Schemes:http
// Host: localhost:3000
// BasePath: /
// Version : 1.0.0
// Contact:
//Consumes:
//  - application/json
//
// Produces:
// - application/json
// swagger:meta
package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Recipe struct {
	ID string `json:"id"`
	// _ID          primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	Name         string    `json:"name"`
	Tags         []string  `json:"tags"`
	Ingredients  []string  `json:"ingredients"`
	Instructions []string  `json:"instructions"`
	PublishedAt  time.Time `json:"publishedAt"`
}

var ctx context.Context
var err error
var client *mongo.Client

var recipes []Recipe

func init() {
	recipes = make([]Recipe, 0)
	file, _ := ioutil.ReadFile("recipes.json")
	_ = json.Unmarshal([]byte(file), &recipes)

	ctx := context.Background()
	clientOpts := options.Client().ApplyURI("mongodb://admin:password@localhost:27017/recipes")
	client, err = mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB")

	var listOfRecipes []interface{}
	for _, recipe := range recipes {
		listOfRecipes = append(listOfRecipes, recipe)
	}
	collection := client.Database("recipes").Collection("recipes")
	estCount, estCountErr := collection.EstimatedDocumentCount(context.TODO())
	if estCountErr != nil {
		panic(estCountErr)
	}
	if estCount == 0 {
		insertManyResult, err := collection.InsertMany(
			ctx, listOfRecipes)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Inserted recipes: ",
			len(insertManyResult.InsertedIDs))
	}
}

// swagger:operation POST /recipes recipes NewRecipeHandler
// Create a recipe
// ---
// consumes:
// - application/json
// produces:
//  - application/json
// responses:
// '200':
//   description: Successful operation
// '400':
//   description: Invalid input
// '404':
//   description: Invalid body
func NewRecipeHandler(c *gin.Context) {
	var recipe Recipe
	err := c.ShouldBindJSON(&recipe)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}
	collection := client.Database("recipes").Collection("recipes")
	recipe.PublishedAt = time.Now()
	_, mongoErr := collection.InsertOne(context.TODO(), recipe)
	if mongoErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": mongoErr.Error(),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Recipe inserted",
	})
}

//swagger:operation GET /recipes recipes ListRecipesHandler
// Returns list of recipes
// ---
// produces:
// - application/json
//response:
// '200':
// description: successful operation
func ListRecipesHandler(c *gin.Context) {
	ctx := context.Background()
	collection := client.Database("recipes").Collection("recipes")
	curr, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
	defer curr.Close(ctx)

	recipes := make([]Recipe, 0)
	for curr.Next(ctx) {
		var recipe Recipe
		curr.Decode(&recipe)
		recipes = append(recipes, recipe)
	}
	c.JSON(http.StatusOK, recipes)
}

// swagger:operation PUT /recipes/{id} recipes UpdateRecipeHandler
// Update an existing recipe
// ---
// parameters:
// - name: id
// in: path
// description: ID of the recipe
// required: true
// type: string
// produces:
// - application/json
// responses:
// '200':
// description: Successful operation
// '400':
// description: Invalid input
// '404':
// description: Invalid recipe ID
func UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe Recipe
	err := c.ShouldBindJSON(&recipe)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	ctx := context.Background()
	collection := client.Database("recipes").Collection("recipes")
	filter := bson.M{"id": id}
	recipe.ID = id
	recipe.PublishedAt = time.Now()
	update := bson.M{
		"$set": recipe,
	}
	upsert := true
	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}
	result := collection.FindOneAndUpdate(ctx, filter, update, &opt)
	if result.Err() != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": result.Err().Error(),
		})
	}

	// 9) Decode the result
	doc := bson.M{}
	_ = result.Decode(&doc)
	c.JSON(http.StatusOK, gin.H{
		"message": doc,
	})
}

// swagger:operation DELETE /recipes/{id} recipes DeleteRecipeHandler
// Delete an existing recipe
// ---
// parameters:
// - name: id
//   in: path
//   description: ID of the recipe
//   required: true
//   type: string
// produces:
//  - application/json
// responses:
// '200':
//   description: Successful operation
// '400':
//   description: Invalid input
// '404':
//   description: Recipe Not found
func DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	collection := client.Database("recipes").Collection("recipes")
	filter := bson.M{
		"id": id,
	}
	result := collection.FindOneAndDelete(context.TODO(), filter)
	if err := result.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "Recipe deleted successfully",
	})

}

// swagger:operation GET /recipes/search recipes SearchRecipeHandler
// Delete an existing recipe
// ---
// parameters:
// - name: tag
//   in: path
//   description: tag keyword to search for
//   required: true
//   type: string
// produces:
//  - application/json
// responses:
// '200':
//   description: Successful operation
// '400':
//   description: Invalid input
// '404':
//   description: Recipes Not found for tag
func SearchRecipeHandler(c *gin.Context) {
	tag := c.Query("tag")
	listOfRecipes := make([]Recipe, 0)

	for i := 0; i < len(recipes); i++ {
		for _, value := range recipes[i].Tags {
			if strings.EqualFold(value, tag) {
				listOfRecipes = append(listOfRecipes, recipes[i])
			}
		}
	}

	c.JSON(http.StatusOK, listOfRecipes)
}

func main() {
	router := gin.Default()
	router.POST("/recipes", NewRecipeHandler)
	router.GET("/recipes", ListRecipesHandler)
	router.PUT("/recipes/:id", UpdateRecipeHandler)
	router.DELETE("/recipes/:id", DeleteRecipeHandler)
	router.GET("/recipes/search", SearchRecipeHandler)
	router.Run(":3000")
}
