package pterodactylGoApi

import (
	"bytes"
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"net/http"
	"strconv"
)

type Client struct {
	/* ServerHostname = "http://xxx.example.com" */
	url string
	/* your application API */
	token string
}

func NewClient(url string, token string) *Client {
	for stringLen := len(url); url[stringLen-1] == '/'; stringLen -= 1 {
	}
	return &Client{url: url, token: token}
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func (this *Client) api(data interface{}, endPoint string, method string) ([]byte, int) {
	/* Send requests to pterodactyl panel */
	url := this.url + "/api/application/" + endPoint
	var res []byte
	var status int
	if method == "POST" || method == "PATCH" {
		ujson, err := json.Marshal(data)
		if err != nil {
			fmt.Print("cant marshal data:" + err.Error())
		}
		ubody := bytes.NewReader(ujson)
		req, _ := http.NewRequest(method, url, ubody)
		req.Header.Set("Authorization", "Bearer "+this.token)
		req.Header.Set("Accept", "Application/vnd.pterodactyl.v1+json")
		req.ContentLength = int64(len(ujson))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			panic("cant Do req:" + err.Error())
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		res = body
		status = resp.StatusCode
	} else {
		req, _ := http.NewRequest(method, url, nil)
		req.Header.Set("Authorization", "Bearer "+this.token)
		req.Header.Set("Accept", "Application/vnd.pterodactyl.v1+json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			panic("cant Do req: " + err.Error())
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		res = body
		status = resp.StatusCode
	}
	return res, status
}

func (this *Client) TestConnection() {
	test, _ := this.api("", "nodes", "GET")
	fmt.Print("PterodactylAPI returns: ", test)
}

func (this *Client) GetUser(ID interface{}, isExternal bool) (User, bool) {
	var endPoint string
	if isExternal {
		endPoint = "users/external/" + ID.(string)
	} else {
		endPoint = "users/" + strconv.Itoa(ID.(int))
	}
	body, status := this.api("", endPoint, "GET")
	if status == 404 || status == 400 {
		return User{}, false
	}
	dec := struct {
		Object     string `json:"object"`
		Attributes User   `json:"attributes"`
	}{}
	if err := json.Unmarshal(body, &dec); err == nil {
		return dec.Attributes, true
	}
	return User{}, false
}

func (this *Client) GetAllUsers() []User {
	body, status := this.api("", "users/", "GET")
	if status != 200 {
		fmt.Print("cant get all users: " + strconv.Itoa(status))
		return []User{}
	}
	dec := struct {
		data []struct {
			Attributes User `json:"attributes"`
		}
	}{}
	var users []User
	if err := json.Unmarshal(body, &dec); err == nil {
		for _, v := range dec.data {
			users = append(users, v.Attributes)
		}
	}
	return users
}

func (this *Client) GetNest(nestID int) Nest {
	body, status := this.api("", "nests/"+strconv.Itoa(nestID), "GET")
	if status != 200 {
		fmt.Print("cant get nest: " + strconv.Itoa(nestID) + " with status code: " + strconv.Itoa(status))
		return Nest{}
	}
	dec := struct {
		Attributes Nest `json:"attributes"`
	}{}
	if err := json.Unmarshal(body, &dec); err == nil {
		return dec.Attributes
	}
	return Nest{}
}

func (this *Client) GetAllNests() []Nest {
	body, status := this.api("", "nests/", "GET")
	if status != 200 {
		fmt.Print("cant get all nests: " + strconv.Itoa(status))
		return []Nest{}
	}
	var ret []Nest
	dec := struct {
		data []struct {
			Attributes Nest `json:"attributes"`
		}
	}{}
	if err := json.Unmarshal(body, &dec); err == nil {
		for _, v := range dec.data {
			ret = append(ret, v.Attributes)
		}
		return ret
	}
	return []Nest{}
}

func (this *Client) GetEgg(nestID int, eggID int) Egg {
	body, status := this.api("", "nests/"+strconv.Itoa(nestID)+"/eggs/"+strconv.Itoa(eggID), "GET")
	if status != 200 {
		return Egg{}
	}
	dec := struct {
		Attributes Egg `json:"attributes"`
	}{}
	if err := json.Unmarshal(body, &dec); err == nil {
		return dec.Attributes
	}
	return Egg{}
}

func (this *Client) GetAllEggs(nestID int) []Egg {
	body, status := this.api("", "nests/"+strconv.Itoa(nestID)+"/eggs/", "GET")
	if status != 200 {
		fmt.Print("cant get all eggs: " + strconv.Itoa(status))
		return []Egg{}
	}
	var ret []Egg
	dec := struct {
		data []struct {
			Attributes Egg `json:"attributes"`
		}
	}{}
	if err := json.Unmarshal(body, &dec); err == nil {
		for _, v := range dec.data {
			ret = append(ret, v.Attributes)
		}
		return ret
	}
	return []Egg{}
}

func (this *Client) GetNode(nodeID int) Node {
	body, status := this.api("", "nodes/"+strconv.Itoa(nodeID), "GET")
	if status != 200 {
		return Node{}
	}
	dec := struct {
		Attributes Node `json:"attributes"`
	}{}
	if err := json.Unmarshal(body, &dec); err == nil {
		return dec.Attributes
	}
	return Node{}
}

func (this *Client) GetAllocations(nodeID int) []Allocation {
	body, status := this.api("", "nodes/"+strconv.Itoa(nodeID)+"/allocations", "GET")
	if status != 200 {
		fmt.Print("cant get allocations with status code: " + strconv.Itoa(status))
		return []Allocation{}
	}
	dec := struct {
		Data []struct {
			Attributes Allocation `json:"attributes"`
		} `json:"data"`
	}{}
	var ret []Allocation
	if err := json.Unmarshal(body, &dec); err == nil {
		for _, v := range dec.Data {
			if !v.Attributes.Assigned {
				ret = append(ret, v.Attributes)
			}
		}
	}
	return ret
}

func (this *Client) GetServer(ID interface{}, isExternal bool) Server {
	var endPoint string
	if isExternal {
		endPoint = "servers/external/" + ID.(string)
	} else {
		endPoint = "servers/" + strconv.Itoa(ID.(int))
	}
	body, status := this.api("", endPoint, "GET")
	if status != 200 {
		return Server{}
	}
	dec := struct {
		Attributes Server `json:"attributes"`
	}{}
	if err := json.Unmarshal(body, &dec); err == nil {
		return dec.Attributes
	} else {
		fmt.Print(err.Error())
	}
	return Server{}
}

func (this *Client) GetAllServers() []Server {
	body, status := this.api("", "servers", "GET")
	if status != 200 {
		return []Server{}
	}
	dec := struct {
		data []struct {
			Attributes Server `json:"attributes"`
		}
	}{}
	var servers []Server
	if err := json.Unmarshal(body, &dec); err == nil {
		for _, v := range dec.data {
			servers = append(servers, v.Attributes)
		}
	}
	return servers
}

func (this *Client) GetServerID(serverExternalID string) int {
	server := this.GetServer(serverExternalID, true)
	if server == (Server{}) {
		return 0
	}
	return server.Id
}

func (this *Client) SuspendServer(serverExternalID string) error {
	serverID := this.GetServerID(serverExternalID)
	if serverID == 0 {
		return errors.New("suspend failed because server not found: " + strconv.Itoa(serverID))
	}
	_, status := this.api("", "servers/"+strconv.Itoa(serverID)+"/suspend", "POST")
	if status != 204 {
		return errors.New("cant suspend server: " + strconv.Itoa(serverID) + " with status code: " + strconv.Itoa(status))
	}
	return nil
}

func (this *Client) UnsuspendServer(serverExternalID string) error {
	serverID := this.GetServerID(serverExternalID)
	if serverID == 0 {
		return errors.New("unsuspend failed because server not found: " + strconv.Itoa(serverID))
	}
	_, status := this.api("", "servers/"+strconv.Itoa(serverID)+"/unsuspend", "POST")
	if status != 204 {
		return errors.New("cant unsuspend server: " + strconv.Itoa(serverID) + " with status code: " + strconv.Itoa(status))
	}
	return nil
}

func (this *Client) ReinstallServer(serverExternalID string) error {
	serverID := this.GetServerID(serverExternalID)
	if serverID == 0 {
		return errors.New("reinstall failed because server not found: " + strconv.Itoa(serverID))
	}
	_, status := this.api("", "servers/"+strconv.Itoa(serverID)+"/reinstall", "POST")
	//fmt.Print(body)
	if status != 204 {
		return errors.New("cant reinstall server: " + strconv.Itoa(serverID) + " with status code: " + strconv.Itoa(status))
	}
	return nil
}

func (this *Client) DeleteServer(serverExternalID string) error {
	serverID := this.GetServerID(serverExternalID)
	if serverID == 0 {
		return errors.New("delete failed because server not found: " + strconv.Itoa(serverID))
	}
	_, status := this.api("", "servers/"+strconv.Itoa(serverID), "DELETE")
	if status != 204 {
		return errors.New("cant delete server: " + strconv.Itoa(serverID) + " with status code: " + strconv.Itoa(status))
	}
	return nil
}

/*_ = PterodactylCreateUser(params, PostPteUser{
ExternalId: "aSTRING",
Username:   "aSTRING",
Email:      "user@example.com",
Language:   "en",
RootAdmin:  false,
Password:   "PASSwd",
FirstName:  "first",
LastName:   "last",
})*/

func (this *Client) CreateUser(userInfo interface{}) error {
	body, status := this.api(userInfo, "users", "POST")
	if status != 201 {
		return errors.New("cant create user with status code: " + strconv.Itoa(status) + " body: " + string(body))
	}
	return nil
}

func (this *Client) DeleteUser(externalID string) error {
	if user, ok := this.GetUser(externalID, true); ok {
		_, status := this.api("", "users/"+strconv.Itoa(user.Uid), "DELETE")
		if status != 204 {
			return errors.New("cant delete user: " + user.UserName + " with status code: " + strconv.Itoa(status))
		}
		return nil
	} else {
		return errors.New("cant get user")
	}
}

func (this *Client) GetEnv(nestID int, eggID int) map[string]string {
	ret := map[string]string{}
	body, status := this.api("", "nests/"+strconv.Itoa(nestID)+"/eggs/"+strconv.Itoa(eggID)+"?include=variables", "GET")
	if status != 200 {
		return map[string]string{}
	}
	dec := struct {
		Attributes struct {
			Relationships struct {
				Variables struct {
					Data []map[string]interface{} `json:"data"`
				} `json:"variables"`
			} `json:"relationships"`
		} `json:"attributes"`
	}{}
	if err := json.Unmarshal(body, &dec); err == nil {
		//fmt.Print(dec.Attributes.Relationships.Variables.Data)
		for _, v := range dec.Attributes.Relationships.Variables.Data {
			keys := v["attributes"].(map[string]interface{})
			key := keys["env_variable"].(string)
			value := keys["default_value"].(string)
			if key != "" {
				ret[key] = value
			}
		}
	} else {
		fmt.Print(err.Error())
	}
	return ret
}

/*_ = PterodactylCreateServer(params, PterodactylServer{
Id:          111,
ExternalId:  "12121",
Uuid:        "",
Identifier:  "",
Name:        "12121",
Description: "12121",
Suspended:   false,
Limits: PterodactylServerLimit{
Memory: 1024,
Swap:   -1,
Disk:   2048,
IO:     500,
CPU:    100,
},
UserId:     1,
NodeId:     5,
Allocation: 517,
NestId:     1,
EggId:      17,
PackId:     0,
})*/

func (this *Client) CreateServer(serverInfo Server) error {
	eggInfo := this.GetEgg(serverInfo.NestId, serverInfo.EggId)
	envInfo := this.GetEnv(serverInfo.NestId, serverInfo.EggId)
	postData := map[string]interface{}{
		"name":         serverInfo.Name,
		"user":         serverInfo.UserId,
		"nest":         serverInfo.NestId,
		"egg":          serverInfo.EggId,
		"docker_image": eggInfo.DockerImage,
		"startup":      eggInfo.StartUp,
		"description":  serverInfo.Description,
		"oom_disabled": true,
		"limits": map[string]int{
			"memory": serverInfo.Limits.Memory,
			"swap":   serverInfo.Limits.Swap,
			"io":     serverInfo.Limits.IO,
			"cpu":    serverInfo.Limits.CPU,
			"disk":   serverInfo.Limits.Disk,
		},
		"feature_limits": map[string]interface{}{
			"databases":   nil,
			"allocations": serverInfo.Allocation,
		},
		"environment":         envInfo,
		"start_on_completion": false,
		"external_id":         serverInfo.ExternalId,
		"allocation": map[string]interface{}{
			"default": serverInfo.Allocation,
		},
	}
	body, status := this.api(postData, "servers", "POST")
	if status == 400 {
		return errors.New("could not find any nodes satisfying the request")
	}
	if status != 201 {
		fmt.Print(body)
		return errors.New("failed to create the server, received the error code: " + strconv.Itoa(status))
	}
	var dec struct {
		Server Server `json:"attributes"`
	}
	if err := json.Unmarshal(body, &dec); err == nil {
		fmt.Print("New server created: ", dec.Server)
	} else {
		return err
	}
	if dec.Server == (Server{}) {
		return errors.New("Pterodactyl API returns empty struct: " + string(body))
	}
	return nil
}

func (this *Client) UpdateServerDetail(externalID string, details PostUpdateDetails) error {
	serverID := this.GetServerID(externalID)
	patchData := map[string]interface{}{
		"user":        details.UserID,
		"description": details.Description,
		"name":        details.ServerName,
		"external_id": details.ExternalID,
	}
	_, status := this.api(patchData, "servers/"+strconv.Itoa(serverID)+"/details", "PATCH")
	if status != 200 {
		return errors.New("cant update server details data: " + externalID)
	}
	return nil
}

func (this *Client) UpdateServerBuild(externalID string, build PostUpdateBuild) error {
	serverID := this.GetServerID(externalID)
	patchData := map[string]interface{}{
		"allocation":   build.Allocation,
		"memory":       build.Memory,
		"io":           build.IO,
		"swap":         build.Swap,
		"cpu":          build.CPU,
		"disk":         build.Disk,
		"oom_disabled": build.OomDisabled,
		"feature_limits": map[string]interface{}{
			"databases":   build.Database,
			"allocations": build.Allocations,
		},
	}
	_, status := this.api(patchData, "servers/"+strconv.Itoa(serverID)+"/build", "PATCH")
	if status != 200 {
		return errors.New("cant update server build data: " + externalID)
	}
	return nil
}

func (this *Client) UpdateServerStartup(externalID string, packID int) error {
	server := this.GetServer(externalID, true)
	eggInfo := this.GetEgg(server.NestId, server.EggId)
	patchData := map[string]interface{}{
		"environment":  this.GetEnv(server.NestId, server.EggId),
		"startup":      eggInfo.StartUp,
		"egg":          server.EggId,
		"pack":         packID,
		"image":        eggInfo.DockerImage,
		"skip_scripts": false,
	}
	_, status := this.api(patchData, "servers/"+strconv.Itoa(server.Id)+"/startup", "PATCH")
	if status != 200 {
		return errors.New("cant update server startup data: " + externalID)
	}
	return this.ReinstallServer(externalID)
}
