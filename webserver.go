package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/davecgh/go-spew/spew"
)

func run() error {
	gin := makeGinRouter()
	httpAddr := os.Getenv("ADDR")
	log.Println("Listening on ", os.Getenv("ADDR"))
	s := &http.Server{
		Addr:           ":" + httpAddr,
		Handler:        gin,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	if err := s.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func makeGinRouter() *gin.Engine {
	ginRouter := gin.New()
	ginRouter.Use(gin.Logger())
	ginRouter.GET("/", getBlockchain)
	ginRouter.POST("/", writeBlockChain)
	return ginRouter
}
func getBlockchain(c *gin.Context) {
	bytes, err := json.MarshalIndent(Blockchain, "", " ")
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(c.Writer, string(bytes))
}

func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}

func writeBlockChain(c *gin.Context) {
	var m Message

	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithJSON(c.Writer, c.Request, http.StatusBadRequest, c.Request.Body)
		return
	}
	defer c.Request.Body.Close()

	newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], m.Grade, m.EduInstitution)
	if err != nil {
		respondWithJSON(c.Writer, c.Request, http.StatusInternalServerError, m)
		return
	}

	if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
		newBlockchain := append(Blockchain, newBlock)
		replaceChain(newBlockchain)
		spew.Dump(Blockchain)
	}

	respondWithJSON(c.Writer, c.Request, http.StatusCreated, newBlock)
}