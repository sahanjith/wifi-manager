package dashboard

import (
	"wislabs.wifi.manager/utils"
	"wislabs.wifi.manager/dao"
	"wislabs.wifi.manager/commons"
	log "github.com/Sirupsen/logrus"
	"strings"
	"strconv"
)

func CreateNewDashboardApp(dashboardAppInfo dao.DashboardAppInfo) {
	appId, err := AddDashboardApp(&dashboardAppInfo)
	if (err == nil) {
		switch (dashboardAppInfo.FilterCriteria){
		case "groupname" :
			AddDashboardAppFilterParams(commons.ADD_DASHBOARD_APP_GROUP, appId, dashboardAppInfo.Parameters)
		case "ssid" :
			AddDashboardAppFilterParams(commons.ADD_DASHBOARD_APP_FILTER_PARAMS, appId, dashboardAppInfo.Parameters)
		}
		AddDashboardAppUsers(&dashboardAppInfo.Users, appId)
		AddDashboardAppMetrics(&dashboardAppInfo.Metrics, appId)
		AddDashboardAppAcls(dashboardAppInfo.Acls, appId)
	}
}

func UpdateDashBoardAppSettings(dashboardAppInfo dao.DashboardAppInfo) {
	UpdateDashboardAppFilterCriteria(&dashboardAppInfo)
	UpdateDashboardAppUsers(&dashboardAppInfo);
	switch (dashboardAppInfo.FilterCriteria){
	case "groupname" :
		UpdateDashboardAppGroups(&dashboardAppInfo);
	case "ssid" :
		UpdateAppFilterParams(&dashboardAppInfo);
	}

	UpdateDashboardAppMetrics(&dashboardAppInfo);
	UpdateDashboardAppAcls(&dashboardAppInfo);
	UpdateDashboardAppAggregateValue(&dashboardAppInfo);
}

func GetAllDashboardAppsOfUser(username string, tenantId int) []dao.DashboardApp {
	dbMap := utils.GetDBConnection("dashboard");
	defer dbMap.Db.Close()

	var apps []dao.DashboardApp
	_, err := dbMap.Select(&apps, commons.GET_DASHBOARD_USER_APPS, username, tenantId)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	return apps
}

func GetDashboardUsersInGroups(tenantId int, appGroups []string) []string {
	var usernames []string
	if (len(appGroups) == 0) {
		return usernames
	}

	dbMap := utils.GetDBConnection(commons.DASHBOARD_DB);
	defer dbMap.Db.Close()
	query := commons.GET_DASHBOARD_USERS_IN_GROUP + " ( "
	for index, value := range appGroups {
		aa := strings.Replace(value, "\"", "", -1)

		query += "'" + strings.Trim(aa, " ") + "'"
		if index < len(appGroups) - 1 {
			query += ","
		}
	}
	_, err := dbMap.Select(&usernames, query + " ))GROUP BY userid HAVING COUNT(DISTINCT groupid) =" + strconv.Itoa(len(appGroups)) + ")", tenantId)
	checkErr(err, "Error occured while getting users of group")
	return usernames
}

func GetDashboardUsersOfApp(appId int) []dao.DashboardAppUser {
	dbMap := utils.GetDBConnection("dashboard");
	defer dbMap.Db.Close()

	var users []dao.DashboardAppUser
	_, err := dbMap.Select(&users, commons.GET_DASHBOARD_APP_USERS, appId)
	if err != nil {
		//panic(err.Error()) // proper error handling instead of panic in your app
		return users
	}
	return users
}

func GetDashboardMetricsOfApp(appId int) []dao.DashboardAppMetric {
	dbMap := utils.GetDBConnection("dashboard");
	defer dbMap.Db.Close()

	var metrics []dao.DashboardAppMetric
	_, err := dbMap.Select(&metrics, commons.GET_DASHBOARD_APP_METRICS, appId)
	if err != nil {
		//panic(err.Error()) // proper error handling instead of panic in your app
		return metrics
	}
	return metrics
}

