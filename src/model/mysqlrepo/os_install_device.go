package mysqlrepo

import (
	"fmt"
	"model"
	"server/osinstallserver/util"
	"strconv"
	"strings"
	"time"
)

//device相关
func (repo *MySQLRepo) AddDevice(batchNumber string, sn string, hostname string, ip string, networkId uint, osId uint, hardwareId uint, systemId uint, location string, locationId uint, assetNumber string, status string, isSupportVm string) (*model.Device, error) {
	mod := model.Device{BatchNumber: batchNumber, Sn: sn, Hostname: hostname, Ip: ip, NetworkID: networkId, OsID: osId, HardwareID: hardwareId, SystemID: systemId, Location: location, LocationID: locationId, AssetNumber: assetNumber, Status: status, IsSupportVm: isSupportVm}
	err := repo.db.Create(&mod).Error
	return &mod, err
}

func (repo *MySQLRepo) UpdateDeviceById(id uint, batchNumber string, sn string, hostname string, ip string, networkId uint, osId uint, hardwareId uint, systemId uint, location string, locationId uint, assetNumber string, status string, isSupportVm string) (*model.Device, error) {
	mod := model.Device{BatchNumber: batchNumber, Sn: sn, Hostname: hostname, Ip: ip, NetworkID: networkId, OsID: osId, HardwareID: hardwareId, SystemID: systemId, Location: location, LocationID: locationId, AssetNumber: assetNumber, Status: status, IsSupportVm: isSupportVm}
	err := repo.db.First(&mod, id).Update("batch_number", batchNumber).Update("sn", sn).Update("hostname", hostname).Update("ip", ip).Update("network_id", networkId).Update("os_id", osId).Update("hardware_id", hardwareId).Update("system_id", systemId).Update("location", location).Update("location_id", locationId).Update("asset_number", assetNumber).Update("status", status).Update("install_progress", 0.0000).Update("install_log", "").Update("is_support_vm", isSupportVm).Error
	return &mod, err
}

func (repo *MySQLRepo) UpdateInstallInfoById(id uint, status string, installProgress float64) (*model.Device, error) {
	mod := model.Device{Status: status, InstallProgress: installProgress}
	err := repo.db.First(&mod, id).Update("status", status).Update("install_progress", installProgress).Error
	return &mod, err
}

func (repo *MySQLRepo) ReInstallDeviceById(id uint) (*model.Device, error) {
	mod := model.Device{}
	err := repo.db.First(&mod, id).Update("status", "pre_install").Update("install_progress", 0.0000).Update("install_log", "").Error
	return &mod, err
}

//device相关
func (repo *MySQLRepo) CreateBatchNumber() (string, error) {
	date := time.Now().Format("2006-01-02")
	var batchNumber string
	//row := repo.db.DB().QueryRow("select count(*) as count from (select batch_number from devices where batch_number like ?) as t", date+"%")
	row := repo.db.DB().QueryRow("select count(*) as count from devices where created_at >= ? and created_at <= ?", date+" 00:00:00", date+" 23:59:59")
	var count = -1
	if err := row.Scan(&count); err != nil {
		return "", err
	}

	if count > 0 {
		var device model.Device
		err := repo.db.Unscoped().Where("created_at >= ? and created_at <= ?", date+" 00:00:00", date+" 23:59:59").Limit(1).Order("id DESC").Find(&device).Error
		if err != nil {
			return "", nil
		}
		fix := util.SubString(device.BatchNumber, 8, len(device.BatchNumber)-8)
		fixNum, err := strconv.Atoi(fix)
		if err != nil {
			return "", err
		}
		batchNumber = strings.Replace(date, "-", "", -1) + fmt.Sprintf("%03d", fixNum+1)
	} else {
		batchNumber = strings.Replace(date, "-", "", -1) + fmt.Sprintf("%03d", 1)
	}

	return batchNumber, nil
}

func (repo *MySQLRepo) DeleteDeviceById(id uint) (*model.Device, error) {
	mod := model.Device{}
	err := repo.db.Unscoped().Where("id = ?", id).Delete(&mod).Error
	return &mod, err
}

func (repo *MySQLRepo) CountDeviceBySn(sn string) (uint, error) {
	mod := model.Device{Sn: sn}
	var count uint
	err := repo.db.Unscoped().Model(mod).Where("sn = ?", sn).Count(&count).Error
	return count, err
}

func (repo *MySQLRepo) CountDeviceByHostname(hostname string) (uint, error) {
	mod := model.Device{Hostname: hostname}
	var count uint
	err := repo.db.Unscoped().Model(mod).Where("hostname = ?", hostname).Count(&count).Error
	return count, err
}

