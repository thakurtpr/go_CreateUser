package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	// "io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"strings"
	"github.com/sethvargo/go-password/password"
	// "github.com/jaswdr/faker/v2"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type User struct {
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Email     string `json:"email"`
	Enabled   bool   `json:"enabled"`
	PhoneNo string `json:"phoneno"`
	// Username  string `json:"username"`
	// Password  string `json:"password"`
}

type Credential struct {
	Type      string `json:"type"`
	Value     string `json:"value"`
	Temporary bool   `json:"temporary"`
}

var client = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

func accessTokenCall() (accessToken interface{}, err error) {

	url := "https://34.93.102.191:18080/auth/realms/camunda-platform/protocol/openid-connect/token"
	method := "POST"

	payload := strings.NewReader("client_id=access_token&client_secret=ZBKi3qEBDKHhszZfwwiFdsvq0pMS3OvH&grant_type=password&username=demo&password=demo")

	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("Error:%v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("Error:%v", err)
	}
	defer res.Body.Close()

	var Token_ResponseData map[string]interface{}
	err=json.NewDecoder(res.Body).Decode(&Token_ResponseData)
	if err!=nil{
		fmt.Println("Error",err)
	}

	return Token_ResponseData["access_token"].(string), nil
}

func getUserId(Username string, accessToken string) (interface{}, error) {
	url := "https://34.93.102.191:18080/auth/admin/realms/camunda-platform/users"
	method := "GET"
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	req.Header.Add("Authorization", accessToken)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
	}
	var userIdExtractor []map[string]interface{}

	checkUserName:=strings.ToLower(Username)
	err=json.NewDecoder(resp.Body).Decode(&userIdExtractor)
	if err!=nil{
		fmt.Println("Error",err)
	}
	for _, value := range userIdExtractor {
		if value["username"] == checkUserName {
			IdUser := value["id"]
			return IdUser, nil
		}
	}
	return nil, err
}

func createUserHandler(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	Token, err := accessTokenCall()
	if err != nil {
		fmt.Println("Error:", err)
	}
	// fmt.Println(Token)

	var incBodyData User


	//Username Generator
	// fake := faker.New()
	// username:=fake.Person().FirstName()
	// fmt.Println(username+" User Generated")


	
	err=json.NewDecoder(request.Body).Decode(&incBodyData)
	if err!=nil{
		fmt.Println("Error",err)
	}
	url := "https://34.93.102.191:18080/auth/admin/realms/camunda-platform/users"
	method := "POST"
	dataToSend := fmt.Sprintf(`{
		"firstName": "%s",
		"lastName": "%s",
		"email": "%s",
		"enabled": true,
		"username": "%s"
	}`, incBodyData.FirstName, incBodyData.LastName, incBodyData.Email,incBodyData.FirstName)

	payload := strings.NewReader(dataToSend)
	// fmt.Println(payload)

	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		fmt.Println("Error:", err)
	}
	accessToken := fmt.Sprintf("Bearer %s", Token)
	req.Header.Add("Authorization", accessToken)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
	}
	var responeCreateUser map[string]interface{}
	err=json.NewDecoder(resp.Body).Decode(&responeCreateUser)
	if err!=nil{
		fmt.Println("Error",err)
	}
	// fmt.Println(responeCreateUser)

	if resp.StatusCode == 201 {
		fmt.Println("User created successfully")

		//password Generator
		resPass,err:=password.Generate(4,4,0,false,false)
		if err!=nil{
			fmt.Println("Error:",err)
		}
		// fmt.Println(resPass)


		userid, err := getUserId(incBodyData.FirstName, accessToken)
		if err != nil {
			fmt.Println("Error:", err)
		}

		userId, ok := userid.(string)
		fmt.Println(userId,"Received")

		if !ok {
			fmt.Println("Error Converting userID")
		}
		url := "https://34.93.102.191:18080/auth/admin/realms/camunda-platform/users/" + userId + "/reset-password"
		method := "PUT"
		send := fmt.Sprintf(`{
			"temporary": false,
			"type": "password",
			"value": "%s"
		}`,resPass)

		payload := strings.NewReader(send)
		req, err := http.NewRequest(method, url, payload)
		if err != nil {
			fmt.Println("Error:", err)
		}
		accessToken := fmt.Sprintf("Bearer %s", Token)
		req.Header.Add("Authorization", accessToken)
		req.Header.Add("Content-Type", "application/json")
		respPassword, err := client.Do(req)
		if err != nil {
			fmt.Println("Error:", err)
		}
		if respPassword.StatusCode == 204 {
			fmt.Println("Password set successfully")
		} else {
			
			fmt.Println("Failed to set password")
		}
		url = "https://34.93.102.191:18080/auth/admin/realms/camunda-platform/users/" + userId + "/role-mappings/realm"
		method = "POST"
		send= `[
			{
				"id": "8ba1339f-ca96-491d-b59f-575a1d248fcd",
				"name": "Tasklist",
				"description": "Grants full access to Tasklist",
				"composite": true,
				"clientRole": false,
				"containerId": "camunda-platform"
			}
		]`
		
		payload = strings.NewReader(send)
		req, err = http.NewRequest(method, url, payload)		
		if err != nil {
			fmt.Println("Error:", err)
		}
		accessToken = fmt.Sprintf("Bearer %s", Token)
		req.Header.Add("Authorization", accessToken)
		req.Header.Add("Content-Type", "application/json")
		respDatacheck, err := client.Do(req)
		if err != nil {
			fmt.Println("Error:", err)
		}
		if respDatacheck.StatusCode == 204 && respPassword.StatusCode == 204 {
			//  Send Details To The User
			auth := smtp.PlainAuth("", "tprop48@gmail.com", "ovgo agtz dsdj bwhq", "smtp.gmail.com")
			to := []string{incBodyData.Email}
			msgStr := fmt.Sprintf("To: %s\r\nSubject: Your Details\r\n\r\nID:%s \r\n Password:%s\r\n", incBodyData.Email,incBodyData.FirstName ,resPass)
			msg := []byte(msgStr)
			err = smtp.SendMail("smtp.gmail.com:587", auth, "tprop48@gmail.com", to, msg)
			if err != nil {
				log.Fatal(err)
			}
			json.NewEncoder(response).Encode(map[string]interface{}{
				"Success": "True",
				"Message": "Check Mail For Id And Password",
			})
			fmt.Println("Role Assigned successfully")
		} else {
			json.NewEncoder(response).Encode(map[string]interface{}{
				"Success": "false",
				"Message": "User Created But Failed To Assign Role || Password",
			})
			fmt.Println("Failed to Assign Role")
		}
	} else {
		json.NewEncoder(response).Encode(map[string]interface{}{
			"Success": "false",
			"Message": responeCreateUser["errorMessage"],
		})
		fmt.Println("Failed to create user")
	}

}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/createUser", createUserHandler).Methods("POST")
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PATCH","PUT","DELETE"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
	})

	handler := c.Handler(r)
	port := ":8086"
	s := &http.Server{
		Addr:    port,
		Handler: handler,
	}

	log.Printf("Server is Running in Port %v", port)
	log.Fatal(s.ListenAndServe())
}