func GetDashboardGroupsOfApp(appId int) []string {
	dbMap := utils.GetDBConnection("dashboard");
	defer dbMap.Db.Close()

	var groups []string
	_, err := dbMap.Select(&groups, commons.GET_DASHBOARD_APP_GROUPS, appId)
	if err != nil {
		//panic(err.Error()) // proper error handling instead of panic in your app
		return groups
	}
	return groups
}

func GetDashboardAclsOfApp(appId int) string {
	dbMap := utils.GetDBConnection("dashboard");
	defer dbMap.Db.Close()

	var acls dao.DashboardAppAcls
	err := dbMap.SelectOne(&acls, commons.GET_DASHBOARD_APP_ACLS, appId)
	if err != nil {
		//panic(err.Error()) // proper error handling instead of panic in your app
		return acls.Acls
	}
	return acls.Acls
}

func GetDashboardAggregateOfApp(appId int) string {
	dbMap := utils.GetDBConnection("dashboard");
	defer dbMap.Db.Close()

	var strAggregate []string
	_, err := dbMap.Select(&strAggregate, commons.GET_DASHBOARD_APP_AGGREGATE, appId)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	return strAggregate[0]
}

func GetFilterParamsOfApp(appId int) []string {
	dbMap := utils.GetDBConnection(commons.DASHBOARD_DB);
	defer dbMap.Db.Close()

	var filterParams []string
	_, err := dbMap.Select(&filterParams, commons.GET_DASHBOARD_APP_FILTER_PARAMS, appId)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	return filterParams
}

func GetFilterCriteriaOfApp(appId int) string {
	dbMap := utils.GetDBConnection(commons.DASHBOARD_DB);
	defer dbMap.Db.Close()

	var filterCriteria string
	err := dbMap.SelectOne(&filterCriteria, commons.GET_DASHBOARD_APP_CRITERIA, appId)
	if err != nil {
		log.Error(err.Error())
		return "none"
	}
	return filterCriteria
}

func GetAllDashboardMetrics(tenantId int) []dao.DashboardMetric {
	dbMap := utils.GetDBConnection("dashboard");
	defer dbMap.Db.Close()

	var metrics []dao.DashboardMetric
	_, err := dbMap.Select(&metrics, commons.GET_ALL_DASHBOARD_METRICS, tenantId)
	if err != nil {
		//panic(err.Error()) // proper error handling instead of panic in your app
		return metrics
	}
	return metrics
}

func GetAllDashboardAclTypes() []string {
	dbMap := utils.GetDBConnection("portal");
	defer dbMap.Db.Close()
	var aclsTypes []string

	_, err := dbMap.Select(&aclsTypes, commons.GET_ALL_DASHBOARD_ACLS)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	return aclsTypes
}

func AddDashboardApp(app *dao.DashboardAppInfo) (int64, error) {
	dbMap := utils.GetDBConnection(commons.DASHBOARD_DB);
	defer dbMap.Db.Close()

	stmtIns, err := dbMap.Db.Prepare(commons.ADD_DASHBOARD_APP)
	defer stmtIns.Close()

	if err != nil {
		return 0, err
	}
	result, err := stmtIns.Exec(app.TenantId, app.Name, app.Aggregate, app.FilterCriteria)
	if err != nil {
		return 0, err
	} else {
		id, err := result.LastInsertId()
		return id, err
	}
}

func AddDashboardAppFilterParams(query string, appId int64, filterParams []string) error {
	dbMap := utils.GetDBConnection(commons.DASHBOARD_DB);
	defer dbMap.Db.Close()

	stmtIns, err := dbMap.Db.Prepare(query)
	defer stmtIns.Close()

	if err != nil {
		return err
	}
	for i := 0; i < len(filterParams); i++ {
		_, err = stmtIns.Exec(appId, filterParams[i])
	}
	return err
}

