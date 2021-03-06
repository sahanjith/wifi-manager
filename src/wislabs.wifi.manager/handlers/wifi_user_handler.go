package handlers

import (
	wifi_controller "wislabs.wifi.manager/controllers/wifi"
	"wislabs.wifi.manager/dao"
	"encoding/json"
	"net/http"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"strconv"
	"wislabs.wifi.manager/authenticator"
)


/**
* POST
* @path /users
*/
func AddUserHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var user dao.PortalUser
	err := decoder.Decode(&user)
	if(err != nil){
		log.Fatalln("Error while decoding wifi user json")
	}
	erradduser := wifi_controller.AddWiFiUser(&user)
	if (erradduser != nil) {
		log.Fatalln("Error while adding wifi user", err)
	}
	w.WriteHeader(http.StatusOK)
}

/**
* GET
* @path /users
* return [{"id":0,"username":"anu","password":"","acctstarttime":"2015-09-20 18:49:32",
*         "acctlastupdatedtime":"2015-09-20 18:49:32","acctactivationtime":"","acctstoptime":"2015-09-20 19:49:32"}]
*/
func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	if(!authenticator.IsAuthorized(authenticator.WIFI_USERS, authenticator.ACTION_READ,r)){
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	tenantId, err := strconv.Atoi(r.Header.Get("tenantid"))
	if (err != nil) {
		log.Fatalln("Error while reading tenantid", err)
	}

	draw, err := strconv.Atoi(r.FormValue("draw"))
	users := wifi_controller.GetAllWiFiUsers(tenantId, draw, r)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(users); err != nil {
		panic(err)
	}
}

/**
* DELETE
* @path /wifi/{tenantid}/users/<user-id>
* delete user from radacct, radcheck and accounting tables
*/
func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	groupname := vars["groupname"]
	tenantid, err := strconv.Atoi(vars["tenantid"])
	if (err != nil) {
		log.Fatalln("Error while reading tenantid", err)
	}
	err = wifi_controller.DeleteUserAccountingSession(username, groupname, tenantid)
	if groupname == "Master" {
		err = wifi_controller.DeleteUserFromRadAcct(username, tenantid)
		err = wifi_controller.DeleteUserFromRadCheck(username, tenantid)
	}
	if err != nil {
		log.Fatalln("Error while deleting user from accounting table" + username + " from DB ", err)
		w.WriteHeader(http.StatusNotFound)
	}else {
		w.WriteHeader(http.StatusOK)
	}
}

func WifiUserExistInGroupNameHanlder(w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
	username := vars["username"]
	groupname := vars["groupname"]
	tenantId, err := strconv.Atoi(r.Header.Get("tenantid"))
	if (err != nil) {
		log.Fatalln("Error while reading tenantid", err)
	}
	var existUser int
	existUser, err = wifi_controller.IsWifiUserExistInGroup(tenantId, username, groupname);
	if err != nil {
		checkErr(err,"Error happening while Wifi user exist in group")
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(existUser); err != nil {
		checkErr(err, "Error happening while JSON encoding.")
	}
}

/**
* POST
* @path /users
*/
func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var user dao.PortalUser
	err := decoder.Decode(&user)
	if(err != nil){
		log.Fatalln("Error while decoding wifi user json")
	}
	wifi_controller.UpdateWiFiUser(&user)
	w.WriteHeader(http.StatusOK)
}

/**
* POST
* @path /wifi/users/count
*
*/
func GetUsersCountFromToHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var constrains dao.Constrains
	err := decoder.Decode(&constrains)

	count, countpre := wifi_controller.GetUsersCountFromTo(constrains)

	changePercentage := getChangePrecentageSummaryDetails(countpre,count)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(changePercentage); err != nil {
		panic(err)
	}
}

/**
* POST
* @path /wifi/users/returncount
*
*/
func GetReturningUsersCountFromToHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var constrains dao.Constrains
	err := decoder.Decode(&constrains)

	count, countpre := wifi_controller.GetReturningUsersCount(constrains)
	changePercentage := getChangePrecentageSummaryDetails(countpre,count)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(changePercentage); err != nil {
		panic(err)
	}
}

/**
* POST
* @path /wifi/users/dailycountseries
*
*/
func GetDailyUsersCountSeriesFromToHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var constrains dao.Constrains
	err := decoder.Decode(&constrains)

	count := wifi_controller.GetDailyUserCountSeriesFromTo(constrains)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(count); err != nil {
		panic(err)
	}
}

/**
* POST
* @path /wifi/users/countbydownlods/{threshold}
*
*/
func GetUserCountOfDownloadsOverHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var constrains dao.Constrains
	decoder.Decode(&constrains)
	vars := mux.Vars(r)
	threshold := vars["threshold"]
	value, _ := strconv.Atoi(threshold)

	count, countpre := wifi_controller.GetUserCountOfDownloadsOver(constrains, value)

	changePercentage := getChangePrecentageSummaryDetails(countpre,count)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(changePercentage); err != nil {
		panic(err)
	}
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Error(msg, err)
	}
}