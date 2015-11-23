package main

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "encoding/json"
    "bytes"
    "github.com/jmoiron/jsonq"
    "gopkg.in/mgo.v2/bson"
    "gopkg.in/mgo.v2"
    "github.com/julienschmidt/httprouter"
    "github.com/r-medina/go-uber"
    "errors"
    "strconv"
    "strings"
    "sort"
    
)

var nextid string
var startid string
var Locids []string
type dataSlice []Data
type eta struct {
	Eta             int         `json:"eta"`
	RequestID       string      `json:"request_id"`
	Status          string      `json:"status"`
	SurgeMultiplier float64         `json:"surge_multiplier"`
}
type Tdata struct {
Id bson.ObjectId `json:"id" bson:"_id"`
Status string  `json:"status" bson:"status"`
Starting_from_location_id string `json:"starting_from_location_id" bson:"starting_from_location_id"`
 Best_route_location_ids []string `json:"best_route_location_ids" bson:"best_route_location_ids"`
  Total_uber_costs int `json:"total_uber_cost" bson:"total_uber_cost"`
  Total_uber_duration int `json:"total_uber_duration" bson:"total_uber_duration"`
  Total_distance float64 `json:"total_distance" bson:"total_distance"` 
}

type Tstatus struct {
Id bson.ObjectId `json:"id" bson:"_id"`
Status string  `json:"status" bson:"status"`
Starting_from_location_id string `json:"starting_from_location_id" bson:"starting_from_location_id"`
 Next_destination_location_id string `json:"next_destination_location_id" bson:"next_destination_location_id"`
 Best_route_location_ids []string `json:"best_route_location_ids" bson:"best_route_location_ids"`
  Total_uber_costs int `json:"total_uber_cost" bson:"total_uber_cost"`
  Total_uber_duration int `json:"total_uber_duration" bson:"total_uber_duration"`
  Total_distance float64 `json:"total_distance" bson:"total_distance"` 
  Uber_wait_time_eta int `json:"uber_wait_time_eta" bson:"uber_wait_time_eta"`
}
type Data struct{
id string
price int
duration int
distance float64
}
type coordinate struct {
    lat float64
    lng float64
}
type request struct {
    LocationIds            []string `json:"location_ids"`
    StartingFromLocationID string   `json:"starting_from_location_id"`
}
type Udata struct {
    Id bson.ObjectId `json:"id" bson:"_id"`
    Name string `json:"name" bson:"name"`
    Address string `json:"address" bson:"address"`
    City string `json:"city" bson:"city"`
    State string `json:"state" bson:"state"`
    Zip string `json:"zip" bson:"zip"`
    Coordinate struct {
        Lat float64 `json:"lat" bson:"lat"`
        Lng float64 `json:"lng" bson:"lng"`
    } `json:"coordinate" bson:"coordinate"`
}

//Create New Location 
func create(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    var u Udata
    URL := "http://maps.google.com/maps/api/geocode/json?address="
    //transfer the data into the local object
    json.NewDecoder(req.Body).Decode(&u)

    //Randomly generated unique ID
    u.Id = bson.NewObjectId()

    URL = URL +u.Address+ " " + u.City + " " + u.State + " " + u.Zip+"&sensor=false"
    URL = strings.Replace(URL, " ", "+", -1)
    fmt.Println("URL "+ URL)

    //call to Google map API
    response, err := http.Get(URL)
    if err != nil {
        return
    }
    defer response.Body.Close()

    resp := make(map[string]interface{})
    body, _ := ioutil.ReadAll(response.Body)
    err = json.Unmarshal(body, &resp)
    if err != nil {
        return
    }

    jq := jsonq.NewQuery(resp)
    status, err := jq.String("status")
    fmt.Println(status)
    if err != nil {
        return
    }
    if status != "OK" {
        err = errors.New(status)
        return
    }

    lat, err := jq.Float("results" ,"0","geometry", "location", "lat")
   if err != nil {
       fmt.Println(err)
        return
    }
    lng, err := jq.Float("results", "0","geometry", "location", "lng")
    if err != nil {
        fmt.Println(err)
        return
    }

    u.Coordinate.Lat = lat
    u.Coordinate.Lng = lng

    //store data into Mongo Lab
    newSession().DB("cmpe273").C("test").Insert(u)

    // interface to JSON struct
    reply, _ := json.Marshal(u)

    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(201)
    fmt.Fprintf(rw, "%s", reply)

}