func AddDashboardAppMetrics(appMetrics *[]dao.DashboardAppMetric, appId int64) error {
	dbMap := utils.GetDBConnection("dashboard");
	defer dbMap.Db.Close()

	stmtIns, err := dbMap.Db.Prepare(commons.ADD_DASHBOARD_APP_METRIC)
	defer stmtIns.Close()

	if err != nil {
		return err
	}
	for i := 0; i < len(*appMetrics); i++ {
		_, err = stmtIns.Exec(appId, (*appMetrics)[i].MetricId)
	}
	return err
}

func AddDashboardAppUsers(appUsers *[]dao.DashboardAppUser, appId int64) error {
	dbMap := utils.GetDBConnection("dashboard");
	defer dbMap.Db.Close()

	stmtIns, err := dbMap.Db.Prepare(commons.ADD_DASHBOARD_APP_USER)
	defer stmtIns.Close()

	if err != nil {
		return err
	}
	for i := 0; i < len(*appUsers); i++ {
		_, err = stmtIns.Exec((*appUsers)[i].TenantId, appId, (*appUsers)[i].UserName)
	}
	return err
}

func AddDashboardAppAcls(acls string, appId int64) error {
	dbMap := utils.GetDBConnection("dashboard");
	defer dbMap.Db.Close()

	stmtIns, err := dbMap.Db.Prepare(commons.ADD_DASHBOARD_ACLS)
	defer stmtIns.Close()

	if err != nil {
		return err
	}
	_, err = stmtIns.Exec(appId, acls)
	if err != nil {
		//panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtIns.Close()
	return err
}

func DeleteDashboardApp(appId int, tenantId int) error {
	dbMap := utils.GetDBConnection("dashboard");
	defer dbMap.Db.Close()

	stmtIns, err := dbMap.Db.Prepare(commons.DELETE_DASHBOARD_APP)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	_, err = stmtIns.Exec(appId, tenantId)
	if err != nil {
		//panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtIns.Close()
	return err
}

func UpdateDashboardAppAggregateValue(dashboardAppInfo  *dao.DashboardAppInfo) {
	dbMap := utils.GetDBConnection("dashboard");
	defer dbMap.Db.Close()

	stmtIns, err := dbMap.Db.Prepare(commons.UPDATE_DB_APP_AGGREGATE_VALUE)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	_, err = stmtIns.Exec(dashboardAppInfo.Aggregate, dashboardAppInfo.AppId)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtIns.Close()

}

func UpdateDashboardAppFilterCriteria(dashboardAppInfo  *dao.DashboardAppInfo) {
	dbMap := utils.GetDBConnection(commons.DASHBOARD_DB);
	defer dbMap.Db.Close()

	stmtIns, err := dbMap.Db.Prepare(commons.UPDATE_DB_APP_FILTER_CRITERIA)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	_, err = stmtIns.Exec(dashboardAppInfo.FilterCriteria, dashboardAppInfo.AppId)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtIns.Close()
}

func UpdateDashboardAppGroups(dashboardAppInfo  *dao.DashboardAppInfo) {
	dbMap := utils.GetDBConnection(commons.DASHBOARD_DB);
	defer dbMap.Db.Close()

	stmtIns, err := dbMap.Db.Prepare(commons.DELETE_DB_APP_GROUPS)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	_, err = stmtIns.Exec(dashboardAppInfo.AppId)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtIns.Close()
	AddDashboardAppFilterParams(commons.ADD_DASHBOARD_APP_GROUP, dashboardAppInfo.AppId, dashboardAppInfo.Parameters)
}

func UpdateDashboardAppMetrics(dashboardAppInfo  *dao.DashboardAppInfo) {
	dbMap := utils.GetDBConnection("dashboard");
	defer dbMap.Db.Close()

	var updatedMetrics *[]dao.DashboardAppMetric
	updatedMetrics = &dashboardAppInfo.Metrics;
	var metrics []dao.DashboardAppMetric
	_, err := dbMap.Select(&metrics, commons.GET_EXIST_DASHBOARD_APP_METRICS, dashboardAppInfo.AppId)
	if err != nil {
		panic(err.Error())
	}

	stmtInsDelete, err := dbMap.Db.Prepare(commons.DELETE_OLD_DB_APP_METRICS)
	defer stmtInsDelete.Close()
	stmtInsAdd, err := dbMap.Db.Prepare(commons.ADD_NEW_DB_APP_METRICS)
	defer stmtInsAdd.Close()
	if err != nil {
		panic(err.Error())
	}

	var Length = (len(*updatedMetrics))
	var countID = len(metrics)
	if ( Length <= countID) {
		for i := 0; i < countID; i++ {
			if !(checkContainsMetrics(metrics[i].MetricId, (*updatedMetrics))) {
				_, err = stmtInsDelete.Exec(&dashboardAppInfo.AppId, metrics[i].MetricId)
			}
		}
	} else {
		for j := 0; j < Length; j++ {
			if !(checkContainsMetrics((*updatedMetrics)[j].MetricId, metrics)) {
				_, err = stmtInsAdd.Exec(&dashboardAppInfo.AppId, (*updatedMetrics)[j].MetricId)
			}
		}
	}
	if err != nil {
		panic(err.Error())
	}
}

func UpdateDashboardAppUsers(dashboardAppInfo  *dao.DashboardAppInfo) {
	dbMap := utils.GetDBConnection("dashboard");
	defer dbMap.Db.Close()

	var updatedUsers *[]dao.DashboardAppUser
	updatedUsers = &dashboardAppInfo.Users;
	var users []dao.DashboardAppUser
	_, err := dbMap.Select(&users, commons.GET_DASHBOARD_APP_USERS, dashboardAppInfo.AppId)
	if err != nil {
		panic(err.Error())
	}

	stmtInsDelete, err := dbMap.Db.Prepare(commons.DELETE_OLD_DB_APP_USERS)
	defer stmtInsDelete.Close()
	stmtInsAdd, err := dbMap.Db.Prepare(commons.ADD_NEW_DB_APP_USERS)
	defer stmtInsAdd.Close()
	if err != nil {
		panic(err.Error())
	}

	var Length = (len(*updatedUsers))
	var countUsers = len(users)
	if ( Length <= countUsers) {
		for i := 0; i < countUsers; i++ {
			if !(checkContainsUsers(users[i].UserName, (*updatedUsers))) {
				_, err = stmtInsDelete.Exec(dashboardAppInfo.AppId, users[i].UserName)
			}
		}
	} else {
		for j := 0; j < Length; j++ {
			if !(checkContainsUsers((*updatedUsers)[j].UserName, users)) {
				_, err = stmtInsAdd.Exec(dashboardAppInfo.TenantId, dashboardAppInfo.AppId, (*updatedUsers)[j].UserName)
			}
		}
	}

	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
}

func UpdateDashboardAppAcls(dashboardAppInfo  *dao.DashboardAppInfo) {
	dbMap := utils.GetDBConnection("dashboard");
	defer dbMap.Db.Close()

	stmtIns, err := dbMap.Db.Prepare(commons.UPDATE_DB_APP_ACLS)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	_, err = stmtIns.Exec(dashboardAppInfo.Acls, dashboardAppInfo.AppId)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtIns.Close()
}

func UpdateAppFilterParams(dashboardAppInfo  *dao.DashboardAppInfo) {
	dbMap := utils.GetDBConnection(commons.DASHBOARD_DB);
	defer dbMap.Db.Close()

	stmtIns, err := dbMap.Db.Prepare(commons.DELETE_DASHBOARD_APP_FILTER_PARAMS)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	_, err = stmtIns.Exec(dashboardAppInfo.AppId)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtIns.Close()
	AddDashboardAppFilterParams(commons.ADD_DASHBOARD_APP_FILTER_PARAMS, dashboardAppInfo.AppId, dashboardAppInfo.Parameters)
}

func checkContainsMetrics(metricid int, groups []dao.DashboardAppMetric) bool {
	for _, v := range groups {
		if v.MetricId == metricid {
			return true
		}
	}
	return false
}

func checkContainsUsers(username string, users []dao.DashboardAppUser) bool {
	for _, v := range users {
		if v.UserName == username {
			return true
		}
	}
	return false
}