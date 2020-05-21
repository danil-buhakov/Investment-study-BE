package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/urfave/negroni"
	"log"
	"net/http"
	"strconv"
)

type (
	API struct {
		router   *mux.Router
		server   *http.Server
		contract *Contract
		accounts *AccountsContainer
	}

	// Route stores an API route data
	Route struct {
		Path   string
		Method string
		Func   func(http.ResponseWriter, *http.Request)
	}
)

func (api *API) Init(contract *Contract, acc *AccountsContainer) {
	api.contract = contract
	api.accounts = acc
	api.router = mux.NewRouter()
	wrapper := negroni.New()
	wrapper.Use(cors.New(cors.Options{
		AllowCredentials: true,
		AllowedMethods:   []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "User-Env"},
	}))

	handleActions(api.router, wrapper, "", []*Route{
		{Path: "/", Method: http.MethodGet, Func: api.Index},
		{Path: "/login", Method: http.MethodPost, Func: api.Login},
		{Path: "/create", Method: http.MethodPost, Func: api.Create},
		{Path: "/balance", Method: http.MethodGet, Func: api.Balance},
		{Path: "/investment/list", Method: http.MethodGet, Func: api.TokenBalances},
		{Path: "/investment/make", Method: http.MethodGet, Func: api.Invest},
	})

	api.server = &http.Server{Addr: fmt.Sprintf(":%d", 8000), Handler: api.router}

	api.server.ListenAndServe()
}

func handleActions(router *mux.Router, wrapper *negroni.Negroni, prefix string, routes []*Route) {
	for _, r := range routes {
		w := wrapper.With()

		w.Use(negroni.Wrap(http.HandlerFunc(r.Func)))
		router.Handle(prefix+r.Path, w).Methods(r.Method, "OPTIONS")
	}
}

func (api *API) Index(w http.ResponseWriter, r *http.Request) {
	Json(w, map[string]interface{}{
		"status": true,
	})
}

func (api *API) Login(w http.ResponseWriter, r *http.Request) {

	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()
	body := map[string]string{}

	err := dec.Decode(&body)
	if err != nil {
		log.Print("Error: ", err)
		return
	}

	address, err := api.accounts.Address(body["login"])
	if err != nil {
		log.Print("Error: ", err)
		return
	}

	Json(w, map[string]interface{}{
		"address": address,
	})
}

func (api *API) Create(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()
	body := map[string]string{}

	err := dec.Decode(&body)
	if err != nil {
		log.Print("Error: ", err)
		return
	}

	address, err := api.accounts.Add(body["login"])
	if err != nil {
		log.Print("Error: ", err)
		return
	}

	Json(w, map[string]interface{}{
		"address": address,
	})
}

func (api *API) Balance(w http.ResponseWriter, r *http.Request) {
	addressHex := r.URL.Query().Get("address")

	bal, err := api.contract.GetBalanceForAddress(addressHex)
	if err != nil {

	}

	Json(w, map[string]interface{}{
		"balance": bal.Int64(),
	})
}

func (api *API) TokenBalances(w http.ResponseWriter, r *http.Request) {
	login := r.URL.Query().Get("login")

	type resp struct {
		Name    string `json:"name"`
		Balance uint64 `json:"balance"`
	}

	var respArr []resp

	for i := range SupportedTokens {

		bal, err := api.contract.GetTokenBalanceForAddress(SupportedTokens[i].Address, login)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		respArr = append(respArr, resp{Name: SupportedTokens[i].Name, Balance: bal.Uint64()})
	}

	Json(w, respArr)
}

func (api *API) Invest(w http.ResponseWriter, r *http.Request) {
	amountStr := r.URL.Query().Get("amount")
	token := r.URL.Query().Get("token")
	login := r.URL.Query().Get("login")

	var tokenAddressHex string
	for key := range SupportedTokens {
		if SupportedTokens[key].Name == token {
			tokenAddressHex = SupportedTokens[key].Address
		}
	}

	amount, _ := strconv.ParseInt(amountStr, 10, 64)

	err := api.contract.InvestOnToken(login, amount, tokenAddressHex)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	Json(w, map[string]interface{}{
		"status": "success",
	})
}

func Json(w http.ResponseWriter, data interface{}) {
	js, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
