package requestHandler

import (
	"context"
	"github.com/jinzhu/gorm"
	memcache "github.com/patrickmn/go-cache"
	"github.com/zwirec/TGChatScanner/TGBotApi"
	"github.com/zwirec/TGChatScanner/clarifaiApi"
	"log"
	"net/http"
)

type RequestHandler struct {
	mux *http.ServeMux
}

type AppContext struct {
	Db               *gorm.DB
	DownloadRequests chan *FileBasic
	CfApi            *clarifaiApi.ClarifaiApi
	BotApi           *TGBotApi.BotApi
	Cache            *memcache.Cache
	Logger           *log.Logger
	ImagesPath       string
	Hostname         string
}

var appContext AppContext

type key int

var loggerContextKey key = 0

func NewRequestHandler() *RequestHandler {
	mux := http.NewServeMux()
	return &RequestHandler{mux: mux}
}

func (r *RequestHandler) RegisterHandlers() {
	r.mux.Handle("/api/v1/images", middleware(middlewareLogin(http.HandlerFunc(getImages))))
	r.mux.Handle("/api/v1/chats", middleware(middlewareLogin(http.HandlerFunc(getChats))))
	//r.mux.Handle("/api/v1/chat", middleware(middlewareLogin(http.HandlerFunc(getChat))))
	r.mux.Handle("/api/v1/tags", middleware(middlewareLogin(http.HandlerFunc(getTags))))
	r.mux.Handle("/"+appContext.BotApi.Token, middleware(http.HandlerFunc(BotUpdateHanlder)))
}

func (r *RequestHandler) SetAppContext(context *AppContext) {
	appContext = *context
}

func AddLogger(ctx context.Context, req *http.Request) context.Context {
	return context.WithValue(ctx, loggerContextKey, appContext.Logger)
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctx := AddLogger(req.Context(), req)
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}

func (r *RequestHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
	return
}
