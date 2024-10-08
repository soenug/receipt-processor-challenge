package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type Receipt struct {
    Retailer        string      `json:"retailer"`
    PurchaseDate    string      `json:"purchaseDate"`
    PurchaseTime    string      `json:"purchaseTime"`
    Items           []Item     `json:"items"`
    Total           string      `json:"total"`
    Points          int         `json:"points"`
}

type Item struct {
    ShortDescription    string      `json:"shortDescription"`
    Price               string      `json:"price"`
}

var receipts = make(map[string]Receipt)

func GeneratedID() string {
    return fmt.Sprintf("%d", rand.Intn(100000000))
}

func CalculatePoints(receipt Receipt) int {
    points := 0

    alphanumeric := regexp.MustCompile("[a-zA-z0-9]")
    points += len(alphanumeric.FindAllString(receipt.Retailer, -1))

    total, err := strconv.ParseFloat(receipt.Total, 64)
    if err == nil && total == float64(int(total)) {
        points += 50
    }

    if total != 0 && total*100 != 0 && int(total*100)%25 == 0 {
        points += 25
    }

    points += (len(receipt.Items) / 2) * 5

    for _, item := range receipt.Items {
        if len(strings.TrimSpace(item.ShortDescription))%3 == 0 {
            price, err := strconv.ParseFloat(item.Price, 64)
            if err == nil {
                points += int(math.Ceil(price*0.2))
            }
        }
    }

    purchaseDate, err := time.Parse("2006-01-02", receipt.PurchaseDate)
    if err == nil && purchaseDate.Day()%2 != 0 {
        points += 6
    }

    purchaseTime, err := time.Parse("15:04", receipt.PurchaseTime)
    if err == nil && purchaseTime.Hour() >= 14 && purchaseTime.Hour() <= 16 {
        points += 10
    }

    return points
}

func ProcessReceipt(w http.ResponseWriter, r *http.Request) {
    var receipt Receipt

    err := json.NewDecoder(r.Body).Decode(&receipt);
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    receipt.Points = CalculatePoints(receipt);

    id := GeneratedID()
    receipts[id] = receipt

    response := map[string]string{"id": id}
    json.NewEncoder(w).Encode(response)
}

func GetPoints(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    receipt, found := receipts[id]
    if !found {
        http.Error(w, "Receipt not found", http.StatusBadRequest)
        return
    }

    response := map[string]int{"points": receipt.Points}
    json.NewEncoder(w).Encode(response)
}

func main() {
    r := mux.NewRouter()

    r.HandleFunc("/receipts/process", ProcessReceipt).Methods("POST")
    r.HandleFunc("/receipts/{id}/points", GetPoints).Methods("GET")

    fmt.Println("Server starting on port 8080...")
    log.Fatal(http.ListenAndServe(":8080", r))
}