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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

type Recipe struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Tags         []string  `json:"tags"`
	Ingredients  []string  `json:"ingredients"`
	Instructions []string  `json:"instructions"`
	PublishedAt  time.Time `json:"publishedAt"`
}

var recipes []Recipe

func init() {
	recipes = make([]Recipe, 0)
	file, _ := ioutil.ReadFile("recipes.json")
	_ = json.Unmarshal([]byte(file), &recipes)
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
	recipe.ID = xid.New().String()
	recipe.PublishedAt = time.Now()
	recipes = append(recipes, recipe)
	c.JSON(http.StatusOK, recipe)
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
	c.JSON(200, recipes)
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
	index := -1
	for i := 0; i < len(recipes); i++ {
		if recipes[i].ID == id {
			index = i
		}
	}

	if index == -1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Recipe not found",
		})
		return
	}

	recipe.ID = id
	recipe.PublishedAt = time.Now()

	recipes[index] = recipe
	c.JSON(http.StatusOK, recipe)
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

	index := -1
	for i := 0; i < len(recipes); i++ {
		if recipes[i].ID == id {
			index = i
		}
	}

	if index == -1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Recipe not found",
		})
		return
	}

	recipes = append(recipes[:index], recipes[index+1:]...)
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("%s recipe deleted", id),
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
