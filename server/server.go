package server

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/PoW-HC/hashcash/pkg/hash"
	"github.com/PoW-HC/hashcash/pkg/pow"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
	"math/rand"
	"powserver/config"
	"powserver/storage"
	"strconv"
	"strings"
	"time"
)

const (
	HttpCodeConnectionClosedByClientBeforeResponse = 499
	TaskHeaderName                                 = "Pow_task"
	HashHeaderName                                 = "Pow_hash"
	MaxIterations                                  = 1 << 20
	SecretPhrase                                   = "secret"
)

var ErrInvalidPowHash = fmt.Errorf("invalid pow hash")

type Storage interface {
	Add(context.Context, string) error
	Get(context.Context, string) (bool, bool, error)
	Mark(context.Context, string) error
}

type Server struct {
	storage Storage
}

func NewServer() *Server {
	return &Server{
		storage: storage.NewTaskStorage(),
	}
}

func (s *Server) Start() error {
	router := fasthttprouter.New()
	router.GET("/hash", s.calcHash)
	router.GET("/random_quote", s.handleRandomQuote)
	err := fasthttp.ListenAndServe(config.HttpServerPort, router.Handler)
	if err != nil {
		return fmt.Errorf("can't start http server: %v", err)
	}

	return nil
}

func (s *Server) calcHash(ctx *fasthttp.RequestCtx) {
	defer ctx.Response.ConnectionClose()
	powTask := string(ctx.Request.Header.Peek(TaskHeaderName))

	hasher, err := hash.NewHasher("sha256")
	if err != nil {
		error500(ctx, err)
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.Response.SetBody([]byte("internal server error"))
		return
	}

	p := pow.New(hasher)

	hashCash, err := pow.InitHashcash(5, powTask, pow.SignExt(SecretPhrase, hasher))
	if err != nil {
		error500(ctx, err)
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.Response.SetBody([]byte("internal server error"))
		return
	}

	solution, err := p.Compute(context.Background(), hashCash, MaxIterations)
	if err != nil {
		error500(ctx, err)
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.Response.SetBody([]byte("internal server error"))
		return
	}

	ctx.Response.SetBody([]byte(solution.String()))
}

func (s *Server) handleRandomQuote(ctx *fasthttp.RequestCtx) {
	defer ctx.Response.ConnectionClose()
	chPowReady := make(chan struct{})

	go s.handlePowDefence(ctx, chPowReady)

	select {
	case <-ctx.Done():
		fmt.Println("SERVER: context cancelled")
		ctx.Response.SetStatusCode(HttpCodeConnectionClosedByClientBeforeResponse)
		ctx.Response.SetBody([]byte("cancelled"))
	case <-time.After(time.Second * 10):
		fmt.Println("SERVER: timeout")
		ctx.TimeoutErrorWithCode("timeout", fasthttp.StatusRequestTimeout)
		ctx.Response.SetBody([]byte("timeout"))
	case <-chPowReady:
	}
}

func (s *Server) handlePowDefence(ctx *fasthttp.RequestCtx, chReady chan struct{}) {
	defer func() {
		chReady <- struct{}{}
	}()
	powTask := string(ctx.Request.Header.Peek(TaskHeaderName))
	powHash := string(ctx.Request.Header.Peek(HashHeaderName))

	if powTask == "" {
		rand.Seed(time.Now().UnixNano())
		powTask = strconv.Itoa(rand.Int())
		ctx.Response.Header.Set(TaskHeaderName, powTask)
		err := s.storage.Add(ctx, powTask)
		if err != nil {
			error500(ctx, err)
		}

		return
	}

	if powHash == "" {
		ctx.Response.Header.Set(TaskHeaderName, powTask)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.Response.SetBody([]byte("add hash header \"Pow_hash\""))

		return
	}

	status, ok, err := s.storage.Get(ctx, powTask)
	if err != nil {
		error500(ctx, err)
		return
	}
	if !ok {
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.Response.SetBody([]byte(fmt.Sprintf("unknown pow task %s. Please, request new task.", powTask)))
		return
	}
	if status {
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.Response.SetBody([]byte(fmt.Sprintf("pow task %s already used. Please, request new task.", powTask)))
		return
	}

	isVerifiedHash, err := verifyPow(powHash, powTask)
	if errors.Is(err, ErrInvalidPowHash) {
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.Response.SetBody([]byte("invalid pow hash"))
		return
	}
	if err != nil {
		error500(ctx, err)
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.Response.SetBody([]byte("internal server error"))
		return
	}

	if !isVerifiedHash {
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.Response.SetBody([]byte("wrong pow hash"))
		return
	}

	quote := GetRandomQuote()
	ctx.Response.SetBody([]byte(fmt.Sprintf("Quote: %s       Author: %s", quote.Phrase, quote.Name)))
	err = s.storage.Mark(ctx, powTask)
	if err != nil {
		error500(ctx, err)
	}
}

func verifyPow(powHash, powTask string) (bool, error) {
	hasher, err := hash.NewHasher("sha256")
	if err != nil {
		return false, err
	}

	p := pow.New(hasher)
	elems := strings.Split(powHash, ":")

	if len(elems) != 7 {
		fmt.Println(errors.New(fmt.Sprintf("SERVER: wrong powHash. %d elements instead 7", len(elems))))
		return false, ErrInvalidPowHash
	}

	version, err := strconv.Atoi(elems[0])
	if err != nil {
		fmt.Println("SERVER:", err)
		return false, ErrInvalidPowHash
	}

	bits, err := strconv.Atoi(elems[1])
	if err != nil {
		fmt.Println("SERVER:", err)
		return false, ErrInvalidPowHash
	}

	i, err := strconv.ParseInt(elems[2], 10, 64)
	if err != nil {
		fmt.Println("SERVER:", err)
		return false, ErrInvalidPowHash
	}

	date := time.Unix(i, 0)
	resource := elems[3]
	ext := elems[4]

	rawDecodedText, err := base64.StdEncoding.DecodeString(elems[5])
	if err != nil {
		fmt.Println("SERVER:", err)
		return false, ErrInvalidPowHash
	}
	random := rawDecodedText

	rawDecodedText, err = base64.StdEncoding.DecodeString(elems[6])
	if err != nil {
		fmt.Println("SERVER:", err)
		return false, ErrInvalidPowHash
	}
	counter, err := strconv.ParseInt(string(rawDecodedText), 16, 64)
	if err != nil {
		fmt.Println("SERVER:", err)
		return false, ErrInvalidPowHash
	}
	hashcash := pow.NewHashcach(int32(version), int32(bits), date, resource, ext, random, counter)

	err = p.Verify(hashcash, powTask)
	if errors.Is(err, pow.ErrWrongChallenge) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func error500(ctx *fasthttp.RequestCtx, err error) {
	ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
	ctx.Response.SetBody([]byte("internal server error"))
	fmt.Println("SERVER:", err)
}
