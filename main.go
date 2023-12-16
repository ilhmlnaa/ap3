package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
)

// SparePart struct represents a spare part with its name and price
type SparePart struct {
	Name  string
	Price float64
}

var spareparts = []SparePart{
	{Name: "Aki", Price: 90.000},
	{Name: "Kampas", Price: 30.000},
	{Name: "Rantai", Price: 80.000},
	{Name: "Ban", Price: 100.000},
	{Name: "Oli", Price: 50.000},
}

var cart = make(map[string]int)

type ViewData struct {
	Spareparts []SparePart
}

func main() {
	r := gin.Default()

	r.Static("/templates", "./templates")
	r.Static("/assets", "./assets")
	

	r.LoadHTMLGlob("templates/*")
	
	// Tambahkan endpoint baru untuk menghapus item dari keranjang
	r.POST("/remove", func(c *gin.Context) {
		removeFromCartHandler(c)
	})

	r.GET("/", func(c *gin.Context) {
		data := ViewData{
			Spareparts: spareparts,
		}
		c.HTML(http.StatusOK, "index.html", data)
	})

	r.GET("/qrcode", func(c *gin.Context) {
		// Retrieve the QR code data
		qrText := c.Query("text")

		// Generate the QR code image file path
		imagePath, err := generateQRCodeImage(qrText)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal Membuat QR Code","endpoint":"/qrcode?text=toqrcode"})
			return
		}

		// Set the content type header to indicate it's an image
		c.Header("Content-Type", "image/png")

		// Serve the QR code image file
		c.File(imagePath)
	})

	r.POST("/add", func(c *gin.Context) {
		addItemToCartHandler(c)
	})

	r.GET("/view", func(c *gin.Context) {
		viewCartHandler(c)
	})

	r.GET("/calculate", func(c *gin.Context) {
		calculateTotalHandler(c)
	})

	r.Run(":8080")
}

func getPartsList() []SparePart {
	return spareparts
}


func addItemToCartHandler(c *gin.Context) {

    // Get the form values
    partName := c.PostForm("part")
    quantity, err := strconv.Atoi(c.PostForm("quantity"))

    if err != nil || quantity < 1 {
        c.HTML(http.StatusBadRequest, "error.html", gin.H{"message": "Invalid quantity."})
        return
    }

    // Check if the part exists
    var selectedPart SparePart
    partExists := false
    for _, part := range spareparts {
        if part.Name == partName {
            selectedPart = part
            partExists = true
            break
        }
    }

    if !partExists {
        c.HTML(http.StatusBadRequest, "error.html", gin.H{"message": "Invalid part name."})
        return
    }

    // Add the item to the cart
    cart[selectedPart.Name] += quantity

    // Check if the request is an AJAX request
    if c.Request.Header.Get("X-Requested-With") == "XMLHttpRequest" {
        // If it's an AJAX request, respond with JSON
        c.JSON(http.StatusOK, gin.H{"success": true, "message": "Barang sukses ditambahkan"})
    } else {
        // If it's a regular form submission, redirect back to the home page
        c.Redirect(http.StatusSeeOther, "/")
    }
}

func viewCartHandler(c *gin.Context) {

	c.HTML(http.StatusOK, "view.html", gin.H{"cart": cart})
}

func removeFromCartHandler(c *gin.Context) {
    // Get the part name and quantity from the request body
    var request map[string]interface{}
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request format."})
        return
    }

    partName, ok := request["part"].(string)
    if !ok {
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid part name."})
        return
    }

    quantity, ok := request["quantity"].(float64)
    if !ok {
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid quantity."})
        return
    }

    // Check if the part exists in the cart
    if _, exists := cart[partName]; !exists {
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Item not found in cart."})
        return
    }

    // Remove the part from the cart or reduce the quantity
    if cart[partName] > int(quantity) {
        cart[partName] -= int(quantity)
    } else {
        delete(cart, partName)
    }

    c.JSON(http.StatusOK, gin.H{"success": true, "message": "Item removed from cart."})
}




func calculateTotalHandler(c *gin.Context) {

	total := 0.0
	for part, quantity := range cart {
		for _, sp := range spareparts {
			if sp.Name == part {
				price := sp.Price
				total += price * float64(quantity)
				break
			}
		}
	}

	intPart := int64(total)
	formattedIntPart := humanize.Comma(intPart)
	decimalPart := total - float64(intPart)
	formattedDecimalPart := fmt.Sprintf("%.3f", decimalPart)[1:]

	qrText := fmt.Sprintf("Payment Amount: Rp.%s%s", formattedIntPart, formattedDecimalPart)
	isEmpty := len(cart) == 0
	c.HTML(http.StatusOK, "payment.html", gin.H{
		"total":                fmt.Sprintf("Rp.%s%s", formattedIntPart, formattedDecimalPart),
		"formattedIntPart":    formattedIntPart,
		"formattedDecimalPart": formattedDecimalPart,
		"qrText":               qrText,
		"isEmpty":              isEmpty,
	})
}

func generateQRCodeImage(data string) (string, error) {
	q, err := qrcode.New(data, qrcode.Low)
	if err != nil {
		return "", err
	}

	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "qrcode*.png")
	if err != nil {
		return "", err
	}
	defer tmpfile.Close()

	// Write the QR code to the temporary file
	if err := q.WriteFile(256, tmpfile.Name()); err != nil {
		return "", err
	}

	return tmpfile.Name(), nil
}