//Get a Location        
func get(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    uniqueid :=  p.ByName("uniqueid")

    if !bson.IsObjectIdHex(uniqueid) {
        rw.WriteHeader(404)
        return
    }

    dataid := bson.ObjectIdHex(uniqueid)

    responseObj := Udata{}

    if err := newSession().DB("cmpe273").C("test").FindId(dataid).One(&responseObj); err != nil {
        rw.WriteHeader(404)
        return
    }

    reply, _ := json.Marshal(responseObj)

    
    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(200)
    fmt.Fprintf(rw, "%s", reply)
}

//Update Location 
func update(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    var u Udata
    uniqueid :=  p.ByName("uniqueid")

    URL := "http://maps.google.com/maps/api/geocode/json?address="

    //transfer the data into the local object
    json.NewDecoder(req.Body).Decode(&u)

    URL = URL +u.Address+ " " + u.City + " " + u.State + " " + u.Zip+"&sensor=false"
    URL = strings.Replace(URL, " ", "+", -1)
    fmt.Println("URL "+ URL)

    //Google map API
    response, err := http.Get(URL)
    if err != nil {
        return
    }
    defer response.Body.Close()

    resp := make(map[string]interface{})
    body, _ := ioutil.ReadAll(response.Body)
    err = json.Unmarshal(body, &resp)
    if err != nil {
        return
    }

    jq := jsonq.NewQuery(resp)
    status, err := jq.String("status")
    fmt.Println(status)
    if err != nil {
        return
    }
    if status != "OK" {
        err = errors.New(status)
        return
    }

    lat, err := jq.Float("results" ,"0","geometry", "location", "lat")
    if err != nil {
        fmt.Println(err)
        return
    }
    lng, err := jq.Float("results", "0","geometry", "location", "lng")
    if err != nil {
        fmt.Println(err)
        return
    }

    u.Coordinate.Lat = lat
    u.Coordinate.Lng = lng

    dataid := bson.ObjectIdHex(uniqueid)
    var data = Udata{
        Address: u.Address,
        City: u.City,
        State: u.State,
        Zip: u.Zip,
    }
    //updatedata
    fmt.Println(data)
    //store data
    newSession().DB("cmpe273").C("test").Update(bson.M{"_id":dataid }, bson.M{"$set": bson.M{ "address": u.Address,
        "city": u.City, "state": u.State,"zip": u.Zip, "coordinate.lat":u.Coordinate.Lat, "coordinate.lng":u.Coordinate.Lng}})

    responseObj := Udata{}

    //retrive the response data
    if err := newSession().DB("cmpe273").C("test").FindId(dataid).One(&responseObj); err != nil {
        rw.WriteHeader(404)
        return
    }
    // interface into JSON struct
    reply, _ := json.Marshal(responseObj)

    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(201)
    fmt.Fprintf(rw, "%s", reply)

}

//Delete a Location
func delete(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    uniqueid :=  p.ByName("uniqueid")

    if !bson.IsObjectIdHex(uniqueid) {
        rw.WriteHeader(404)
        return
    }

    dataid := bson.ObjectIdHex(uniqueid)

    // delete user
    if err := newSession().DB("cmpe273").C("test").RemoveId(dataid); err != nil {
        rw.WriteHeader(404)
        return
    }

    rw.WriteHeader(200)
}

//creating new MongoDB session 
func newSession() *mgo.Session {
    //Connecting to Mongo Lab
    s, err := mgo.Dial("mongodb://dbuser:dbpswd@ds041934.mongolab.com:41934/cmpe273")

    // Check if mongo server running
    if err != nil {
        panic(err)
    }
    return s
}