func (repo *MySQLRepo) CountDeviceByHostnameAndId(hostname string, id uint) (uint, error) {
	mod := model.Device{Hostname: hostname}
	var count uint
	err := repo.db.Unscoped().Model(mod).Where("hostname = ? and id != ?", hostname, id).Count(&count).Error
	return count, err
}

func (repo *MySQLRepo) CountDeviceByIp(ip string) (uint, error) {
	mod := model.Device{Ip: ip}
	var count uint
	err := repo.db.Unscoped().Model(mod).Where("ip = ?", ip).Count(&count).Error
	return count, err
}

func (repo *MySQLRepo) CountDeviceByIpAndId(ip string, id uint) (uint, error) {
	mod := model.Device{Ip: ip}
	var count uint
	err := repo.db.Unscoped().Model(mod).Where("ip = ? and id != ?", ip, id).Count(&count).Error
	return count, err
}

func (repo *MySQLRepo) CountDevice(where string) (int, error) {
	row := repo.db.DB().QueryRow("SELECT count(t1.id) as count FROM devices t1 left join networks t2 on t1.network_id = t2.id left join os_configs t3 on t1.os_id = t3.id left join hardwares t4 on t1.hardware_id = t4.id left join system_configs t5 on t1.system_id = t5.id " + where)
	var count = -1
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (repo *MySQLRepo) CountDeviceByWhere(where string) (int, error) {
	row := repo.db.DB().QueryRow("SELECT count(*) as count FROM devices where " + where)
	var count = -1
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (repo *MySQLRepo) GetDeviceByWhere(where string) ([]model.Device, error) {
	var result []model.Device
	sql := "SELECT * FROM devices where " + where
	err := repo.db.Raw(sql).Scan(&result).Error
	return result, err
}

func (repo *MySQLRepo) GetDeviceListWithPage(limit uint, offset uint, where string) ([]model.DeviceFull, error) {
	/*
		var mods []model.Device
		err := repo.db.Limit(limit).Offset(offset).Find(&mods).Error
		return mods, err
	*/

	var result []model.DeviceFull
	sql := "SELECT t1.*,t2.network as network_name,t3.name as os_name,concat(t4.company,'-',t4.product,'-',t4.model_name) as hardware_name,t5.name as system_name FROM devices t1 left join networks t2 on t1.network_id = t2.id left join os_configs t3 on t1.os_id = t3.id left join hardwares t4 on t1.hardware_id = t4.id left join system_configs t5 on t1.system_id = t5.id " + where + " limit " + fmt.Sprintf("%d", limit)
	if offset > 0 {
		sql += "," + fmt.Sprintf("%d", offset)
	}

	err := repo.db.Raw(sql).Scan(&result).Error
	return result, err
}

func (repo *MySQLRepo) GetFullDeviceById(id uint) (*model.DeviceFull, error) {
	var result model.DeviceFull
	err := repo.db.Raw("SELECT t1.*,t2.network as network_name,t3.name as os_name,concat(t4.company,'-',t4.product,'-',t4.model_name) as hardware_name,t5.name as system_name FROM devices t1 left join networks t2 on t1.network_id = t2.id left join os_configs t3 on t1.os_id = t3.id left join hardwares t4 on t1.hardware_id = t4.id left join system_configs t5 on t1.system_id = t5.id where t1.id = ?", id).Scan(&result).Error
	return &result, err
}

func (repo *MySQLRepo) GetDeviceById(id uint) (*model.Device, error) {
	var mod model.Device
	err := repo.db.Unscoped().Where("id = ?", id).Find(&mod).Error
	return &mod, err
}

func (repo *MySQLRepo) GetDeviceIdBySn(sn string) (uint, error) {
	mod := model.Device{Sn: sn}
	err := repo.db.Unscoped().Where("sn = ?", sn).Find(&mod).Error
	return mod.ID, err
}

func (repo *MySQLRepo) GetSystemBySn(sn string) (*model.SystemConfig, error) {
	var mod model.SystemConfig
	err := repo.db.Joins("inner join devices on devices.system_id = system_configs.id").Where("devices.sn = ?", sn).Find(&mod).Error
	return &mod, err
}

func (repo *MySQLRepo) GetNetworkBySn(sn string) (*model.Network, error) {
	var mod model.Network
	err := repo.db.Joins("inner join devices on devices.network_id = networks.id").Where("devices.sn = ?", sn).Find(&mod).Error
	return &mod, err
}