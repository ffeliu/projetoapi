package main

import (
	"net/http"
	"strings"
	"time"

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
		auth.PUT("/", refreshToken)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.Run()
}

func formatSwagger() {
	//http://localhost:8080/swagger/index.html
	docs.SwaggerInfo.Title = "API de avaliações"
	docs.SwaggerInfo.Description = "Essa api permite manter todas as avaliações realizadas."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
}

// @Summary Recupera as avaliações
// @Description Exibe a lista, sem todos os campos, de todas as avaliações
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @param Authorization header string true "Token"
// @Success 200 {array} model.EvaluationPartialView
// @Router /evaluation [get]
// @Failure 404 "Not found"
func getAllEvaluation(c *gin.Context) {
	var evaluations []model.Evaluation
	var _evaluationPartialView []model.EvaluationPartialView

	if !validateToken(c) {
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
// @Security BearerAuth
// @param Bearer header string true "Token"
// @Param id path int true "Evaluation ID"
// @Success 200 {object} model.EvaluationFullView
// @Router /evaluation/{id} [get]
// @Failure 404 "Not found"
func getEvaluationById(c *gin.Context) {
	var evaluation model.Evaluation
	id := c.Param("id")
	ratingDesc := ""

	if !validateToken(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "message": "Acesso não autorizado"})
		return
	}

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
// @Security BearerAuth
// @param Bearer header string true "Token"
// @Param evaluation body model.EvaluationUpd true "Udpdate evaluation"
// @Param id path int true "Evaluation ID"
// @Router /evaluation/{id} [put]
// @Success 200 {object} model.Evaluation
// @Failure 400 "Bad request"
// @Failure 404 "Not found"
func updateEvaluation(c *gin.Context) {
	var updEvaluation model.EvaluationUpd
	var evaluation model.Evaluation

	if !validateToken(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "message": "Acesso não autorizado"})
		return
	}

	if err := c.ShouldBindJSON(&updEvaluation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Sintaxe de requisição mal formada!"})
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
// @Security BearerAuth
// @param Bearer header string true "Token"
// @Param id path int true "Evaluation ID"
// @Router /evaluation/{id} [delete]
// @Success 200 {object} model.Evaluation
// @Failure 404 "Not found"
func deleteEvaluation(c *gin.Context) {
	var evaluation model.Evaluation

	if !validateToken(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "message": "Acesso não autorizado"})
		return
	}

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
// @Security BearerAuth
// @param Bearer header string true "Token"
// @Param evaluation body model.EvaluationAdd true "Add evaluation"
// @Router /evaluation [post]
// @Success 201 {object} model.Evaluation
// @Failure 400 "Bad request"
// @Failure 404 "Not found"
func addEvaluation(c *gin.Context) {

	var addEvaluation model.EvaluationAdd

	if !validateToken(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "message": "Acesso não autorizado"})
		return
	}

	if err := c.ShouldBindJSON(&addEvaluation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Sintaxe de requisição mal formada!"})
		return
	}

	rating := addEvaluation.Rating
	note := addEvaluation.Note

	evaluation := model.Evaluation{Rating: rating, Note: note}

	db.Save(&evaluation)

	c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "Avaliação criada com sucesso!", "resourceId": evaluation.Id})
}

// @Summary Realizar autenticação
// @Description Autentica o usuário e gera o token para os próximos acessos
// @Accept  json
// @Produce  json
// @Router /auth [post]
// @Param evaluation body model.Users true "Do login"
// @Success 200 {object} model.Users
// @Failure 400 "Bad request"
// @Failure 401 "Unauthorized"
func doAuthentication(c *gin.Context) {
	var creds model.Users
	var usr model.Users

	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Sintaxe de requisição mal formada!"})
		return
	}

	db.Find(&usr, "username = ? and password = ?", creds.Username, creds.Password)
	if usr.Username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "message": "Usuário não encontrado"})
		return
	}

	token := getToken(creds)

	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "message": "Acesso não autorizado"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "message": "Login realizado com sucesso!", "token": token})
}

func getToken(creds model.Users) string {

	// Set expiration time of the token
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

func validateToken(c *gin.Context) bool {

	var token string

	reqToken := c.Request.Header.Get("Authorization")
	if strings.Contains(reqToken, "Bearer") {
		if strings.TrimSpace(reqToken) == "" {
			return false
		}

		splitToken := strings.Split(reqToken, "Bearer")
		token = strings.TrimSpace(splitToken[1])
	} else {
		token = strings.TrimSpace(reqToken)
	}

	claims := &Claims{}

	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return false
		}
	}

	if !tkn.Valid {
		return false
	}

	return true
}

// @Summary Atualiza token de autenticação
// @Description Atualiza o token de autenticação do usuário
// @Accept  json
// @Produce  json
// @Router /auth [put]
// @Param evaluation body model.Users true "Refresh token"
// @Success 204 {object} model.Users
// @Failure 400 "Bad request"
// @Failure 401 "Unauthorized"
func refreshToken(c *gin.Context) {
	var user model.Users
	var usr model.Users

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Sintaxe de requisição mal formada!"})
		return
	}

	db.Find(&usr, "username = ? and password = ?", user.Username, user.Password)
	if usr.Username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "message": "Acesso não autorizado"})
		return
	}

	token := getToken(user)

	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "message": "Acesso não autorizado"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": http.StatusNoContent, "message": "Token atualizado com sucesso!", "token": token})
}