//function to get coordinates of objectid
func getdetails(x string) (y coordinate) {
    responseObj := Udata{}
//dataid := bson.ObjectId(x)
    if err := newSession().DB("cmpe273").C("test").Find(bson.M{"_id": bson.ObjectIdHex(x)}).One(&responseObj); err != nil {
        z := coordinate{}
    return z
}
    p := coordinate{
    lat: responseObj.Coordinate.Lat,
    lng: responseObj.Coordinate.Lng,
    }
    return p
    
}
//function to call /estimate/price endpoint
func getprice(x string, z string)(y Data, p int){
response, err := http.Get(x)
    if (err != nil || response.StatusCode==422){
        da := Data{}
        return da,1
    }else{    
   defer response.Body.Close()
    var price []int
    resp := make(map[string]interface{})
    body, _ := ioutil.ReadAll(response.Body)
    err = json.Unmarshal(body, &resp)
    if err != nil {
    da := Data{}
        return da,1
    }
    ptr := resp["prices"].([]interface{})
    jq := jsonq.NewQuery(resp)
     for i, _ := range ptr {
     pr,_ := jq.Int("prices",fmt.Sprintf("%d", i),"low_estimate")
     price = append(price,pr)
	}
     min := price[0]
     for j, _ := range price {
     if(price[j]<=min && price[j]!=0){
     min = price[j]
     }
     }
     du,_:=jq.Int("prices","0","duration")
     dist,_:=jq.Float("prices","0","distance")
     data := Data{
     id:z,
     price:min,
     duration:du,
     distance:dist,
     }
    return data,0    
}
}
//function to get price from last location to start point
func getpricetostart(x string)(y Data){
var price []int
response, err := http.Get(x)
    if err != nil {
        return
    }
    defer response.Body.Close()
    resp := make(map[string]interface{})
    body, _ := ioutil.ReadAll(response.Body)
    err = json.Unmarshal(body, &resp)
    if err != nil {
        return
    }
    ptr := resp["prices"].([]interface{})
    jq := jsonq.NewQuery(resp)
     for i, _ := range ptr {
     pr,_ := jq.Int("prices",fmt.Sprintf("%d", i),"low_estimate")
     price = append(price,pr)
	}
     min := price[0]
     for j, _ := range price {
     if(price[j]<=min && price[j]!=0){
     min = price[j]
     }
     }
     du,_:=jq.Int("prices","0","duration")
     dist,_:=jq.Float("prices","0","distance")
     d := Data{
     id:"",
     price : min,
     duration:du,
     distance:dist,
}
return d
}
// Len is part of sort.Interface.
func (d dataSlice) Len() int {
	return len(d)
}

// Swap is part of sort.Interface.
func (d dataSlice) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

// Less is part of sort.Interface. We use count as the value to sort by
func (d dataSlice) Less(i, j int) bool {
	return d[i].price < d[j].price 
}
//function to find least price among the given set
func sortdata(x map[string]Data)(y Data) {
	m := x
	s := make(dataSlice, 0, len(m))
	for _, d := range m {
		s = append(s, d)
	}		
	sort.Sort(s)
	return s[0]
}

//delete specified id from list
func deleteid(s []string, p string)(x []string) {
    var r []string
    for _, str := range s {
        if str != p {
            r = append(r, str)
        }
    }
    return r
}

//return sum of float64 array
func Sumfloat(a []float64) (sum float64) {
    for _, v := range a {
        sum += v
    }
    return
}
//return sum of int array
func Sumint(a []int) (sum int) {
    for _, v := range a {
        sum += v
    }
    return
}

