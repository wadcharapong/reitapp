package services

import (
	"github.com/wadcharapong/reitapp/app"
	"github.com/wadcharapong/reitapp/models"
	"github.com/wadcharapong/reitapp/config"
	"encoding/json"
	"fmt"
	"github.com/night-codes/mgo-ai"
	"github.com/olivere/elastic"
	"golang.org/x/net/context"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type ReitServicer interface {
	GetReitBySymbol(symbol string) (models.ReitItem, error)
	GetReitAll() ([]models.ReitItem, error)
	SaveReitFavorite(userId string, symbol string) error
	DeleteReitFavorite(userId string, ticker string) error
	GetReitFavoriteByUserIDJoin(userId string) []*models.FavoriteInfo
	GetUserProfileByCriteria(userId string, site string ) models.UserProfile
	SaveUserProfile(profile *models.UserProfile) string
	CreateNewUserProfile(facebook models.Facebook,google models.Google ) string
	SearchElastic(query string) []models.ReitItem
	SearchMap(lat float64 ,lon float64) models.PlaceInfo
}

type Reit_Service struct {
	reitItems []models.ReitItem
	reitItem models.ReitItem
	reitFavorite []*models.FavoriteInfo
	userProfile models.UserProfile
	locationInfo models.PlaceInfo
	err error
}

func (self Reit_Service) GetReitAll() ([]models.ReitItem, error) {
	session := *app.GetDocumentMongo()
	defer session.Close()
	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	document := session.DB(config.Mongo_DB).C("REIT")
	//self.err = document.Find(nil).All(&self.reitItems)
	query := []bson.M{{
		"$lookup": bson.M{ // lookup the documents table here
			"from":         "MajorShareholders",
			"localField":   "symbol",
			"foreignField": "symbol",
			"as":           "majorShareholders",
		}},
		{
			"$lookup": bson.M{ // lookup the documents table here
				"from":         "Place",
				"localField":   "symbol",
				"foreignField": "symbol",
				"as":           "place",
			}}}

	pipe := document.Pipe(query)
	self.err = pipe.All(&self.reitItems)
	if self.err != nil {
		// TODO: Do something about the error
		fmt.Printf("error : ", self.err)
	} else {

	}
	return self.reitItems, self.err
}


func (self Reit_Service) GetReitBySymbol(symbol string) (models.ReitItem, error) {
	session := *app.GetDocumentMongo()
	defer session.Close()
	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	document := session.DB(config.Mongo_DB).C("REIT")
	//self.err = document.Find(bson.M{"symbol": symbol}).One(&self.reitItem)
	query := []bson.M{{
		"$lookup": bson.M{ // lookup the documents table here
			"from":         "MajorShareholders",
			"localField":   "symbol",
			"foreignField": "symbol",
			"as":           "majorShareholders",
		}},
		{
			"$lookup": bson.M{ // lookup the documents table here
				"from":         "Place",
				"localField":   "symbol",
				"foreignField": "symbol",
				"as":           "place",
			}},
		{"$match": bson.M{
			"symbol": symbol,
		}}}

	pipe := document.Pipe(query)
	self.err = pipe.One(&self.reitItem)
	return self.reitItem, self.err
}



func (self Reit_Service) SaveReitFavorite(userId string, symbol string) error {
	fmt.Println("start : GetReitAll")
	session := *app.GetDocumentMongo()
	defer session.Close()
	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	ai.Connect(session.DB(config.Mongo_DB).C("counters"))
	document := session.DB(config.Mongo_DB).C("Favorite")
	favorite := models.Favorite{ID:ai.Next("Favorite"),UserId: userId, Symbol: symbol}
	self.err = document.Insert(&favorite)
	if self.err != nil {
		// TODO: Do something about the error
		fmt.Printf("error : ", self.err)
	} else {

	}
	return self.err
}

func (self Reit_Service) DeleteReitFavorite(userId string, ticker string) error{
	fmt.Println("start : GetReitAll")
	session := *app.GetDocumentMongo()
	defer session.Close()
	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	document := session.DB(config.Mongo_DB).C("Favorite")
	//favorite := models.Favorite{UserId: userId, Symbol: ticker}
	self.err = document.Remove(bson.M{"symbol":ticker,"userId":userId})
	if self.err != nil {
		// TODO: Do something about the error
		fmt.Printf("error : ", self.err)
	} else {

	}
	return self.err
}

func (self Reit_Service) GetReitFavoriteByUserIDJoin(userId string) []*models.FavoriteInfo {
	fmt.Println("start : GetReitAll")
	session := *app.GetDocumentMongo()
	defer session.Close()
	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	document := session.DB(config.Mongo_DB).C("Favorite")

	query := []bson.M{{
		"$lookup": bson.M{ // lookup the documents table here
			"from":         "REIT",
			"localField":   "symbol",
			"foreignField": "symbol",
			"as":           "Reit",
		}},
		{"$match": bson.M{
			"userId": userId,
		}}}

	pipe := document.Pipe(query)
	err := pipe.All(&self.reitFavorite)
	if err != nil {
		// TODO: Do something about the error
		fmt.Printf("error : ", err)
	} else {

	}
	return self.reitFavorite
}

func (self Reit_Service) CreateNewUserProfile(facebook models.Facebook,google models.Google ) string {
	var reitServicer ReitServicer
	reitServicer = Reit_Service{}
	var message string
	if (facebook != models.Facebook{}){
		userProfile := reitServicer.GetUserProfileByCriteria(facebook.ID, "facebook")
		if(userProfile == models.UserProfile{}){
			userProfile =  models.UserProfile{
				UserID: facebook.ID,
				UserName: facebook.Name,
				FullName:facebook.Name,
				Email:facebook.Email,
				Image:facebook.Picture.Data.URL,
				Site:"facebook"}
			message = reitServicer.SaveUserProfile(&userProfile)
		}
	} else if (google != models.Google{}){
		userProfile := reitServicer.GetUserProfileByCriteria(google.ID, "google")
		if(userProfile == models.UserProfile{}){
			userProfile =  models.UserProfile{
				UserID: google.ID,
				UserName: google.Name,
				FullName:google.Name,
				Image:google.Picture,
				Email:google.Email,
				Site:"google"}
			message = reitServicer.SaveUserProfile(&userProfile)
		}
	}
	return message
}

func (self Reit_Service) SaveUserProfile(profile *models.UserProfile) string {
	session := *app.GetDocumentMongo()
	defer session.Close()
	// Optional. Switch the session to a monotonic behavior.
	ai.Connect(session.DB(config.Mongo_DB).C("counters"))
	session.SetMode(mgo.Monotonic, true)
	document := session.DB(config.Mongo_DB).C("UserProfile")
	profile.ID = ai.Next("userProfile")
	err := document.Insert(&profile)
	if err != nil {
		return "fail"
	} else {
		return "success"
	}
}

func (self Reit_Service) GetUserProfileByCriteria(userId string, site string ) models.UserProfile {
	session := *app.GetDocumentMongo()
	defer session.Close()
	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	document := session.DB(config.Mongo_DB).C("UserProfile")
	err := document.Find(bson.M{"userID": userId,"site": site}).One(&self.userProfile)
	if err != nil {
		// TODO: Do something about the error
		fmt.Printf("error : ", err)
	} else {

	}
	return self.userProfile
}

func (self Reit_Service) SearchElastic(query string) []models.ReitItem {
	ctx := context.Background()
	client := app.GetElasticSearch()

	// Search with a term query
	termQuery := elastic.NewMultiMatchQuery(query,"nickName","symbol","reitManager").Type("phrase_prefix")
	//termQuery := elastic.NewTermQuery("nickName",query)
	searchResult, err := client.Search().
		Index(config.ElasticIndexName).   // search in index "reitapp"
		Query(termQuery).   // specify the query
		//Sort("user", true). // sort by "user" field, ascending
		From(0).Size(10).   // take documents 0-9
		Pretty(true).       // pretty print request and response JSON
		Do(ctx)             // execute
	if err != nil {
		// Handle error
		panic(err)
	}

	// TotalHits is another convenience function that works even when something goes wrong.
	fmt.Printf("Found a total of %d reits\n", searchResult.TotalHits())

	// Here's how you iterate through results with full control over each step.
	if searchResult.Hits.TotalHits > 0 {
		// Iterate through results
		for _, hit := range searchResult.Hits.Hits {
			// hit.Index contains the name of the index

			// Deserialize hit.Source into a Tweet (could also be just a map[string]interface{}).
			var t models.ReitItem
			err := json.Unmarshal(*hit.Source, &t)
			if err != nil {
				// Deserialization failed
			}
			self.reitItems = append(self.reitItems,t)
			// Work with tweet
			fmt.Printf("reit by %s: %s\n", t.Symbol, t.NickName)

		}
		return self.reitItems
	} else {
		// No hits
		fmt.Print("Found no reit\n")
	}

	return  nil
}

func (self Reit_Service) SearchMap(lat float64 ,lon float64) models.PlaceInfo {
	fmt.Println("start : SearchMap")
	session := *app.GetDocumentMongo()
	defer session.Close()
	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	document := session.DB(config.Mongo_DB).C("Place")
	scope := 100
	query := []bson.M{{
		"$geoNear": bson.M{
			"near": bson.M{
				"type":        "Point",
				"coordinates": []float64{lon,lat},
			},
			"distanceField": "dist.calculated",
			"maxDistance": scope, // miles to meter
			"spherical": "true",
		}},
		{"$lookup": bson.M{ // lookup the documents table here
			"from":         "REIT",
			"localField":   "symbol",
			"foreignField": "symbol",
			"as":           "Reit",
		}}}

	pipe := document.Pipe(query)
	err := pipe.One(&self.locationInfo)
	if err != nil {
		// TODO: Do something about the error
		fmt.Printf("error : ", err)
	} else {

	}
	return self.locationInfo
}

func AddDataElastic(reit models.ReitItem) error {
	ctx := context.Background()
	client := app.GetElasticSearch()

	CheckIndex()
	//Search with a term query// Index a tweet (using JSON serialization)
	_, err := client.Index().
		Index(config.ElasticIndexName).
		Type("reit").
		//Id("1").
		BodyJson(&reit).
		Do(ctx)
	if err != nil {
		// Handle error
		panic(err)
	}

	return nil
}

func CheckIndex(){
	ctx := context.Background()
	client := app.GetElasticSearch()
	exists, err := client.IndexExists(config.ElasticIndexName).Do(ctx)
	if err != nil {
		// Handle error
		panic(err)
	}
	if !exists {
		// Create a new index.
		createIndex, err := client.CreateIndex(config.ElasticIndexName).BodyString(models.Mapping).Do(ctx)
		if err != nil {
			// Handle error
			panic(err)
		}
		if !createIndex.Acknowledged {
			// Not acknowledged
		}
	}
}