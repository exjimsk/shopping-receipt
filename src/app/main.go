package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	states = map[string]State{}
	items  = map[string]Item{}
)

type State struct {
	Name       string          `json:"name"`
	TaxRate    float64         `json:"rate"`
	Exemptions map[string]bool `json:"exemptions"`
}

type Item struct {
	Name     string  `json:"name"`
	Category string  `json:"category"`
	Price    float64 `json:"price"`
	Currency string  `json:"currency"`
	Quantity int     `json:"quantity"`
}

type Receipt struct {
	State    string          `json:"state"`
	Cart     []string        `json:"cart"`
	Items    map[string]Item `json:"items"`
	Subtotal float64         `json:"subtotal"`
	Tax      float64         `json:"tax"`
	Total    float64         `json:"total"`
}

func (rcp *Receipt) Compute() error {
	state, ok := states[rcp.State]
	if !ok {
		return fmt.Errorf("Invalid state")
	}

	rate := state.TaxRate

	cartItems := []Item{}

	for _, cartItem := range rcp.Cart {
		item, ok := items[cartItem]
		if !ok {
			return fmt.Errorf("Invalid item")
		}

		cartItems = append(cartItems, item)
	}
	for _, item := range cartItems {
		rcp.Subtotal = rcp.Subtotal + item.Price

		if _, exempt := state.Exemptions[item.Category]; !exempt {
			tax := rate * item.Price
			rcp.Tax = rcp.Tax + tax
		}

		_item, exists := rcp.Items[item.Name]
		if !exists {
			item.Quantity = 1
			rcp.Items[item.Name] = item
		} else {
			_item.Quantity = _item.Quantity + 1
			rcp.Items[item.Name] = _item
		}
	}

	rcp.Subtotal = RoundCent(rcp.Subtotal)
	rcp.Tax = RoundNickel(rcp.Tax)

	rcp.Total = rcp.Subtotal + rcp.Tax

	return nil
}

func RoundCent(x float64) float64 {
	return math.Ceil(x*100) / 100
}

func RoundNickel(x float64) float64 {
	return math.Ceil(x*20) / 20
}

func main() {

	if err := readStates(); err != nil {
		panic(err)
	}

	if err := readItems(); err != nil {
		panic(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadFile("index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(b)
	})

	http.HandleFunc("/states", func(w http.ResponseWriter, r *http.Request) {
		b, err := json.Marshal(states)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	})

	http.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		b, err := json.Marshal(items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	})

	http.HandleFunc("/total", func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			fmt.Println(err)
			return
		}

		rcp := &Receipt{Items: map[string]Item{}}

		if err := json.Unmarshal(b, rcp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := rcp.Compute(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		b, err = json.Marshal(rcp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	})

	port := fmt.Sprintf(":%v", os.Getenv("PORT"))

	fmt.Println("app started on", port)
	http.ListenAndServe(port, nil)
}

func readStates() error {
	file, err := os.Open("_states.csv")
	if err != nil {
		return err
	}

	r := csv.NewReader(file)

	r.Read()

	for {
		row, err := r.Read()
		if err != nil {
			break
		} else if len(row) != 3 {
			return fmt.Errorf("Wrong CSV format")
		}

		taxRate, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			return fmt.Errorf("Invalid tax rate")
		}

		exemptions := strings.Split(row[2], ";")
		if err != nil {
			return fmt.Errorf("Invalid exemptions")
		}
		_exemptions := map[string]bool{}
		for _, e := range exemptions {
			_exemptions[e] = true
		}

		s := State{
			Name:       row[0],
			TaxRate:    taxRate,
			Exemptions: _exemptions,
		}

		states[s.Name] = s
	}

	return nil
}

func readItems() error {
	file, err := os.Open("_items.csv")
	if err != nil {
		return err
	}

	r := csv.NewReader(file)

	r.Read()

	for {
		row, err := r.Read()
		if err != nil {
			break
		} else if len(row) != 4 {
			return fmt.Errorf("Wrong CSV format")
		}

		price, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			return fmt.Errorf("Invalid price")
		}

		i := Item{
			Name:     row[0],
			Category: row[1],
			Price:    price,
			Currency: row[3],
		}

		items[i.Name] = i
	}

	return nil
}