//Shortroute function accepting http request
func shortroute (rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
   var n int
    decoder := json.NewDecoder(req.Body)
    var t request
    err := decoder.Decode(&t)
    if err != nil {
        panic(err)
    }
    Start := t.StartingFromLocationID
    LocIds := t.LocationIds
    var T Tdata
    var z coordinate
    var tp []int
    var td []float64
    var tdu []int
   for arraylength:=len(LocIds); arraylength>0; arraylength--{
   if(n!=1){
    z = getdetails(Start)
    start_lat := z.lat
    start_lng := z.lng
    x := []coordinate{}
    for i := 0; i < len(LocIds); i++ {
       y := getdetails(LocIds[i])
       x = append(x,y)
   }
   tdata := map[string]Data{}
      for i:=0;i<len(x);i++{
      url := fmt.Sprintf("https://sandbox-api.uber.com/v1/estimates/price?start_latitude=%f&start_longitude=%f&end_latitude=%f&end_longitude=%f&server_token=tyC8DGEaUgBO68yd4rE9RIKF4PyweeZq0uH-bz9-",start_lat,start_lng,x[i].lat,x[i].lng)
      d,e :=getprice(url, LocIds[i])
      if(e!=1){
      tdata[LocIds[i]] = d
      }else{
      //error handling if if location address is invalid or location is more than 100 miles in distance 
      n=1
      }
      }
   da:=sortdata(tdata)
  T.Best_route_location_ids = append(T.Best_route_location_ids,da.id)
   tp = append(tp,da.price)
   td = append(td,da.distance)
   tdu = append(tdu,da.duration)
   LocIds=deleteid(LocIds,da.id)
   Start=da.id
   }else{
   rw.Header().Set("Content-Type", "text/plain")
    rw.WriteHeader(http.StatusBadRequest)
    reply:="Distance to one of location from starting point is more than 100 miles or you have entered address is Invalid, check and try again!!!"
    fmt.Fprintf(rw, "%s", reply)
    return
    }}
    if(n!=1){
   if(LocIds==nil){
   z = getdetails(Start)
    start_lat := z.lat
    start_lng := z.lng
    x := coordinate{}
    y := getdetails(t.StartingFromLocationID)
    x.lat=y.lat
    x.lng=y.lng
       tdata := map[string]Data{}
      url := fmt.Sprintf("https://sandbox-api.uber.com/v1/estimates/price?start_latitude=%f&start_longitude=%f&end_latitude=%f&end_longitude=%f&server_token=tyC8DGEaUgBO68yd4rE9RIKF4PyweeZq0uH-bz9-",start_lat,start_lng,x.lat,x.lng)
      d:=getpricetostart(url)
      tdata[Start] = d
   tp = append(tp,d.price)
   td = append(td,d.distance)
   tdu = append(tdu,d.duration)
   }
T.Id = bson.NewObjectId()
T.Status = "Planning"
T.Starting_from_location_id= t.StartingFromLocationID
 T.Best_route_location_ids = T.Best_route_location_ids
T.Total_uber_costs = Sumint(tp)
T.Total_uber_duration = Sumint(tdu)
T.Total_distance = Sumfloat(td) 
newSession().DB("cmpe273").C("test1").Insert(T)
    // interface to JSON struct
    reply, _ := json.Marshal(T)
    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(201)
    fmt.Fprintf(rw, "%s", reply)
} }
//function to retrive trip details by tripId
func gettrip(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    tripid :=  p.ByName("tripid")

    if !bson.IsObjectIdHex(tripid) {
        rw.WriteHeader(404)
        return
    }

    dataid := bson.ObjectIdHex(tripid)

    responseObj := Tdata{}

    if err := newSession().DB("cmpe273").C("test1").FindId(dataid).One(&responseObj); err != nil {
        rw.WriteHeader(404)
        return
    }

    reply, _ := json.Marshal(responseObj)

    
    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(200)
    fmt.Fprintf(rw, "%s", reply)
}


