package main

import (
	"net/http"
	"time"

	//"time"

	"projetoapi/docs"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"projetoapi/model"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var db *gorm.DB

// Create the JWT key used to create the signature
var jwtKey = []byte("my_secret_key")

var users = map[string]string{
	"user1": "password1",
	"user2": "password2",
}

// Create a struct that will be encoded to a JWT.
// We add jwt.StandardClaims as an embedded type, to provide fields like expiry time
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func init() {
	// http://gorm.io/docs/connecting_to_the_database.html
	//open a db connection
	var err error
	db, err = gorm.Open("postgres", "host=localhost port=5432 user=postgres dbname=EvaluationDB password=1020304050 sslmode=disable")

	if err != nil {
		panic("failed to connect database")
	}
}

func main() {

	formatSwagger()

	// Creates a gin router with default middleware:
	// logger and recovery (crash-free) middleware
	router := gin.Default()

	evalution := router.Group("/api/v1/evaluation")
	{
		evalution.POST("/", addEvaluation)
		evalution.GET("/", getAllEvaluation)
		evalution.GET("/:id", getEvaluationById)
		evalution.PUT("/:id", updateEvaluation)
		evalution.DELETE("/:id", deleteEvaluation)
	}

	auth := router.Group("api/v1/auth")
	{
		auth.POST("/", doAuthentication)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.Run()
}

func formatSwagger() {
	// programatically set swagger info
	docs.SwaggerInfo.Title = "API de avaliações"
	docs.SwaggerInfo.Description = "Essa api permite manter todas as avaliações realizadas."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8081"
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
}

// @Summary Recupera as avaliações
// @Description Exibe a lista, sem todos os campos, de todas as avaliações
// @Accept  json
// @Produce  json
// @Success 200 {array} model.EvaluationPartialView
// @Router /evaluation [get]
// @Failure 404 "Not found"
func getAllEvaluation(c *gin.Context) {
	var evaluations []model.Evaluation
	var _evaluationPartialView []model.EvaluationPartialView

	token := c.Request.Header.Get("Authorization")

	if !validateToken(token) {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "message": "Acesso não autorizado"})
		return
	}

	db.Find(&evaluations)

	if len(evaluations) <= 0 {
		c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "Nao foram encontradas avaliacoes!"})
		return
	}

	for _, item := range evaluations {

		ratingDesc := ""

		switch item.Rating {
		case 1:
			ratingDesc = "Péssima"
		case 2:
			ratingDesc = "Ruim"
		case 3:
			ratingDesc = "Aceitável"
		case 4:
			ratingDesc = "Boa"
		case 5:
			ratingDesc = "Ótima"
		}

		_evaluationPartialView = append(_evaluationPartialView,
			model.EvaluationPartialView{Id: item.Id, Rating: ratingDesc})

	}

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": _evaluationPartialView})
}

// @Summary Recupera uma avaliação pelo id
// @Description Exibe os detalhes de uma avaliação pelo ID
// @ID get-string-by-int
// @Accept  json
// @Produce  json
// @Param id path int true "Evaluation ID"
// @Success 200 {object} model.EvaluationFullView
// @Router /evaluation/{id} [get]
// @Failure 404 "Not found"
func getEvaluationById(c *gin.Context) {
	var evaluation model.Evaluation
	id := c.Param("id")
	ratingDesc := ""

	db.First(&evaluation, id)
	if evaluation.Id == 0 {
		c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "Avaliação não encontrada!"})
		return
	}

	switch evaluation.Rating {
	case 1:
		ratingDesc = "Péssima"
	case 2:
		ratingDesc = "Ruim"
	case 3:
		ratingDesc = "Aceitável"
	case 4:
		ratingDesc = "Boa"
	case 5:
		ratingDesc = "Ótima"
	}

	_todo := model.EvaluationFullView{Id: evaluation.Id, Rating: ratingDesc, Note: evaluation.Note}
	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": _todo})
}

// @Summary Atualiza uma avaliação
// @Description Atualiza uma avaliação sobre a utilização da aplicação
// @Accept  json
// @Produce  json
// @Param evaluation body model.EvaluationUpd true "Udpdate evaluation"
// @Param id path int true "Evaluation ID"
// @Router /evaluation/{id} [put]
// @Success 200 {object} model.Evaluation
// @Failure 400 "Bad request"
// @Failure 404 "Not found"
func updateEvaluation(c *gin.Context) {
	var updEvaluation model.EvaluationUpd
	var evaluation model.Evaluation

	if err := c.ShouldBindJSON(&updEvaluation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Bad request!"})
		return
	}

	id := c.Param("id")

	db.First(&evaluation, id)

	if evaluation.Id == 0 {
		c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "Nao encontrou avaliação!"})
		return
	}

	evaluation.Rating = updEvaluation.Rating
	evaluation.Note = updEvaluation.Note

	db.Save(&evaluation)

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "message": "Avaliação atualizada com sucesso!"})
}

// @Summary Exclui uma avaliação pelo ID
// @Description Exclui uma avaliação realizada
// @ID get-string-by-int
// @Accept  json
// @Produce  json
// @Param id path int true "Evaluation ID"
// @Router /evaluation/{id} [delete]
// @Success 200 {object} model.Evaluation
// @Failure 404 "Not found"
func deleteEvaluation(c *gin.Context) {
	var evaluation model.Evaluation

	id := c.Param("id")

	db.First(&evaluation, id)

	if evaluation.Id == 0 {
		c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "Avaliação não encontrada!"})
		return
	}

	db.Delete(&evaluation)

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "message": "Avaliação excluida com sucesso!"})
}

// @Summary Adicionar uma avaliação
// @Description Cria uma avaliação sobre a utilização da aplicação
// @Accept  json
// @Produce  json
// @Param evaluation body model.EvaluationAdd true "Add evaluation"
// @Router /evaluation [post]
// @Success 201 {object} model.Evaluation
// @Failure 400 "Bad request"
// @Failure 404 "Not found"
func addEvaluation(c *gin.Context) {

	var addEvaluation model.EvaluationAdd

	if err := c.ShouldBindJSON(&addEvaluation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Bad request!"})
		return
	}

	rating := addEvaluation.Rating
	note := addEvaluation.Note

	evaluation := model.Evaluation{Rating: rating, Note: note}

	db.Save(&evaluation)

	c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "Avaliação criada com sucesso!", "resourceId": evaluation.Id})
}

func doAuthentication(c *gin.Context) {
	var creds model.Credentials

	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Bad request!"})
		return
	}

	// Get the expected password from our in memory map
	expectedPassword, ok := users[creds.Username]

	// If a password exists for the given user
	// AND, if it is the same as the password we received, the we can move ahead
	// if NOT, then we return an "Unauthorized" status
	if !ok || expectedPassword != creds.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "message": "Acesso não autorizado"})
		return
	}

	token := getToken(creds)

	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "message": "Acesso não autorizado"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "message": "Login realizadom com sucesso!", "token": token})
}

func getToken(creds model.Credentials) string {

	// Declare the expiration time of the token
	// here, we have kept it as 5 minutes
	expirationTime := time.Now().Add(5 * time.Minute)

	// Create the JWT claims, which includes the username and expiry time
	claims := &Claims{
		Username: creds.Username,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Create the JWT string
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return ""
	}

	return tokenString
}

func validateToken(token string) bool {
	return false
}
