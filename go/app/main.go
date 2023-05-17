package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	ImgDir = "images"
)

type Response struct {
	Message string `json:"message"`
}

type Item struct {
	Name          string `json:"name"`
	Category      string `json:"category"`
	ImageFilename string `json:"image_filename"`
}

type ItemList struct {
	Items []Item `json:"items"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	c.Logger().Infof("Receive item name: %s", name)
	category := c.FormValue("category")
	c.Logger().Infof("Receive item category: %s", category)
	imageFileName := c.FormValue("image")
	c.Logger().Infof("Receive item image: %s", imageFileName)

	// sha256 handle image
	h := sha256.New()
	h.Write([]byte(imageFileName))
	hashImageFileName := hex.EncodeToString(h.Sum(nil)) + ".jpg"

	// Create item and add it to the end of items
	item := Item{Name: name, Category: category, ImageFilename: hashImageFileName}
	items, err := readItemsJson(c)
	if err != nil {
		return err
	}
	items = append(items, item)
	itemList := ItemList{Items: items}

	// Write data to items.json
	jsonFile, err := os.Open("jsons/items.json")
	if err != nil {
		res := Response{Message: "error opening json file"}
		return c.JSON(http.StatusBadRequest, res)
	}
	defer jsonFile.Close()
	jsonData, err := json.MarshalIndent(itemList, "", "\t")
	if err != nil {
		res := Response{Message: "error converting json"}
		return c.JSON(http.StatusBadRequest, res)
	}
	err = os.WriteFile("jsons/items.json", jsonData, 0777)
	if err != nil {
		res := Response{Message: "error writing json file"}
		return c.JSON(http.StatusBadRequest, res)
	}

	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

func getAllItems(c echo.Context) error {
	items, err := readItemsJson(c)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, items)
}

func readItemsJson(c echo.Context) ([]Item, error) {
	jsonData, err := os.ReadFile("jsons/items.json")
	if err != nil {
		res := Response{"error reading json file aa"}
		return nil, c.JSON(http.StatusBadRequest, res)
	}
	var items ItemList
	json.Unmarshal(jsonData, &items)

	return items.Items, nil
}

func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Debugf("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.INFO)

	front_url := os.Getenv("FRONT_URL")
	if front_url == "" {
		front_url = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{front_url},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Routes
	e.GET("/", root)
	e.GET("/items", getAllItems)
	e.POST("/items", addItem)
	e.GET("/image/:imageFilename", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