//function which return uber_estimated_time_of_arrival
func geteta(x float64,y float64,z string)(p int){
lat := strconv.FormatFloat(x, 'E', -1, 64)
lng := strconv.FormatFloat(y, 'E', -1, 64)
 url := "https://sandbox-api.uber.com/v1/requests"
    var jsonStr = []byte(`{
"start_latitude":"`+lat+`",
"start_longitude":"`+lng+`",
"product_id":"`+z+`",
}`)
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
    req.Header.Set("Authorization", "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzY29wZXMiOlsicHJvZmlsZSIsInJlcXVlc3RfcmVjZWlwdCIsInJlcXVlc3QiLCJoaXN0b3J5X2xpdGUiXSwic3ViIjoiMzVjYzM1NzctMjZjZC00YjVjLWFiYzEtODZiOGVlNWZlZTQ3IiwiaXNzIjoidWJlci11czEiLCJqdGkiOiIxZjZmOGIzZS1hMWJhLTQ1NDgtOTNkYS1jNDYxNWY0YWNhYTgiLCJleHAiOjE0NTA0Mzg5MDQsImlhdCI6MTQ0Nzg0NjkwMywidWFjdCI6IkxQSXB4d3V5eUE1UW9YNzRtWlBRWHVkZ01pUTZUeiIsIm5iZiI6MTQ0Nzg0NjgxMywiYXVkIjoiOFlaRWNXc21zX0d3QU1zTlFtOHhyMF94aWFOazhpa3UifQ.BNrFhtQxxFddp542eyaSxDD4peLzdbUaDs7fDeeRixAjyhGtvcyUmDZuNN4lAuYU9ETqbmsUx6AcRat0Pc9ZczuISPDlDxMS9bzJgFIlHFotaMOflkUDJS-ffaB35VzH-j1y2EXyFFQvYNesX5BOQVuwieLXS7sjef1Efz36UL6_MX36_Lq4p0QmO2HtDgo7YHXFo2z4n4DnaHIgIEMFrm0T9nK4D6Zlf0BySf5CPu5AfuOpNj46MY6ZFh3WlqLJFCdWgX7Wyd5U4rh9zJyrwopcwFfP3C0QddcxR-cuxDQuYaHX-OHDcWsXyf2NSmhEo_tw1caAt_xRfK3xhaTOPw")
     req.Header.Set("Content-Type", "application/json")
var resp1 eta
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    body, _ := ioutil.ReadAll(resp.Body)
    err = json.Unmarshal(body,&resp1)
if err != nil {
  panic(err)
}

rid:= resp1.Eta
return rid

}
//fuction to handle PUT /trips/tripid/request
func triptrack(rw http.ResponseWriter, req *http.Request, p httprouter.Params){
client := uber.NewClient("tyC8DGEaUgBO68yd4rE9RIKF4PyweeZq0uH-bz9-")
tripid :=  p.ByName("tripid")
    if !bson.IsObjectIdHex(tripid) {
        rw.WriteHeader(404)
        return
    }

    dataid := bson.ObjectIdHex(tripid)

    responseObj := Tdata{}

    if err := newSession().DB("cmpe273").C("test1").FindId(dataid).One(&responseObj); err != nil {
        rw.WriteHeader(404)
        return
    }
    if(nextid==""){
    startid =responseObj.Starting_from_location_id
     Locids =responseObj. Best_route_location_ids
    z := getdetails(responseObj.Starting_from_location_id)
    start_lat := z.lat
    start_lng := z.lng
products,_ := client.GetProducts(start_lat,start_lng)
productid := products[0].ProductID
eta:=geteta(start_lat,start_lng,productid)
nextid = Locids[0]
    reply := Tstatus{
    Id:responseObj.Id,
    Starting_from_location_id :startid, 
    Best_route_location_ids:responseObj. Best_route_location_ids,
    Total_uber_costs:responseObj.Total_uber_costs,
    Total_uber_duration:responseObj.Total_uber_duration,
    Total_distance:responseObj.Total_distance,
    Uber_wait_time_eta: eta,
     Status : "Requesting",
     Next_destination_location_id: nextid,
  }
  newSession().DB("cmpe273").C("test1").Update(bson.M{"_id":dataid }, bson.M{"$set": bson.M{ "status": "Requesting"}})
  startid = nextid
  Locids=deleteid(Locids,nextid)
  if(Locids!=nil){
  nextid = Locids[0]
  }else{
  nextid = "empty"
  }
res, _ := json.Marshal(reply)
    rw.Header().Set("Content-Type", "application/json")
    //rw.WriteHeader(201)
    fmt.Fprintf(rw, "%s", res)
    }else if(Locids!=nil){
    if(nextid!="empty"){
    z := getdetails(startid)
    start_lat := z.lat
    start_lng := z.lng
products,_ := client.GetProducts(start_lat,start_lng)
productid := products[0].ProductID
eta:=geteta(start_lat,start_lng,productid)
    reply := Tstatus{
    Id:responseObj.Id,
    Starting_from_location_id :startid, 
    Best_route_location_ids:responseObj. Best_route_location_ids,
    Total_uber_costs:responseObj.Total_uber_costs,
    Total_uber_duration:responseObj.Total_uber_duration,
    Total_distance:responseObj.Total_distance,
    Uber_wait_time_eta: eta,
     Status : "Requesting",
     Next_destination_location_id: nextid,
     }
     newSession().DB("cmpe273").C("test1").Update(bson.M{"_id":dataid }, bson.M{"$set": bson.M{ "status": "Requesting"}})
     startid = nextid
  Locids=deleteid(Locids,nextid)
  if(Locids!=nil){
  nextid = Locids[0]
  }else{
  nextid = "empty"
  }
  res, _ := json.Marshal(reply)
    rw.Header().Set("Content-Type", "application/json")
    //rw.WriteHeader(201)
    fmt.Fprintf(rw, "%s", res)
    }
    }else if(nextid=="empty"){
    z := getdetails(startid)
    start_lat := z.lat
    start_lng := z.lng
products,_ := client.GetProducts(start_lat,start_lng)
productid := products[0].ProductID
eta:=geteta(start_lat,start_lng,productid)
    reply := Tstatus{
    Id:responseObj.Id,
    Starting_from_location_id :startid, 
    Best_route_location_ids:responseObj. Best_route_location_ids,
    Total_uber_costs:responseObj.Total_uber_costs,
    Total_uber_duration:responseObj.Total_uber_duration,
    Total_distance:responseObj.Total_distance,
    Uber_wait_time_eta: eta,
     Status : "Requesting",
     Next_destination_location_id: responseObj.Starting_from_location_id,
     }
     newSession().DB("cmpe273").C("test1").Update(bson.M{"_id":dataid }, bson.M{"$set": bson.M{ "status": "Requesting"}})
     nextid="complete"
  res, _ := json.Marshal(reply)
    rw.Header().Set("Content-Type", "application/json")
    //rw.WriteHeader(201)
    fmt.Fprintf(rw, "%s", res)
    }else{
    reply := Tstatus{
    Id:responseObj.Id,
    Starting_from_location_id :responseObj.Starting_from_location_id, 
    Best_route_location_ids:responseObj. Best_route_location_ids,
    Total_uber_costs:responseObj.Total_uber_costs,
    Total_uber_duration:responseObj.Total_uber_duration,
    Total_distance:responseObj.Total_distance,
    Uber_wait_time_eta: 0 ,
     Status : "Finished",
     Next_destination_location_id: "",
     }
     newSession().DB("cmpe273").C("test1").Update(bson.M{"_id":dataid }, bson.M{"$set": bson.M{ "status": "Finished"}})
     nextid=""
  res, _ := json.Marshal(reply)
    rw.Header().Set("Content-Type", "application/json")
    //rw.WriteHeader(201)
    fmt.Fprintf(rw, "%s", res)
    }
    }
  
    


func main()  {
    mux := httprouter.New()
    mux.GET("/locations/:uniqueid", get)
    mux.POST("/locations", create)
    mux.PUT("/locations/:uniqueid", update)
    mux.DELETE("/locations/:uniqueid", delete)
    mux.POST("/trips",shortroute)
    mux.GET("/trips/:tripid", gettrip)
   mux.PUT("/trips/:tripid/request", triptrack)
        server := http.Server{
        Addr:        "0.0.0.0:5555",
        Handler: mux,
    }
    server.ListenAndServe()
}