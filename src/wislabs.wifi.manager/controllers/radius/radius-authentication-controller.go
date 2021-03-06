package radius

import (
	"wislabs.wifi.manager/dao"
	"github.com/kirves/goradius"
	"wislabs.wifi.manager/utils"
	"wislabs.wifi.manager/commons"
	"gopkg.in/gorp.v1"
	"database/sql"
	"strconv"
	"errors"
	"net"
	"strings"
	"bytes"
)

func TestAuthenticationOnUser(nasClientInfo dao.NasClientTestInfo) bool{
	auth := goradius.Authenticator(nasClientInfo.ServerIp , nasClientInfo.AuthPort , nasClientInfo.Secret)
	authenticateStatus, err := auth.Authenticate(nasClientInfo.UserName, nasClientInfo.Password, nasClientInfo.NASClientName)

	if err != nil {
		return authenticateStatus;
	}
	return authenticateStatus
}
func CreateRadiusServer(rServer dao.RadiusServer) error{
	dbMap := utils.GetDBConnection("dashboard");
	defer dbMap.Db.Close()

	stmtIns, err := dbMap.Db.Prepare(commons.ADD_RADIUS_SERVER)
	defer stmtIns.Close()
	if err != nil {
		return errors.New("Error occourred while insert radius server into radiusservers table| Stack : " + err.Error() )
	}
	_, err = stmtIns.Exec(rServer.TenantId, rServer.DBHostName, rServer.DBHostIp, rServer.DBSchemaName, rServer.DBHostPort, rServer.DBUserName, rServer.DBPassword);
	return err
}

func GetAllRadiusDetails(tenantId int) ([]dao.RadiusServer, error) {
	dbMap := utils.GetDBConnection("dashboard");
	defer dbMap.Db.Close()
	var radiusConfigs []dao.RadiusServer
	_, err := dbMap.Select(&radiusConfigs, commons.GET_ALL_RADIUS_CONFIGS, tenantId)
	if err != nil{
		return nil,	errors.New("Error occourred while getting all radius server configs| Stack : " + err.Error() )
	}
	return radiusConfigs, err
}

func GetInstanceConfigsById(instanceId int, tenantId int) (dao.RadiusServer, error){
	dbMap := utils.GetDBConnection("dashboard");
	defer dbMap.Db.Close()

	var radiusServerConfigs dao.RadiusServer

	err := dbMap.SelectOne(&radiusServerConfigs, commons.GET_SERVERCONFIGS_BY_INSTANCEID, instanceId, tenantId)
	if err != nil {
		return radiusServerConfigs, errors.New("Error occourred while getting server configs by id | Stack : " + err.Error() )
	}
	return radiusServerConfigs, err
}

func CreateNASClient(radiusServerConfigs dao.RadiusServer, nasClientInfo dao.NasClient) error{
	dbMap, errDb := GetRadiusServerDBConnection(radiusServerConfigs);
	defer dbMap.Db.Close()
	if errDb != nil{
		return errors.New("Error occourred while connection to DB server |"+" DB Server IP :"+radiusServerConfigs.DBHostIp +":"+strconv.Itoa(radiusServerConfigs.DBHostPort)+" | Stack : " + errDb.Error() )
	}
	stmtIns, err := dbMap.Db.Prepare(commons.ADD_NAS_CLIENT)
	defer stmtIns.Close()
	if err != nil {
		return errors.New("Error occourred while creating nas client | Stack : " + err.Error() )
	}
	_, err = stmtIns.Exec(nasClientInfo.NasName, nasClientInfo.ShortName, nasClientInfo.NasType, nasClientInfo.NasPorts, nasClientInfo.Secret);
	return err
}

func UpdateNASClient(radiusServerConfigs dao.RadiusServer, nasClientInfo dao.NasClient) error{
	dbMap, errDb := GetRadiusServerDBConnection(radiusServerConfigs);
	defer dbMap.Db.Close()
	if errDb != nil{
		return errors.New("Error occourred while connection to DB server |"+" DB Server IP :"+radiusServerConfigs.DBHostIp +":"+strconv.Itoa(radiusServerConfigs.DBHostPort)+" | Stack : " + errDb.Error() )
	}
	stmtIns, err := dbMap.Db.Prepare(commons.UPDATE_NAS_CLIENT)
	defer stmtIns.Close()
	if err != nil {
		return errors.New("Error occourred while update nas client | Stack : " + err.Error() )
	}
	_, err = stmtIns.Exec(nasClientInfo.ShortName, nasClientInfo.NasType, nasClientInfo.NasPorts, nasClientInfo.Secret, nasClientInfo.NasClientID);
	return err
}

func DeleteNASClient(radiusServerConfigs dao.RadiusServer, nasClientId int) error{
	dbMap, errDb := GetRadiusServerDBConnection(radiusServerConfigs);
	defer dbMap.Db.Close()
	if errDb != nil{
		return errors.New("Error occourred while connection to DB server |"+" DB Server IP :"+radiusServerConfigs.DBHostIp +":"+strconv.Itoa(radiusServerConfigs.DBHostPort)+" | Stack : " + errDb.Error() )
	}
	_, err := dbMap.Exec(commons.DELETE_NAS_CLIENT, nasClientId)
	if (err != nil) {
		return err
	}else {
		return nil
	}
}

