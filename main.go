package main

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
)

type receipt struct {
	ID           string `json:"id"`
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items        []struct {
		ShortDescription string  `json:"shortDescription"`
		Price            float64 `json:"price"`
	} `json:"items"`
	Total float64 `json:"total"`
}

// getReceipts responds with the list of all receipts as JSON.
func getReceipts(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, receipts)
}

func postReceipts(c *gin.Context) {
	var newReceipt receipt
	var newID = generateID()

	if err := c.BindJSON(&newReceipt); err != nil {
		return
	}

	newReceipt.ID = newID
	receipts = append(receipts, newReceipt)
	c.IndentedJSON(http.StatusCreated, "id: "+newReceipt.ID)
}

func getReceiptByID(c *gin.Context) {
	id := c.Param("id")
	points := 0

	for _, a := range receipts {
		if a.ID == id {
			//1 Point for every alphanumeric character in Retailer name
			for _, char := range a.Retailer {
				if unicode.IsLetter(char) || unicode.IsDigit(char) {
					points++
				}
			}
			//50 Points for a Total with a Round dollar amount
			if isMultiple(a.Total, 1.00) {
				points += 50
			}
			//25 Points for a Total that is a multiple of 0.25
			if isMultiple(a.Total, 0.25) {
				points += 25
			}
			//5 Points for every 2 receipt items
			points += 5 * (len(a.Items) / 2)
			//If the trimmed length of the item description is a multiple of 3,
			//multiply the price by 0.2 and round up to the nearest integer.
			//The result is the number of points earned.
			for _, item := range a.Items {
				//Trim the space from each side of the string
				trimmedDescription := strings.TrimLeft(item.ShortDescription, " ")
				trimmedDescription = strings.TrimRight(trimmedDescription, " ")

				if len(trimmedDescription)%3 == 0 {
					points += int(math.Ceil(item.Price * 0.2))
				}
			}
			//6 Points if the purchase date is odd
			d := string(a.PurchaseDate[8:10])
			if d[0] == '0' {
				d = string(d[1])
			}
			day, err := strconv.Atoi(d)
			if err != nil {
				c.IndentedJSON(gin.ErrorTypeNu, gin.H{"message": "variable day to int failed."})
			} else {
				if day%2 == 1 {
					points += 6
				}
			}
			//10 Points if purchase time is after 2:00pm and before 4:00pm
			//Converting the hours to an int
			h := string(a.PurchaseTime[0:2])
			hour, err := strconv.Atoi(h)
			if err != nil {
				c.IndentedJSON(gin.ErrorTypeNu, gin.H{"message": "variable hour to int failed."})
			} else {
				//Converting the minutes to an int
				m := string(a.PurchaseTime[3:5])
				minute, err := strconv.Atoi(m)
				if err != nil {
					c.IndentedJSON(gin.ErrorTypeNu, gin.H{"message": "variable minute to int failed."})
				} else {
					//Check for correct time
					if hour == 15 || hour == 14 && minute > 0 {
						points += 10
					}
				}
			}
			c.IndentedJSON(http.StatusOK, "points: "+strconv.Itoa(points))
		}
	}
	// c.IndentedJSON(http.StatusNotFound, gin.H{"message": "receipt not found"})
}

func isMultiple(f float64, n float64) bool {
	if n == 0 {
		return false
	}
	remainder := math.Mod(f, n)
	return math.Abs(remainder) < 1e-9
}

// empty list of receipts
var receipts = []receipt{}

func generateID() string {
	return strconv.Itoa(len(receipts))
}

func main() {
	router := gin.Default()
	router.GET("/receipts/process", getReceipts)
	router.GET("/receipts/:id/points", getReceiptByID)
	router.POST("/receipts/process", postReceipts)

	router.Run("localhost:8080")
}