func GetRadiusServerClients(radiusServerConfig dao.RadiusServer) ([]dao.NasClient, error){
	dbMap, errDb := GetRadiusServerDBConnection(radiusServerConfig);
	defer dbMap.Db.Close()
	var nasClients []dao.NasClient
	if errDb != nil{
		return nil, errors.New("Error occourred while connection to DB server |"+" DB Server IP :"+radiusServerConfig.DBHostIp +":"+strconv.Itoa(radiusServerConfig.DBHostPort)+" | Stack : " + errDb.Error() )

	}
	_, err := dbMap.Select(&nasClients, commons.GET_NAS_CLIENTS_INSERVER)
	if err != nil {
		return nil, errors.New("Error occourred while getting NAS clients On "+"DB Server IP :"+radiusServerConfig.DBHostIp +":"+strconv.Itoa(radiusServerConfig.DBHostPort)+" | Stack : " + err.Error() )
	}
	return nasClients, err
}

func DeleteRadiusInstance(tenantid int,username string) error {
	dbMap := utils.GetDBConnection("dashboard");
	defer dbMap.Db.Close()

	_, err := dbMap.Exec(commons.DELETE_RADIUS_SERVER_INST, tenantid, username)
	if (err != nil) {
		return err
	}else {
		return nil
	}
}

func UpdateRadiusServerInstance(config dao.RadiusServer, tenantId int) error{
	dbMap := utils.GetDBConnection("dashboard");
	defer dbMap.Db.Close()
	_, err := dbMap.Exec(commons.UPDATE_RADIUS_SERVER_INST, config.DBHostName, config.DBHostPort, config.DBSchemaName, config.DBUserName, config.DBPassword, tenantId, config.InstanceId)
	if (err != nil) {
		return err
	}else {
		return nil
	}
}

func IsWifiUserValidInRadius(tenantId int, username string) (int, error) {
	dbMap := utils.GetDBConnection(commons.RADIUS_DB);
	defer dbMap.Db.Close()

	var checkUser int
	err := dbMap.SelectOne(&checkUser, commons.IS_VALID_USER_IN_RADIUS, username, username)
	if err != nil {
		return checkUser, err
	}
	return checkUser, nil
}

func IsNASIpExistsInRadius(radiusServerConfigs dao.RadiusServer, ipAddress string, length string) (bool, error)  {
	dbMap, errDb := GetRadiusServerDBConnection(radiusServerConfigs);
	defer dbMap.Db.Close()
	if errDb != nil{
		return false, errors.New("Error occourred while connection to DB server |"+" DB Server IP :"+radiusServerConfigs.DBHostIp +":"+strconv.Itoa(radiusServerConfigs.DBHostPort)+" | Stack : " + errDb.Error() )
	}
	rangeSize, _ := strconv.Atoi(length)
	var allNasNames []string
	_, err := dbMap.Select(&allNasNames, commons.GET_ALL_NASNAMES)
	if err != nil {
		return false, err
	}
	for _, nasName := range allNasNames {
		result, err := checkIpBetweenIPRange(ipAddress,rangeSize,nasName)
		println(result)
		if err != nil {
		}
		if(result){
			return true, nil
		}
	}
	return false, nil
}

// Get Database connection for each radius server.That database connection for get NAS clients of each radius server
func GetRadiusServerDBConnection(radiusServerConfig dao.RadiusServer) (*gorp.DbMap, error){
	var connectionUrl string
	connectionUrl = radiusServerConfig.DBUserName + ":" + radiusServerConfig.DBPassword + "@tcp(" + radiusServerConfig.DBHostIp + ":" + strconv.Itoa(radiusServerConfig.DBHostPort)  + ")/" + radiusServerConfig.DBSchemaName
	db, err := sql.Open("mysql", connectionUrl)
	if err != nil {
		return nil, errors.New("Error occourred while conecting to radius server ip:"+ radiusServerConfig.DBHostIp + "stack : " + err.Error() )
	}
	dbmap := &gorp.DbMap{Db: db, Dialect:gorp.MySQLDialect{"InnoDB", "UTF8"}}
	return dbmap, err
}


func checkIpBetweenIPRange(requestIP string,rangeSize int, existsIP string) (bool, error){
	var slices []string
	if rangeSize != 0 {
		var requestIPString string
		requestIPString = requestIP+"/"+strconv.Itoa(rangeSize)
		_, requestNet, _ := net.ParseCIDR(requestIPString)
		slices = strings.Split(existsIP, "/")
		if len(slices)>1 {
			_, existsNet, _ := net.ParseCIDR(existsIP)
			return intersectBetweenCIDR(requestNet, existsNet),nil
		} else {
			existsNet := net.ParseIP(slices[0])
			return intersectBetweenCIDRandIP(existsNet,requestNet),nil
		}
	}else{
		requestNet := net.ParseIP(requestIP)
		slices = strings.Split(existsIP, "/")
		if len(slices)>1 {
			_, existsNet, _ := net.ParseCIDR(existsIP)
			return intersectBetweenCIDRandIP(requestNet, existsNet),nil
		} else {
			existsNet := net.ParseIP(slices[0])
			return intersectBetweenIP(requestNet, existsNet),nil
		}
	}
}

func intersectBetweenCIDR(CIDR1, CIDR2 *net.IPNet) bool {
	return CIDR2.Contains(CIDR1.IP) || CIDR1.Contains(CIDR2.IP)
}

func intersectBetweenCIDRandIP(IP net.IP, CIDR *net.IPNet) bool {
	return CIDR.Contains(IP)
}

func intersectBetweenIP(IP1 net.IP, IP2 net.IP) bool {
	return bytes.Compare(IP1, IP2) == 0
}
